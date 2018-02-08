package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/session"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

// Driver stores all the configuration data
type Driver struct {
	ctx          context.Context
	client       *govmomi.Client
	finder       *find.Finder
	datacenter   *object.Datacenter
	ResourcePool *object.ResourcePool
}

// NewDriver creates a new vSphere connection
func NewDriver(config *ConnectConfig) (*Driver, error) {
	ctx := context.TODO()

	vcenterURL, err := url.Parse(fmt.Sprintf("https://%v/sdk", config.VCenterServer))
	if err != nil {
		return nil, err
	}
	credentials := url.UserPassword(config.Username, config.Password)
	vcenterURL.User = credentials

	soapClient := soap.NewClient(vcenterURL, config.InsecureConnection)
	vimClient, err := vim25.NewClient(ctx, soapClient)
	if err != nil {
		return nil, err
	}

	vimClient.RoundTripper = session.KeepAlive(vimClient.RoundTripper, 10*time.Minute)
	client := &govmomi.Client{
		Client:         vimClient,
		SessionManager: session.NewManager(vimClient),
	}

	err = client.SessionManager.Login(ctx, credentials)
	if err != nil {
		return nil, err
	}

	finder := find.NewFinder(client.Client, false)
	datacenter, err := finder.DatacenterOrDefault(ctx, config.Datacenter)
	if err != nil {
		return nil, err
	}
	finder.SetDatacenter(datacenter)

	d := Driver{
		ctx:        ctx,
		client:     client,
		datacenter: datacenter,
		finder:     finder,
	}
	return &d, nil
}

// CreateVM creates the VM
func (d *Driver) CreateVM(config *CreateConfig) (*object.VirtualMachine, error) {

	var devices object.VirtualDeviceList
	var err error

	spec := &types.VirtualMachineConfigSpec{
		Name:       config.VMName,
		GuestId:    config.GuestOS,
		NumCPUs:    int32(config.CPU),
		MemoryMB:   int64(config.RAM),
		Annotation: config.Annotation,
		Version:    config.HardwareVersion,
	}

	// Storage configuration
	var bytesRegexp = regexp.MustCompile(`^(?i)(\d+)([BKMGTPE]?)(ib|b)?$`)
	m := bytesRegexp.FindStringSubmatch(config.Disk)
	i, _ := strconv.ParseInt(m[1], 10, 64)
	diskByteSize := i * 1024 * 1024 * 1024

	data := config.IsoDatastore
	datafile := config.IsoFile

	devices, err = d.addStorage(nil, data, datafile, diskByteSize)
	if err != nil {
		return nil, err
	}

	// Network configuration
	networkname := config.Network
	netadaptertype := config.NetworkAdapter
	devices, err = d.addNetwork(devices, networkname, netadaptertype)
	if err != nil {
		return nil, err
	}

	deviceChange, err := devices.ConfigSpec(types.VirtualDeviceConfigSpecOperationAdd)
	if err != nil {
		return nil, err
	}

	spec.DeviceChange = deviceChange

	folder, err := d.finder.FolderOrDefault(d.ctx, fmt.Sprintf("/%v/vm/%v", d.datacenter.Name(), config.Folder))
	if err != nil {
		return nil, err
	}

	var relocateSpec types.VirtualMachineRelocateSpec

	pool, err := d.finder.ResourcePoolOrDefault(d.ctx, config.ResourcePool)
	if err != nil {
		return nil, err
	}
	poolRef := pool.Reference()
	relocateSpec.Pool = &poolRef

	datastore, err := d.finder.Datastore(d.ctx, config.Datastore)
	datastoreRef := datastore.Reference()
	relocateSpec.Datastore = &datastoreRef

	spec.Files = &types.VirtualMachineFileInfo{
		VmPathName: fmt.Sprintf("[%s]", datastore.Name()),
	}

	task, err := folder.CreateVM(d.ctx, *spec, pool, nil)
	if err != nil {
		return nil, err
	}

	info, err := task.WaitForResult(d.ctx, nil)
	if err != nil {
		return nil, err
	}

	vm := object.NewVirtualMachine(d.client.Client, info.Result.(types.ManagedObjectReference))
	return vm, nil
}

// DestroyVM destroys the VM
func (d *Driver) DestroyVM(vm *object.VirtualMachine) error {
	task, err := vm.Destroy(d.ctx)
	if err != nil {
		return err
	}
	_, err = task.WaitForResult(d.ctx, nil)
	return err
}

// ConfigureVM configures the VM
func (d *Driver) ConfigureVM(vm *object.VirtualMachine, config *HardwareConfig) error {
	var confSpec types.VirtualMachineConfigSpec
	confSpec.NumCPUs = config.CPUs
	confSpec.MemoryMB = config.RAM

	var cpuSpec types.ResourceAllocationInfo
	cpuSpec.Reservation = &config.CPUReservation
	cpuSpec.Limit = &config.CPULimit
	confSpec.CpuAllocation = &cpuSpec

	var ramSpec types.ResourceAllocationInfo
	ramSpec.Reservation = &config.RAMReservation
	confSpec.MemoryAllocation = &ramSpec

	confSpec.MemoryReservationLockedToMax = &config.RAMReserveAll

	task, err := vm.Reconfigure(d.ctx, confSpec)
	if err != nil {
		return err
	}
	_, err = task.WaitForResult(d.ctx, nil)
	return err
}

// PowerOn powers on the VM
func (d *Driver) PowerOn(vm *object.VirtualMachine) error {
	task, err := vm.PowerOn(d.ctx)
	if err != nil {
		return err
	}
	_, err = task.WaitForResult(d.ctx, nil)
	return err
}

// WaitForIP waits for the IP address to become available via VMware Tools
func (d *Driver) WaitForIP(vm *object.VirtualMachine) (string, error) {
	ip, err := vm.WaitForIP(d.ctx)
	if err != nil {
		return "", err
	}
	return ip, nil
}

// PowerOff powers of the VM
func (d *Driver) PowerOff(vm *object.VirtualMachine) error {
	state, err := vm.PowerState(d.ctx)
	if err != nil {
		return err
	}

	if state == types.VirtualMachinePowerStatePoweredOff {
		return nil
	}

	task, err := vm.PowerOff(d.ctx)
	if err != nil {
		return err
	}
	_, err = task.WaitForResult(d.ctx, nil)
	return err
}

// StartShutdown starts the guest shutdown process
func (d *Driver) StartShutdown(vm *object.VirtualMachine) error {
	err := vm.ShutdownGuest(d.ctx)
	return err
}

// WaitForShutdown waits for the VM to shutdown
func (d *Driver) WaitForShutdown(vm *object.VirtualMachine, timeout time.Duration) error {
	shutdownTimer := time.After(timeout)
	for {
		powerState, err := vm.PowerState(d.ctx)
		if err != nil {
			return err
		}
		if powerState == "poweredOff" {
			break
		}

		select {
		case <-shutdownTimer:
			err := errors.New("Timeout while waiting for machine to shut down.")
			return err
		default:
			time.Sleep(1 * time.Second)
		}
	}
	return nil
}

// CreateSnapshot creates a snapshot of the VM
func (d *Driver) CreateSnapshot(vm *object.VirtualMachine) error {
	task, err := vm.CreateSnapshot(d.ctx, "Created by Packer", "", false, false)
	if err != nil {
		return err
	}
	_, err = task.WaitForResult(d.ctx, nil)
	return err
}

// ConvertToTemplate converts the VM to a template
func (d *Driver) ConvertToTemplate(vm *object.VirtualMachine) error {
	err := vm.MarkAsTemplate(d.ctx)
	return err
}

// Device handles the complex calls to configure the network adapter
func (d *Driver) Device(networkname string, netadaptertype string) (types.BaseVirtualDevice, error) {

	network, err := d.finder.Network(d.ctx, networkname)

	backing, err := network.EthernetCardBackingInfo(context.TODO())
	if err != nil {
		return nil, err
	}

	device, err := object.EthernetCardTypes().CreateEthernetCard(netadaptertype, backing)
	if err != nil {
		return nil, err
	}

	/*	if config.NetworkMacAddress != "" {
			card := device.(types.BaseVirtualEthernetCard).GetVirtualEthernetCard()
			card.AddressType = string(types.VirtualEthernetCardMacTypeManual)
			card.MacAddress = config.NetworkMacAddress
		}
	*/
	return device, nil
}

// addNetwork adds the network adapter to the VM
func (d *Driver) addNetwork(devices object.VirtualDeviceList, networkname string, netadaptertype string) (object.VirtualDeviceList, error) {
	netdev, err := d.Device(networkname, netadaptertype)
	if err != nil {
		return nil, err
	}

	devices = append(devices, netdev)
	return devices, nil
}

// addStorage adds the CD-ROM and Hard Disk to the VM
func (d *Driver) addStorage(devices object.VirtualDeviceList, isopath string, isofile string, diskbytesize int64) (object.VirtualDeviceList, error) {

	// Create SCSI Controller for Hard Disk
	scsi, err := devices.CreateSCSIController("scsi")
	if err != nil {
		return nil, err
	}

	devices = append(devices, scsi)

	// Create IDE Controller for CD-ROM
	idecontroller, err := devices.CreateIDEController()
	if err != nil {
		return nil, err
	}

	devices = append(devices, idecontroller)

	// Add a CD-ROM to the IDE Controller

	// Find the IDE controller
	ide, err := devices.FindIDEController("")
	if err != nil {
		return nil, err
	}

	// Create the CD-ROM Drive
	cdrom, err := devices.CreateCdrom(ide)
	if err != nil {
		return nil, err
	}

	// Find the datastore the specified for the ISO
	isodatastore, err := d.finder.Datastore(d.ctx, isopath)
	cdrom = devices.InsertIso(cdrom, isodatastore.Path(isofile))
	devices = append(devices, cdrom)

	// Add Hard Disk
	controllername := devices.Name(scsi)

	controller, err := devices.FindDiskController(controllername)
	if err != nil {
		return nil, err
	}

	disk := &types.VirtualDisk{
		VirtualDevice: types.VirtualDevice{
			Key: devices.NewKey(),
			Backing: &types.VirtualDiskFlatVer2BackingInfo{
				DiskMode:        string(types.VirtualDiskModePersistent),
				ThinProvisioned: types.NewBool(true),
			},
		},
		CapacityInKB: diskbytesize / 1024,
	}

	devices.AssignController(disk, controller)
	devices = append(devices, disk)

	return devices, nil
}

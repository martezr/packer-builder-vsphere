package main

import (
	"github.com/mitchellh/multistep"
	"github.com/vmware/govmomi/object"
	"github.com/hashicorp/packer/packer"
	"fmt"
)

type CloneConfig struct {
	VMName         string `mapstructure:"vm_name"`
	Folder         string `mapstructure:"folder"`
  GuestOS        string `mapstructure:"guest_os_type"`
	CPU            int32  `mapstructure:"cpu"`
	RAM            int64  `mapstructure:"ram"`


  Network        string `mapstructure:"network"`
  NetworkAdapter string `mapstructure:"network_adapter"`

  disk           string `mapstructure:"disk_size"`
	iso            string `mapstructure:"iso"`
	isoDatastore   string `mapstructure:"iso_datastore"`
	Host           string `mapstructure:"host"`
	ResourcePool   string `mapstructure:"resource_pool"`
	Datastore      string `mapstructure:"datastore"`
}

func (c *CreateConfig) Prepare() []error {
	var errs []error

	if c.VMName == "" {
		errs = append(errs, fmt.Errorf("Target VM name is required"))
	}
	if c.Host == "" {
		errs = append(errs, fmt.Errorf("vSphere host is required"))
	}

	return errs
}

type StepCreateVM struct {
	config *CreateConfig
}

func (s *StepCreateVM) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	d := state.Get("driver").(*Driver)

	ui.Say("Creating VM...")

	vm, err := d.CreateVM(s.config)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	state.Put("vm", vm)
	return multistep.ActionContinue
}

func (s *StepCreateVM) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if !cancelled && !halted {
		return
	}

	if vm, ok := state.GetOk("vm"); ok {
		ui := state.Get("ui").(packer.Ui)
		d := state.Get("driver").(*Driver)

		ui.Say("Destroying VM...")

		err := d.DestroyVM(vm.(*object.VirtualMachine))
		if err != nil {
			ui.Error(err.Error())
		}
	}
}

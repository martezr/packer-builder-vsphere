# Packer Builder for VMware vSphere

This a plugin for [HashiCorp Packer](https://www.packer.io/). It uses native vSphere API, and creates virtual machines remotely.

- VMware Player is not required
- Builds are incremental are create from scratch and require an automated method of installing the operating system.
- Official vCenter API is used, no ESXi host [modification](https://www.packer.io/docs/builders/vmware-iso.html#building-on-a-remote-vsphere-hypervisor) is required

## Usage
* Download the plugin from [Releases](https://github.com/martezr/packer-builder-vsphere/releases) page
* [Install](https://www.packer.io/docs/extending/plugins.html#installing-plugins) the plugin, or simply put it into the same directory with configuration files

## Linux Example

```json
{
  "builders": [
    {
      "type": "vsphere-iso",

      "vcenter_server": "vcenter.domain.com",
      "insecure_connection": "true",
      "username": "root",
      "password": "secret",
      "cluster": "cluster01",
      "host": "esxi-1.domain.com",

      "vm_name":  "vm-1",
      "convert_to_template": "true",
      "folder": "templates",
      "cpu": "1",
      "ram": "2048",
      "network": "VM Network",
      "network_adapter": "vmxnet3",
      "guest_os_type": "otherGuest",
      "datastore": "Local_Storage",
      "disk_size": "5GB",
      "iso": "ISOS/CentOS7.ISO",
      "iso_datastore": "Local_Storage",
      "communicator": "ssh",
      "ssh_username": "root",
      "ssh_password": "password",
      "annotation": "CentOS 7 with ChefDK"
    }
  ],
    "provisioners": [
        {
            "type": "shell",
            "inline":[
                "rpm -Uvh https://packages.chef.io/files/stable/chefdk/2.4.17/el/7/chefdk-2.4.17-1.el7.x86_64.rpm",
                "chef -v"
            ]
        }
    ]
}
```

## Windows Example

```json
{
  "builders": [
    {
      "type": "vsphere-iso",

      "vcenter_server": "vcenter.domain.com",
      "insecure_connection": "true",
      "username": "root",
      "password": "secret",
      "cluster": "cluster01",
      "host": "esxi-1.domain.com",

      "vm_name":  "vm-1",
      "convert_to_template": "true",
      "folder": "templates",
      "cpu": "1",
      "ram": "2048",
      "network": "VM Network",
      "network_adapter": "e1000",
      "guest_os_type": "otherGuest",
      "datastore": "Local_Storage",
      "disk_size": "5GB",
      "iso": "ISOS/WIN2K12.ISO",
      "iso_datastore": "Local_Storage",
      "communicator": "winrm",
      "winrm_username": "root",
      "winrm_password": "secret"
    }
  ]
}
```

## Parameters

Connection:
* `vcenter_server` - [**mandatory**] vCenter server hostname.
* `username` - [**mandatory**] vSphere username.
* `password` - [**mandatory**] vSphere password.
* `insecure_connection` - do not validate server's TLS certificate. `false` by default.
* `datacenter` - required if there are several datacenters.

Location:
* `vm_name` - [**mandatory**] name of target VM.
* `folder` - VM folder where target VM is created.
* `host` - vSphere host or cluster where target VM is created. If hosts are groupped into folders, full path should be specified: `folder/host`.
* `resource_pool` - by default a root of vSphere host.
* `datastore` - required if target is a cluster, or a host with multiple datastores.

Hardware customization:
* `hardware_version` -
* `guest_os_type` - Guest Operating System identifier.
* `cpu` - number of CPU sockets. Inherited from source VM by default.
* `CPU_reservation` - Amount of reserved CPU resources in MHz. Inherited from source VM by default.
* `CPU_limit` - Upper limit of available CPU resources in MHz. Inherited from source VM by default, set to `-1` for reset.
* `ram` - Amount of RAM in megabytes. Inherited from source VM by default.
* `RAM_reservation` - Amount of reserved RAM in MB. Inherited from source VM by default.
* `RAM_reserve_all` - Reserve all available RAM (bool). `false` by default. Cannot be used together with `RAM_reservation`.
* `disk_size` - The size of the hard disk.
* `iso_datastore` - The datastore the ISO file is stored on.
* `iso` - The path of the ISO file, full path should be specified: `folder/file`
* `network` - The virtual network the VM is attached to.
* `network_adapter` - The network adapter type for the VM.

Provisioning:
* `communicator` - [mandatory]
* `ssh_username` - [mandatory] username in guest OS.
* `ssh_password or ssh_private_key_file` - [mandatory] password or SSH-key filename to access a guest OS.

Post-processing:
* `create_snapshot` - add a snapshot, so VM can be used as a base for linked clones. `false` by default.
* `convert_to_template` - convert VM to a template. `false` by default.

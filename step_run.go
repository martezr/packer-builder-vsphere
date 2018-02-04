package main

import (
	"fmt"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
	"github.com/vmware/govmomi/object"
)

// StepRun stores the configuration for the run process
type StepRun struct {
}

// Run powers on the VM and waits for the IP address
func (s *StepRun) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	d := state.Get("driver").(*Driver)
	vm := state.Get("vm").(*object.VirtualMachine)

	ui.Say("Power on VM...")
	err := d.PowerOn(vm)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Say("Waiting for IP...")
	ip, err := d.WaitForIP(vm)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}
	state.Put("ip", ip)
	ui.Say(fmt.Sprintf("IP address: %v", ip))

	return multistep.ActionContinue
}

// Cleanup the run process
func (s *StepRun) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if !cancelled && !halted {
		return
	}

	ui := state.Get("ui").(packer.Ui)
	d := state.Get("driver").(*Driver)
	vm := state.Get("vm").(*object.VirtualMachine)

	ui.Say("Power off VM...")
	err := d.PowerOff(vm)
	if err != nil {
		ui.Error(err.Error())
	}
}

package main

import (
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
	"github.com/vmware/govmomi/object"
)

// StepConvertToTemplate stores the boolean for whether the VM is converted to a template
type StepConvertToTemplate struct {
	ConvertToTemplate bool
}

// Run the template creation process
func (s *StepConvertToTemplate) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	d := state.Get("driver").(*Driver)
	vm := state.Get("vm").(*object.VirtualMachine)

	if s.ConvertToTemplate {
		ui.Say("Convert VM into template...")
		err := d.ConvertToTemplate(vm)
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

// Cleanup the template creation process
func (s *StepConvertToTemplate) Cleanup(state multistep.StateBag) {}

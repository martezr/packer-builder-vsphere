package main

import (
	"fmt"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
	"github.com/vmware/govmomi/object"
)

// HardwareConfig stores all the details for post creation hardware configuration.
type HardwareConfig struct {
	CPUs           int32 `mapstructure:"CPUs"`
	CPUReservation int64 `mapstructure:"CPU_reservation"`
	CPULimit       int64 `mapstructure:"CPU_limit"`
	RAM            int64 `mapstructure:"RAM"`
	RAMReservation int64 `mapstructure:"RAM_reservation"`
	RAMReserveAll  bool  `mapstructure:"RAM_reserve_all"`
}

// Prepare for the hardware configuration process
func (c *HardwareConfig) Prepare() []error {
	var errs []error

	if c.RAMReservation > 0 && c.RAMReserveAll != false {
		errs = append(errs, fmt.Errorf("'RAM_reservation' and 'RAM_reserve_all' cannot be used together"))
	}

	return errs
}

// StepConfigureHardware defines the hardware configuration step
type StepConfigureHardware struct {
	config *HardwareConfig
}

// Run configures the VM hardware
func (s *StepConfigureHardware) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	d := state.Get("driver").(*Driver)
	vm := state.Get("vm").(*object.VirtualMachine)

	if *s.config != (HardwareConfig{}) {
		ui.Say("Customizing hardware parameters...")

		err := d.ConfigureVM(vm, s.config)
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

// Cleanup the hardware configuration process
func (s *StepConfigureHardware) Cleanup(multistep.StateBag) {}

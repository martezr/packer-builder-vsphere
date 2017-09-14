package main

import (
	"errors"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
	"github.com/vmware/govmomi/object"
)

type Builder struct {
	config *Config
	runner multistep.Runner
}

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	c, warnings, errs := NewConfig(raws...)
	if errs != nil {
		return warnings, errs
	}
	b.config = c

	return warnings, nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	state := new(multistep.BasicStateBag)
	state.Put("hook", hook)
	state.Put("ui", ui)

	steps := []multistep.Step{
		&StepConnect{
			config: &b.config.ConnectConfig,
		},
		&StepCreateVM{
			config: &b.config.CreateConfig,
		},
		&StepConfigureHardware{
			config: &b.config.HardwareConfig,
		},
		&StepRun{},
		&common.StepProvision{},
		&StepShutdown{},
		&StepCreateSnapshot{
			createSnapshot: b.config.CreateSnapshot,
		},
		&StepConvertToTemplate{
			ConvertToTemplate: b.config.ConvertToTemplate,
		},
	}

	// Run!
	b.runner = common.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	// If we were interrupted or cancelled, then just exit.
	if _, ok := state.GetOk(multistep.StateCancelled); ok {
		return nil, errors.New("Build was cancelled.")
	}

	if _, ok := state.GetOk(multistep.StateHalted); ok {
		return nil, errors.New("Build was halted.")
	}

	artifact := &Artifact{
		Name: b.config.VMName,
		VM:   state.Get("vm").(*object.VirtualMachine),
	}
	return artifact, nil
}

func (b *Builder) Cancel() {
	if b.runner != nil {
		b.runner.Cancel()
	}
}

package main

import (
	"fmt"
	"github.com/mitchellh/multistep"
)

// ConnectConfig holds all the details for the vSphere connection process.
type ConnectConfig struct {
	VCenterServer      string `mapstructure:"vcenter_server"`
	Username           string `mapstructure:"username"`
	Password           string `mapstructure:"password"`
	InsecureConnection bool   `mapstructure:"insecure_connection"`
	Datacenter         string `mapstructure:"datacenter"`
}

// Prepare the vCenter connection
func (c *ConnectConfig) Prepare() []error {
	var errs []error

	if c.VCenterServer == "" {
		errs = append(errs, fmt.Errorf("vCenter hostname is required"))
	}
	if c.Username == "" {
		errs = append(errs, fmt.Errorf("Username is required"))
	}
	if c.Password == "" {
		errs = append(errs, fmt.Errorf("Password is required"))
	}

	return errs
}

// StepConnect defines the connection step
type StepConnect struct {
	config *ConnectConfig
}

// Run connects to the vCenter server
func (s *StepConnect) Run(state multistep.StateBag) multistep.StepAction {
	driver, err := NewDriver(s.config)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}
	state.Put("driver", driver)

	return multistep.ActionContinue
}

// Cleanup the connection process
func (s *StepConnect) Cleanup(multistep.StateBag) {}

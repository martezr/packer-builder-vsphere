package main

import (
	"context"
	"github.com/vmware/govmomi/object"
)

// BuilderId for the local artifacts
const BuilderId = "martezr.vsphere-iso"

// Artifact is the result of running the vsphere-iso builder, namely a set
// of files associated with the resulting machine.
type Artifact struct {
	Name string
	VM   *object.VirtualMachine
}

// BuilderId returns the builder ID.
func (a *Artifact) BuilderId() string {
	return BuilderId
}

// Files returns the files represented by the artifact.
func (a *Artifact) Files() []string {
	return []string{}
}

// Id returns the name of the artifact.
func (a *Artifact) Id() string {
	return a.Name
}

// String returns the string representation of the artifact.
func (a *Artifact) String() string {
	return a.Name
}

// State returns specific details from the artifact.
func (a *Artifact) State(name string) interface{} {
	return nil
}

// Destroy the vSphere VM represented by the artifact.
func (a *Artifact) Destroy() error {
	ctx := context.TODO()
	task, err := a.VM.Destroy(ctx)
	if err != nil {
		return err
	}
	_, err = task.WaitForResult(ctx, nil)
	return err
}

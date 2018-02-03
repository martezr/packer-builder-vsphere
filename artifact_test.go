package main

import (
	"testing"

	"github.com/hashicorp/packer/packer"
)

func TestArtifact_Impl(t *testing.T) {
	var _ packer.Artifact = new(Artifact)
}

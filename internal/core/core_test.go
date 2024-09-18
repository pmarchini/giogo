package core_test

import (
	"testing"

	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pmarchini/giogo/internal/core"
)

func TestRunCommand(t *testing.T) {
	resources := specs.LinuxResources{}
	coreModule := core.NewCore(resources)

	// Test with a valid command
	err := coreModule.RunCommand([]string{"echo", "Hello, World!"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test with an invalid command
	err = coreModule.RunCommand([]string{"invalid_command"})
	if err == nil {
		t.Errorf("Expected error for invalid command, got nil")
	}
}

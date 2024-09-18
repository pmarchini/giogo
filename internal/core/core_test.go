package core_test

import (
	"fmt"
	"os/exec"
	"testing"

	"github.com/pmarchini/giogo/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestRunCommand tests the Core RunCommand method using a mocked CgroupManager
func TestRunCommand(t *testing.T) {
	mockManager := new(core.MockCgroupManager)

	// Setup expectations
	mockManager.On("AddProcess", mock.AnythingOfType("int")).Return(nil)
	mockManager.On("Delete").Return(nil)

	// Create a Core instance with the mock manager
	core := &core.Core{
		CgroupManager: mockManager,
	}

	// Mock command
	exec.Command("echo", "hello")

	// Run the command
	err := core.RunCommand([]string{"echo", "hello"})

	// Assert no errors
	assert.NoError(t, err)

	// Assert expectations on the mock
	mockManager.AssertCalled(t, "AddProcess", mock.AnythingOfType("int"))
	mockManager.AssertCalled(t, "Delete")
}

// TestRunCommand_AddProcessError tests when adding a process to the cgroup fails
func TestRunCommand_AddProcessError(t *testing.T) {
	mockManager := new(core.MockCgroupManager)

	// Setup expectations: AddProcess will return an error
	mockManager.On("AddProcess", mock.AnythingOfType("int")).Return(fmt.Errorf("failed to add process"))
	mockManager.On("Delete").Return(nil)

	// Create a Core instance with the mock manager
	core := &core.Core{
		CgroupManager: mockManager,
	}

	// Run the command
	err := core.RunCommand([]string{"echo", "hello"})

	// Assert that we received an error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to add process")

	// Assert that Delete is still called even when AddProcess fails
	mockManager.AssertCalled(t, "AddProcess", mock.AnythingOfType("int"))
	mockManager.AssertCalled(t, "Delete")
}

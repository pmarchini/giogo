package executor_test

import (
	"testing"

	"github.com/pmarchini/giogo/internal/executor"
	"github.com/pmarchini/giogo/internal/limiter"
)

func TestExecutorRunCommand(t *testing.T) {
	cpuLimiter, _ := limiter.NewCPULimiter("0.5")
	memLimiter, _ := limiter.NewMemoryLimiter("128m")
	limiters := []limiter.ResourceLimiter{cpuLimiter, memLimiter}

	exec := executor.NewExecutor(limiters)

	// Test with a valid command
	err := exec.RunCommand([]string{"echo", "Hello, Executor!"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test with an invalid command
	err = exec.RunCommand([]string{"invalid_command"})
	if err == nil {
		t.Errorf("Expected error for invalid command, got nil")
	}
}

package executor

import (
	"fmt"
	"os"

	"github.com/pmarchini/giogo/internal/core"
	"github.com/pmarchini/giogo/internal/limiter"

	specs "github.com/opencontainers/runtime-spec/specs-go"
)

type Executor struct {
	Limiters       []limiter.ResourceLimiter
	NetworkLimiter *limiter.NetworkLimiter
}

func NewExecutor(limiters []limiter.ResourceLimiter) *Executor {
	executor := &Executor{
		Limiters: limiters,
	}
	
	// Extract NetworkLimiter if present for special handling
	for _, l := range limiters {
		if netLimiter, ok := l.(*limiter.NetworkLimiter); ok {
			executor.NetworkLimiter = netLimiter
			break
		}
	}
	
	return executor
}

func (e *Executor) RunCommand(args []string) error {
	var resources specs.LinuxResources
	for _, l := range e.Limiters {
		l.Apply(&resources)
	}

	// Set up traffic control if network limiter with bandwidth is configured
	if e.NetworkLimiter != nil && e.NetworkLimiter.MaxBandwidth > 0 {
		// Get default interface - in production this could be configurable
		iface := limiter.GetDefaultInterface()
		
		// Setup tc before running the command
		if err := e.NetworkLimiter.SetupTrafficControl(iface); err != nil {
			return fmt.Errorf("failed to setup traffic control: %v", err)
		}
		
		// Ensure cleanup happens when we're done
		defer func() {
			if err := e.NetworkLimiter.CleanupTrafficControl(iface); err != nil {
				// Non-fatal: log the error but don't fail the command
				fmt.Fprintf(os.Stderr, "Warning: failed to cleanup traffic control: %v\n", err)
			}
		}()
	}

	coreModule, err := core.NewCore(resources)
	if err != nil {
		return err
	}
	return coreModule.RunCommand(args)
}

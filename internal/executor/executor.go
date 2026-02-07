package executor

import (
	"fmt"
	"os"

	"github.com/pmarchini/giogo/internal/core"
	"github.com/pmarchini/giogo/internal/limiter"

	specs "github.com/opencontainers/runtime-spec/specs-go"
)

type Executor struct {
	Limiters []limiter.ResourceLimiter
}

func NewExecutor(limiters []limiter.ResourceLimiter) *Executor {
	return &Executor{
		Limiters: limiters,
	}
}

func (e *Executor) RunCommand(args []string) error {
	var resources specs.LinuxResources
	for _, l := range e.Limiters {
		l.Apply(&resources)
	}

	// Track successfully setup lifecycle limiters for cleanup
	var setupLimiters []limiter.LifecycleLimiter
	
	// Setup any limiters that implement lifecycle methods
	for _, l := range e.Limiters {
		if lifecycle, ok := l.(limiter.LifecycleLimiter); ok {
			if err := lifecycle.Setup(); err != nil {
				// Cleanup already setup limiters before returning error
				for i := len(setupLimiters) - 1; i >= 0; i-- {
					if cleanupErr := setupLimiters[i].Cleanup(); cleanupErr != nil {
						fmt.Fprintf(os.Stderr, "Warning: failed to cleanup limiter during error recovery: %v\n", cleanupErr)
					}
				}
				return fmt.Errorf("failed to setup limiter: %v", err)
			}
			setupLimiters = append(setupLimiters, lifecycle)
		}
	}
	
	// Ensure cleanup happens for all successfully setup limiters
	defer func() {
		for i := len(setupLimiters) - 1; i >= 0; i-- {
			if err := setupLimiters[i].Cleanup(); err != nil {
				// Non-fatal: log the error but don't fail the command
				fmt.Fprintf(os.Stderr, "Warning: failed to cleanup limiter: %v\n", err)
			}
		}
	}()

	coreModule, err := core.NewCore(resources)
	if err != nil {
		return err
	}
	return coreModule.RunCommand(args)
}

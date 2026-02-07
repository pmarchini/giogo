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

	// Setup any limiters that implement lifecycle methods
	for _, l := range e.Limiters {
		if lifecycle, ok := l.(limiter.LifecycleLimiter); ok {
			if err := lifecycle.Setup(); err != nil {
				return fmt.Errorf("failed to setup limiter: %v", err)
			}
			// Ensure cleanup happens when we're done
			defer func(lc limiter.LifecycleLimiter) {
				if err := lc.Cleanup(); err != nil {
					// Non-fatal: log the error but don't fail the command
					fmt.Fprintf(os.Stderr, "Warning: failed to cleanup limiter: %v\n", err)
				}
			}(lifecycle)
		}
	}

	coreModule, err := core.NewCore(resources)
	if err != nil {
		return err
	}
	return coreModule.RunCommand(args)
}

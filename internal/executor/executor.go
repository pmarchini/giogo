package executor

import (
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
	// Prepare resources
	var resources specs.LinuxResources
	for _, l := range e.Limiters {
		l.Apply(&resources)
	}

	// Utilize the core module to run the command
	coreModule := core.NewCore(resources)
	return coreModule.RunCommand(args)
}

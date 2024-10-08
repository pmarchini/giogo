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
	var resources specs.LinuxResources
	for _, l := range e.Limiters {
		l.Apply(&resources)
	}

	coreModule, err := core.NewCore(resources)
	if err != nil {
		return err
	}
	return coreModule.RunCommand(args)
}

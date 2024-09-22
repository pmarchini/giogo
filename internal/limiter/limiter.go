package limiter

import (
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

type ResourceLimiter interface {
	Apply(resources *specs.LinuxResources)
}

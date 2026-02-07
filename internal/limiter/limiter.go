package limiter

import (
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

type ResourceLimiter interface {
	Apply(resources *specs.LinuxResources)
}

// LifecycleLimiter is an optional interface for limiters that need
// setup and cleanup operations beyond the cgroup configuration
type LifecycleLimiter interface {
	ResourceLimiter
	Setup() error
	Cleanup() error
}

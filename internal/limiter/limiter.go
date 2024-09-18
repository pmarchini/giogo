package limiter

import (
    "github.com/containerd/cgroups"
    specs "github.com/opencontainers/runtime-spec/specs-go"
)

type ResourceLimiter interface {
    Apply(resources *specs.LinuxResources)
}

type CgroupManager interface {
    CreateCgroup(path string, resources *specs.LinuxResources) (cgroups.Cgroup, error)
}

package core

import (
	"github.com/containerd/cgroups/v3/cgroup1"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

type CgroupV1Manager struct {
	control cgroup1.Cgroup
}

func NewCgroupV1Manager(path string, resources specs.LinuxResources) (CgroupManager, error) {
	control, err := cgroup1.New(cgroup1.StaticPath(path), &resources)
	if err != nil {
		return nil, err
	}
	return &CgroupV1Manager{control: control}, nil
}

// AddProcess adds a process to the cgroup v1
func (m *CgroupV1Manager) AddProcess(pid int) error {
	return m.control.Add(cgroup1.Process{Pid: pid})
}

// Delete deletes the cgroup v1
func (m *CgroupV1Manager) Delete() error {
	return m.control.Delete()
}

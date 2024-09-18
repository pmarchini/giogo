package core

import (
	"fmt"
	"log"

	"github.com/containerd/cgroups/v3/cgroup2"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// CgroupV2Manager manages cgroup v2
type CgroupV2Manager struct {
	manager *cgroup2.Manager
}

// NewCgroupV2Manager creates a new CgroupV2Manager
func NewCgroupV2Manager(path string, resources specs.LinuxResources) (CgroupManager, error) {
	fmt.Printf("Creating cgroup v2 manager for path %s", path)
	manager, err := cgroup2.NewSystemd("/", path, -1, cgroup2.ToResources(&resources))
	if err != nil {
		log.Printf("Error creating cgroup v2 manager: %v", err)
		return nil, err
	}
	return &CgroupV2Manager{manager: manager}, nil
}

// AddProcess adds a process to the cgroup v2
func (m *CgroupV2Manager) AddProcess(pid int) error {
	return m.manager.AddProc(uint64(pid))
}

// Delete deletes the cgroup v2
func (m *CgroupV2Manager) Delete() error {
	return m.manager.DeleteSystemd()
}

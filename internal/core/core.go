package core

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/containerd/cgroups"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

type Core struct {
	Resources specs.LinuxResources
}

func NewCore(resources specs.LinuxResources) *Core {
	return &Core{
		Resources: resources,
	}
}

func (c *Core) RunCommand(args []string) error {
	// Determine cgroups mode (v1 or v2)
	cgroupMode := cgroups.Mode()

	// Create a unique cgroup path
	cgroupPath := fmt.Sprintf("/giogo_cgroup_%d", os.Getpid())

	var cgroup cgroups.Cgroup
	var err error

	if cgroupMode == cgroups.Unified {
		cgroup, err = cgroups.New(cgroups.Systemd, cgroups.StaticPath(cgroupPath), &c.Resources)
	} else {
		cgroup, err = cgroups.New(cgroups.V1, cgroups.StaticPath(cgroupPath), &c.Resources)
	}
	if err != nil {
		return fmt.Errorf("error creating cgroup: %v", err)
	}
	defer cgroup.Delete()

	// Prepare the command to execute
	execCmd := exec.Command(args[0], args[1:]...)
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	execCmd.Stdin = os.Stdin

	// Start the command
	err = execCmd.Start()
	if err != nil {
		return fmt.Errorf("error starting command: %v", err)
	}

	// Add the process to the cgroup
	err = cgroup.Add(cgroups.Process{Pid: execCmd.Process.Pid})
	if err != nil {
		return fmt.Errorf("error adding process to cgroup: %v", err)
	}

	// Wait for the command to finish
	err = execCmd.Wait()
	if err != nil {
		return fmt.Errorf("command exited with error: %v", err)
	}

	return nil
}

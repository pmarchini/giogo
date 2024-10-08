package core

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"

	"github.com/containerd/cgroups/v3"
	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pmarchini/giogo/ft"
)

// Core struct holds the resources and the CgroupManager
type Core struct {
	Resources     specs.LinuxResources
	CgroupManager CgroupManager
}

func IsValidSystemdSlice(path string) bool {
	regexPattern := `^[a-zA-Z0-9]+(-[a-zA-Z0-9]+)*$`

	// Compile the regex
	regex := regexp.MustCompile(regexPattern)

	// Check if the path matches the regex
	return regex.MatchString(path)
}

func GenerateCgroupPath() string {
	cgroupSystemdSlice := fmt.Sprintf("giogo-cgroup-%d", os.Getpid())
	if !IsValidSystemdSlice(cgroupSystemdSlice) {
		panic("Invalid systemd slice path")
	}
	return cgroupSystemdSlice
}

// NewCore returns a new Core instance and initializes the appropriate CgroupManager based on the cgroup version
func NewCore(resources specs.LinuxResources) (*Core, error) {
	cgroupMode := cgroups.Mode()
	cgroupPath := GenerateCgroupPath()

	fmt.Printf("Creating core for groupPath %s\n", cgroupPath)

	var manager CgroupManager
	var err error

	// Choose the appropriate cgroup manager (v1 or v2)
	if cgroupMode == cgroups.Unified {
		manager, err = NewCgroupV2Manager(cgroupPath, resources)
	} else {
		if !ft.CheckFeature("FT_CGROUP_V1_SUPPORT") {
			return nil, fmt.Errorf("cgroup v1 is not supported")
		}
		manager, err = NewCgroupV1Manager(cgroupPath, resources)
	}
	if err != nil {
		return nil, fmt.Errorf("error initializing cgroup manager: %v", err)
	}

	return &Core{
		Resources:     resources,
		CgroupManager: manager,
	}, nil
}

// RunCommand runs the command in a cgroup, ensuring the process is added to the cgroup and the cgroup is deleted after execution
func (c *Core) RunCommand(args []string) error {
	// Ensure the cgroup is always deleted when the function exits
	defer func() {
		if err := c.CgroupManager.Delete(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to delete cgroup: %v\n", err)
		}
	}()
	// Prepare the command to execute
	execCmd := exec.Command(args[0], args[1:]...)
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	execCmd.Stdin = os.Stdin

	// Start the command
	err := execCmd.Start()
	if err != nil {
		return fmt.Errorf("error starting command: %v", err)
	}

	// Add the process to the cgroup
	err = c.CgroupManager.AddProcess(execCmd.Process.Pid)
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

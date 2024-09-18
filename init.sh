#!/bin/bash

# Set the module path (replace with your own)
MODULE_PATH="github.com/yourusername/giogo"

# Create the project directory
mkdir -p giogo
cd giogo || exit

# Initialize the Go module
go mod init "$MODULE_PATH"

# Create the directory structure
mkdir -p cmd/giogo
mkdir -p internal/cli
mkdir -p internal/core
mkdir -p internal/executor
mkdir -p internal/limiter
mkdir -p internal/utils

# Create and write main.go
cat << 'EOF' > cmd/giogo/main.go
package main

import (
    "giogo/internal/cli"
)

func main() {
    cli.Execute()
}
EOF

# Create and write cli.go
cat << 'EOF' > internal/cli/cli.go
package cli

import (
    "fmt"
    "os"

    "giogo/internal/executor"
    "giogo/internal/limiter"

    "github.com/spf13/cobra"
)

var (
    ram        string
    cpu        string
    ioReadMax  string
    ioWriteMax string
)

func Execute() {
    var rootCmd = &cobra.Command{
        Use:   "giogo [flags] -- command [args...]",
        Short: "Giogo runs commands with specified cgroup resource limits",
        RunE:  runCommand,
        Args:  cobra.MinimumNArgs(1),
    }

    // Define flags
    rootCmd.Flags().StringVar(&ram, "ram", "", "Memory limit (e.g., 128m, 1g)")
    rootCmd.Flags().StringVar(&cpu, "cpu", "", "CPU limit as a fraction between 0 and 1 (e.g., 0.5)")
    rootCmd.Flags().StringVar(&ioReadMax, "io-read-max", "", "IO read max bandwidth (e.g., 128k, 1m)")
    rootCmd.Flags().StringVar(&ioWriteMax, "io-write-max", "", "IO write max bandwidth (e.g., 128k, 1m)")

    // Execute the root command
    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}

func runCommand(cmd *cobra.Command, args []string) error {
    var limiters []limiter.ResourceLimiter

    if cpu != "" {
        cpuLimiter, err := limiter.NewCPULimiter(cpu)
        if err != nil {
            return fmt.Errorf("invalid CPU value: %v", err)
        }
        limiters = append(limiters, cpuLimiter)
    }

    if ram != "" {
        memLimiter, err := limiter.NewMemoryLimiter(ram)
        if err != nil {
            return fmt.Errorf("invalid RAM value: %v", err)
        }
        limiters = append(limiters, memLimiter)
    }

    // TODO: Implement IO limiters when ready

    exec := executor.NewExecutor(limiters)
    if err := exec.RunCommand(args); err != nil {
        return err
    }

    return nil
}
EOF

# Create and write core.go
cat << 'EOF' > internal/core/core.go
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
        cgroup, err = cgroups.New(cgroups.Unified, cgroups.StaticPath(cgroupPath), &c.Resources)
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
EOF

# Create and write executor.go
cat << 'EOF' > internal/executor/executor.go
package executor

import (
    "giogo/internal/core"
    "giogo/internal/limiter"

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
EOF

# Create and write limiter.go
cat << 'EOF' > internal/limiter/limiter.go
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
EOF

# Create and write cpu.go
cat << 'EOF' > internal/limiter/cpu.go
package limiter

import (
    "strconv"

    specs "github.com/opencontainers/runtime-spec/specs-go"
)

type CPULimiter struct {
    Fraction float64
}

func (c *CPULimiter) Apply(resources *specs.LinuxResources) {
    period := uint64(100000)
    quota := int64(c.Fraction * float64(period))
    resources.CPU = &specs.LinuxCPU{
        Period: &period,
        Quota:  &quota,
    }
}

func NewCPULimiter(value string) (*CPULimiter, error) {
    fraction, err := strconv.ParseFloat(value, 64)
    if err != nil || fraction <= 0 || fraction > 1 {
        return nil, err
    }
    return &CPULimiter{Fraction: fraction}, nil
}
EOF

# Create and write memory.go
cat << 'EOF' > internal/limiter/memory.go
package limiter

import (
    "giogo/internal/utils"

    specs "github.com/opencontainers/runtime-spec/specs-go"
)

type MemoryLimiter struct {
    Limit int64
}

func (m *MemoryLimiter) Apply(resources *specs.LinuxResources) {
    resources.Memory = &specs.LinuxMemory{
        Limit: &m.Limit,
    }
}

func NewMemoryLimiter(value string) (*MemoryLimiter, error) {
    limit, err := utils.ParseMemory(value)
    if err != nil {
        return nil, err
    }
    return &MemoryLimiter{Limit: limit}, nil
}
EOF

# Create and write io.go (empty implementation for now)
cat << 'EOF' > internal/limiter/io.go
package limiter

// TODO: Implement IO limiter
EOF

# Create and write utils.go
cat << 'EOF' > internal/utils/utils.go
package utils

import (
    "strconv"
    "strings"
)

func ParseMemory(s string) (int64, error) {
    s = strings.TrimSpace(s)
    var multiplier int64 = 1
    if strings.HasSuffix(s, "g") || strings.HasSuffix(s, "G") {
        multiplier = 1024 * 1024 * 1024
        s = s[:len(s)-1]
    } else if strings.HasSuffix(s, "m") || strings.HasSuffix(s, "M") {
        multiplier = 1024 * 1024
        s = s[:len(s)-1]
    } else if strings.HasSuffix(s, "k") || strings.HasSuffix(s, "K") {
        multiplier = 1024
        s = s[:len(s)-1]
    } else {
        multiplier = 1
    }
    value, err := strconv.ParseFloat(s, 64)
    if err != nil {
        return 0, err
    }
    return int64(value * float64(multiplier)), nil
}
EOF


echo "Setup complete. You can now run ./giogo --help to see the available commands."

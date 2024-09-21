package cli

import (
	"fmt"
	"os"

	"github.com/pmarchini/giogo/internal/executor"
	"github.com/pmarchini/giogo/internal/limiter"

	"github.com/spf13/cobra"
)

var (
	ram        string
	cpu        string
	ioReadMax  string
	ioWriteMax string
)

func SetupRootCommand(rootCmd *cobra.Command) {
	rootCmd.Use = "giogo [flags] -- command [args...]"
	rootCmd.Short = "Giogo runs commands with specified cgroup resource limits"
	rootCmd.RunE = runCommand
	rootCmd.Args = cobra.MinimumNArgs(1)

	// Define flags
	rootCmd.Flags().StringVar(&ram, "ram", "", "Memory limit (e.g., 128m, 1g)")
	rootCmd.Flags().StringVar(&cpu, "cpu", "", "CPU limit as a fraction between 0 and 1 (e.g., 0.5)")
	rootCmd.Flags().StringVar(&ioReadMax, "io-read-max", "-1", "IO read max bandwidth (e.g., 128k, 1m)")
	rootCmd.Flags().StringVar(&ioWriteMax, "io-write-max", "-1", "IO write max bandwidth (e.g., 128k, 1m)")
}

func Execute() {
	var rootCmd = &cobra.Command{}
	SetupRootCommand(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// TODO: This logic should be moved to a separate package as it's part of the core functionality
func CreateLimiters(cpu, ram, ioReadMax, ioWriteMax string) ([]limiter.ResourceLimiter, error) {
	var limiters []limiter.ResourceLimiter

	if cpu != "" {
		cpuLimiter, err := limiter.NewCPULimiter(cpu)
		if err != nil {
			return nil, fmt.Errorf("invalid CPU value: %v", err)
		}
		limiters = append(limiters, cpuLimiter)
	}

	if ram != "" {
		memLimiter, err := limiter.NewMemoryLimiter(ram)
		if err != nil {
			return nil, fmt.Errorf("invalid RAM value: %v", err)
		}
		limiters = append(limiters, memLimiter)
	}

	if ioReadMax != "" || ioWriteMax != "" {
		ioInit := limiter.IOLimiterInitializer{
			ReadThrottle:  ioReadMax,
			WriteThrottle: ioWriteMax,
		}
		ioLimiter, err := limiter.NewIOLimiter(&ioInit)
		if err != nil {
			return nil, fmt.Errorf("invalid IO value: %v", err)
		}
		limiters = append(limiters, ioLimiter)
		// https://andrestc.com/post/cgroups-io/
		// I/O, by default, uses Kernel caching, which means that the I/O is not directly written to the disk, but to the Kernel cache. This cache is then written to the disk in the background. This is done to improve performance, as writing to the disk is much slower than writing to memory.
		// For this reason we need to limit also the memory in the cgroup if not already done.
		if ram == "" {
			// pick ioWriteMax with fallback to ioReadMax
			if ioWriteMax != limiter.UnlimitedIOValue {
				ram = ioWriteMax
			} else {
				ram = ioReadMax
			}
			memLimiter, err := limiter.NewMemoryLimiter(ram)
			if err != nil {
				return nil, fmt.Errorf("invalid RAM value: %v", err)
			}
			limiters = append(limiters, memLimiter)
		}
		// Known issue: a minimum amount of memory is required to start a process, so if the memory limit is too low, the process will not start.
	}

	return limiters, nil
}

func runCommand(cmd *cobra.Command, args []string) error {
	limiters, err := CreateLimiters(cpu, ram, ioReadMax, ioWriteMax)
	if err != nil {
		return err
	}

	exec := executor.NewExecutor(limiters)
	if err := exec.RunCommand(args); err != nil {
		return err
	}

	return nil
}

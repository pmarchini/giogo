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

	if ioReadMax != "" || ioWriteMax != "" {
		ioInit := limiter.IOLimiterInitializer{
			ReadThrottle:  ioReadMax,
			WriteThrottle: ioWriteMax,
		}
		ioLimiter, err := limiter.NewIOLimiter(&ioInit)
		if err != nil {
			return fmt.Errorf("invalid IO value: %v", err)
		}
		limiters = append(limiters, ioLimiter)
	}

	exec := executor.NewExecutor(limiters)
	if err := exec.RunCommand(args); err != nil {
		return err
	}

	return nil
}

package cli_test

import (
	"bytes"
	"testing"

	"github.com/pmarchini/giogo/internal/cli"
	"github.com/pmarchini/giogo/internal/limiter"
	"github.com/pmarchini/giogo/internal/utils"
	"github.com/spf13/cobra"
)

func TestExecuteHelp(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd := &cobra.Command{}
	cli.SetupRootCommand(rootCmd)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Errorf("Expected help output, got empty string")
	}
}

func TestExecuteCommand(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd := &cobra.Command{}
	cli.SetupRootCommand(rootCmd)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--cpu=0.5", "--ram=128m", "--", "echo", "Hello, kobra!"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

// Test IO limits
func TestExecuteIOLimits(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd := &cobra.Command{}
	cli.SetupRootCommand(rootCmd)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--io-read-max=128k", "--io-write-max=128k", "--", "echo", "Hello, IOLimiter!"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestCreateLimiters_IOLimitsAndNoRAM(t *testing.T) {
	cpu := ""
	ram := ""
	ioReadMax := "10m"
	ioWriteMax := "20m"

	ioWriteMaxBytes, err := utils.BytesStringToBytes(ioWriteMax)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	limiters, err := cli.CreateLimiters(cpu, ram, ioReadMax, ioWriteMax)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var memLimiter *limiter.MemoryLimiter
	for _, l := range limiters {
		if ml, ok := l.(*limiter.MemoryLimiter); ok {
			memLimiter = ml
			break
		}
	}

	if memLimiter == nil {
		t.Errorf("expected memory limiter to be included when IO limits are set and RAM is not")
	} else if memLimiter.Limit != ioWriteMaxBytes {
		t.Errorf("expected memory limiter to have the same value as ioWriteMax, got %v", memLimiter.Limit)
	}
}

func TestCreateLimiters_IOLimitsAndNoRAM_ReadLimit(t *testing.T) {
	cpu := ""
	ram := ""
	ioReadMax := "10m"
	ioWriteMax := limiter.UnlimitedIOValue

	ioReadMaxBytes, err := utils.BytesStringToBytes(ioReadMax)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	limiters, err := cli.CreateLimiters(cpu, ram, ioReadMax, ioWriteMax)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var memLimiter *limiter.MemoryLimiter
	for _, l := range limiters {
		if ml, ok := l.(*limiter.MemoryLimiter); ok {
			memLimiter = ml
			break
		}
	}

	if memLimiter == nil {
		t.Errorf("expected memory limiter to be included when IO limits are set and RAM is not")
	} else if memLimiter.Limit != ioReadMaxBytes {
		t.Errorf("expected memory limiter to have the same value as ioReadMax, got %v", memLimiter.Limit)
	}
}

func TestCreateLimiters_IOLimitsAndRAM(t *testing.T) {
	cpu := ""
	ram := "128m"
	ioReadMax := "10m"
	ioWriteMax := "20m"

	expectedRam, err := utils.BytesStringToBytes(ram)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	limiters, err := cli.CreateLimiters(cpu, ram, ioReadMax, ioWriteMax)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var memLimiter *limiter.MemoryLimiter
	for _, l := range limiters {
		if ml, ok := l.(*limiter.MemoryLimiter); ok {
			memLimiter = ml
			break
		}
	}

	if memLimiter == nil {
		t.Errorf("expected memory limiter to be included when user specifies both RAM and IO limits")
	} else if memLimiter.Limit != expectedRam {
		t.Errorf("expected memory limiter to be the one provided by the user, got %v", memLimiter.Limit)
	}
}

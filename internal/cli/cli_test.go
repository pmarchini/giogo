package cli_test

import (
	"bytes"
	"testing"

	"github.com/pmarchini/giogo/internal/cli"
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

	output := buf.String()
	if len(output) == 0 {
		t.Errorf("Expected command output, got empty string")
	}
}

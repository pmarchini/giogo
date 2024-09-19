package limiter_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pmarchini/giogo/internal/limiter"
)

// CreateMockBlockDevicesHelper creates mock block devices in the given directory
func CreateMockBlockDevicesHelper(t *testing.T, tempDir string, blockDevices []limiter.BlockDevice) {
	t.Helper() // Mark this function as a helper

	for _, device := range blockDevices {
		// Create a directory for each block device
		deviceDir := filepath.Join(tempDir, device.Name)
		if err := os.Mkdir(deviceDir, 0755); err != nil {
			t.Fatalf("Failed to create device directory: %v", err)
		}

		// Create the "dev" file with major:minor number
		devFilePath := filepath.Join(deviceDir, "dev")
		devContent := fmt.Sprintf("%d:%d", device.Major, device.Minor)
		if err := os.WriteFile(devFilePath, []byte(devContent), 0644); err != nil {
			t.Fatalf("Failed to write dev file: %v", err)
		}
	}
}

func setupMockBlockDevices(t *testing.T, blockDevices []limiter.BlockDevice) (string, func(), error) {
	t.Helper() // Mark this function as a helper
	// Create a temporary directory to simulate /sys/block
	tempDir, err := os.MkdirTemp("", "block-devices-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
		return "", nil, err
	}

	// Create the mock block devices
	CreateMockBlockDevicesHelper(t, tempDir, blockDevices)

	// Return the temp directory and a cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup, nil
}

func TestGetBlockDevices(t *testing.T) {
	// Define the mock block devices
	mockDevices := []limiter.BlockDevice{
		{Name: "sda", Major: 8, Minor: 0},
		{Name: "sdb", Major: 8, Minor: 16},
	}

	// Use the helper function to create the mock block devices
	tempDir, cleanup, err := setupMockBlockDevices(t, mockDevices)
	if err != nil {
		t.Fatalf("Failed to set up mock block devices: %v", err)
	}
	defer cleanup()

	// Call GetBlockDevices with the temporary directory
	devices, err := limiter.GetBlockDevices(tempDir)
	if err != nil {
		t.Fatalf("Error retrieving block devices: %v", err)
	}

	// Verify the number of devices retrieved
	if len(devices) != len(mockDevices) {
		t.Fatalf("Expected %d devices, got %d", len(mockDevices), len(devices))
	}

	// Verify each device's name, major, and minor numbers
	for i, device := range devices {
		expected := mockDevices[i]
		if device.Name != expected.Name || device.Major != expected.Major || device.Minor != expected.Minor {
			t.Errorf("Unexpected device at index %d: got %+v, expected %+v", i, device, expected)
		}
	}
}

// Test unparsable throttle values
func TestNewIOLimiterUnparsableThrottleValues(t *testing.T) {
	// Define the mock block devices
	mockDevices := []limiter.BlockDevice{
		{Name: "sda", Major: 8, Minor: 0},
		{Name: "sdb", Major: 8, Minor: 16},
	}

	// Use the helper function to create the mock block devices
	tempDir, cleanup, err := setupMockBlockDevices(t, mockDevices)
	if err != nil {
		t.Fatalf("Failed to set up mock block devices: %v", err)
	}
	defer cleanup()
	// Define the test cases
	tests := []struct {
		readThrottle  string
		writeThrottle string
		expectedError string
	}{
		{"invalid", "invalid", "unparsable ReadThrottle value"},
		{"invalid", "1000", "unparsable ReadThrottle value"},
		{"1000", "invalid", "unparsable WriteThrottle value"},
	}
	// Run the test cases
	for _, tt := range tests {
		t.Run(fmt.Sprintf("ReadThrottle=%s,WriteThrottle=%s", tt.readThrottle, tt.writeThrottle), func(t *testing.T) {
			init := &limiter.IOLimiterInitializer{
				ReadThrottle:           tt.readThrottle,
				WriteThrottle:          tt.writeThrottle,
				OverrideSystemBlockDir: tempDir,
			}
			_, err := limiter.NewIOLimiter(init)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if ioLimiterErr, ok := err.(*limiter.IOLimiterError); !ok {
				t.Fatalf("unexpected error type: %T", err)
			} else if ioLimiterErr.Message != tt.expectedError {
				t.Fatalf("unexpected error message: got %s, want %s", ioLimiterErr.Message, tt.expectedError)
			}
		})
	}
}

// Test invalid system block directory
func TestNewIOLimiterInvalidSystemBlockDir(t *testing.T) {
	// Define an invalid system block directory
	invalidDir := "/nonexistent"
	// Define the test cases
	tests := []struct {
		systemBlockDir string
		expectedError  string
	}{
		{"", "error retrieving block devices"},
		{invalidDir, "error retrieving block devices"},
	}
	// Run the test cases
	for _, tt := range tests {
		t.Run(fmt.Sprintf("SystemBlockDir=%s", tt.systemBlockDir), func(t *testing.T) {
			init := &limiter.IOLimiterInitializer{
				ReadThrottle:           "1M",
				WriteThrottle:          "1M",
				OverrideSystemBlockDir: tt.systemBlockDir,
			}
			_, err := limiter.NewIOLimiter(init)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if ioLimiterErr, ok := err.(*limiter.IOLimiterError); !ok {
				t.Fatalf("unexpected error type: %T", err)
			} else if ioLimiterErr.Message != tt.expectedError {
				t.Fatalf("unexpected error message: got %s, want %s", ioLimiterErr.Message, tt.expectedError)
			}
		})
	}
}

// test that the limiter package is able to create a new IOLimiter
func TestNewIOLimiter(t *testing.T) {
	// Define the mock block devices
	mockDevices := []limiter.BlockDevice{
		{Name: "sda", Major: 8, Minor: 0},
		{Name: "sdb", Major: 8, Minor: 16},
	}

	// Use the helper function to create the mock block devices
	tempDir, cleanup, err := setupMockBlockDevices(t, mockDevices)
	if err != nil {
		t.Fatalf("Failed to set up mock block devices: %v", err)
	}
	defer cleanup()
	// create a new IOLimiterInitializer
	init := &limiter.IOLimiterInitializer{
		ReadThrottle:           "1M",
		WriteThrottle:          "1M",
		OverrideSystemBlockDir: tempDir,
	}
	// create a new IOLimiter
	limiter, err := limiter.NewIOLimiter(init)
	// check if there was an error
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// check if the ReadThrottle is correct
	if limiter.ReadThrottle != 1024*1024 {
		t.Fatalf("unexpected ReadThrottle: %d", limiter.ReadThrottle)
	}
	// check if the WriteThrottle is correct
	if limiter.WriteThrottle != 1024*1024 {
		t.Fatalf("unexpected WriteThrottle: %d", limiter.WriteThrottle)
	}
	// check if the BlockDevices is correct
	if len(limiter.BlockDevices) != len(mockDevices) {
		t.Fatalf("unexpected number of BlockDevices: %d", len(limiter.BlockDevices))
	}
}

// Test Apply method of IOLimiter
func TestIOLimiterApply(t *testing.T) {
	// Define the mock block devices
	mockDevices := []limiter.BlockDevice{
		{Name: "sda", Major: 8, Minor: 0},
		{Name: "sdb", Major: 8, Minor: 16},
	}

	// Use the helper function to create the mock block devices
	tempDir, cleanup, err := setupMockBlockDevices(t, mockDevices)
	if err != nil {
		t.Fatalf("Failed to set up mock block devices: %v", err)
	}
	defer cleanup()
	// create a new IOLimiterInitializer
	init := &limiter.IOLimiterInitializer{
		ReadThrottle:           "1M",
		WriteThrottle:          "1M",
		OverrideSystemBlockDir: tempDir,
	}
	// create a new IOLimiter
	limiter, err := limiter.NewIOLimiter(init)
	// check if there was an error
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// create a new LinuxResources
	resources := &specs.LinuxResources{}
	// apply the limiter to the resources
	limiter.Apply(resources)
	// check if the ThrottleReadBpsDevice is correct
	if len(resources.BlockIO.ThrottleReadBpsDevice) != len(mockDevices) {
		t.Fatalf("unexpected number of ThrottleReadBpsDevice: %d", len(resources.BlockIO.ThrottleReadBpsDevice))
	}
	// check if the ThrottleWriteBpsDevice is correct
	if len(resources.BlockIO.ThrottleWriteBpsDevice) != len(mockDevices) {
		t.Fatalf("unexpected number of ThrottleWriteBpsDevice: %d", len(resources.BlockIO.ThrottleWriteBpsDevice))
	}
	// Check that the Major and Minor numbers are correct
	for i, device := range resources.BlockIO.ThrottleReadBpsDevice {
		expected := mockDevices[i]
		if device.Major != expected.Major || device.Minor != expected.Minor {
			t.Errorf("Unexpected device at index %d: got %+v, expected %+v", i, device, expected)
		}
	}
}

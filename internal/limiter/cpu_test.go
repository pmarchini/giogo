package limiter_test

import (
	"testing"

	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pmarchini/giogo/internal/limiter"
)

func TestNewCPULimiter(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
		wantErr  bool
	}{
		{"0.5", 0.5, false},
		{"1", 1.0, false},
		{"0", 0.0, true},
		{"-0.1", -0.1, true},
		{"1.1", 1.1, true},
		{"invalid", 0.0, true},
	}

	for _, tt := range tests {
		limiter, err := limiter.NewCPULimiter(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("NewCPULimiter(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && limiter.Fraction != tt.expected {
			t.Errorf("NewCPULimiter(%q) = %v, expected %v", tt.input, limiter.Fraction, tt.expected)
		}
	}
}

func TestCPULimiterApply(t *testing.T) {
	cpuLimiter := &limiter.CPULimiter{Fraction: 0.5}
	var resources specs.LinuxResources
	cpuLimiter.Apply(&resources)

	if resources.CPU == nil {
		t.Errorf("CPU resources not set")
	} else {
		expectedPeriod := uint64(100000)
		expectedQuota := int64(50000)
		if *resources.CPU.Period != expectedPeriod {
			t.Errorf("CPU Period = %d, expected %d", *resources.CPU.Period, expectedPeriod)
		}
		if *resources.CPU.Quota != expectedQuota {
			t.Errorf("CPU Quota = %d, expected %d", *resources.CPU.Quota, expectedQuota)
		}
	}
}

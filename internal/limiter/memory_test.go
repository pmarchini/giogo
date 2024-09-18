package limiter_test

import (
	"testing"

	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pmarchini/giogo/internal/limiter"
)

func TestNewMemoryLimiter(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		wantErr  bool
	}{
		{"128k", 128 * 1024, false},
		{"256M", 256 * 1024 * 1024, false},
		{"1G", 1 * 1024 * 1024 * 1024, false},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		limiter, err := limiter.NewMemoryLimiter(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("NewMemoryLimiter(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && limiter.Limit != tt.expected {
			t.Errorf("NewMemoryLimiter(%q) = %v, expected %v", tt.input, limiter.Limit, tt.expected)
		}
	}
}

func TestMemoryLimiterApply(t *testing.T) {
	memoryLimiter := &limiter.MemoryLimiter{Limit: 128 * 1024 * 1024}
	var resources specs.LinuxResources
	memoryLimiter.Apply(&resources)

	if resources.Memory == nil {
		t.Errorf("Memory resources not set")
	} else {
		expectedLimit := int64(128 * 1024 * 1024)
		if *resources.Memory.Limit != expectedLimit {
			t.Errorf("Memory Limit = %d, expected %d", *resources.Memory.Limit, expectedLimit)
		}
	}
}

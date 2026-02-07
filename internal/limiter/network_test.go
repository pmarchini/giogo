package limiter_test

import (
	"testing"

	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pmarchini/giogo/internal/limiter"
)

func TestNewNetworkLimiter(t *testing.T) {
	tests := []struct {
		name           string
		init           *limiter.NetworkLimiterInitializer
		wantClassID    *uint32
		wantPriority   *uint32
		wantErr        bool
	}{
		{
			name: "valid class ID only",
			init: &limiter.NetworkLimiterInitializer{
				ClassID: "100",
			},
			wantClassID:  uint32Ptr(100),
			wantPriority: nil,
			wantErr:      false,
		},
		{
			name: "valid priority only",
			init: &limiter.NetworkLimiterInitializer{
				Priority: "50",
			},
			wantClassID:  nil,
			wantPriority: uint32Ptr(50),
			wantErr:      false,
		},
		{
			name: "valid class ID and priority",
			init: &limiter.NetworkLimiterInitializer{
				ClassID:  "200",
				Priority: "75",
			},
			wantClassID:  uint32Ptr(200),
			wantPriority: uint32Ptr(75),
			wantErr:      false,
		},
		{
			name: "invalid class ID",
			init: &limiter.NetworkLimiterInitializer{
				ClassID: "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid priority",
			init: &limiter.NetworkLimiterInitializer{
				Priority: "invalid",
			},
			wantErr: true,
		},
		{
			name: "negative class ID",
			init: &limiter.NetworkLimiterInitializer{
				ClassID: "-1",
			},
			wantErr: true,
		},
		{
			name: "empty values",
			init: &limiter.NetworkLimiterInitializer{
				ClassID:  "",
				Priority: "",
			},
			wantClassID:  nil,
			wantPriority: nil,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			netLimiter, err := limiter.NewNetworkLimiter(tt.init)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewNetworkLimiter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			
			if tt.wantClassID != nil {
				if netLimiter.ClassID == nil {
					t.Errorf("Expected ClassID to be %v, got nil", *tt.wantClassID)
				} else if *netLimiter.ClassID != *tt.wantClassID {
					t.Errorf("ClassID = %v, expected %v", *netLimiter.ClassID, *tt.wantClassID)
				}
			} else if netLimiter.ClassID != nil {
				t.Errorf("Expected ClassID to be nil, got %v", *netLimiter.ClassID)
			}
			
			if tt.wantPriority != nil {
				if netLimiter.Priority == nil {
					t.Errorf("Expected Priority to be %v, got nil", *tt.wantPriority)
				} else if *netLimiter.Priority != *tt.wantPriority {
					t.Errorf("Priority = %v, expected %v", *netLimiter.Priority, *tt.wantPriority)
				}
			} else if netLimiter.Priority != nil {
				t.Errorf("Expected Priority to be nil, got %v", *netLimiter.Priority)
			}
		})
	}
}

func TestNetworkLimiterApply(t *testing.T) {
	tests := []struct {
		name     string
		limiter  *limiter.NetworkLimiter
		checkFn  func(*testing.T, *specs.LinuxResources)
	}{
		{
			name: "apply class ID only",
			limiter: &limiter.NetworkLimiter{
				ClassID: uint32Ptr(100),
			},
			checkFn: func(t *testing.T, resources *specs.LinuxResources) {
				if resources.Network == nil {
					t.Error("Network resources not set")
					return
				}
				if resources.Network.ClassID == nil {
					t.Error("ClassID not set")
				} else if *resources.Network.ClassID != 100 {
					t.Errorf("ClassID = %d, expected 100", *resources.Network.ClassID)
				}
			},
		},
		{
			name: "apply priority only",
			limiter: &limiter.NetworkLimiter{
				Priority: uint32Ptr(50),
			},
			checkFn: func(t *testing.T, resources *specs.LinuxResources) {
				if resources.Network == nil {
					t.Error("Network resources not set")
					return
				}
				if len(resources.Network.Priorities) == 0 {
					t.Error("Priorities not set")
				} else if resources.Network.Priorities[0].Priority != 50 {
					t.Errorf("Priority = %d, expected 50", resources.Network.Priorities[0].Priority)
				}
			},
		},
		{
			name: "apply both class ID and priority",
			limiter: &limiter.NetworkLimiter{
				ClassID:  uint32Ptr(200),
				Priority: uint32Ptr(75),
			},
			checkFn: func(t *testing.T, resources *specs.LinuxResources) {
				if resources.Network == nil {
					t.Error("Network resources not set")
					return
				}
				if resources.Network.ClassID == nil {
					t.Error("ClassID not set")
				} else if *resources.Network.ClassID != 200 {
					t.Errorf("ClassID = %d, expected 200", *resources.Network.ClassID)
				}
				if len(resources.Network.Priorities) == 0 {
					t.Error("Priorities not set")
				} else if resources.Network.Priorities[0].Priority != 75 {
					t.Errorf("Priority = %d, expected 75", resources.Network.Priorities[0].Priority)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resources specs.LinuxResources
			tt.limiter.Apply(&resources)
			tt.checkFn(t, &resources)
		})
	}
}

// Helper function to create a pointer to uint32
func uint32Ptr(v uint32) *uint32 {
	return &v
}

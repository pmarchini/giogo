package limiter

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// Base error for NetworkLimiter
var ErrInvalidNetworkValue = errors.New("invalid network limiter value")

// NetworkLimiterError represents a custom error with a specific message and underlying cause
type NetworkLimiterError struct {
	Message string
	Cause   error
}

func (e *NetworkLimiterError) Error() string {
	return fmt.Sprintf("%s: %v", e.Message, e.Cause)
}

func (e *NetworkLimiterError) Is(target error) bool {
	return errors.Is(e.Cause, target)
}

// Custom errors
var (
	ErrUnparsableClassID   = &NetworkLimiterError{Message: "unparsable class ID value", Cause: ErrInvalidNetworkValue}
	ErrUnparsablePriority  = &NetworkLimiterError{Message: "unparsable priority value", Cause: ErrInvalidNetworkValue}
	ErrUnparsableBandwidth = &NetworkLimiterError{Message: "unparsable bandwidth value", Cause: ErrInvalidNetworkValue}
)

// NetworkLimiter applies network resource limits
type NetworkLimiter struct {
	ClassID           *uint32
	Priority          *uint32
	MaxBandwidth      uint64 // Maximum egress bandwidth in bytes per second (0 means unlimited)
	MaxBandwidthIngress uint64 // Maximum ingress bandwidth in bytes per second (0 means unlimited)
	interfaceName     string // Network interface to apply tc rules to
	ifbDeviceName     string // IFB device name for ingress shaping
}

// Apply the network limits to the provided Linux resources
func (n *NetworkLimiter) Apply(resources *specs.LinuxResources) {
	if resources.Network == nil {
		resources.Network = &specs.LinuxNetwork{}
	}
	
	if n.ClassID != nil {
		resources.Network.ClassID = n.ClassID
	}
	
	if n.Priority != nil {
		// Note: The spec uses LinuxInterfacePriority which requires an interface name
		// We use an empty string to apply the priority to all interfaces
		// This can be extended in the future to support per-interface priorities
		resources.Network.Priorities = []specs.LinuxInterfacePriority{
			{
				Name:     "", // Empty string applies to all interfaces
				Priority: *n.Priority,
			},
		}
	}
}

// NetworkLimiterInitializer holds the initialization parameters for NetworkLimiter
type NetworkLimiterInitializer struct {
	ClassID             string
	Priority            string
	MaxBandwidth        string
	MaxBandwidthIngress string
}

// NewNetworkLimiter creates a new NetworkLimiter with validation and error handling
func NewNetworkLimiter(init *NetworkLimiterInitializer) (*NetworkLimiter, error) {
	iface := GetDefaultInterface()
	limiter := &NetworkLimiter{
		interfaceName: iface,
		ifbDeviceName: fmt.Sprintf("ifb-%s", iface), // IFB device for ingress shaping
	}
	
	if init.ClassID != "" {
		classID, err := strconv.ParseUint(init.ClassID, 10, 32)
		if err != nil {
			return nil, ErrUnparsableClassID
		}
		classIDValue := uint32(classID)
		limiter.ClassID = &classIDValue
	}
	
	if init.Priority != "" {
		priority, err := strconv.ParseUint(init.Priority, 10, 32)
		if err != nil {
			return nil, ErrUnparsablePriority
		}
		priorityValue := uint32(priority)
		limiter.Priority = &priorityValue
	}
	
	if init.MaxBandwidth != "" {
		bandwidth, err := parseBandwidth(init.MaxBandwidth)
		if err != nil {
			return nil, &NetworkLimiterError{Message: "unparsable egress bandwidth value", Cause: err}
		}
		limiter.MaxBandwidth = bandwidth
	}
	
	if init.MaxBandwidthIngress != "" {
		bandwidth, err := parseBandwidth(init.MaxBandwidthIngress)
		if err != nil {
			return nil, &NetworkLimiterError{Message: "unparsable ingress bandwidth value", Cause: err}
		}
		limiter.MaxBandwidthIngress = bandwidth
	}
	
	return limiter, nil
}

// parseBandwidth parses bandwidth string (e.g., "1m", "500k") to bytes per second
func parseBandwidth(s string) (uint64, error) {
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
	return uint64(value * float64(multiplier)), nil
}

// Setup implements the LifecycleLimiter interface
// Sets up traffic control rules for bandwidth limiting
func (n *NetworkLimiter) Setup() error {
	// Validate interface name
	if n.interfaceName == "" {
		return fmt.Errorf("network interface name not set")
	}
	
	// Setup egress (outgoing) traffic control if bandwidth limit is set
	if n.ClassID != nil && n.MaxBandwidth > 0 {
		if err := setupHTB(n.interfaceName, *n.ClassID, n.MaxBandwidth); err != nil {
			return err
		}
	}
	
	// Setup ingress (incoming) traffic control if bandwidth limit is set
	if n.ClassID != nil && n.MaxBandwidthIngress > 0 {
		if err := setupIngressHTB(n.interfaceName, n.ifbDeviceName, *n.ClassID, n.MaxBandwidthIngress); err != nil {
			// Cleanup egress if it was setup
			if n.MaxBandwidth > 0 {
				cleanupHTB(n.interfaceName)
			}
			return err
		}
	}
	
	return nil
}

// Cleanup implements the LifecycleLimiter interface
// Removes traffic control rules set up by Setup
func (n *NetworkLimiter) Cleanup() error {
	var firstError error
	
	// Cleanup egress if it was setup
	if n.ClassID != nil && n.MaxBandwidth > 0 {
		if err := cleanupHTB(n.interfaceName); err != nil && firstError == nil {
			firstError = err
		}
	}
	
	// Cleanup ingress if it was setup
	if n.ClassID != nil && n.MaxBandwidthIngress > 0 {
		if err := cleanupIngressHTB(n.interfaceName, n.ifbDeviceName); err != nil && firstError == nil {
			firstError = err
		}
	}
	
	return firstError
}

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
	ClassID     *uint32
	Priority    *uint32
	MaxBandwidth uint64 // Maximum bandwidth in bytes per second (0 means unlimited)
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
	ClassID      string
	Priority     string
	MaxBandwidth string
}

// NewNetworkLimiter creates a new NetworkLimiter with validation and error handling
func NewNetworkLimiter(init *NetworkLimiterInitializer) (*NetworkLimiter, error) {
	limiter := &NetworkLimiter{}
	
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
			return nil, &NetworkLimiterError{Message: "unparsable bandwidth value", Cause: err}
		}
		limiter.MaxBandwidth = bandwidth
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

// SetupTrafficControl sets up tc (traffic control) rules for bandwidth limiting
// This method should be called after the cgroup is created and classID is set
func (n *NetworkLimiter) SetupTrafficControl(interfaceName string) error {
	// Only setup tc if we have both classID and bandwidth limit
	if n.ClassID == nil || n.MaxBandwidth == 0 {
		return nil
	}
	
	return setupHTB(interfaceName, *n.ClassID, n.MaxBandwidth)
}

// CleanupTrafficControl removes tc rules set up by SetupTrafficControl
func (n *NetworkLimiter) CleanupTrafficControl(interfaceName string) error {
	// Only cleanup if we have a classID (indicating we set up tc)
	if n.ClassID == nil || n.MaxBandwidth == 0 {
		return nil
	}
	
	return cleanupHTB(interfaceName)
}

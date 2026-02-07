package limiter

import (
	"errors"
	"fmt"
	"strconv"

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
	ErrUnparsableClassID = &NetworkLimiterError{Message: "unparsable class ID value", Cause: ErrInvalidNetworkValue}
	ErrUnparsablePriority = &NetworkLimiterError{Message: "unparsable priority value", Cause: ErrInvalidNetworkValue}
)

// NetworkLimiter applies network resource limits
type NetworkLimiter struct {
	ClassID  *uint32
	Priority *uint32
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
	ClassID  string
	Priority string
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
	
	return limiter, nil
}

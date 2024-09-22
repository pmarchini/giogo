package limiter

import (
	"errors"
	"fmt"
	"strconv"

	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// Base error for CPULimiter
var ErrInvalidFraction = errors.New("invalid CPU limiter fraction")

// CPULimiterError represents a custom error with a specific message and underlying cause
type CPULimiterError struct {
	Message string
	Cause   error
}

func (e *CPULimiterError) Error() string {
	return fmt.Sprintf("%s: %v", e.Message, e.Cause)
}

func (e *CPULimiterError) Is(target error) bool {
	return errors.Is(e.Cause, target)
}

// Custom errors
var (
	ErrUnparsableValue = &CPULimiterError{Message: "unparsable value", Cause: ErrInvalidFraction}
	ErrFractionTooLow  = &CPULimiterError{Message: "fraction too low", Cause: ErrInvalidFraction}
	ErrFractionTooHigh = &CPULimiterError{Message: "fraction too high", Cause: ErrInvalidFraction}
)

// CPULimiter applies CPU resource limits based on a fraction of usage
type CPULimiter struct {
	Fraction float64
}

// Apply the CPU limits to the provided Linux resources
func (c *CPULimiter) Apply(resources *specs.LinuxResources) {
	period := uint64(100000)
	quota := int64(c.Fraction * float64(period))
	resources.CPU = &specs.LinuxCPU{
		Period: &period,
		Quota:  &quota,
	}
}

// NewCPULimiter creates a new CPULimiter with validation and error handling
func NewCPULimiter(value string) (*CPULimiter, error) {
	fraction, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, ErrUnparsableValue
	}

	if fraction <= 0 {
		return nil, ErrFractionTooLow
	}
	if fraction > 1 {
		return nil, ErrFractionTooHigh
	}

	return &CPULimiter{Fraction: fraction}, nil
}

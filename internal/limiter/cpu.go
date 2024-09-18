package limiter

import (
    "strconv"

    specs "github.com/opencontainers/runtime-spec/specs-go"
)

type CPULimiter struct {
    Fraction float64
}

func (c *CPULimiter) Apply(resources *specs.LinuxResources) {
    period := uint64(100000)
    quota := int64(c.Fraction * float64(period))
    resources.CPU = &specs.LinuxCPU{
        Period: &period,
        Quota:  &quota,
    }
}

func NewCPULimiter(value string) (*CPULimiter, error) {
    fraction, err := strconv.ParseFloat(value, 64)
    if err != nil || fraction <= 0 || fraction > 1 {
        return nil, err
    }
    return &CPULimiter{Fraction: fraction}, nil
}

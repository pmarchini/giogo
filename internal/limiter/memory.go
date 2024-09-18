package limiter

import (
    "github.com/pmarchini/giogo/internal/utils"

    specs "github.com/opencontainers/runtime-spec/specs-go"
)

type MemoryLimiter struct {
    Limit int64
}

func (m *MemoryLimiter) Apply(resources *specs.LinuxResources) {
    resources.Memory = &specs.LinuxMemory{
        Limit: &m.Limit,
    }
}

func NewMemoryLimiter(value string) (*MemoryLimiter, error) {
    limit, err := utils.ParseMemory(value)
    if err != nil {
        return nil, err
    }
    return &MemoryLimiter{Limit: limit}, nil
}

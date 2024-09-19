package limiter

import (
	"github.com/pmarchini/giogo/internal/utils"

	specs "github.com/opencontainers/runtime-spec/specs-go"
)

type MemoryLimiter struct {
	Limit uint64
}

func (m *MemoryLimiter) Apply(resources *specs.LinuxResources) {
	limit := int64(m.Limit)
	resources.Memory = &specs.LinuxMemory{
		Limit: &limit,
	}
}

func NewMemoryLimiter(value string) (*MemoryLimiter, error) {
	limit, err := utils.BytesStringToBytes(value)
	if err != nil {
		return nil, err
	}
	return &MemoryLimiter{Limit: limit}, nil
}

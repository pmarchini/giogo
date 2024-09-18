package core

import (
	"github.com/stretchr/testify/mock"
	// Replace with your actual package path
)

// MockCgroupManager is a mock implementation of CgroupManager for testing purposes
type MockCgroupManager struct {
	mock.Mock
}

func (m *MockCgroupManager) AddProcess(pid int) error {
	args := m.Called(pid)
	return args.Error(0)
}

func (m *MockCgroupManager) Delete() error {
	args := m.Called()
	return args.Error(0)
}

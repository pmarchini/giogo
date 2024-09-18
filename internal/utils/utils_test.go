package utils_test

import (
	"testing"

	"github.com/pmarchini/giogo/internal/utils"
)

func TestParseMemory(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		wantErr  bool
	}{
		{"128k", 128 * 1024, false},
		{"256K", 256 * 1024, false},
		{"512m", 512 * 1024 * 1024, false},
		{"1G", 1 * 1024 * 1024 * 1024, false},
		{"1.5g", int64(1.5 * 1024 * 1024 * 1024), false},
		{"1024", 1024, false},
		{"", 0, true},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		result, err := utils.ParseMemory(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseMemory(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && result != tt.expected {
			t.Errorf("ParseMemory(%q) = %d, expected %d", tt.input, result, tt.expected)
		}
	}
}

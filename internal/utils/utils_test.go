package utils_test

import (
	"testing"

	"github.com/pmarchini/giogo/internal/utils"
)

func TestBytesStringToBytes(t *testing.T) {
	tests := []struct {
		input    string
		expected uint64
		wantErr  bool
	}{
		{"128k", 128 * 1024, false},
		{"256K", 256 * 1024, false},
		{"512m", 512 * 1024 * 1024, false},
		{"1G", 1 * 1024 * 1024 * 1024, false},
		{"1.5g", uint64(1.5 * 1024 * 1024 * 1024), false},
		{"1024", 1024, false},
		{"", 0, true},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		result, err := utils.BytesStringToBytes(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("BytesStringToBytes(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && result != tt.expected {
			t.Errorf("BytesStringToBytes(%q) = %d, expected %d", tt.input, result, tt.expected)
		}
	}
}

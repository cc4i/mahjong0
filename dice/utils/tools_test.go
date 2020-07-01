package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRandString(t *testing.T) {
	tests := []struct {
		name   string
		input  int
		output int
	}{
		{"Testing length = 0", 0, 0},
		{"Testing length = 3", 3, 3},
		{"Testing length = 5", 5, 5},
		{"Testing length = 8", 8, 8},
		{"Testing length = -1", -1, 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := RandString(test.input)
			assert.Equal(t, len(s), test.output)
		})
	}
}

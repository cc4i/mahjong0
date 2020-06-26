package engine

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
func TestFindPair(t *testing.T) {
	tests := []struct {
		name string
		input string
		key string
		value string
	}{
		{"string with space1", "abc= efg", "abc", "efg"},
		{"string with space2", "abc = efg", "abc", "efg"},
		{"string without space", "abc=efg", "abc", "efg"},
		{"string with special character", "abc= efg=", "abc", "efg="},
		{"string with special character2", "abc= efg====", "abc", "efg===="},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			k,v, err := FindPair(test.input)
			assert.NoError(t, err)
			assert.Equal(t, test.key, k)
			assert.Equal(t, test.value, v)
		})
	}
}

func TestExecutionPlan_ExtractValue(t *testing.T) {

}

func TestExecutionPlan_ScanOutput(t *testing.T) {

}
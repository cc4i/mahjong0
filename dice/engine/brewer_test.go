package engine

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFindPair(t *testing.T) {
	tests := []struct {
		name  string
		input string
		key   string
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
			k, v, err := FindPair(test.input)
			assert.NoError(t, err)
			assert.Equal(t, test.key, k)
			assert.Equal(t, test.value, v)
		})
	}
}

var sid = []string{"000-000-001", "000-000-002"}
var tsEmpty = Ts{}
var tsNormal = Ts{}
var plan = &ExecutionPlan{}

func BuildData() {

}

func TestExecutionPlan_ExtractValue(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output map[string]string
	}{
		{"normal output", ``, map[string]string{"x": "y", "x1": "y1"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(testing *testing.T) {

		})
	}
}

func TestExecutionPlan_ScanOutput(t *testing.T) {

}

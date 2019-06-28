package common

import (
	"strings"
	"testing"
)

func TestValidContractName(t *testing.T) {
	testCases := []struct {
		in       string
		expected string
	}{
		{
			in:       "000",
			expected: "contract name length expect",
		},
		{
			in:       "0000000000000000",
			expected: "contract name does not fit the rule of contract",
		},
		{
			in:       "_11111111111111.",
			expected: "contract name does not fit the rule of contract",
		},
		{
			in:       "a11111111111111.",
			expected: "contract name does not fit the rule of contract",
		},
		{
			in:       "_11111111111111_",
			expected: "",
		},
	}
	for index := range testCases {
		t.Log("index ", index)
		actual := ValidContractName(testCases[index].in)
		expected := testCases[index].expected
		if actual == nil && expected == "" {
			continue
		}

		if actual == nil && expected != "" {
			t.Error("expected:", expected, "actual:", actual)
		}
		if !strings.HasPrefix(actual.Error(), expected) {
			t.Error("expected:", expected, ",actual:", actual)
		}
	}
}

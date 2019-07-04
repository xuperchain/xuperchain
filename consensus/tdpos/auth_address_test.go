package tdpos

import (
	"testing"
)

type tempStruct struct {
	initiator   string
	authRequire []string
}

func TestAuthAddress(t *testing.T) {
	target := "ak1"
	tp := &TDpos{}
	testCases := []struct {
		in       *tempStruct
		expected bool
	}{
		{
			in: &tempStruct{
				initiator:   "ak1",
				authRequire: []string{"ak1", "ak2"},
			},
			expected: true,
		},
		{
			in: &tempStruct{
				initiator:   "ak2",
				authRequire: []string{"ak3"},
			},
			expected: false,
		},
		{
			in: &tempStruct{
				initiator:   "",
				authRequire: nil,
			},
			expected: false,
		},
		{
			in: &tempStruct{
				initiator:   "ak1",
				authRequire: nil,
			},
			expected: true,
		},
	}
	for index := range testCases {
		actual := tp.isAuthAddress(target, testCases[index].in.initiator, testCases[index].in.authRequire)
		expected := testCases[index].expected
		if actual != expected {
			t.Errorf("expect %v, actual %v, target:%s not auth, index:%d", expected, actual, target, index)
		}
	}
}

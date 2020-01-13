package probe

import (
	"github.com/xuperchain/log15"
	"testing"
)

func TestNewSpeedCalc(t *testing.T) {
	sc := NewSpeedCalc("test")
	if sc.start != 0 {
		t.Errorf("test NewSpeedCalc failed, expected %v actual %v", 0, sc.start)
	}
}

func TestClear(t *testing.T) {
	testCases := []struct {
		in       *SpeedCalc
		expected int64
	}{
		{
			in: &SpeedCalc{
				start: 5,
				dict: map[string]int64{
					"test": 5,
				},
			},
			expected: 0,
		},
	}

	for index := range testCases {
		testCases[index].in.Clear()
		if testCases[index].in.start != testCases[index].expected {
			t.Errorf("test Clear failed, expected %v actual %v", testCases[index].expected, testCases[index].in.start)
		}
	}
}

func TestAdd(t *testing.T) {
	testCases := map[string]struct {
		in       *SpeedCalc
		expected int64
	}{
		"test flag exist": {
			in: &SpeedCalc{
				start: 5,
				dict: map[string]int64{
					"test flag exist": 5,
				},
			},
			expected: 6,
		},

		"test flag not exist": {
			in: &SpeedCalc{
				start: 5,
				dict: map[string]int64{
					"test flag exist": 5,
				},
			},
			expected: 1,
		},
	}

	for testName, testCase := range testCases {
		testCase.in.Add(testName)
		if testCase.in.dict[testName] != testCase.expected {
			t.Errorf("test Add failed, expected %v actual %v", testCase.expected, testCase.in.dict[testName])
		}
	}
}

func TestShowInfo(t *testing.T) {
	testCases := []struct {
		in *SpeedCalc
	}{
		{
			in: &SpeedCalc{
				start: 5,
				dict: map[string]int64{
					"test": 5,
				},
				maxSpeed: map[string]float64{
					"test": 300,
				},
			},
		},
	}

	for index := range testCases {
		testCases[index].in.ShowInfo(log.New())
	}
}

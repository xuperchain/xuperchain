package jstest

import "testing"

func TestAssert(t *testing.T) {
	runner, err := NewRunner(&RunOption{
		InGoTest: true,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()

	err = runner.RunFile("./testdata/jstest.test.js")
	if err != nil {
		t.Fatal(err)
	}
}

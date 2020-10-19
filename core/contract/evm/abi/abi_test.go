package abi

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestNewAbi(t *testing.T) {
	abiFile := "abi_test.bin"
	method := "getUint"

	abiBuf, err := ioutil.ReadFile(abiFile)
	if err != nil {
		t.Error(err)
	}

	args := make(map[string]interface{})
	enc1, err := New(abiBuf)
	if err != nil {
		t.Error(err)
	}
	input1, err := enc1.Encode(method, args)
	if err != nil {
		t.Error(err)
	}

	// [0 2 103 164]
	fmt.Printf("%v\n", input1)

	enc2, err := LoadFile(abiFile)
	input2, err := enc2.Encode(method, args)
	if err != nil {
		t.Error(err)
	}

	// [0 2 103 164]
	fmt.Printf("%v\n", input2)
}

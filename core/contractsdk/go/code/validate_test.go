package code

import (
	"math/big"
	"testing"
)

func TestInt(t *testing.T) {
	args := &struct {
		Count *big.Int `json:"count" validate:"lte=30"`
	}{}

	data1 := map[string][]byte{
		"count": []byte(big.NewInt(30).String()),
	}
	if err := Unmarshal(data1, args); err != nil {
		t.Error(err)
	}
	data2 := map[string][]byte{
		"count": []byte(big.NewInt(31).String()),
	}
	if err := Unmarshal(data2, args); err == nil {
		t.Error("error")
	}
}

func TestNil(t *testing.T) {
	type testcase struct {
		Name *big.Int `json:"name" validate:"required"`
	}
	case1 := &testcase{}
	case2 := &testcase{}

	data := map[string][]byte{
		"name": []byte("100"),
	}
	if err := Unmarshal(data, case1); err != nil {
		t.Error(err)
	}
	if err := Unmarshal(map[string][]byte{}, case2); err == nil {
		t.Error("want error,get nil")
	}
}

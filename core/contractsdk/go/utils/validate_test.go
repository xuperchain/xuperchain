package utils

import (
	"fmt"
	"math/big"
	"testing"
)

func TestValidate(t *testing.T) {
	args := &struct {
		Count *big.Int   `json:"count"`
		Value *big.Float `json:"value"`
		Name  string     `json:"name"`
	}{}
	data := map[string][]byte{
		"count": []byte("1234"),
		"value": []byte("12.34"),
		"name":  []byte("zhagnsan"),
	}
	err := Validate(data, args)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(args.Count)
	fmt.Println(args.Value)
	fmt.Println(args.Name)
}

func TestValidate2(t *testing.T) {

	args := &struct {
		Count *big.Int   `json:"count" required:"true"`
		Value *big.Float `json:"value" required:"true"`
		Name  string     `json:"name" required:"true"`
	}{}
	data := map[string][]byte{
		"count": []byte("1234"),
		"value": []byte("12.34"),
		//"name":  []byte("zhagnsan"),
	}
	err := Validate(data, args)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(args.Count)
	fmt.Println(args.Value)
	fmt.Println(args.Name)
}

package abi

import (
	"fmt"

	"golang.org/x/crypto/sha3"
)

// FunctionIDSize is the length of the function selector
const FunctionIDSize = 4

type FunctionSpec struct {
	Name       string
	FunctionID FunctionID
	Constant   bool
	Inputs     []Argument
	Outputs    []Argument
}

type FunctionID [FunctionIDSize]byte

func NewFunctionSpec(name string, inputs, outputs []Argument) *FunctionSpec {
	sig := Signature(name, inputs)
	return &FunctionSpec{
		Name:       name,
		FunctionID: GetFunctionID(sig),
		Constant:   false,
		Inputs:     inputs,
		Outputs:    outputs,
	}
}

func GetFunctionID(signature string) (id FunctionID) {
	hash := sha3.NewLegacyKeccak256()
	hash.Write([]byte(signature))
	copy(id[:], hash.Sum(nil)[:4])
	return
}

func Signature(name string, args []Argument) string {
	return name + argsToSignature(args, false)
}

// Sets this function as constant
func (f *FunctionSpec) SetConstant() *FunctionSpec {
	f.Constant = true
	return f
}

func (f *FunctionSpec) String() string {
	return f.Name + argsToSignature(f.Inputs, true) +
		" returns " + argsToSignature(f.Outputs, true)
}

func (fs FunctionID) Bytes() []byte {
	return fs[:]
}

func argsToSignature(args []Argument, addIndexedName bool) (str string) {
	str = "("
	for i, a := range args {
		if i > 0 {
			str += ","
		}
		str += a.EVM.GetSignature()
		if addIndexedName && a.Indexed {
			str += " indexed"
		}
		if a.IsArray {
			if a.ArrayLength > 0 {
				str += fmt.Sprintf("[%d]", a.ArrayLength)
			} else {
				str += "[]"
			}
		}
		if addIndexedName && a.Name != "" {
			str += " " + a.Name
		}
	}
	str += ")"
	return
}

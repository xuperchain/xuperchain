package abi

import (
	"fmt"

	"github.com/hyperledger/burrow/execution/evm/abi"
)

type ABI struct {
	spec *abi.Spec
}

func LoadFile(fpath string) (*ABI, error) {
	spec, err := abi.LoadPath(fpath)
	if err != nil {
		return nil, err
	}
	return newABI(spec), nil
}

func New(buf []byte) (*ABI, error) {
	spec, err := abi.ReadSpec(buf)
	if err != nil {
		return nil, err
	}
	return newABI(spec), nil
}

func newABI(spec *abi.Spec) *ABI {
	return &ABI{
		spec: spec,
	}
}

func (a *ABI) Encode(methodName string, args map[string]interface{}) ([]byte, error) {
	if methodName == "" {
		if a.spec.Constructor != nil {
			return a.encodeMethod(a.spec.Constructor, args)
		}
		return nil, nil
	}
	method, ok := a.spec.Functions[methodName]
	if !ok {
		return nil, fmt.Errorf("method %s not found", methodName)
	}
	return a.encodeMethod(method, args)
}

func (a *ABI) encodeMethod(method *abi.FunctionSpec, args map[string]interface{}) ([]byte, error) {
	var inputs []interface{}
	for _, input := range method.Inputs {
		v, ok := args[input.Name]
		if !ok {
			return nil, fmt.Errorf("arg name %s not found", input.Name)
		}
		// v, err := encode(input.Type, v)
		// if err != nil {
		// 	return nil, err
		// }
		// fmt.Printf("encode %s => %v\n", input.Name, v)
		inputs = append(inputs, v)
	}
	out, _, err := a.spec.Pack(method.Name, inputs...)
	return out, err
}

//func decodeHex(str string) []byte {
//	var buf []byte
//	n, err := fmt.Sscanf(str, "0x%x", &buf)
//	if err != nil {
//		panic(err)
//	}
//	if n != 1 {
//		panic("bad address")
//	}
//	return buf
//}

// func encodeInt(x interface{}, size int) interface{} {
// 	if size <= 64 {
// 		panic(fmt.Sprintf("unsupported int size %d", size))
// 	}
// 	return new(big.Int).SetBytes(decodeHex(x.(string)))
// }

// func encodeAddress(x interface{}) interface{} {
// 	var addr common.Address
// 	buf := decodeHex(x.(string))
// 	copy(addr[:], buf)
// 	return addr
// }

// func encode(t abi.Type, x interface{}) (interface{}, error) {
// 	switch t.T {
// 	case abi.IntTy, abi.UintTy:
// 		return encodeInt(x, t.Size), nil
// 	case abi.BoolTy, abi.StringTy, abi.SliceTy, abi.ArrayTy:
// 		return x, nil
// 	case abi.AddressTy:
// 		return encodeAddress(x), nil
// 	default:
// 		panic("Invalid type")
// 	}
// }

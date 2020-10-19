package abi

import (
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"path"
	"path/filepath"
	"reflect"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/deploy/compile"
	"github.com/hyperledger/burrow/logging"
)

// Variable exist to unpack return values into, so have both the return
// value and its name
type Variable struct {
	Name  string
	Value string
}

// LoadPath loads one abi file or finds all files in a directory
func LoadPath(abiFileOrDirs ...string) (*Spec, error) {
	if len(abiFileOrDirs) == 0 {
		return nil, fmt.Errorf("no ABI file or directory provided")
	}

	specs := make([]*Spec, 0)

	for _, dir := range abiFileOrDirs {
		err := filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("error returned while walking abiDir '%s': %v", dir, err)
			}
			ext := filepath.Ext(path)
			if fi.IsDir() || !(ext == ".bin" || ext == ".abi") {
				return nil
			}
			abiSpc, err := ReadSpecFile(path)
			if err != nil {
				return fmt.Errorf("error parsing abi file at %s: %v", path, err)
			}
			specs = append(specs, abiSpc)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return MergeSpec(specs), nil
}

// EncodeFunctionCallFromFile ABI encodes a function call based on ABI in file, and the
// arguments specified as strings.
// The abiFileName specifies the name of the ABI file, and abiPath the path where it can be found.
// The fname specifies which function should called, if
// it doesn't exist exist the fallback function will be called. If fname is the empty
// string, the constructor is called. The arguments must be specified in args. The count
// must match the function being called.
// Returns the ABI encoded function call, whether the function is constant according
// to the ABI (which means it does not modified contract state)
func EncodeFunctionCallFromFile(abiFileName, abiPath, funcName string, logger *logging.Logger, args ...interface{}) ([]byte, *FunctionSpec, error) {
	abiSpecBytes, err := readAbi(abiPath, abiFileName, logger)
	if err != nil {
		return []byte{}, nil, err
	}

	return EncodeFunctionCall(abiSpecBytes, funcName, logger, args...)
}

// EncodeFunctionCall ABI encodes a function call based on ABI in string abiData
// and the arguments specified as strings.
// The fname specifies which function should called, if
// it doesn't exist exist the fallback function will be called. If fname is the empty
// string, the constructor is called. The arguments must be specified in args. The count
// must match the function being called.
// Returns the ABI encoded function call, whether the function is constant according
// to the ABI (which means it does not modified contract state)
func EncodeFunctionCall(abiData, funcName string, logger *logging.Logger, args ...interface{}) ([]byte, *FunctionSpec, error) {
	logger.TraceMsg("Packing Call via ABI",
		"spec", abiData,
		"function", funcName,
		"arguments", fmt.Sprintf("%v", args),
	)

	abiSpec, err := ReadSpec([]byte(abiData))
	if err != nil {
		logger.InfoMsg("Failed to decode abi spec",
			"abi", abiData,
			"error", err.Error(),
		)
		return nil, nil, err
	}

	packedBytes, funcSpec, err := abiSpec.Pack(funcName, args...)
	if err != nil {
		logger.InfoMsg("Failed to encode abi spec",
			"abi", abiData,
			"error", err.Error(),
		)
		return nil, nil, err
	}

	return packedBytes, funcSpec, nil
}

// DecodeFunctionReturnFromFile ABI decodes the return value from a contract function call.
func DecodeFunctionReturnFromFile(abiLocation, binPath, funcName string, resultRaw []byte, logger *logging.Logger) ([]*Variable, error) {
	abiSpecBytes, err := readAbi(binPath, abiLocation, logger)
	if err != nil {
		return nil, err
	}
	logger.TraceMsg("ABI Specification (Decode)", "spec", abiSpecBytes)

	// Unpack the result
	return DecodeFunctionReturn(abiSpecBytes, funcName, resultRaw)
}

func DecodeFunctionReturn(abiData, name string, data []byte) ([]*Variable, error) {
	abiSpec, err := ReadSpec([]byte(abiData))
	if err != nil {
		return nil, err
	}

	var args []Argument

	if name == "" {
		args = abiSpec.Constructor.Outputs
	} else {
		if _, ok := abiSpec.Functions[name]; ok {
			args = abiSpec.Functions[name].Outputs
		} else {
			args = abiSpec.Fallback.Outputs
		}
	}

	if args == nil {
		return nil, fmt.Errorf("no such function")
	}
	vars := make([]*Variable, len(args))

	if len(args) == 0 {
		return nil, nil
	}

	vals := make([]interface{}, len(args))
	for i := range vals {
		vals[i] = new(string)
	}
	err = Unpack(args, data, vals...)
	if err != nil {
		return nil, err
	}

	for i, a := range args {
		if a.Name != "" {
			vars[i] = &Variable{Name: a.Name, Value: *(vals[i].(*string))}
		} else {
			vars[i] = &Variable{Name: fmt.Sprintf("%d", i), Value: *(vals[i].(*string))}
		}
	}

	return vars, nil
}

// Spec

// ReadSpecFile reads an ABI file from a file
func ReadSpecFile(filename string) (*Spec, error) {
	specBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return ReadSpec(specBytes)
}

// Struct reflection

// SpecFromStructReflect generates a FunctionSpec where the arguments and return values are
// described a struct. Both args and rets should be set to the return value of reflect.TypeOf()
// with the respective struct as an argument.
func SpecFromStructReflect(fname string, args reflect.Type, rets reflect.Type) *FunctionSpec {
	inputs := make([]Argument, args.NumField())
	outputs := make([]Argument, rets.NumField())

	for i := 0; i < args.NumField(); i++ {
		f := args.Field(i)
		a := typeFromReflect(f.Type)
		a.Name = f.Name
		inputs[i] = a
	}

	for i := 0; i < rets.NumField(); i++ {
		f := rets.Field(i)
		a := typeFromReflect(f.Type)
		a.Name = f.Name
		outputs[i] = a
	}

	return NewFunctionSpec(fname, inputs, outputs)
}

func SpecFromFunctionReflect(fname string, v reflect.Value, skipIn, skipOut int) *FunctionSpec {
	t := v.Type()

	if t.Kind() != reflect.Func {
		panic(fmt.Sprintf("%s is not a function", t.Name()))
	}

	inputs := make([]Argument, t.NumIn()-skipIn)
	outputs := make([]Argument, t.NumOut()-skipOut)

	for i := range inputs {
		inputs[i] = typeFromReflect(t.In(i + skipIn))
	}

	for i := range outputs {
		outputs[i] = typeFromReflect(t.Out(i))
	}

	return NewFunctionSpec(fname, inputs, outputs)
}

func GetPackingTypes(args []Argument) []interface{} {
	res := make([]interface{}, len(args))

	for i, a := range args {
		if a.IsArray {
			t := reflect.TypeOf(a.EVM.getGoType())
			res[i] = reflect.MakeSlice(reflect.SliceOf(t), int(a.ArrayLength), 0).Interface()
		} else {
			res[i] = a.EVM.getGoType()
		}
	}

	return res
}

func typeFromReflect(v reflect.Type) Argument {
	arg := Argument{Name: v.Name()}

	if v == reflect.TypeOf(crypto.Address{}) {
		arg.EVM = EVMAddress{}
	} else if v == reflect.TypeOf(big.Int{}) {
		arg.EVM = EVMInt{M: 256}
	} else {
		if v.Kind() == reflect.Array {
			arg.IsArray = true
			arg.ArrayLength = uint64(v.Len())
			v = v.Elem()
		} else if v.Kind() == reflect.Slice {
			arg.IsArray = true
			v = v.Elem()
		}

		switch v.Kind() {
		case reflect.Bool:
			arg.EVM = EVMBool{}
		case reflect.String:
			arg.EVM = EVMString{}
		case reflect.Uint64:
			arg.EVM = EVMUint{M: 64}
		case reflect.Int64:
			arg.EVM = EVMInt{M: 64}
		default:
			panic(fmt.Sprintf("no mapping for type %v", v.Kind()))
		}
	}

	return arg
}

func readAbi(root, contract string, logger *logging.Logger) (string, error) {
	p := path.Join(root, stripHex(contract))
	if _, err := os.Stat(p); err != nil {
		logger.TraceMsg("abifile not found", "tried", p)
		p = path.Join(root, stripHex(contract)+".bin")
		if _, err = os.Stat(p); err != nil {
			logger.TraceMsg("abifile not found", "tried", p)
			return "", fmt.Errorf("abi doesn't exist for =>\t%s", p)
		}
	}
	logger.TraceMsg("Found ABI file", "path", p)
	sol, err := compile.LoadSolidityContract(p)
	if err != nil {
		return "", err
	}
	return string(sol.Abi), nil
}

func stripHex(s string) string {
	if len(s) > 1 {
		if s[:2] == "0x" {
			s = s[2:]
			if len(s)%2 != 0 {
				s = "0" + s
			}
			return s
		}
	}
	return s
}

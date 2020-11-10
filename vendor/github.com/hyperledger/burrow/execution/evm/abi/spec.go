package abi

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/hyperledger/burrow/crypto"
)

// Token to use in deploy yaml in order to indicate call to the fallback function.
const FallbackFunctionName = "()"

// Spec is the ABI for contract decoded.
type Spec struct {
	Constructor  *FunctionSpec
	Fallback     *FunctionSpec
	Functions    map[string]*FunctionSpec
	EventsByName map[string]*EventSpec
	EventsByID   map[EventID]*EventSpec
}

type specJSON struct {
	Name            string
	Type            string
	Inputs          []argumentJSON
	Outputs         []argumentJSON
	StateMutability string
	Anonymous       bool
}

func NewSpec() *Spec {
	return &Spec{
		// Zero value for constructor and fallback function is assumed when those functions are not present
		Constructor:  &FunctionSpec{},
		Fallback:     &FunctionSpec{},
		EventsByName: make(map[string]*EventSpec),
		EventsByID:   make(map[EventID]*EventSpec),
		Functions:    make(map[string]*FunctionSpec),
	}
}

// ReadSpec takes an ABI and decodes it for futher use
func ReadSpec(specBytes []byte) (*Spec, error) {
	var specJ []specJSON
	err := json.Unmarshal(specBytes, &specJ)
	if err != nil {
		// The abi spec file might a bin file, with the Abi under the Abi field in json
		var binFile struct {
			Abi []specJSON
		}
		err = json.Unmarshal(specBytes, &binFile)
		if err != nil {
			return nil, err
		}
		specJ = binFile.Abi
	}

	abiSpec := NewSpec()

	for _, s := range specJ {
		switch s.Type {
		case "constructor":
			abiSpec.Constructor.Inputs, err = readArgSpec(s.Inputs)
			if err != nil {
				return nil, err
			}
		case "fallback":
			abiSpec.Fallback.Inputs = make([]Argument, 0)
			abiSpec.Fallback.Outputs = make([]Argument, 0)
			abiSpec.Fallback.SetConstant()
			abiSpec.Functions[FallbackFunctionName] = abiSpec.Fallback
		case "event":
			ev := new(EventSpec)
			err = ev.unmarshalSpec(&s)
			if err != nil {
				return nil, err
			}
			abiSpec.EventsByName[ev.Name] = ev
			abiSpec.EventsByID[ev.ID] = ev
		case "function":
			inputs, err := readArgSpec(s.Inputs)
			if err != nil {
				return nil, err
			}
			outputs, err := readArgSpec(s.Outputs)
			if err != nil {
				return nil, err
			}
			abiSpec.Functions[s.Name] = NewFunctionSpec(s.Name, inputs, outputs).SetConstant()
		}
	}

	return abiSpec, nil
}

// MergeSpec takes multiple Specs and merges them into once structure. Note that
// the same function name or event name can occur in different abis, so there might be
// some information loss.
func MergeSpec(abiSpec []*Spec) *Spec {
	newSpec := NewSpec()

	for _, s := range abiSpec {
		for n, f := range s.Functions {
			newSpec.Functions[n] = f
		}

		// Different Abis can have the Event name, but with a different signature
		// Loop over the signatures, as these are less likely to have collisions
		for _, e := range s.EventsByID {
			newSpec.EventsByName[e.Name] = e
			newSpec.EventsByID[e.ID] = e
		}
	}

	return newSpec
}

func (spec *Spec) GetEventAbi(id EventID, addresses crypto.Address) (*EventSpec, error) {
	eventSpec, ok := spec.EventsByID[id]
	if !ok {
		return nil, fmt.Errorf("could not find ABI for event with ID %v", id)
	}
	return eventSpec, nil
}

// Pack ABI encodes a function call. The fname specifies which function should called, if
// it doesn't exist exist the fallback function will be called. If fname is the empty
// string, the constructor is called. The arguments must be specified in args. The count
// must match the function being called.
// Returns the ABI encoded function call, whether the function is constant according
// to the ABI (which means it does not modified contract state)
func (spec *Spec) Pack(fname string, args ...interface{}) ([]byte, *FunctionSpec, error) {
	var funcSpec *FunctionSpec
	var argSpec []Argument
	if fname != "" {
		if _, ok := spec.Functions[fname]; ok {
			funcSpec = spec.Functions[fname]
		} else {
			return nil, nil, fmt.Errorf("unknown function in Pack: %s", fname)
		}
	} else {
		if spec.Constructor.Inputs != nil {
			funcSpec = spec.Constructor
		} else {
			return nil, nil, fmt.Errorf("contract does not have a constructor")
		}
	}

	argSpec = funcSpec.Inputs

	packed := make([]byte, 0)

	if fname != "" {
		packed = funcSpec.FunctionID[:]
	}

	packedArgs, err := Pack(argSpec, args...)
	if err != nil {
		return nil, nil, err
	}

	return append(packed, packedArgs...), funcSpec, nil
}

// Unpack decodes the return values from a function call
func (spec *Spec) Unpack(data []byte, fname string, args ...interface{}) error {
	var funcSpec *FunctionSpec
	var argSpec []Argument
	if fname != "" {
		if _, ok := spec.Functions[fname]; ok {
			funcSpec = spec.Functions[fname]
		} else {
			funcSpec = spec.Fallback
		}
	} else {
		funcSpec = spec.Constructor
	}

	argSpec = funcSpec.Outputs

	if argSpec == nil {
		return fmt.Errorf("unknown function in Unpack: %s", fname)
	}

	return unpack(argSpec, data, func(i int) interface{} {
		return args[i]
	})
}

func (spec *Spec) UnpackWithID(data []byte, args ...interface{}) error {
	var argSpec []Argument

	var id FunctionID
	copy(id[:], data)
	for _, fspec := range spec.Functions {
		if id == fspec.FunctionID {
			argSpec = fspec.Outputs
		}
	}

	if argSpec == nil {
		return fmt.Errorf("unknown function in UnpackWithID: %x", id)
	}

	return unpack(argSpec, data[4:], func(i int) interface{} {
		return args[i]
	})
}

func readArgSpec(argsJ []argumentJSON) ([]Argument, error) {
	args := make([]Argument, len(argsJ))
	var err error

	for i, a := range argsJ {
		args[i].Name = a.Name
		args[i].Indexed = a.Indexed

		baseType := a.Type
		isArray := regexp.MustCompile(`(.*)\[([0-9]+)\]`)
		m := isArray.FindStringSubmatch(a.Type)
		if m != nil {
			args[i].IsArray = true
			args[i].ArrayLength, err = strconv.ParseUint(m[2], 10, 32)
			if err != nil {
				return nil, err
			}
			baseType = m[1]
		} else if strings.HasSuffix(a.Type, "[]") {
			args[i].IsArray = true
			baseType = strings.TrimSuffix(a.Type, "[]")
		}

		isM := regexp.MustCompile("(bytes|uint|int)([0-9]+)")
		m = isM.FindStringSubmatch(baseType)
		if m != nil {
			M, err := strconv.ParseUint(m[2], 10, 32)
			if err != nil {
				return nil, err
			}
			switch m[1] {
			case "bytes":
				if M < 1 || M > 32 {
					return nil, fmt.Errorf("bytes%d is not valid type", M)
				}
				args[i].EVM = EVMBytes{M}
			case "uint":
				if M < 8 || M > 256 || (M%8) != 0 {
					return nil, fmt.Errorf("uint%d is not valid type", M)
				}
				args[i].EVM = EVMUint{M}
			case "int":
				if M < 8 || M > 256 || (M%8) != 0 {
					return nil, fmt.Errorf("uint%d is not valid type", M)
				}
				args[i].EVM = EVMInt{M}
			}
			continue
		}

		isMxN := regexp.MustCompile("(fixed|ufixed)([0-9]+)x([0-9]+)")
		m = isMxN.FindStringSubmatch(baseType)
		if m != nil {
			M, err := strconv.ParseUint(m[2], 10, 32)
			if err != nil {
				return nil, err
			}
			N, err := strconv.ParseUint(m[3], 10, 32)
			if err != nil {
				return nil, err
			}
			if M < 8 || M > 256 || (M%8) != 0 {
				return nil, fmt.Errorf("%s is not valid type", baseType)
			}
			if N == 0 || N > 80 {
				return nil, fmt.Errorf("%s is not valid type", baseType)
			}
			if m[1] == "fixed" {
				args[i].EVM = EVMFixed{N: N, M: M, signed: true}
			} else if m[1] == "ufixed" {
				args[i].EVM = EVMFixed{N: N, M: M, signed: false}
			} else {
				panic(m[1])
			}
			continue
		}
		switch baseType {
		case "uint":
			args[i].EVM = EVMUint{M: 256}
		case "int":
			args[i].EVM = EVMInt{M: 256}
		case "address":
			args[i].EVM = EVMAddress{}
		case "bool":
			args[i].EVM = EVMBool{}
		case "fixed":
			args[i].EVM = EVMFixed{M: 128, N: 8, signed: true}
		case "ufixed":
			args[i].EVM = EVMFixed{M: 128, N: 8, signed: false}
		case "bytes":
			args[i].EVM = EVMBytes{M: 0}
		case "string":
			args[i].EVM = EVMString{}
		default:
			// Assume it is a type of Contract
			args[i].EVM = EVMAddress{}
		}
	}

	return args, nil
}

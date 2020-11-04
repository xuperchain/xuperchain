package native

import (
	"fmt"
	"math/big"
	"reflect"
	"runtime"
	"strings"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/engine"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/permission"
)

// Function is metadata for native functions. Act as call targets
// for the EVM when collected into an Contract. Can be used to generate
// bindings in a smart contract languages.
type Function struct {
	// Comment describing function's purpose, parameters, and return value
	Comment string
	// Permissions required to call function
	PermFlag permission.PermFlag
	// Whether this function writes to state
	Pure bool
	// Native function to which calls will be dispatched when a containing
	F interface{}
	// Following fields are for only for memoization
	// The name of the contract to which this function belongs (if any)
	contractName string
	// Function name (used to form signature)
	name string
	// The abi
	abi *abi.FunctionSpec
	// Address of containing contract
	address   crypto.Address
	externals engine.Dispatcher
	logger    *logging.Logger
}

var _ Native = &Function{}

// Context is the first argument to any native function. This struct carries
// all the context an native needs to access e.g. state in burrow.
type Context struct {
	State engine.State
	engine.CallParams
	// TODO: this allows us to call back to EVM contracts if we wish - make use of it somewhere...
	externals engine.Dispatcher
	Logger    *logging.Logger
}

// Created a new function mounted directly at address (i.e. no Solidity contract or function selection)
func NewFunction(comment string, address crypto.Address, permFlag permission.PermFlag, f interface{}) (*Function, error) {
	function := &Function{
		Comment:  comment,
		PermFlag: permFlag,
		F:        f,
	}
	err := function.init(address)
	if err != nil {
		return nil, err
	}
	return function, nil
}

func (f *Function) SetExternals(externals engine.Dispatcher) {
	// Wrap it to treat nil dispatcher as empty list
	f.externals = engine.NewDispatchers(externals)
}

func (f *Function) Call(state engine.State, params engine.CallParams,
	transfer func(from, to crypto.Address, amount *big.Int) error) ([]byte, error) {
	return Call(state, params, f.execute, transfer)
}

func (f *Function) execute(state engine.State, params engine.CallParams, transfer func(crypto.Address, crypto.Address, *big.Int) error) ([]byte, error) {
	// check if we have permission to call this function
	hasPermission, err := HasPermission(state.CallFrame, params.Caller, f.PermFlag)
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, &errors.LacksNativePermission{Address: params.Caller, NativeName: f.name}
	}

	ctx := Context{
		State:      state,
		CallParams: params,
		externals:  f.externals,
		Logger:     f.logger,
	}
	fnv := reflect.ValueOf(f.F)
	fnt := fnv.Type()

	args := []reflect.Value{reflect.ValueOf(ctx)}

	if f.abi != nil {
		arguments := reflect.New(fnt.In(1))
		err = abi.Unpack(f.abi.Inputs, params.Input, arguments.Interface())
		if err != nil {
			return nil, err
		}
		args = append(args, arguments.Elem())
	}

	rets := fnv.Call(args)
	if !rets[1].IsNil() {
		return nil, rets[1].Interface().(error)
	}

	ret := rets[0].Interface()
	if f.abi != nil {
		return abi.Pack(f.abi.Outputs, ret)
	}

	output, ok := ret.([]byte)
	if !ok {
		return nil, fmt.Errorf("function has no associated ABI but returns %T instead of []byte", ret)
	}
	return output, nil
}

func (f *Function) FullName() string {
	if f.contractName != "" {
		return f.contractName + "." + f.name
	}
	return f.name
}

func (f *Function) Address() crypto.Address {
	return f.address
}

// Signature returns the function signature as would be used for ABI hashing
func (f *Function) Signature() string {
	argTypeNames := make([]string, len(f.abi.Inputs))
	for i, arg := range f.abi.Inputs {
		argTypeNames[i] = arg.EVM.GetSignature()
	}
	return fmt.Sprintf("%s(%s)", f.name, strings.Join(argTypeNames, ","))
}

// For templates
func (f *Function) Name() string {
	return f.name
}

// NArgs returns the number of function arguments
func (f *Function) NArgs() int {
	return len(f.abi.Inputs)
}

// Abi returns the FunctionSpec for this function
func (f *Function) Abi() *abi.FunctionSpec {
	return f.abi
}

func (f *Function) ContractMeta() []*acm.ContractMeta {
	// FIXME: can we do something here - a function is not a contract...
	return nil
}

func (f *Function) String() string {
	return fmt.Sprintf("SNativeFunction{Name: %s; Inputs: %d; Outputs: %d}",
		f.name, len(f.abi.Inputs), len(f.abi.Outputs))
}

func (f *Function) init(address crypto.Address) error {
	// Get name of function
	t := reflect.TypeOf(f.F)
	v := reflect.ValueOf(f.F)
	// v.String() for functions returns the empty string
	fullyQualifiedName := runtime.FuncForPC(v.Pointer()).Name()
	a := strings.Split(fullyQualifiedName, ".")
	f.name = a[len(a)-1]

	if t.NumIn() != 1 && t.NumIn() != 2 {
		return fmt.Errorf("native function %s must have a one or two arguments", fullyQualifiedName)
	}

	if t.NumOut() != 2 {
		return fmt.Errorf("native function %s must return a single struct and an error", fullyQualifiedName)
	}

	if t.In(0) != reflect.TypeOf(Context{}) {
		return fmt.Errorf("first agument of %s must be struct Context", fullyQualifiedName)
	}

	if t.NumIn() == 2 {
		f.abi = abi.SpecFromStructReflect(f.name, t.In(1), t.Out(0))
	}
	f.address = address
	return nil
}

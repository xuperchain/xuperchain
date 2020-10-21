// Copyright Monax Industries Limited
// SPDX-License-Identifier: Apache-2.0

package native

import (
	"fmt"
	"math/big"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/engine"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/hyperledger/burrow/logging"
)

//
// Native (go) contracts are dispatched based on account permissions and can access
// and modify an account's permissions
//

// Instructions on adding an native function. First declare a function like so:
//
// func unsetBase(context Context, args unsetBaseArgs) (unsetBaseRets, error) {
// }
//
// The name of the function will be used as the name of the function in solidity. The
// first arguments is Context; this will give you access to state, and the logger
// etc. The second arguments must be a struct type. The members of this struct must be
// exported (start with uppercase letter), and they will be converted into arguments
// for the solidity function, with the same types. The first return value is a struct
// which defines the return values from solidity just like the arguments.
//
// The second return value must be error. If non-nil is returned for error, then
// the current transaction will be aborted and the execution will stop.
//
// For each contract you will need to create a Contract{} struct,
// with the function listed. Only the PermFlag and the function F needs to be filled
// in for each Function. Add this to the SNativeContracts() function.

// Contract is metadata for native contract. Acts as a call target
// from the EVM. Can be used to generate bindings in a smart contract languages.
type Contract struct {
	// Comment describing purpose of native contract and reason for assembling
	// the particular functions
	Comment string
	// Name of the native contract
	Name          string
	functionsByID map[abi.FunctionID]*Function
	functions     []*Function
	address       crypto.Address
	logger        *logging.Logger
}

var _ Native = &Contract{}

// Create a new native contract description object by passing a comment, name
// and a list of member functions descriptions
func NewContract(name string, comment string, logger *logging.Logger, fs ...Function) (*Contract, error) {
	address := AddressFromName(name)
	functionsByID := make(map[abi.FunctionID]*Function, len(fs))
	functions := make([]*Function, len(fs))
	logger = logger.WithScope("NativeContract")
	for i, f := range fs {
		function := f
		err := function.init(address)
		if err != nil {
			return nil, err
		}
		if function.abi == nil {
			return nil, fmt.Errorf("could not establish ABI for function - contract functions must have a " +
				"struct second argument in order to establish ABI")
		}
		function.contractName = name
		function.logger = logger
		fid := function.abi.FunctionID
		otherF, ok := functionsByID[fid]
		if ok {
			return nil, fmt.Errorf("function with ID %x already defined: %s", fid, otherF.Signature())
		}
		functionsByID[fid] = &function
		functions[i] = &function
	}
	return &Contract{
		Comment:       comment,
		Name:          name,
		functionsByID: functionsByID,
		functions:     functions,
		address:       address,
		logger:        logger,
	}, nil
}

// Dispatch is designed to be called from the EVM once a native contract
// has been selected.
func (c *Contract) Call(state engine.State, params engine.CallParams,
	transfer func(from, to crypto.Address, amount *big.Int) error) (output []byte, err error) {
	if len(params.Input) < abi.FunctionIDSize {
		return nil, errors.Errorf(errors.Codes.NativeFunction,
			"Burrow Native dispatch requires a 4-byte function identifier but arguments are only %v bytes long",
			len(params.Input))
	}

	var id abi.FunctionID
	copy(id[:], params.Input)
	function, err := c.FunctionByID(id)
	if err != nil {
		return nil, err
	}

	params.Input = params.Input[abi.FunctionIDSize:]

	return function.Call(state, params, transfer)
}

func (c *Contract) SetExternals(externals engine.Dispatcher) {
	for _, f := range c.functions {
		f.SetExternals(externals)
	}
}

func (c *Contract) FullName() string {
	return c.Name
}

// We define the address of an native contact as the last 20 bytes of the sha3
// hash of its name
func (c *Contract) Address() crypto.Address {
	return c.address
}

// Get function by calling identifier FunctionSelector
func (c *Contract) FunctionByID(id abi.FunctionID) (*Function, errors.CodedError) {
	f, ok := c.functionsByID[id]
	if !ok {
		return nil,
			errors.Errorf(errors.Codes.NativeFunction, "unknown native function with ID %x", id)
	}
	return f, nil
}

// Get function by name
func (c *Contract) FunctionByName(name string) *Function {
	for _, f := range c.functions {
		if f.name == name {
			return f
		}
	}
	return nil
}

// Get functions in order of declaration
func (c *Contract) Functions() []*Function {
	functions := make([]*Function, len(c.functions))
	copy(functions, c.functions)
	return functions
}

func (c *Contract) ContractMeta() []*acm.ContractMeta {
	// FIXME: make this return actual ABI metadata
	metadata := "{}"
	metadataHash := acmstate.GetMetadataHash(metadata)
	return []*acm.ContractMeta{
		{
			CodeHash:     []byte(c.Name),
			MetadataHash: metadataHash[:],
			Metadata:     metadata,
		},
	}
}

func AddressFromName(name string) (address crypto.Address) {
	hash := crypto.Keccak256([]byte(name))
	copy(address[:], hash[len(hash)-crypto.AddressLength:])
	return
}

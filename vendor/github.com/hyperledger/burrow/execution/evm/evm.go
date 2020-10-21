// Copyright Monax Industries Limited
// SPDX-License-Identifier: Apache-2.0

package evm

import (
	"fmt"
	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/execution/engine"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/execution/native"
	"github.com/hyperledger/burrow/logging"
)

const (
	DataStackInitialCapacity    = 1024
	MaximumAllowedBlockLookBack = 256
	uint64Length                = 8
)

type EVM struct {
	options  Options
	sequence uint64
	// Provide any foreign dispatchers to allow calls between VMs
	externals engine.Dispatcher
	// User dispatcher.CallableProvider to get access to other VMs
	logger *logging.Logger
}

// Options are parameters that are generally stable across a burrow configuration.
// Defaults will be used for any zero values.
type Options struct {
	MemoryProvider           func(errors.Sink) Memory
	Natives                  *native.Natives
	Nonce                    []byte
	DebugOpcodes             bool
	DumpTokens               bool
	CallStackMaxDepth        uint64
	DataStackInitialCapacity uint64
	DataStackMaxDepth        uint64
	Logger                   *logging.Logger
}

func New(options Options) *EVM {
	// Set defaults
	if options.MemoryProvider == nil {
		options.MemoryProvider = DefaultDynamicMemoryProvider
	}
	if options.Logger == nil {
		options.Logger = logging.NewNoopLogger()
	}
	if options.Natives == nil {
		options.Natives = native.MustDefaultNatives()
	}
	vm := &EVM{
		options: options,
	}
	// TODO: ultimately this wiring belongs a level up, but for the time being it is convenient to handle it here
	// since we need to both intercept backend state to serve up natives AND connect the external dispatchers
	engine.Connect(vm, options.Natives)
	vm.logger = options.Logger.WithScope("NewVM").With("evm_nonce", options.Nonce)
	return vm
}

func Default() *EVM {
	return New(Options{})
}

// Initiate an EVM call against the provided state pushing events to eventSink. code should contain the EVM bytecode,
// input the CallData (readable by CALLDATALOAD), value the amount of native token to transfer with the call
// an quantity metering the number of computational steps available to the execution according to the gas schedule.
func (vm *EVM) Execute(st acmstate.ReaderWriter, blockchain engine.Blockchain, eventSink exec.EventSink,
	params engine.CallParams, code []byte) ([]byte, error) {

	// Make it appear as if natives are stored in state
	st = native.NewState(vm.options.Natives, st)

	state := engine.State{
		CallFrame:  engine.NewCallFrame(st).WithMaxCallStackDepth(vm.options.CallStackMaxDepth),
		Blockchain: blockchain,
		EventSink:  eventSink,
	}

	output, err := vm.Contract(code).Call(state, params, st.Transfer)
	if err == nil {
		// Only sync back when there was no exception
		err = state.CallFrame.Sync()
	}
	// Always return output - we may have a reverted exception for which the return is meaningful
	return output, err
}

// Sets a new nonce and resets the sequence number. Nonces should only be used once!
// A global counter or sufficient randomness will work.
func (vm *EVM) SetNonce(nonce []byte) {
	vm.options.Nonce = nonce
	vm.sequence = 0
}

func (vm *EVM) SetLogger(logger *logging.Logger) {
	vm.logger = logger
}

func (vm *EVM) Dispatch(acc *acm.Account) engine.Callable {
	// Try external calls then fallback to EVM
	callable := vm.externals.Dispatch(acc)
	if callable != nil {
		return callable
	}
	// This supports empty code calls
	return vm.Contract(acc.EVMCode)
}

func (vm *EVM) SetExternals(externals engine.Dispatcher) {
	vm.externals = externals
}

func (vm *EVM) Contract(code []byte) *Contract {
	return &Contract{
		EVM:  vm,
		Code: NewCode(code),
	}
}

func (vm *EVM) debugf(format string, a ...interface{}) {
	if vm.options.DebugOpcodes {
		fmt.Printf(format, a...)
	}
}

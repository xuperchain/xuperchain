package engine

import (
	"math/big"
	"time"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/exec"
)

type Blockchain interface {
	LastBlockHeight() uint64
	LastBlockTime() time.Time
	BlockHash(height uint64) ([]byte, error)
}

type CallParams struct {
	CallType exec.CallType
	Origin   crypto.Address
	Caller   crypto.Address
	Callee   crypto.Address
	Input    []byte
	Value    *big.Int
	Gas      *uint64
}

// Effectively a contract, but can either represent a single function or a contract with multiple functions and a selector
type Callable interface {
	Call(state State, params CallParams, transfer func(crypto.Address, crypto.Address, *big.Int) error) (output []byte, err error)
	// Fully qualified name of the callable, including any contract qualification
}

type CallableFunc func(st State, params CallParams) (output []byte, err error)

func (c CallableFunc) Call(state State, params CallParams) (output []byte, err error) {
	return c(state, params)
}

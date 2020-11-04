// Copyright Monax Industries Limited
// SPDX-License-Identifier: Apache-2.0

package evm

import (
	"fmt"
	"math"
	"math/big"

	. "github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/native"
)

// Not goroutine safe
type Stack struct {
	slice       []Word256
	maxCapacity uint64
	ptr         int

	gas     *uint64
	errSink errors.Sink
}

func NewStack(errSink errors.Sink, initialCapacity uint64, maxCapacity uint64, gas *uint64) *Stack {
	return &Stack{
		slice:       make([]Word256, initialCapacity),
		ptr:         0,
		maxCapacity: maxCapacity,
		gas:         gas,
		errSink:     errSink,
	}
}

func (st *Stack) Push(d Word256) {
	st.useGas(native.GasStackOp)
	err := st.ensureCapacity(uint64(st.ptr) + 1)
	if err != nil {
		st.pushErr(errors.Codes.DataStackOverflow)
		return
	}
	st.slice[st.ptr] = d
	st.ptr++
}

func (st *Stack) Pop() Word256 {
	st.useGas(native.GasStackOp)
	if st.ptr == 0 {
		st.pushErr(errors.Codes.DataStackUnderflow)
		return Zero256
	}
	st.ptr--
	return st.slice[st.ptr]
}

// currently only called after sha3.Sha3
func (st *Stack) PushBytes(bz []byte) {
	if len(bz) != 32 {
		panic("Invalid bytes size: expected 32")
	}
	st.Push(LeftPadWord256(bz))
}

func (st *Stack) PushAddress(address crypto.Address) {
	st.Push(address.Word256())
}

func (st *Stack) Push64(i uint64) {
	st.Push(Uint64ToWord256(i))
}

func (st *Stack) Pop64() uint64 {
	d := st.Pop()
	if Is64BitOverflow(d) {
		st.pushErr(errors.Errorf(errors.Codes.IntegerOverflow, "uint64 overflow from word: %v", d))
		return 0
	}
	return Uint64FromWord256(d)
}

// Pushes the bigInt as a Word256 encoding negative values in 32-byte twos complement and returns the encoded result
func (st *Stack) PushBigInt(bigInt *big.Int) Word256 {
	word := LeftPadWord256(U256(bigInt).Bytes())
	st.Push(word)
	return word
}

func (st *Stack) PopBigIntSigned() *big.Int {
	return S256(st.PopBigInt())
}

func (st *Stack) PopBigInt() *big.Int {
	d := st.Pop()
	return new(big.Int).SetBytes(d[:])
}

func (st *Stack) PopBytes() []byte {
	return st.Pop().Bytes()
}

func (st *Stack) PopAddress() crypto.Address {
	return crypto.AddressFromWord256(st.Pop())
}

func (st *Stack) Len() int {
	return st.ptr
}

func (st *Stack) Swap(n int) {
	st.useGas(native.GasStackOp)
	if st.ptr < n {
		st.pushErr(errors.Codes.DataStackUnderflow)
		return
	}
	st.slice[st.ptr-n], st.slice[st.ptr-1] = st.slice[st.ptr-1], st.slice[st.ptr-n]
}

func (st *Stack) Dup(n int) {
	st.useGas(native.GasStackOp)
	if st.ptr < n {
		st.pushErr(errors.Codes.DataStackUnderflow)
		return
	}
	st.Push(st.slice[st.ptr-n])
}

// Not an opcode, costs no gas.
func (st *Stack) Peek() Word256 {
	if st.ptr == 0 {
		st.pushErr(errors.Codes.DataStackUnderflow)
		return Zero256
	}
	return st.slice[st.ptr-1]
}

func (st *Stack) Print(n int) {
	fmt.Println("### stack ###")
	if st.ptr > 0 {
		nn := n
		if st.ptr < n {
			nn = st.ptr
		}
		for j, i := 0, st.ptr-1; i > st.ptr-1-nn; i-- {
			fmt.Printf("%-3d  %X\n", j, st.slice[i])
			j += 1
		}
	} else {
		fmt.Println("-- empty --")
	}
	fmt.Println("#############")
}

func Is64BitOverflow(word Word256) bool {
	for i := 0; i < len(word)-8; i++ {
		if word[i] != 0 {
			return true
		}
	}
	return false
}

// Ensures the current stack can hold a new element. Will only grow the
// backing array (will not shrink).
func (st *Stack) ensureCapacity(newCapacity uint64) error {
	// Maximum length of a slice that allocates memory is the same as the native int max size
	// We could rethink this limit, but we don't want different validators to disagree on
	// transaction validity so we pick the lowest common denominator
	if newCapacity > math.MaxInt32 {
		// If we ever did want more than an int32 of space then we would need to
		// maintain multiple pages of memory
		return fmt.Errorf("cannot address memory beyond a maximum index "+
			"with int32 width (%v bytes)", math.MaxInt32)
	}
	newCapacityInt := int(newCapacity)
	// We're already big enough so return
	if newCapacityInt <= len(st.slice) {
		return nil
	}
	if st.maxCapacity > 0 && newCapacity > st.maxCapacity {
		return fmt.Errorf("cannot grow memory because it would exceed the "+
			"current maximum limit of %v bytes", st.maxCapacity)
	}
	// Ensure the backing array of slice is big enough
	// Grow the memory one word at time using the pre-allocated zeroWords to avoid
	// unnecessary allocations. Use append to make use of any spare capacity in
	// the slice's backing array.
	for newCapacityInt > cap(st.slice) {
		// We'll trust Go exponentially grow our arrays (at first).
		st.slice = append(st.slice, Zero256)
	}
	// Now we've ensured the backing array of the slice is big enough we can
	// just re-slice (even if len(mem.slice) < newCapacity)
	st.slice = st.slice[:newCapacity]
	return nil
}

func (st *Stack) useGas(gasToUse uint64) {
	if *st.gas > gasToUse {
		*st.gas -= gasToUse
	} else {
		st.pushErr(errors.Codes.InsufficientGas)
	}
}

func (st *Stack) pushErr(err errors.CodedError) {
	st.errSink.PushError(err)
}

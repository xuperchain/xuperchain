package js

import (
	"fmt"
	"unsafe"
)

const (
	nanHead = 0x7FF80000
)

const (
	// ValueUndefined is the ref of Undefined
	ValueUndefined = 0
)

const (
	// ValueNaN is the ref of Nan
	ValueNaN Ref = nanHead<<32 | iota
	// ValueZero is the ref of number 0
	ValueZero
	// ValueNull is the ref of Null
	ValueNull
	// ValueTrue is the ref of True
	ValueTrue
	// ValueFalse is the ref of False
	ValueFalse
	// ValueGlobal is the ref of global
	ValueGlobal
	// ValueMemory is the ref of wasm Memory object
	ValueMemory
	// ValueGo is the ref of Go object
	ValueGo
)

// Ref represents the unique id of a js object
type Ref int64

// Number return ref as a number, if ref not a number, false will be returned
func (r Ref) Number() (int64, bool) {
	f := *(*float64)(unsafe.Pointer(&r))
	if f == f {
		return int64(f), true
	}
	return 0, false
}

// ID return the id of ref
func (r Ref) ID() int64 {
	id := uint32(r)
	return int64(id)
}

// String return the debug string of ref
func (r Ref) String() string {
	n, ok := r.Number()
	if ok {
		return fmt.Sprintf("%d", n)
	}
	return fmt.Sprintf("0x%x", int64(r))
}

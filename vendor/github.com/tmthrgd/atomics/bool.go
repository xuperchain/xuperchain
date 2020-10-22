// Copyright 2017 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License that can be found in
// the LICENSE file.

package atomics

import (
	"strconv"
	"sync/atomic"
)

func boolToUint32(v bool) uint32 {
	if v {
		return 1
	}

	return 0
}

// Bool provides an atomic bool.
type Bool struct {
	noCopy noCopy
	val    uint32
}

// NewBool returns an atomic bool with a given value.
func NewBool(val bool) *Bool {
	return &Bool{val: boolToUint32(val)}
}

// Raw returns a pointer to the underlying uint32.
//
// It is only safe to access the pointer with methods from the
// sync/atomic package. Use caution if manually dereferencing.
//
// The true value is stored as one, false is stored as zero.
//
// The behaviour of Bool is undefined if this value is set
// to anything other than zero or one.
func (b *Bool) Raw() *uint32 {
	return &b.val
}

// Load returns the value of the bool.
func (b *Bool) Load() (val bool) {
	return atomic.LoadUint32(&b.val) != 0
}

// Store sets the value of the bool.
func (b *Bool) Store(val bool) {
	atomic.StoreUint32(&b.val, boolToUint32(val))
}

// Swap sets the value of the bool and returns the old value.
func (b *Bool) Swap(new bool) (old bool) {
	return atomic.SwapUint32(&b.val, boolToUint32(new)) != 0
}

// CompareAndSwap sets the value of the bool to new but only
// if it currently has the value old. It returns true if the swap
// succeeded.
func (b *Bool) CompareAndSwap(old, new bool) (swapped bool) {
	return atomic.CompareAndSwapUint32(&b.val, boolToUint32(old), boolToUint32(new))
}

// Set is a wrapper for Swap(true).
func (b *Bool) Set() (old bool) {
	return b.Swap(true)
}

// Reset is a wrapper for Swap(false).
func (b *Bool) Reset() (old bool) {
	return b.Swap(false)
}

// String implements fmt.Stringer.
func (b *Bool) String() string {
	return strconv.FormatBool(b.Load())
}

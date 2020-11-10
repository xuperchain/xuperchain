// Copyright 2017 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License that can be found in
// the LICENSE file.

package atomics

import (
	"sync/atomic"
	"unsafe"
)

func pointerToString(val unsafe.Pointer) string {
	if val != nil {
		return *(*string)(val)
	}

	return ""
}

func addressOfString(val string) unsafe.Pointer {
	return unsafe.Pointer(&val)
}

func stringToPointer(val string) unsafe.Pointer {
	if val != "" {
		return addressOfString(val)
	}

	return nil
}

// String provides an atomic string.
type String struct {
	noCopy noCopy
	val    *string
}

type stringPtr struct {
	val unsafe.Pointer
}

// NewString returns an atomic string with a given value.
func NewString(val string) *String {
	return &String{val: &val}
}

// Load returns the value of the string.
func (s *String) Load() string {
	p := (*stringPtr)(unsafe.Pointer(s))
	return pointerToString(atomic.LoadPointer(&p.val))
}

// Store sets the value of the string.
func (s *String) Store(val string) {
	p := (*stringPtr)(unsafe.Pointer(s))
	atomic.StorePointer(&p.val, stringToPointer(val))
}

// Swap sets the value of the string and returns the old value.
func (s *String) Swap(new string) (old string) {
	p := (*stringPtr)(unsafe.Pointer(s))
	return pointerToString(atomic.SwapPointer(&p.val, stringToPointer(new)))
}

// Reset is a wrapper for Swap("").
func (s *String) Reset() (old string) {
	return s.Swap("")
}

// String implements fmt.Stringer.
func (s *String) String() string {
	return s.Load()
}

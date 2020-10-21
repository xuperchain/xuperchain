// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !amd64 gccgo appengine

// Package bitwise provides an efficient implementation of a & b == b.
package bitwise

import (
	"runtime"
	"unsafe"
)

const wordSize = int(unsafe.Sizeof(uintptr(0)))
const supportsUnaligned = runtime.GOARCH == "386" || runtime.GOARCH == "amd64"

func fastAndEqBytes(a, b []byte) bool {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}

	w := n / wordSize
	if w > 0 {
		aw := *(*[]uintptr)(unsafe.Pointer(&a))
		bw := *(*[]uintptr)(unsafe.Pointer(&b))

		for i := 0; i < w; i++ {
			if aw[i] & bw[i] != bw[i] {
				return false
			}
		}
	}

	for i := (n - n%wordSize); i < n; i++ {
		if a[i] & b[i] != b[i] {
			return false
		}
	}

	return true
}

func safeAndEqBytes(a, b []byte) bool {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}

	for i := 0; i < n; i++ {
		if a[i] & b[i] != b[i] {
			return false
		}
	}

	return true
}

// AndEq returns true iff a & b == b
func AndEq(a, b []byte) bool {
	if supportsUnaligned {
		return fastAndEqBytes(a, b)
	}

	// TODO: if (a, b) have common alignment
	// we could still try fastAndEqBytes.
	return safeAndEqBytes(a, b)
}

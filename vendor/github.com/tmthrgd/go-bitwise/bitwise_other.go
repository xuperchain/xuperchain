// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !amd64 gccgo appengine

// Package bitwise provides efficient implementations of xor/xnor/and/and-not/nand/or/nor/not.
package bitwise

import (
	"runtime"
	"unsafe"
)

const wordSize = int(unsafe.Sizeof(uintptr(0)))
const supportsUnaligned = runtime.GOARCH == "386" || runtime.GOARCH == "amd64"

func fastXORBytes(dst, a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	if len(dst) < n {
		n = len(dst)
	}

	w := n / wordSize
	if w > 0 {
		dw := *(*[]uintptr)(unsafe.Pointer(&dst))
		aw := *(*[]uintptr)(unsafe.Pointer(&a))
		bw := *(*[]uintptr)(unsafe.Pointer(&b))

		for i := 0; i < w; i++ {
			dw[i] = aw[i] ^ bw[i]
		}
	}

	for i := n - n%wordSize; i < n; i++ {
		dst[i] = a[i] ^ b[i]
	}

	return n
}

func safeXORBytes(dst, a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	if len(dst) < n {
		n = len(dst)
	}

	for i := 0; i < n; i++ {
		dst[i] = a[i] ^ b[i]
	}

	return n
}

// XOR sets each element in according to dst[i] = a[i] XOR b[i]
func XOR(dst, a, b []byte) int {
	if supportsUnaligned {
		return fastXORBytes(dst, a, b)
	}

	// TODO: if (dst, a, b) have common alignment
	// we could still try fastXORBytes.
	return safeXORBytes(dst, a, b)
}

func fastXNORBytes(dst, a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	if len(dst) < n {
		n = len(dst)
	}

	w := n / wordSize
	if w > 0 {
		dw := *(*[]uintptr)(unsafe.Pointer(&dst))
		aw := *(*[]uintptr)(unsafe.Pointer(&a))
		bw := *(*[]uintptr)(unsafe.Pointer(&b))

		for i := 0; i < w; i++ {
			dw[i] = ^(aw[i] ^ bw[i])
		}
	}

	for i := n - n%wordSize; i < n; i++ {
		dst[i] = ^(a[i] ^ b[i])
	}

	return n
}

func safeXNORBytes(dst, a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	if len(dst) < n {
		n = len(dst)
	}

	for i := 0; i < n; i++ {
		dst[i] = ^(a[i] ^ b[i])
	}

	return n
}

// XNOR sets each element in according to dst[i] = NOT (a[i] XOR b[i])
func XNOR(dst, a, b []byte) int {
	if supportsUnaligned {
		return fastXNORBytes(dst, a, b)
	}

	// TODO: if (dst, a, b) have common alignment
	// we could still try fastXNORBytes.
	return safeXNORBytes(dst, a, b)
}

func fastAndBytes(dst, a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	if len(dst) < n {
		n = len(dst)
	}

	w := n / wordSize
	if w > 0 {
		dw := *(*[]uintptr)(unsafe.Pointer(&dst))
		aw := *(*[]uintptr)(unsafe.Pointer(&a))
		bw := *(*[]uintptr)(unsafe.Pointer(&b))

		for i := 0; i < w; i++ {
			dw[i] = aw[i] & bw[i]
		}
	}

	for i := n - n%wordSize; i < n; i++ {
		dst[i] = a[i] & b[i]
	}

	return n
}

func safeAndBytes(dst, a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	if len(dst) < n {
		n = len(dst)
	}

	for i := 0; i < n; i++ {
		dst[i] = a[i] & b[i]
	}

	return n
}

// And sets each element in according to dst[i] = a[i] AND b[i]
func And(dst, a, b []byte) int {
	if supportsUnaligned {
		return fastAndBytes(dst, a, b)
	}

	// TODO: if (dst, a, b) have common alignment
	// we could still try fastAndBytes.
	return safeAndBytes(dst, a, b)
}

func fastAndNotBytes(dst, a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	if len(dst) < n {
		n = len(dst)
	}

	w := n / wordSize
	if w > 0 {
		dw := *(*[]uintptr)(unsafe.Pointer(&dst))
		aw := *(*[]uintptr)(unsafe.Pointer(&a))
		bw := *(*[]uintptr)(unsafe.Pointer(&b))

		for i := 0; i < w; i++ {
			dw[i] = aw[i] &^ bw[i]
		}
	}

	for i := n - n%wordSize; i < n; i++ {
		dst[i] = a[i] &^ b[i]
	}

	return n
}

func safeAndNotBytes(dst, a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	if len(dst) < n {
		n = len(dst)
	}

	for i := 0; i < n; i++ {
		dst[i] = a[i] &^ b[i]
	}

	return n
}

// AndNot sets each element in according to dst[i] = a[i] AND (NOT b[i])
func AndNot(dst, a, b []byte) int {
	if supportsUnaligned {
		return fastAndNotBytes(dst, a, b)
	}

	// TODO: if (dst, a, b) have common alignment
	// we could still try fastAndNotBytes.
	return safeAndNotBytes(dst, a, b)
}

func fastNotAndBytes(dst, a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	if len(dst) < n {
		n = len(dst)
	}

	w := n / wordSize
	if w > 0 {
		dw := *(*[]uintptr)(unsafe.Pointer(&dst))
		aw := *(*[]uintptr)(unsafe.Pointer(&a))
		bw := *(*[]uintptr)(unsafe.Pointer(&b))

		for i := 0; i < w; i++ {
			dw[i] = ^(aw[i] & bw[i])
		}
	}

	for i := n - n%wordSize; i < n; i++ {
		dst[i] = ^(a[i] & b[i])
	}

	return n
}

func safeNotAndBytes(dst, a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	if len(dst) < n {
		n = len(dst)
	}

	for i := 0; i < n; i++ {
		dst[i] = ^(a[i] & b[i])
	}

	return n
}

// NotAnd sets each element in according to dst[i] = NOT (a[i] AND b[i])
func NotAnd(dst, a, b []byte) int {
	if supportsUnaligned {
		return fastNotAndBytes(dst, a, b)
	}

	// TODO: if (dst, a, b) have common alignment
	// we could still try fastNotAndBytes.
	return safeNotAndBytes(dst, a, b)
}

func fastOrBytes(dst, a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	if len(dst) < n {
		n = len(dst)
	}

	w := n / wordSize
	if w > 0 {
		dw := *(*[]uintptr)(unsafe.Pointer(&dst))
		aw := *(*[]uintptr)(unsafe.Pointer(&a))
		bw := *(*[]uintptr)(unsafe.Pointer(&b))

		for i := 0; i < w; i++ {
			dw[i] = aw[i] | bw[i]
		}
	}

	for i := n - n%wordSize; i < n; i++ {
		dst[i] = a[i] | b[i]
	}

	return n
}

func safeOrBytes(dst, a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	if len(dst) < n {
		n = len(dst)
	}

	for i := 0; i < n; i++ {
		dst[i] = a[i] | b[i]
	}

	return n
}

// Or sets each element in according to dst[i] = a[i] OR b[i]
func Or(dst, a, b []byte) int {
	if supportsUnaligned {
		return fastOrBytes(dst, a, b)
	}

	// TODO: if (dst, a, b) have common alignment
	// we could still try fastOrBytes.
	return safeOrBytes(dst, a, b)
}

func fastNotOrBytes(dst, a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	if len(dst) < n {
		n = len(dst)
	}

	w := n / wordSize
	if w > 0 {
		dw := *(*[]uintptr)(unsafe.Pointer(&dst))
		aw := *(*[]uintptr)(unsafe.Pointer(&a))
		bw := *(*[]uintptr)(unsafe.Pointer(&b))

		for i := 0; i < w; i++ {
			dw[i] = ^(aw[i] | bw[i])
		}
	}

	for i := n - n%wordSize; i < n; i++ {
		dst[i] = ^(a[i] | b[i])
	}

	return n
}

func safeNotOrBytes(dst, a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	if len(dst) < n {
		n = len(dst)
	}

	for i := 0; i < n; i++ {
		dst[i] = ^(a[i] | b[i])
	}

	return n
}

// NotOr sets each element in according to dst[i] = NOT (a[i] OR b[i])
func NotOr(dst, a, b []byte) int {
	if supportsUnaligned {
		return fastNotOrBytes(dst, a, b)
	}

	// TODO: if (dst, a, b) have common alignment
	// we could still try fastNotOrBytes.
	return safeNotOrBytes(dst, a, b)
}

func fastNotBytes(dst, src []byte) int {
	n := len(src)
	if len(dst) < n {
		n = len(dst)
	}

	w := n / wordSize
	if w > 0 {
		dw := *(*[]uintptr)(unsafe.Pointer(&dst))
		sw := *(*[]uintptr)(unsafe.Pointer(&src))

		for i := 0; i < w; i++ {
			dw[i] = ^sw[i]
		}
	}

	for i := n - n%wordSize; i < n; i++ {
		dst[i] = ^src[i]
	}

	return n
}

func safeNotBytes(dst, src []byte) int {
	n := len(src)
	if len(dst) < n {
		n = len(dst)
	}

	for i := 0; i < n; i++ {
		dst[i] = ^src[i]
	}

	return n
}

// Not sets each element in according to dst[i] = NOT src[i]
func Not(dst, src []byte) int {
	if supportsUnaligned {
		return fastNotBytes(dst, src)
	}

	// TODO: if (dst, src) have common alignment
	// we could still try fastNotBytes.
	return safeNotBytes(dst, src)
}

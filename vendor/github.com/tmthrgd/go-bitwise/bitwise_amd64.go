// Copyright 2017 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License license that can be found in
// the LICENSE file.

// +build amd64,!gccgo,!appengine

// Package bitwise provides efficient implementations of xor/xnor/and/and-not/nand/or/nor/not.
package bitwise

// XOR sets each element in according to dst[i] = a[i] XOR b[i]
func XOR(dst, a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	if len(dst) < n {
		n = len(dst)
	}

	if n == 0 {
		return 0
	}

	xorASM(&dst[0], &a[0], &b[0], uint64(n))
	return n
}

// XNOR sets each element in according to dst[i] = NOT (a[i] XOR b[i])
func XNOR(dst, a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	if len(dst) < n {
		n = len(dst)
	}

	if n == 0 {
		return 0
	}

	xnorASM(&dst[0], &a[0], &b[0], uint64(n))
	return n
}

// And sets each element in according to dst[i] = a[i] AND b[i]
func And(dst, a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	if len(dst) < n {
		n = len(dst)
	}

	if n == 0 {
		return 0
	}

	andASM(&dst[0], &a[0], &b[0], uint64(n))
	return n
}

// AndNot sets each element in according to dst[i] = a[i] AND (NOT b[i])
func AndNot(dst, a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	if len(dst) < n {
		n = len(dst)
	}

	if n == 0 {
		return 0
	}

	andNotASM(&dst[0], &a[0], &b[0], uint64(n))
	return n
}

// NotAnd sets each element in according to dst[i] = NOT (a[i] AND b[i])
func NotAnd(dst, a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	if len(dst) < n {
		n = len(dst)
	}

	if n == 0 {
		return 0
	}

	nandASM(&dst[0], &a[0], &b[0], uint64(n))
	return n
}

// Or sets each element in according to dst[i] = a[i] OR b[i]
func Or(dst, a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	if len(dst) < n {
		n = len(dst)
	}

	if n == 0 {
		return 0
	}

	orASM(&dst[0], &a[0], &b[0], uint64(n))
	return n
}

// NotOr sets each element in according to dst[i] = NOT (a[i] OR b[i])
func NotOr(dst, a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	if len(dst) < n {
		n = len(dst)
	}

	if n == 0 {
		return 0
	}

	norASM(&dst[0], &a[0], &b[0], uint64(n))
	return n
}

// Not sets each element in according to dst[i] = NOT src[i]
func Not(dst, src []byte) int {
	n := len(src)
	if len(dst) < n {
		n = len(dst)
	}

	if n == 0 {
		return 0
	}

	notASM(&dst[0], &src[0], uint64(n))
	return n
}

//go:generate go run asm_gen.go

// This function is implemented in bitwise_xor_amd64.s
//go:noescape
func xorASM(dst, a, b *byte, len uint64)

// This function is implemented in bitwise_xnor_amd64.s
//go:noescape
func xnorASM(dst, a, b *byte, len uint64)

// This function is implemented in bitwise_and_amd64.s
//go:noescape
func andASM(dst, a, b *byte, len uint64)

// This function is implemented in bitwise_andnot_amd64.s
//go:noescape
func andNotASM(dst, a, b *byte, len uint64)

// This function is implemented in bitwise_nand_amd64.s
//go:noescape
func nandASM(dst, a, b *byte, len uint64)

// This function is implemented in bitwise_or_amd64.s
//go:noescape
func orASM(dst, a, b *byte, len uint64)

// This function is implemented in bitwise_nor_amd64.s
//go:noescape
func norASM(dst, a, b *byte, len uint64)

// This function is implemented in bitwise_not_amd64.s
//go:noescape
func notASM(dst, src *byte, len uint64)

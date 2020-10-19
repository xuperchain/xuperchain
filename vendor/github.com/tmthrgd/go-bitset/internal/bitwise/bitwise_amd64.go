// Copyright 2017 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License license that can be found in
// the LICENSE file.

// +build amd64,!gccgo,!appengine

// Package bitwise provides an efficient implementation of a & b == b.
package bitwise

// AndEq returns true iff a & b == b
func AndEq(a, b []byte) bool {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}

	if n == 0 {
		return true
	}

	return andeqASM(&a[0], &b[0], uint64(n))
}

// This function is implemented in bitwise_andeq_amd64.s
//go:noescape
func andeqASM(a, b *byte, len uint64) (ret bool)

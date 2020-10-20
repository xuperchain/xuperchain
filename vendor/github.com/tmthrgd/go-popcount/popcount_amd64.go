// Copyright 2016 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License license that can be found in
// the LICENSE file.

// +build amd64,!gccgo,!appengine

package popcount

import "unsafe"

var usePOPCNT = hasPOPCNT()

// CountBytes function counts number of non-zero bits in slice of 8bit unsigned integers.
func CountBytes(s []byte) uint64 {
	if len(s) == 0 {
		return 0
	}

	if !usePOPCNT {
		return countBytesGo(s)
	}

	return countBytesASM(&s[0], uint64(len(s)))
}

// CountSlice64 function counts number of non-zero bits in slice of 64bit unsigned integers.
func CountSlice64(s []uint64) uint64 {
	if len(s) == 0 {
		return 0
	}

	if !usePOPCNT {
		return countSlice64Go(s)
	}

	return countBytesASM((*byte)(unsafe.Pointer(&s[0])), uint64(len(s)*8))
}

// This function is implemented in popcnt_amd64.s
//go:noescape
func hasPOPCNT() (ret bool)

// This function is implemented in popcount_amd64.s
//go:noescape
func countBytesASM(src *byte, len uint64) (ret uint64)

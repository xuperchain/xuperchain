// Copyright 2015 Hideaki Ohno. All rights reserved.
// Use of this source code is governed by an MIT License
// that can be found in the LICENSE file.

// Package popcount is a population count implementation for Golang.
package popcount

import "math/bits"

// Count64 function counts the number of non-zero bits in a 64bit unsigned integer.
//
// Deprecated: use math/bits.OnesCount64 instead.
func Count64(x uint64) uint64 {
	// While math/bits.OnesCount64 is faster, due to it's use of POPCNTQ, it's
	// slower if used here as the cost of checking for POPCNT support cannot be
	// amalgamated by the compiler. Thus we stick to this Golang routine below.

	x = (x & 0x5555555555555555) + ((x & 0xAAAAAAAAAAAAAAAA) >> 1)
	x = (x & 0x3333333333333333) + ((x & 0xCCCCCCCCCCCCCCCC) >> 2)
	x = (x & 0x0F0F0F0F0F0F0F0F) + ((x & 0xF0F0F0F0F0F0F0F0) >> 4)
	x *= 0x0101010101010101
	return ((x >> 56) & 0xFF)
}

func countSlice64Go(s []uint64) (count uint64) {
	for _, x := range s {
		count += uint64(bits.OnesCount64(x))
	}

	return
}

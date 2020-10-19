// Copyright 2017 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License that can be found in
// the LICENSE file.

package bitset

import "github.com/tmthrgd/atomics"

func (a Atomic) index(bit uint) (ptr *atomics.Uint64, mask uint64) {
	return &a[bit/64], 1 << (bit & 63)
}

func atomicMask1(start, end uint) (mask uint64) {
	const max = ^uint64(0)
	return ((max << (start & 63)) ^ (max << (end - start&^63))) & ((1 >> (start & 63)) - 1)
}

func atomicMask2(start, end uint) (mask uint64) {
	const shiftBy = 31 + 32*(^uint(0)>>63)
	return ((1 << (end & 63)) - 1) & uint64((((end&^63-start)>>shiftBy)&1)-1)
}

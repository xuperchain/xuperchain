// Copyright 2017 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License license that can be found in
// the LICENSE file.

package bitset

func mask1(start, end uint) byte {
	const max = ^byte(0)
	return ((max << (start & 7)) ^ (max << (end - start&^7))) & ((1 >> (start & 7)) - 1)
}

func mask2(start, end uint) byte {
	const shiftBy = 31 + 32*(^uint(0)>>63)
	return ((1 << (end & 7)) - 1) & byte((((end&^7-start)>>shiftBy)&1)-1)
}

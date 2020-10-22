// Copyright 2017 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License license that can be found in
// the LICENSE file.

package bitset

import "github.com/tmthrgd/go-byte-test"

func (b Bitset) IsSet(bit uint) bool {
	if bit > b.Len() {
		panic(errOutOfRange)
	}

	return b[bit>>3]&(1<<(bit&7)) != 0
}

func (b Bitset) IsClear(bit uint) bool {
	return !b.IsSet(bit)
}

func (b Bitset) IsRangeSet(start, end uint) bool {
	if start > end {
		panic(errEndLessThanStart)
	}

	if end > b.Len() {
		panic(errOutOfRange)
	}

	if mask := mask1(start, end); mask != 0 {
		if b[start>>3]&mask != mask {
			return false
		}
	}

	if start := (start + 7) &^ 7; start < end {
		if !bytetest.Test(b[start>>3:end>>3], 0xff) {
			return false
		}
	}

	if mask := mask2(start, end); mask != 0 {
		return b[end>>3]&mask == mask
	}

	return true
}

func (b Bitset) IsRangeClear(start, end uint) bool {
	if start > end {
		panic(errEndLessThanStart)
	}

	if end > b.Len() {
		panic(errOutOfRange)
	}

	if mask := mask1(start, end); mask != 0 {
		if b[start>>3]&mask != 0 {
			return false
		}
	}

	if start := (start + 7) &^ 7; start < end {
		if !bytetest.Test(b[start>>3:end>>3], 0) {
			return false
		}
	}

	if mask := mask2(start, end); mask != 0 {
		return b[end>>3]&mask == 0
	}

	return true
}

func (b Bitset) All() bool {
	return bytetest.Test(b, 0xff)
}

func (b Bitset) None() bool {
	return bytetest.Test(b, 0)
}

func (b Bitset) Any() bool {
	return !b.None()
}

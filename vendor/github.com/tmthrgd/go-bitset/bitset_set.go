// Copyright 2017 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License license that can be found in
// the LICENSE file.

package bitset

import "github.com/tmthrgd/go-memset"

func (b Bitset) Set(bit uint) {
	if bit > b.Len() {
		panic(errOutOfRange)
	}

	b[bit>>3] |= 1 << (bit & 7)
}

func (b Bitset) Clear(bit uint) {
	if bit > b.Len() {
		panic(errOutOfRange)
	}

	b[bit>>3] &^= 1 << (bit & 7)
}

func (b Bitset) Invert(bit uint) {
	if bit > b.Len() {
		panic(errOutOfRange)
	}

	b[bit>>3] ^= 1 << (bit & 7)
}

func (b Bitset) SetRange(start, end uint) {
	if start > end {
		panic(errEndLessThanStart)
	}

	if end > b.Len() {
		panic(errOutOfRange)
	}

	if mask := mask1(start, end); mask != 0 {
		b[start>>3] |= mask
	}

	if start := (start + 7) &^ 7; start < end {
		memset.Memset(b[start>>3:end>>3], 0xff)
	}

	if mask := mask2(start, end); mask != 0 {
		b[end>>3] |= mask
	}
}

func (b Bitset) ClearRange(start, end uint) {
	if start > end {
		panic(errEndLessThanStart)
	}

	if end > b.Len() {
		panic(errOutOfRange)
	}

	if mask := mask1(start, end); mask != 0 {
		b[start>>3] &^= mask
	}

	if start := (start + 7) &^ 7; start < end {
		memset.Memset(b[start>>3:end>>3], 0)
	}

	if mask := mask2(start, end); mask != 0 {
		b[end>>3] &^= mask
	}
}

func (b Bitset) InvertRange(start, end uint) {
	b.ComplementRange(b, start, end)
}

func (b Bitset) SetTo(bit uint, value bool) {
	if value {
		b.Set(bit)
	} else {
		b.Clear(bit)
	}
}

func (b Bitset) SetRangeTo(start, end uint, value bool) {
	if value {
		b.SetRange(start, end)
	} else {
		b.ClearRange(start, end)
	}
}

func (b Bitset) SetAll() {
	memset.Memset(b, 0xff)
}

func (b Bitset) ClearAll() {
	memset.Memset(b, 0)
}

func (b Bitset) InvertAll() {
	b.Complement(b)
}

func (b Bitset) SetAllTo(value bool) {
	if value {
		b.SetAll()
	} else {
		b.ClearAll()
	}
}

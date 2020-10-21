// Copyright 2017 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License license that can be found in
// the LICENSE file.

package bitset

func (b Bitset) Copy(b1 Bitset) {
	copy(b, b1)
}

func (b Bitset) CopyRange(b1 Bitset, start, end uint) {
	if start > end {
		panic(errEndLessThanStart)
	}

	if end > b.Len() || end > b1.Len() {
		panic(errOutOfRange)
	}

	if mask := mask1(start, end); mask != 0 {
		b[start>>3] = b[start>>3]&^mask | b1[start>>3]&mask
	}

	if start := (start + 7) &^ 7; start < end {
		copy(b[start>>3:end>>3], b1[start>>3:end>>3])
	}

	if mask := mask2(start, end); mask != 0 {
		b[end>>3] = b[end>>3]&^mask | b1[end>>3]&mask
	}
}

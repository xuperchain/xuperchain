// Copyright 2017 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License license that can be found in
// the LICENSE file.

package bitset

import "bytes"

func (b Bitset) Equal(b1 Bitset) bool {
	return bytes.Equal(b, b1)
}

func (b Bitset) EqualRange(b1 Bitset, start, end uint) bool {
	if start > end {
		panic(errEndLessThanStart)
	}

	if end > b.Len() || end > b1.Len() {
		panic(errOutOfRange)
	}

	if mask := mask1(start, end); mask != 0 {
		if b[start>>3]&mask != b1[start>>3]&mask {
			return false
		}
	}

	if start := (start + 7) &^ 7; start < end {
		if !bytes.Equal(b[start>>3:end>>3], b1[start>>3:end>>3]) {
			return false
		}
	}

	if mask := mask2(start, end); mask != 0 {
		return b[end>>3]&mask == b1[end>>3]&mask
	}

	return true
}

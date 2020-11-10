// Copyright 2017 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License license that can be found in
// the LICENSE file.

package bitset

var useShiftFastPath = true // for testing

func (b Bitset) ShiftLeft(b1 Bitset, shift uint) {
	if shift > b1.Len() {
		panic(errOutOfRange)
	}

	if shift&7 == 0 && useShiftFastPath {
		// fast path
		copy(b, b1[shift>>3:])
	} else {
		// slow path
		l := b1.Len() - shift
		if b.Len() < l {
			l = b.Len()
		}

		for i := uint(0); i < l; i++ {
			b.SetTo(i, b1.IsSet(i+shift))
		}
	}
}

func (b Bitset) ShiftRight(b1 Bitset, shift uint) {
	if shift > b.Len() {
		panic(errOutOfRange)
	}

	if shift&7 == 0 && useShiftFastPath {
		// fast path
		copy(b[shift>>3:], b1)
	} else {
		// slow path
		l := b.Len()
		if b1.Len() < l-shift {
			l = b1.Len() + shift
		}

		for i := l - 1; i >= shift; i-- {
			b.SetTo(i, b1.IsSet(i-shift))
		}
	}
}

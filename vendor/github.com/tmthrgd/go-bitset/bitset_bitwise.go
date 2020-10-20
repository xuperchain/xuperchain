// Copyright 2017 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License license that can be found in
// the LICENSE file.

package bitset

import "github.com/tmthrgd/go-bitwise"

func (b Bitset) Complement(b1 Bitset) {
	bitwise.Not(b, b1)
}

func (b Bitset) Union(b1, b2 Bitset) {
	bitwise.Or(b, b1, b2)
}

func (b Bitset) Intersection(b1, b2 Bitset) {
	bitwise.And(b, b1, b2)
}

func (b Bitset) Difference(b1, b2 Bitset) {
	bitwise.AndNot(b, b1, b2)
}

func (b Bitset) SymmetricDifference(b1, b2 Bitset) {
	bitwise.XOR(b, b1, b2)
}

func (b Bitset) ComplementRange(b1 Bitset, start, end uint) {
	if start > end {
		panic(errEndLessThanStart)
	}

	if end > b.Len() || end > b1.Len() {
		panic(errOutOfRange)
	}

	if mask := mask1(start, end); mask != 0 {
		b[start>>3] = b[start>>3]&^mask | (^b1[start>>3])&mask
	}

	if start := (start + 7) &^ 7; start < end {
		bitwise.Not(b[start>>3:end>>3], b1[start>>3:end>>3])
	}

	if mask := mask2(start, end); mask != 0 {
		b[end>>3] = b[end>>3]&^mask | (^b1[end>>3])&mask
	}
}

func (b Bitset) UnionRange(b1, b2 Bitset, start, end uint) {
	if start > end {
		panic(errEndLessThanStart)
	}

	if end > b.Len() || end > b1.Len() || end > b2.Len() {
		panic(errOutOfRange)
	}

	if mask := mask1(start, end); mask != 0 {
		b[start>>3] = b[start>>3]&^mask | (b1[start>>3]|b2[start>>3])&mask
	}

	if start := (start + 7) &^ 7; start < end {
		bitwise.Or(b[start>>3:end>>3], b1[start>>3:end>>3], b2[start>>3:end>>3])
	}

	if mask := mask2(start, end); mask != 0 {
		b[end>>3] = b[end>>3]&^mask | (b1[end>>3]|b2[end>>3])&mask
	}
}

func (b Bitset) IntersectionRange(b1, b2 Bitset, start, end uint) {
	if start > end {
		panic(errEndLessThanStart)
	}

	if end > b.Len() || end > b1.Len() || end > b2.Len() {
		panic(errOutOfRange)
	}

	if mask := mask1(start, end); mask != 0 {
		b[start>>3] = b[start>>3]&^mask | (b1[start>>3]&b2[start>>3])&mask
	}

	if start := (start + 7) &^ 7; start < end {
		bitwise.And(b[start>>3:end>>3], b1[start>>3:end>>3], b2[start>>3:end>>3])
	}

	if mask := mask2(start, end); mask != 0 {
		b[end>>3] = b[end>>3]&^mask | (b1[end>>3]&b2[end>>3])&mask
	}
}

func (b Bitset) DifferenceRange(b1, b2 Bitset, start, end uint) {
	if start > end {
		panic(errEndLessThanStart)
	}

	if end > b.Len() || end > b1.Len() || end > b2.Len() {
		panic(errOutOfRange)
	}

	if mask := mask1(start, end); mask != 0 {
		b[start>>3] = b[start>>3]&^mask | (b1[start>>3]&^b2[start>>3])&mask
	}

	if start := (start + 7) &^ 7; start < end {
		bitwise.AndNot(b[start>>3:end>>3], b1[start>>3:end>>3], b2[start>>3:end>>3])
	}

	if mask := mask2(start, end); mask != 0 {
		b[end>>3] = b[end>>3]&^mask | (b1[end>>3]&^b2[end>>3])&mask
	}
}

func (b Bitset) SymmetricDifferenceRange(b1, b2 Bitset, start, end uint) {
	if start > end {
		panic(errEndLessThanStart)
	}

	if end > b.Len() || end > b1.Len() || end > b2.Len() {
		panic(errOutOfRange)
	}

	if mask := mask1(start, end); mask != 0 {
		b[start>>3] = b[start>>3]&^mask | (b1[start>>3]^b2[start>>3])&mask
	}

	if start := (start + 7) &^ 7; start < end {
		bitwise.XOR(b[start>>3:end>>3], b1[start>>3:end>>3], b2[start>>3:end>>3])
	}

	if mask := mask2(start, end); mask != 0 {
		b[end>>3] = b[end>>3]&^mask | (b1[end>>3]^b2[end>>3])&mask
	}
}

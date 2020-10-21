// Copyright 2017 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License license that can be found in
// the LICENSE file.

package bitset

import (
	"errors"

	"github.com/tmthrgd/go-hex"
)

var (
	errEndLessThanStart = errors.New("go-bitset: cannot range backwards")
	errOutOfRange       = errors.New("go-bitset: out of range")
)

type Bitset []byte

func New(size uint) Bitset {
	size = (size + 7) &^ 7
	return make(Bitset, size>>3)
}

func (b Bitset) Len() uint {
	return uint(len(b)) << 3
}

func (b Bitset) ByteLen() int {
	return len(b)
}

func (b Bitset) Slice(start, end uint) Bitset {
	if start > end {
		panic(errEndLessThanStart)
	}

	if end > b.Len() {
		panic(errOutOfRange)
	}

	if start&7 != 0 || end&7 != 0 {
		panic(errors.New("go-bitset: cannot slice inside a byte"))
	}

	return b[start>>3 : end>>3]
}

func (b Bitset) Clone() Bitset {
	return append(Bitset(nil), b...)
}

func (b Bitset) CloneRange(start, end uint) Bitset {
	if start > end {
		panic(errEndLessThanStart)
	}

	if end > b.Len() {
		panic(errOutOfRange)
	}

	b1 := New(end - start)
	b1.ShiftLeft(b, start)
	b1.ClearRange(end-start, b1.Len())
	return b1
}

func (b Bitset) String() string {
	const maxSize = 128

	if len(b) > maxSize {
		return "Bitset{" + hex.EncodeToString(b[:maxSize]) + "...}"
	}

	return "Bitset{" + hex.EncodeToString(b) + "}"
}

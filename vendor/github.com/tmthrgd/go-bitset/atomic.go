// Copyright 2017 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License that can be found in
// the LICENSE file.

package bitset

import (
	"errors"
	"fmt"

	"github.com/tmthrgd/atomics"
)

type Atomic []atomics.Uint64

func NewAtomic(size uint) Atomic {
	size = (size + 63) &^ 63
	return make(Atomic, size/64)
}

func (a Atomic) Len() uint {
	return uint(len(a)) * 64
}

func (a Atomic) Uint64Len() int {
	return len(a)
}

func (a Atomic) Slice(start, end uint) Atomic {
	if start > end {
		panic(errEndLessThanStart)
	}

	if end > a.Len() {
		panic(errOutOfRange)
	}

	if start&63 != 0 || end&63 != 0 {
		panic(errors.New("go-bitset: cannot slice inside a uint64"))
	}

	return a[start/64 : end/64]
}

func (a Atomic) String() string {
	return fmt.Sprintf("Atomic{%p,%d}", &a, a.Len())
}

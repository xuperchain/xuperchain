// Copyright 2017 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License that can be found in
// the LICENSE file.

package bitset

func (a Atomic) IsSet(bit uint) bool {
	if bit > a.Len() {
		panic(errOutOfRange)
	}

	ptr, mask := a.index(bit)
	return ptr.Load()&mask != 0
}

func (a Atomic) IsClear(bit uint) bool {
	return !a.IsSet(bit)
}

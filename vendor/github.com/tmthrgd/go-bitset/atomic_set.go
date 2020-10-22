// Copyright 2017 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License that can be found in
// the LICENSE file.

package bitset

func (a Atomic) Set(bit uint) {
	if bit > a.Len() {
		panic(errOutOfRange)
	}

	ptr, mask := a.index(bit)
	old := ptr.Load()
	for !ptr.CompareAndSwap(old, old|mask) {
		old = ptr.Load()
	}
}

func (a Atomic) Clear(bit uint) {
	if bit > a.Len() {
		panic(errOutOfRange)
	}

	ptr, mask := a.index(bit)
	old := ptr.Load()
	for !ptr.CompareAndSwap(old, old&^mask) {
		old = ptr.Load()
	}
}

func (a Atomic) Invert(bit uint) {
	if bit > a.Len() {
		panic(errOutOfRange)
	}

	ptr, mask := a.index(bit)
	old := ptr.Load()
	for !ptr.CompareAndSwap(old, old^mask) {
		old = ptr.Load()
	}
}

func (a Atomic) SetRange(start, end uint) {
	if start > end {
		panic(errEndLessThanStart)
	}

	if end > a.Len() {
		panic(errOutOfRange)
	}

	if mask := atomicMask1(start, end); mask != 0 {
		ptr, _ := a.index(start)
		old := ptr.Load()
		for !ptr.CompareAndSwap(old, old|mask) {
			old = ptr.Load()
		}
	}

	for i := (start + 63) &^ 63; i < end&^63; i += 64 {
		ptr, _ := a.index(i)
		ptr.Store(^uint64(0))
	}

	if mask := atomicMask2(start, end); mask != 0 {
		ptr, _ := a.index(end)
		old := ptr.Load()
		for !ptr.CompareAndSwap(old, old|mask) {
			old = ptr.Load()
		}
	}
}

func (a Atomic) ClearRange(start, end uint) {
	if start > end {
		panic(errEndLessThanStart)
	}

	if end > a.Len() {
		panic(errOutOfRange)
	}

	if mask := atomicMask1(start, end); mask != 0 {
		ptr, _ := a.index(start)
		old := ptr.Load()
		for !ptr.CompareAndSwap(old, old&^mask) {
			old = ptr.Load()
		}
	}

	for i := (start + 63) &^ 63; i < end&^63; i += 64 {
		ptr, _ := a.index(i)
		ptr.Store(0)
	}

	if mask := atomicMask2(start, end); mask != 0 {
		ptr, _ := a.index(end)
		old := ptr.Load()
		for !ptr.CompareAndSwap(old, old&^mask) {
			old = ptr.Load()
		}
	}
}

func (a Atomic) SetTo(bit uint, value bool) {
	if value {
		a.Set(bit)
	} else {
		a.Clear(bit)
	}
}

func (a Atomic) SetRangeTo(start, end uint, value bool) {
	if value {
		a.SetRange(start, end)
	} else {
		a.ClearRange(start, end)
	}
}

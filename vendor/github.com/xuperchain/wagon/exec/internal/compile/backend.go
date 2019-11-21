// Copyright 2019 The go-interpreter Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package compile

type dirtyState uint8

const (
	stateScratch           dirtyState = iota // We don't care about the value.
	stateStackLen                            // Stores the stack len (dirty).
	stateStackFirstElem                      // Caches a pointer to the stack array.
	stateLocalFirstElem                      // Caches a pointer to the locals array.
	stateGlobalSliceHeader                   // Caches a pointer to the globals slice header.
)

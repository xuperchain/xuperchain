// Copyright 2019 The go-interpreter Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !appengine

package compile

import "unsafe"

func jitcall(asm unsafe.Pointer, stack, locals, globals *[]uint64, mem *[]byte) uint64

// Copyright 2019 The go-interpreter Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !debugstack

package exec

// debugStackDepth enables runtime checks of the stack depth. If
// the stack every would exceed or underflow its expected bounds,
// a panic is thrown.
const debugStackDepth = false

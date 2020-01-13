// Copyright 2019 The go-interpreter Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package compile

// NativeCodeUnit represents compiled native code.
type NativeCodeUnit interface {
	Invoke(stack, locals, globals *[]uint64, mem *[]byte) JITExitSignal
}

// CompletionStatus describes the final status of a native execution.
type CompletionStatus uint64

// Valid completion statuses.
const (
	CompletionOK CompletionStatus = iota
	CompletionBadBounds
	CompletionUnreachable
	CompletionFatalInternalError
	CompletionDivideZero
)

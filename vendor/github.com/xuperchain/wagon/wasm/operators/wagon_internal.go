// Copyright 2019 The go-interpreter Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package operators

import "github.com/xuperchain/wagon/wasm"

// These opcodes implement optimizations in wagon execution, and are invalid
// opcodes for any uses other than internal use. Expect them to change at any
// time.
// If these opcodes are ever used in future wasm instructions, feel free to
// reassign them to other free opcodes.
var (
	internalOpcodes = map[byte]bool{
		WagonNativeExec: true,
	}

	CheckGas        = newOp(0xfd, "wagon.checkGas", []wasm.ValueType{}, noReturn)
	WagonNativeExec = newOp(0xfe, "wagon.nativeExec", []wasm.ValueType{wasm.ValueTypeI64}, noReturn)
)

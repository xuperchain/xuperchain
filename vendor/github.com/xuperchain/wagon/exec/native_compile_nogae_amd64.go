// Copyright 2019 The go-interpreter Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !appengine

package exec

import (
	"encoding/binary"

	"github.com/xuperchain/wagon/exec/internal/compile"
)

func init() {
	supportedNativeArchs = append(supportedNativeArchs, nativeArch{
		Arch: "amd64",
		OS:   "linux",
		make: makeAMD64NativeBackend,
	}, nativeArch{
		Arch: "amd64",
		OS:   "windows",
		make: makeAMD64NativeBackend,
	}, nativeArch{
		Arch: "amd64",
		OS:   "darwin",
		make: makeAMD64NativeBackend,
	})
}

func makeAMD64NativeBackend(endianness binary.ByteOrder) *nativeCompiler {
	be := &compile.AMD64Backend{EmitBoundsChecks: debugStackDepth}
	return &nativeCompiler{
		Builder:   be,
		Scanner:   be.Scanner(),
		allocator: &compile.MMapAllocator{},
	}
}

// Copyright 2019 The go-interpreter Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build amd64 !appengine

#include "funcdata.h"
#include "textflag.h"

// jitcall(*asm, *stackSlice, *localSlice, *globalSlice, *memSlice) uint64
TEXT ·jitcall(SB),NOSPLIT|NOFRAME,$0-48
        GO_ARGS
        MOVQ asm+0(FP),      AX  // Load the address of the assembly section.
        MOVQ stack+8(FP),    R10 // Load the address of the stack.
        MOVQ locals+16(FP),  R11 // Load the address of the locals.
        MOVQ mem+32(FP),     SI  // Load the address of main memory.
        MOVQ 0(AX),          AX  // Deference pointer to native code.
        JMP AX                   // Jump to native code.

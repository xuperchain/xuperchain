// Copyright 2015 Hideaki Ohno. All rights reserved.
// Use of this source code is governed by an MIT License
// that can be found in the LICENSE file.

// +build amd64,!gccgo,!appengine

#include "textflag.h"

// func hasPOPCNT() bool
TEXT ·hasPOPCNT(SB),NOSPLIT,$0
	XORQ AX, AX
	INCL AX
	CPUID
	SHRQ $23, CX
	ANDQ $1, CX
	MOVB CX, ret+0(FP)
	RET

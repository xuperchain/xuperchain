// Copyright 2016 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License license that can be found in
// the LICENSE file.

// +build amd64,!gccgo,!appengine

#include "textflag.h"

// Important performance information can be found at:
// http://stackoverflow.com/a/25089720
//
// POPCNT has a false-dependency bug that causes a performance
// hit. Thus, in bigloop four separate destination registers are
// used to allow intra-loop parallelization, and in loop the
// destination register is cleared (with no practical effect)
// before POPCNT to allow inter-loop parallelization.

TEXT ·countBytesASM(SB),NOSPLIT,$0
	MOVQ src+0(FP), SI
	MOVQ len+8(FP), BX

	XORQ AX, AX

	CMPQ BX, $8
	JB tail

	CMPQ BX, $32
	JB loop

bigloop:
	POPCNTQ -8(SI)(BX*1), R11
	POPCNTQ -16(SI)(BX*1), R10
	POPCNTQ -24(SI)(BX*1), R9
	POPCNTQ -32(SI)(BX*1), R8

	ADDQ R11, AX
	ADDQ R10, AX
	ADDQ R9, AX
	ADDQ R8, AX

	SUBQ $32, BX
	JZ ret

	CMPQ BX, $32
	JAE bigloop

	CMPQ BX, $8
	JB tail

loop:
	XORQ DX, DX
	POPCNTQ -8(SI)(BX*1), DX

	ADDQ DX, AX

	SUBQ $8, BX
	JZ ret

	CMPQ BX, $8
	JAE loop

tail:
	XORQ DX, DX

	CMPQ BX, $4
	JB tail_2

	MOVL -4(SI)(BX*1), DX

	SUBQ $4, BX
	JZ tail_4

tail_2:
	CMPQ BX, $2
	JB tail_3

	SHLQ $16, DX
	ORW -2(SI)(BX*1), DX

	SUBQ $2, BX
	JZ tail_4

tail_3:
	SHLQ $8, DX
	ORB -1(SI)(BX*1), DX

tail_4:
	POPCNTQ DX, DX

	ADDQ DX, AX

ret:
	MOVQ AX, ret+16(FP)
	RET

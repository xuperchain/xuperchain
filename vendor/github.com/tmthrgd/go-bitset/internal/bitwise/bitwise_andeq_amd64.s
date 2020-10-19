// Copyright 2017 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License license that can be found in
// the LICENSE file.

// +build amd64,!gccgo,!appengine

#include "textflag.h"

TEXT ·andeqASM(SB),NOSPLIT,$0
	MOVQ a+0(FP), SI
	MOVQ b+8(FP), DI
	MOVQ len+16(FP), BX

	CMPQ BX, $16
	JB loop

bigloop:
	MOVOU -16(SI)(BX*1), X0
	MOVOU -16(DI)(BX*1), X1

	PAND X1, X0
	PXOR X1, X0

	PTEST X0, X0
	JNZ ret_false

	SUBQ $16, BX
	JZ ret_true

	CMPQ BX, $16
	JAE bigloop

loop:
	MOVB -1(SI)(BX*1), AX
	MOVB -1(DI)(BX*1), DX

	ANDB DX, AX

	CMPB DX, AX
	JNE ret_false

	SUBQ $1, BX
	JNZ loop

ret_true:
	MOVB $1, ret+24(FP)
	RET

ret_false:
	MOVB $0, ret+24(FP)
	RET

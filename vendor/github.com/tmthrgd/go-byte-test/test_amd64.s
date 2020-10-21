// Copyright 2017 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License license that can be found in
// the LICENSE file.

// +build amd64,!gccgo,!appengine

#include "textflag.h"

TEXT ·testAsm(SB),NOSPLIT,$0
	MOVQ src+0(FP), SI
	MOVQ len+8(FP), BX
	MOVB value+16(FP), AX

	CMPQ BX, $16
	JB loop

	PINSRB $0, AX, X0
	PXOR X1, X1
	PSHUFB X1, X0

	CMPQ BX, $64
	JB bigloop

	CMPQ BX, $128
	JB hugeloop

massiveloop:
	MOVOU -16(SI)(BX*1), X1
	MOVOU -32(SI)(BX*1), X2
	MOVOU -48(SI)(BX*1), X3
	MOVOU -64(SI)(BX*1), X4
	MOVOU -80(SI)(BX*1), X5
	MOVOU -96(SI)(BX*1), X6
	MOVOU -112(SI)(BX*1), X7
	MOVOU -128(SI)(BX*1), X8

	PXOR X0, X1
	PXOR X0, X2
	PXOR X0, X3
	PXOR X0, X4
	PXOR X0, X5
	PXOR X0, X6
	PXOR X0, X7
	PXOR X0, X8

	POR X2, X1
	POR X3, X1
	POR X4, X1
	POR X5, X1
	POR X6, X1
	POR X7, X1
	POR X8, X1

	PTEST X1, X1
	JNZ ret_false

	SUBQ $128, BX
	JZ ret_true

	CMPQ BX, $128
	JAE massiveloop

	CMPQ BX, $16
	JB loop

	CMPQ BX, $64
	JB bigloop

hugeloop:
	MOVOU -16(SI)(BX*1), X1
	MOVOU -32(SI)(BX*1), X2
	MOVOU -48(SI)(BX*1), X3
	MOVOU -64(SI)(BX*1), X4

	PXOR X0, X1
	PXOR X0, X2
	PXOR X0, X3
	PXOR X0, X4

	POR X2, X1
	POR X3, X1
	POR X4, X1

	PTEST X1, X1
	JNZ ret_false

	SUBQ $64, BX
	JZ ret_true

	CMPQ BX, $64
	JAE hugeloop

	CMPQ BX, $16
	JB loop

bigloop:
	MOVOU -16(SI)(BX*1), X1

	PXOR X0, X1

	PTEST X1, X1
	JNZ ret_false

	SUBQ $16, BX
	JZ ret_true

	CMPQ BX, $16
	JAE bigloop

loop:
	MOVB -1(SI)(BX*1), R15

	CMPB AX, R15
	JNE ret_false

	DECQ BX
	JNZ loop

ret_true:
	MOVB $1, ret+24(FP)
	RET

ret_false:
	MOVB $0, ret+24(FP)
	RET

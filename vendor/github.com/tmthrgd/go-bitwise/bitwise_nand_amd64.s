// Copyright 2017 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License license that can be found in
// the LICENSE file.
//
// This file is auto-generated - do not modify

// +build amd64,!gccgo,!appengine

#include "textflag.h"

TEXT ·nandASM(SB),NOSPLIT,$0
	MOVQ dst+0(FP), DI
	MOVQ a+8(FP), SI
	MOVQ b+16(FP), DX
	MOVQ len+24(FP), BX
	CMPQ BX, $16
	JB loop
	PCMPEQL X15, X15
	CMPQ BX, $64
	JB bigloop
hugeloop:
	MOVOU -16(SI)(BX*1), X0
	MOVOU -32(SI)(BX*1), X2
	MOVOU -48(SI)(BX*1), X4
	MOVOU -64(SI)(BX*1), X6
	MOVOU -16(DX)(BX*1), X1
	MOVOU -32(DX)(BX*1), X3
	MOVOU -48(DX)(BX*1), X5
	MOVOU -64(DX)(BX*1), X7
	PAND X0, X1
	PXOR X15, X1
	PAND X2, X3
	PXOR X15, X3
	PAND X4, X5
	PXOR X15, X5
	PAND X6, X7
	PXOR X15, X7
	MOVOU X1, -16(DI)(BX*1)
	MOVOU X3, -32(DI)(BX*1)
	MOVOU X5, -48(DI)(BX*1)
	MOVOU X7, -64(DI)(BX*1)
	SUBQ $64, BX
	JZ ret
	CMPQ BX, $64
	JAE hugeloop
	CMPQ BX, $16
	JB loop
bigloop:
	MOVOU -16(SI)(BX*1), X0
	MOVOU -16(DX)(BX*1), X1
	PAND X0, X1
	PXOR X15, X1
	MOVOU X1, -16(DI)(BX*1)
	SUBQ $16, BX
	JZ ret
	CMPQ BX, $16
	JAE bigloop
loop:
	MOVB -1(SI)(BX*1), AX
	ANDB -1(DX)(BX*1), AX
	NOTB AX
	MOVB AX, -1(DI)(BX*1)
	SUBQ $1, BX
	JNZ loop
ret:
	RET

// Copyright 2017 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License license that can be found in
// the LICENSE file.
//
// This file is auto-generated - do not modify

// +build amd64,!gccgo,!appengine

#include "textflag.h"

TEXT ·notASM(SB),NOSPLIT,$0
	MOVQ dst+0(FP), DI
	MOVQ src+8(FP), SI
	MOVQ len+16(FP), BX
	CMPQ BX, $16
	JB loop
	PCMPEQL X0, X0
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
	MOVOU X1, -16(DI)(BX*1)
	MOVOU X2, -32(DI)(BX*1)
	MOVOU X3, -48(DI)(BX*1)
	MOVOU X4, -64(DI)(BX*1)
	SUBQ $64, BX
	JZ ret
	CMPQ BX, $64
	JAE hugeloop
	CMPQ BX, $16
	JB loop
bigloop:
	MOVOU -16(SI)(BX*1), X1
	PXOR X0, X1
	MOVOU X1, -16(DI)(BX*1)
	SUBQ $16, BX
	JZ ret
	CMPQ BX, $16
	JAE bigloop
loop:
	MOVB -1(SI)(BX*1), AX
	NOTB AX
	MOVB AX, -1(DI)(BX*1)
	SUBQ $1, BX
	JNZ loop
ret:
	RET

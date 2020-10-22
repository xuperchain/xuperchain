// Copyright 2016 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License license that can be found in
// the LICENSE file.

// +build amd64,!gccgo,!appengine

#include "textflag.h"

TEXT ·memsetAsm(SB),NOSPLIT,$0
	MOVQ dst+0(FP), DI
	MOVQ len+8(FP), BX
	MOVB value+16(FP), SI

	CMPQ BX, $16
	JB loop

	PINSRB $0, SI, X0
	PXOR X1, X1
	PSHUFB X1, X0

	CMPB ·useAVX(SB), $1
	JNE bigloop

	CMPQ BX, $64
	JB bigloop

	VINSERTF128 $1, X0, Y0, Y0

	CMPQ BX, $0x1000000
	JAE hugeloop_nt_preheader

hugeloop:
	VMOVDQU Y0, -32(DI)(BX*1)
	VMOVDQU Y0, -64(DI)(BX*1)

	SUBQ $64, BX
	JZ ret_after_y0

	CMPQ BX, $64
	JAE hugeloop

	VZEROUPPER

	CMPQ BX, $16
	JB loop

bigloop:
	MOVOU X0, -16(DI)(BX*1)

	SUBQ $16, BX
	JZ ret

	CMPQ BX, $16
	JAE bigloop

loop:
	MOVB SI, -1(DI)(BX*1)

	DECQ BX
	JNZ loop

ret:
	RET

ret_after_y0:
	VZEROUPPER
	RET

hugeloop_nt_preheader:
	VMOVDQU Y0, -32(DI)(BX*1)

	ADDQ DI, BX
	ANDQ $~31, BX
	SUBQ DI, BX

hugeloop_nt:
	VMOVNTDQ Y0, -32(DI)(BX*1)
	VMOVNTDQ Y0, -64(DI)(BX*1)
	VMOVNTDQ Y0, -96(DI)(BX*1)
	VMOVNTDQ Y0, -128(DI)(BX*1)

	SUBQ $128, BX
	JZ ret_after_nt

	CMPQ BX, $128
	JAE hugeloop_nt

	SFENCE
	VZEROUPPER

	CMPQ BX, $16
	JAE bigloop

	JMP loop

ret_after_nt:
	SFENCE
	VZEROUPPER
	RET

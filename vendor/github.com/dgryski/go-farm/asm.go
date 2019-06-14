// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
)

const k0 uint64 = 0xc3a5c85c97cb3127
const k1 uint64 = 0xb492b66fbe98f273
const k2 uint64 = 0x9ae16a3b2f90404f

func shiftMix(val Register) Register {
	r := GP64()
	MOVQ(val, r)
	SHRQ(Imm(47), r)
	XORQ(val, r)
	return r
}

func shiftMix64(val uint64) uint64 {
	return val ^ (val >> 47)
}

func hashLen16MulLine(a, b, c, d, k, mul Register) Register {
	tmpa := GP64()
	MOVQ(a, tmpa)

	ADDQ(b, tmpa)
	RORQ(Imm(43), tmpa)
	ADDQ(d, tmpa)
	tmpc := GP64()
	MOVQ(c, tmpc)
	RORQ(Imm(30), tmpc)
	ADDQ(tmpc, tmpa)

	ADDQ(c, a)
	ADDQ(k, b)
	RORQ(Imm(18), b)
	ADDQ(b, a)

	r := hashLen16Mul(tmpa, a, mul)
	return r
}

func hashLen16Mul(u, v, mul Register) Register {
	XORQ(v, u)
	IMULQ(mul, u)
	a := shiftMix(u)

	XORQ(a, v)
	IMULQ(mul, v)
	b := shiftMix(v)

	IMULQ(mul, b)

	return b
}

func hashLen0to16(sbase, slen Register) {
	CMPQ(slen, Imm(8))
	JL(LabelRef("check4"))
	{
		a := GP64()
		MOVQ(Mem{Base: sbase}, a)

		b := GP64()
		t := GP64()
		MOVQ(slen, t)
		SUBQ(Imm(8), t)
		ADDQ(sbase, t)
		MOVQ(Mem{Base: t}, b)

		rk2 := GP64()
		MOVQ(Imm(k2), rk2)

		ADDQ(rk2, a)

		mul := slen
		SHLQ(Imm(1), mul)
		ADDQ(rk2, mul)

		c := GP64()
		MOVQ(b, c)
		RORQ(Imm(37), c)
		IMULQ(mul, c)
		ADDQ(a, c)

		d := GP64()
		MOVQ(a, d)
		RORQ(Imm(25), d)
		ADDQ(b, d)
		IMULQ(mul, d)

		r := hashLen16Mul(c, d, mul)
		Store(r, ReturnIndex(0))
		RET()
		JMP(LabelRef("fp64ret"))
	}

	Label("check4")

	CMPQ(slen, Imm(4))
	JL(LabelRef("check0"))
	{
		rk2 := GP64()
		MOVQ(Imm(k2), rk2)

		mul := GP64()
		MOVQ(slen, mul)
		SHLQ(Imm(1), mul)
		ADDQ(rk2, mul)

		a := GP64()
		MOVL(Mem{Base: sbase}, a.As32())

		SHLQ(Imm(3), a)
		ADDQ(slen, a)

		b := GP64()
		SUBQ(Imm(4), slen)
		ADDQ(slen, sbase)
		MOVL(Mem{Base: sbase}, b.As32())
		r := hashLen16Mul(a, b, mul)

		Store(r, ReturnIndex(0))
		RET()
		JMP(LabelRef("fp64ret"))
	}

	Label("check0")
	TESTQ(slen, slen)
	JZ(LabelRef("empty"))
	{

		a := GP64()
		MOVBQZX(Mem{Base: sbase}, a)

		base := GP64()
		MOVQ(slen, base)
		SHRQ(Imm(1), base)

		b := GP64()
		ADDQ(sbase, base)
		MOVBQZX(Mem{Base: base}, b)

		MOVQ(slen, base)
		SUBQ(Imm(1), base)
		c := GP64()
		ADDQ(sbase, base)
		MOVBQZX(Mem{Base: base}, c)

		SHLQ(Imm(8), b)
		ADDQ(b, a)
		y := a

		SHLQ(Imm(2), c)
		ADDQ(c, slen)
		z := slen

		rk0 := GP64()
		MOVQ(Imm(k0), rk0)
		IMULQ(rk0, z)

		rk2 := GP64()
		MOVQ(Imm(k2), rk2)

		IMULQ(rk2, y)
		XORQ(y, z)

		r := shiftMix(z)
		IMULQ(rk2, r)

		Store(r, ReturnIndex(0))
		RET()
		JMP(LabelRef("fp64ret"))
	}

	Label("empty")

	ret := GP64()
	MOVQ(Imm(k2), ret)
	Store(ret, ReturnIndex(0))
	RET()
	JMP(LabelRef("fp64ret"))
}

func hashLen17to32(sbase, slen Register) {
	mul := GP64()
	MOVQ(slen, mul)
	SHLQ(Imm(1), mul)

	rk2 := GP64()
	MOVQ(Imm(k2), rk2)
	ADDQ(rk2, mul)

	a := GP64()
	MOVQ(Mem{Base: sbase}, a)

	rk1 := GP64()
	MOVQ(Imm(k1), rk1)
	IMULQ(rk1, a)

	b := GP64()
	MOVQ(Mem{Base: sbase, Disp: 8}, b)

	base := GP64()
	MOVQ(slen, base)
	SUBQ(Imm(16), base)
	ADDQ(sbase, base)

	c := GP64()
	MOVQ(Mem{Base: base, Disp: 8}, c)
	IMULQ(mul, c)

	d := GP64()
	MOVQ(Mem{Base: base}, d)
	IMULQ(rk2, d)

	r := hashLen16MulLine(a, b, c, d, rk2, mul)
	Store(r, ReturnIndex(0))
	RET()
	JMP(LabelRef("fp64ret"))
}

// Return an 8-byte hash for 33 to 64 bytes.
func hashLen33to64(sbase, slen Register) {
	mul := GP64()
	MOVQ(slen, mul)
	SHLQ(Imm(1), mul)

	rk2 := GP64()
	MOVQ(Imm(k2), rk2)
	ADDQ(rk2, mul)

	a := GP64()
	MOVQ(Mem{Base: sbase}, a)
	IMULQ(rk2, a)

	b := GP64()
	MOVQ(Mem{Base: sbase, Disp: 8}, b)

	base := GP64()
	MOVQ(slen, base)
	SUBQ(Imm(16), base)
	ADDQ(sbase, base)

	c := GP64()
	MOVQ(Mem{Base: base, Disp: 8}, c)
	IMULQ(mul, c)

	d := GP64()
	MOVQ(Mem{Base: base}, d)
	IMULQ(rk2, d)

	y := GP64()
	MOVQ(a, y)

	ADDQ(b, y)
	RORQ(Imm(43), y)
	ADDQ(d, y)
	tmpc := GP64()
	MOVQ(c, tmpc)
	RORQ(Imm(30), tmpc)
	ADDQ(tmpc, y)

	ADDQ(a, c)
	ADDQ(rk2, b)
	RORQ(Imm(18), b)
	ADDQ(b, c)

	tmpy := GP64()
	MOVQ(y, tmpy)
	z := hashLen16Mul(tmpy, c, mul)

	e := GP64()
	MOVQ(Mem{Base: sbase, Disp: 16}, e)
	IMULQ(mul, e)

	f := GP64()
	MOVQ(Mem{Base: sbase, Disp: 24}, f)

	base = GP64()
	MOVQ(slen, base)
	SUBQ(Imm(32), base)
	ADDQ(sbase, base)
	g := GP64()
	MOVQ(Mem{Base: base}, g)
	ADDQ(y, g)
	IMULQ(mul, g)

	h := GP64()
	MOVQ(Mem{Base: base, Disp: 8}, h)
	ADDQ(z, h)
	IMULQ(mul, h)

	r := hashLen16MulLine(e, f, g, h, a, mul)
	Store(r, ReturnIndex(0))
	RET()
	JMP(LabelRef("fp64ret"))
}

// Return a 16-byte hash for s[0] ... s[31], a, and b.  Quick and dirty.
func weakHashLen32WithSeeds(sbase Register, disp int, a, b Register) (Register, Register) {

	w := Mem{Base: sbase, Disp: disp + 0}
	x := Mem{Base: sbase, Disp: disp + 8}
	y := Mem{Base: sbase, Disp: disp + 16}
	z := Mem{Base: sbase, Disp: disp + 24}

	// a += w
	ADDQ(w, a)

	// b = bits.RotateLeft64(b+a+z, -21)
	ADDQ(a, b)
	ADDQ(z, b)
	RORQ(Imm(21), b)

	// c := a
	c := GP64()
	MOVQ(a, c)

	// a += x
	// a += y
	ADDQ(x, a)
	ADDQ(y, a)

	// b += bits.RotateLeft64(a, -44)
	atmp := GP64()
	MOVQ(a, atmp)
	RORQ(Imm(44), atmp)
	ADDQ(atmp, b)

	// a += z
	// b += c
	ADDQ(z, a)
	ADDQ(c, b)

	r1, r2 := GP64(), GP64()
	MOVQ(a, r1)
	MOVQ(b, r2)
	return r1, r2

	return a, b

}

func loopBody(x, y, z, vlo, vhi, wlo, whi, sbase GPVirtual, mul1 GPVirtual, mul2 uint64) {
	ADDQ(y, x)
	ADDQ(vlo, x)
	ADDQ(Mem{Base: sbase, Disp: 8}, x)
	RORQ(Imm(37), x)

	IMULQ(mul1, x)

	ADDQ(vhi, y)
	ADDQ(Mem{Base: sbase, Disp: 48}, y)
	RORQ(Imm(42), y)
	IMULQ(mul1, y)

	if mul2 != 1 {
		t := GP64()
		MOVQ(U32(mul2), t)
		IMULQ(whi, t)
		XORQ(t, x)
	} else {
		XORQ(whi, x)
	}

	if mul2 != 1 {
		t := GP64()
		MOVQ(U32(mul2), t)
		IMULQ(vlo, t)
		ADDQ(t, y)
	} else {
		ADDQ(vlo, y)
	}

	ADDQ(Mem{Base: sbase, Disp: 40}, y)

	ADDQ(wlo, z)
	RORQ(Imm(33), z)
	IMULQ(mul1, z)

	{
		IMULQ(mul1, vhi)
		MOVQ(x, vlo)
		ADDQ(wlo, vlo)
		a, b := weakHashLen32WithSeeds(sbase, 0, vhi, vlo)
		MOVQ(a, vlo)
		MOVQ(b, vhi)
	}

	{
		ADDQ(z, whi)
		MOVQ(y, wlo)
		ADDQ(Mem{Base: sbase, Disp: 16}, wlo)
		a, b := weakHashLen32WithSeeds(sbase, 32, whi, wlo)
		MOVQ(a, wlo)
		MOVQ(b, whi)
	}

	XCHGQ(z, x)
}

func main() {

	ConstraintExpr("amd64,!purego")

	TEXT("Fingerprint64", NOSPLIT, "func(s []byte) uint64")

	slen := GP64()
	sbase := GP64()

	Load(Param("s").Base(), sbase)
	Load(Param("s").Len(), slen)

	CMPQ(slen, Imm(16))
	JG(LabelRef("check32"))
	hashLen0to16(sbase, slen)

	Label("check32")
	CMPQ(slen, Imm(32))
	JG(LabelRef("check64"))
	hashLen17to32(sbase, slen)

	Label("check64")
	CMPQ(slen, Imm(64))
	JG(LabelRef("long"))
	hashLen33to64(sbase, slen)

	Label("long")

	seed := uint64(81)

	vlo, vhi, wlo, whi := GP64(), GP64(), GP64(), GP64()
	XORQ(vlo, vlo)
	XORQ(vhi, vhi)
	XORQ(wlo, wlo)
	XORQ(whi, whi)

	x := GP64()

	eightOne := uint64(81)

	MOVQ(Imm(eightOne*k2), x)
	ADDQ(Mem{Base: sbase}, x)

	y := GP64()
	y64 := uint64(seed*k1) + 113
	MOVQ(Imm(y64), y)

	z := GP64()
	MOVQ(Imm(shiftMix64(y64*k2+113)*k2), z)

	endIdx := GP64()
	MOVQ(slen, endIdx)
	tmp := GP64()
	SUBQ(Imm(1), endIdx)
	MOVQ(U64(^uint64(63)), tmp)
	ANDQ(tmp, endIdx)
	last64Idx := GP64()
	MOVQ(slen, last64Idx)
	SUBQ(Imm(1), last64Idx)
	ANDQ(Imm(63), last64Idx)
	SUBQ(Imm(63), last64Idx)
	ADDQ(endIdx, last64Idx)

	last64 := GP64()
	MOVQ(last64Idx, last64)
	ADDQ(sbase, last64)

	end := GP64()
	MOVQ(slen, end)

	Label("loop")

	rk1 := GP64()
	MOVQ(Imm(k1), rk1)

	loopBody(x, y, z, vlo, vhi, wlo, whi, sbase, rk1, 1)

	ADDQ(Imm(64), sbase)
	SUBQ(Imm(64), end)
	CMPQ(end, Imm(64))
	JG(LabelRef("loop"))

	MOVQ(last64, sbase)

	mul := GP64()
	MOVQ(z, mul)
	ANDQ(Imm(0xff), mul)
	SHLQ(Imm(1), mul)
	ADDQ(rk1, mul)

	MOVQ(last64, sbase)

	SUBQ(Imm(1), slen)
	ANDQ(Imm(63), slen)
	ADDQ(slen, wlo)

	ADDQ(wlo, vlo)
	ADDQ(vlo, wlo)

	loopBody(x, y, z, vlo, vhi, wlo, whi, sbase, mul, 9)

	{
		a := hashLen16Mul(vlo, wlo, mul)
		ADDQ(z, a)
		b := shiftMix(y)
		rk0 := GP64()
		MOVQ(Imm(k0), rk0)
		IMULQ(rk0, b)
		ADDQ(b, a)

		c := hashLen16Mul(vhi, whi, mul)
		ADDQ(x, c)

		r := hashLen16Mul(a, c, mul)
		Store(r, ReturnIndex(0))
	}

	Label("fp64ret")
	RET()

	Generate()
}

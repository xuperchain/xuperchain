// Copyright 2019 The go-interpreter Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package compile

import (
	"fmt"

	asm "github.com/twitchyliquid64/golang-asm"
	"github.com/twitchyliquid64/golang-asm/obj"
	"github.com/twitchyliquid64/golang-asm/obj/x86"
)

const (
	// APushWasmStack is a symbolic instruction representing the movement
	// of a value in an x86-64 register, to the top of the WASM stack.
	APushWasmStack = x86.ALAST + iota
	// APopWasmStack is a symbolic instruction representing the movement
	// of the top of the WASM stack, to a value in an x86-64 register.
	APopWasmStack
	// ALoadGlobalsSliceHeader is a symbolic instruction representing that
	// the slice header for wasm globals should be loaded into R15. This
	// allows us to defer instruction generation till later phases, so
	// this instruction can be a NOP if already loaded.
	ALoadGlobalsSliceHeader
	// ALoadLocalsFirstElem is a symbolic instruction representing that
	// the first wasm should be loaded into R15. This allows us to defer
	// instruction generation till later phases, so this instruction can
	// be a NOP if already loaded.
	ALoadLocalsFirstElem
	// AFlushStackLength is a symbolic instruction representing that
	// the WASM stack length should be flushed from registers to main memory,
	// if it is dirty.
	AFlushStackLength
)

// dirtyRegs tracks registers which hold values.
type dirtyRegs struct {
	R13 dirtyState
	R14 dirtyState
	R15 dirtyState
}

func (regs *dirtyRegs) flush(inst *obj.Prog, builder *asm.Builder, reg uint16) {
	var regState *dirtyState
	switch reg {
	case x86.REG_R13:
		regState = &regs.R13
	default:
		panic(fmt.Sprintf("compile: unknown register: %v", reg))
	}

	switch *regState {
	case stateScratch, stateStackFirstElem, stateLocalFirstElem, stateGlobalSliceHeader:
		inst.As = obj.ANOP
		return // Value does not change - no need to write back.
	case stateStackLen:
		inst.As = x86.AMOVQ
		inst.From.Type = obj.TYPE_REG
		inst.From.Reg = x86.REG_R13
		inst.To.Type = obj.TYPE_MEM
		inst.To.Reg = x86.REG_R10
		inst.To.Offset = 8
	default:
		panic(fmt.Sprintf("compile: unknown regState: %v", regState))
	}

	*regState = stateScratch
}

// lowerAMD64 converts symbolic instructions into concrete x86-64 instructions.
func (b *AMD64Backend) lowerAMD64(builder *asm.Builder) {
	var (
		regs = &dirtyRegs{}
		inst = builder.Root()
	)

	for inst = inst.Link; inst.Link != nil; inst = inst.Link {

		switch inst.As {
		case AFlushStackLength:
			regs.flush(inst, builder, x86.REG_R13)

		case ALoadGlobalsSliceHeader:
			b.emitLoadGlobalsSliceHeader(inst, builder, regs)
		case ALoadLocalsFirstElem:
			b.emitLoadLocalsFirstElem(inst, builder, regs)

		case APushWasmStack:
			b.emitWasmStackPush(inst, builder, regs)
		case APopWasmStack:
			b.emitWasmStackLoad(inst, builder, regs)
		}
	}
}

func (b *AMD64Backend) emitLoadLocalsFirstElem(inst *obj.Prog, builder *asm.Builder, regs *dirtyRegs) {
	if regs.R15 != stateLocalFirstElem {
		inst.As = x86.AMOVQ
		inst.To.Type = obj.TYPE_REG
		inst.To.Reg = x86.REG_R15
		inst.From.Type = obj.TYPE_MEM
		inst.From.Reg = x86.REG_R11
		regs.R15 = stateLocalFirstElem
	} else {
		inst.As = obj.ANOP
	}
}

func (b *AMD64Backend) emitLoadGlobalsSliceHeader(inst *obj.Prog, builder *asm.Builder, regs *dirtyRegs) {
	if regs.R15 != stateGlobalSliceHeader {
		inst.As = x86.AMOVQ
		inst.To.Type = obj.TYPE_REG
		inst.To.Reg = x86.REG_R15
		inst.From.Type = obj.TYPE_MEM
		inst.From.Reg = x86.REG_SP
		inst.From.Offset = 32

		prog := builder.NewProg()
		prog.As = x86.AMOVQ
		prog.To.Type = obj.TYPE_REG
		prog.To.Reg = x86.REG_R15
		prog.From.Type = obj.TYPE_MEM
		prog.From.Reg = x86.REG_R15

		prog.Link = inst.Link
		inst.Link = prog
		regs.R15 = stateGlobalSliceHeader
	} else {
		inst.As = obj.ANOP
	}
}

func (b *AMD64Backend) emitWasmStackLoad(inst *obj.Prog, builder *asm.Builder, regs *dirtyRegs) {
	// movq r13,     [r10+8] (if not already loaded)
	// decq r13
	// movq r14,     [r10] (if not already loaded)
	// leaq r12,     [r14 + r13*8]
	// movq reg,     [r12]
	to := inst.To
	nextInst := inst.Link
	ci := inst.From.Val.(currentInstruction)

	if regs.R13 != stateStackLen {
		inst.As = x86.AMOVQ
		inst.To.Type = obj.TYPE_REG
		inst.To.Reg = x86.REG_R13
		inst.From.Type = obj.TYPE_MEM
		inst.From.Reg = x86.REG_R10
		inst.From.Offset = 8
		regs.R13 = stateStackLen
	} else {
		inst.As = obj.ANOP
	}
	prev := inst

	prog := builder.NewProg()
	prog.As = x86.ADECQ
	prog.To.Type = obj.TYPE_REG
	prog.To.Reg = x86.REG_R13
	prev.Link = prog
	prev = prog

	if b.EmitBoundsChecks {
		// movq r12, [r10+16]
		// cmp r12, r13
		// ja endbounds
		// <emitExit() code>
		// endbounds:
		prog = builder.NewProg()
		prog.As = x86.AMOVQ
		prog.To.Type = obj.TYPE_REG
		prog.To.Reg = x86.REG_R12
		prog.From.Type = obj.TYPE_MEM
		prog.From.Reg = x86.REG_R10
		prog.From.Offset = 16
		prev.Link = prog
		prev = prog

		prog = builder.NewProg()
		prog.As = x86.ACMPQ
		prog.To.Type = obj.TYPE_REG
		prog.To.Reg = x86.REG_R13
		prog.From.Type = obj.TYPE_REG
		prog.From.Reg = x86.REG_R12
		prev.Link = prog
		prev = prog

		jmp := builder.NewProg()
		jmp.As = x86.AJHI
		jmp.To.Type = obj.TYPE_BRANCH
		prev.Link = jmp
		prev = jmp

		retValue, ret := b.exitInstructions(builder, CompletionBadBounds|makeExitIndex(ci.idx))
		prev.Link = retValue
		retValue.Link = ret
		prev = ret

		prog = builder.NewProg()
		prog.As = obj.ANOP // branch target - assembler will optimize out.
		jmp.Pcond = prog
		prev.Link = prog
		prev = prog
	}

	if regs.R14 != stateStackFirstElem {
		prog = builder.NewProg()
		prog.As = x86.AMOVQ
		prog.To.Type = obj.TYPE_REG
		prog.To.Reg = x86.REG_R14
		prog.From.Type = obj.TYPE_MEM
		prog.From.Reg = x86.REG_R10
		prev.Link = prog
		prev = prog
		regs.R14 = stateStackFirstElem
	}

	prog = builder.NewProg()
	prog.As = x86.ALEAQ
	prog.To.Type = obj.TYPE_REG
	prog.To.Reg = x86.REG_R12
	prog.From.Type = obj.TYPE_MEM
	prog.From.Reg = x86.REG_R14
	prog.From.Scale = 8
	prog.From.Index = x86.REG_R13
	prev.Link = prog
	prev = prog

	prog = builder.NewProg()
	prog.As = x86.AMOVQ
	prog.From.Type = obj.TYPE_MEM
	prog.From.Reg = x86.REG_R12
	prog.To = to
	prev.Link = prog
	prog.Link = nextInst
}

func (b *AMD64Backend) emitWasmStackPush(inst *obj.Prog, builder *asm.Builder, regs *dirtyRegs) {
	// movq r14,     [r10] (if not already loaded)
	// movq r13,     [r10+8] (if not already loaded)
	// leaq r12,     [r14 + r13*8]
	// movq [r12],   <data>
	// incq r13
	from := inst.From
	nextInst := inst.Link
	ci := inst.To.Val.(currentInstruction)

	var prog *obj.Prog
	if regs.R14 != stateStackFirstElem {
		inst.As = x86.AMOVQ
		inst.To.Type = obj.TYPE_REG
		inst.To.Reg = x86.REG_R14
		inst.From.Type = obj.TYPE_MEM
		inst.From.Reg = x86.REG_R10
		regs.R14 = stateStackFirstElem
	} else {
		inst.As = obj.ANOP
	}

	prev := inst
	if regs.R13 != stateStackLen {
		prog = builder.NewProg()
		prev.Link = prog
		prog.As = x86.AMOVQ
		prog.To.Type = obj.TYPE_REG
		prog.To.Reg = x86.REG_R13
		prog.From.Type = obj.TYPE_MEM
		prog.From.Reg = x86.REG_R10
		prog.From.Offset = 8
		prev = prog
		regs.R13 = stateStackLen
	}

	if b.EmitBoundsChecks {
		// movq r12, [r10+16]
		// cmp r12, r13
		// ja endbounds
		// <emitExit() code>
		// endbounds:
		prog = builder.NewProg()
		prev.Link = prog
		prog.As = x86.AMOVQ
		prog.To.Type = obj.TYPE_REG
		prog.To.Reg = x86.REG_R12
		prog.From.Type = obj.TYPE_MEM
		prog.From.Reg = x86.REG_R10
		prog.From.Offset = 16
		prev = prog

		prog = builder.NewProg()
		prev.Link = prog
		prog.As = x86.ACMPQ
		prog.To.Type = obj.TYPE_REG
		prog.To.Reg = x86.REG_R13
		prog.From.Type = obj.TYPE_REG
		prog.From.Reg = x86.REG_R12
		prev = prog

		jmp := builder.NewProg()
		prev.Link = jmp
		jmp.As = x86.AJHI
		jmp.To.Type = obj.TYPE_BRANCH
		prev = jmp

		retValue, ret := b.exitInstructions(builder, CompletionBadBounds|makeExitIndex(ci.idx))
		prev.Link = retValue
		retValue.Link = ret
		prev = ret

		prog = builder.NewProg()
		prog.As = obj.ANOP // branch target - assembler will optimize out.
		jmp.Pcond = prog
		prev.Link = prog
		prev = prog
	}

	prog = builder.NewProg()
	prog.As = x86.ALEAQ
	prog.To.Type = obj.TYPE_REG
	prog.To.Reg = x86.REG_R12
	prog.From.Type = obj.TYPE_MEM
	prog.From.Reg = x86.REG_R14
	prog.From.Scale = 8
	prog.From.Index = x86.REG_R13
	prev.Link = prog
	prev = prog

	prog = builder.NewProg()
	prog.As = x86.AMOVQ
	prog.To.Type = obj.TYPE_MEM
	prog.To.Reg = x86.REG_R12
	prog.From = from
	prev.Link = prog
	prev = prog

	prog = builder.NewProg()
	prog.As = x86.AINCQ
	prog.To.Type = obj.TYPE_REG
	prog.To.Reg = x86.REG_R13
	prev.Link = prog
	prog.Link = nextInst
}

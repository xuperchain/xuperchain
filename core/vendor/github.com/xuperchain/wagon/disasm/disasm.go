// Copyright 2017 The go-interpreter Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package disasm provides functions for disassembling WebAssembly bytecode.
package disasm

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"

	"github.com/xuperchain/wagon/internal/stack"
	"github.com/xuperchain/wagon/wasm"
	"github.com/xuperchain/wagon/wasm/leb128"
	ops "github.com/xuperchain/wagon/wasm/operators"
)

// Instr describes an instruction, consisting of an operator, with its
// appropriate immediate value(s).
type Instr struct {
	Op ops.Op

	// Immediates are arguments to an operator in the bytecode stream itself.
	// Valid value types are:
	// - (u)(int/float)(32/64)
	// - wasm.BlockType
	Immediates  []interface{}
	NewStack    *StackInfo // non-nil if the instruction creates or unwinds a stack.
	Block       *BlockInfo // non-nil if the instruction starts or ends a new block.
	Unreachable bool       // whether the operator can be reached during execution
	// IsReturn is true if executing this instruction will result in the
	// function returning. This is true for branches (br, br_if) to
	// the depth <max_relative_depth> + 1, or the return operator itself.
	// If true, NewStack for this instruction is nil.
	IsReturn bool
	// If the operator is br_table (ops.BrTable), this is a list of StackInfo
	// fields for each of the blocks/branches referenced by the operator.
	Branches []StackInfo
}

// StackInfo stores details about a new stack created or unwound by an instruction.
type StackInfo struct {
	StackTopDiff int64 // The difference between the stack depths at the end of the block
	PreserveTop  bool  // Whether the value on the top of the stack should be preserved while unwinding
	IsReturn     bool  // Whether the unwind is equivalent to a return
}

// BlockInfo stores details about a block created or ended by an instruction.
type BlockInfo struct {
	Start     bool           // If true, this instruction starts a block. Else this instruction ends it.
	Signature wasm.BlockType // The block signature

	// Indices to the accompanying control operator.
	// For 'if', this is the index to the 'else' operator.
	IfElseIndex int
	// For 'else', this is the index to the 'if' operator.
	ElseIfIndex int
	// The index to the `end' operator for if/else/loop/block.
	EndIndex int
	// For end, it is the index to the operator that starts the block.
	BlockStartIndex int
}

// Disassembly is the result of disassembling a WebAssembly function.
type Disassembly struct {
	Code     []Instr
	MaxDepth int // The maximum stack depth that can be reached while executing this function
}

func (d *Disassembly) checkMaxDepth(depth int) {
	if depth > d.MaxDepth {
		d.MaxDepth = depth
	}
}

func pushPolymorphicOp(indexStack [][]int, index int) {
	indexStack[len(indexStack)-1] = append(indexStack[len(indexStack)-1], index)
}

func isInstrReachable(indexStack [][]int) bool {
	return len(indexStack[len(indexStack)-1]) == 0
}

var ErrStackUnderflow = errors.New("disasm: stack underflow")

// NewDisassembly disassembles the given function. It also takes the function's
// parent module as an argument for locating any other functions referenced by
// fn.
func NewDisassembly(fn wasm.Function, module *wasm.Module) (*Disassembly, error) {
	return NewDisassemblyWithGas(fn, module, nil)
}

// NewDisassemblyWithGas like NewDisassembly but add checkGas instructions.
func NewDisassemblyWithGas(fn wasm.Function, module *wasm.Module, gasMapper GasMapper) (*Disassembly, error) {
	code := fn.Body.Code
	instrs, err := Disassemble(code)
	if err != nil {
		return nil, err
	}
	if gasMapper != nil {
		instrs = AddGasInstr(instrs, gasMapper)
	}
	disas := &Disassembly{
		Code: make([]Instr, 0, len(instrs)),
	}

	// A stack of int arrays holding indices to instructions that make the stack
	// polymorphic. Each block has its corresponding array. We start with one
	// array for the root stack
	blockPolymorphicOps := [][]int{{}}
	// a stack of current execution stack depth values, so that the depth for each
	// stack is maintained independently for calculating discard values
	stackDepths := &stack.Stack{}
	stackDepths.Push(0)
	blockIndices := &stack.Stack{} // a stack of indices to operators which start new blocks
	curIndex := 0

	for _, instr := range instrs {
		//logger.Printf("stack top is %d", stackDepths.Top())
		opStr := instr.Op
		op := opStr.Code
		if op == ops.End || op == ops.Else {
			// There are two possible cases here:
			// 1. The corresponding block/if/loop instruction
			// *is* reachable, and an instruction somewhere in this
			// block (and NOT in a nested block) makes the stack
			// polymorphic. In this case, this end/else is reachable.
			//
			// 2. The corresponding block/if/loop instruction
			// is *not* reachable, which makes this end/else unreachable
			// too.
			isUnreachable := blockIndices.Len() != len(blockPolymorphicOps)-1
			instr.Unreachable = isUnreachable
		} else {
			instr.Unreachable = !isInstrReachable(blockPolymorphicOps)
		}

		var blockStartIndex uint64
		switch op {
		case ops.End, ops.Else:
			blockStartIndex = blockIndices.Pop()
			if op == ops.Else {
				blockIndices.Push(uint64(curIndex))
			}
		case ops.Block, ops.Loop, ops.If:
			blockIndices.Push(uint64(curIndex))
		}

		if instr.Unreachable {
			continue
		}

		// logger.Printf("op: %s, unreachable: %v", opStr.Name, instr.Unreachable)
		if !opStr.Polymorphic {
			top := int(stackDepths.Top())
			top -= len(opStr.Args)
			stackDepths.SetTop(uint64(top))
			if top < 0 {
				panic("underflow during validation")
			}
			if opStr.Returns != wasm.ValueType(wasm.BlockTypeEmpty) {
				top++
				stackDepths.SetTop(uint64(top))
			}
			disas.checkMaxDepth(top)
		}

		switch op {
		case ops.Unreachable:
			pushPolymorphicOp(blockPolymorphicOps, curIndex)
		case ops.Drop:
			stackDepths.SetTop(stackDepths.Top() - 1)
		case ops.Select:
			stackDepths.SetTop(stackDepths.Top() - 2)
		case ops.Return:
			stackDepths.SetTop(stackDepths.Top() - uint64(len(fn.Sig.ReturnTypes)))
			pushPolymorphicOp(blockPolymorphicOps, curIndex)
		case ops.End, ops.Else:
			blockSig := disas.Code[blockStartIndex].Block.Signature
			instr.Block = &BlockInfo{
				Start:     false,
				Signature: blockSig,
			}
			if op == ops.End {
				instr.Block.BlockStartIndex = int(blockStartIndex)
				disas.Code[blockStartIndex].Block.EndIndex = curIndex
			} else { // ops.Else
				instr.Block.ElseIfIndex = int(blockStartIndex)
				disas.Code[blockStartIndex].Block.IfElseIndex = int(curIndex)
			}

			// The max depth reached while execing the last block
			// If the signature of the current block is not empty,
			// this will be incremented.
			// Same with ops.Br/BrIf, we subtract 2 instead of 1
			// to get the depth of the *parent* block of the branch
			// we want to take.
			prevDepthIndex := stackDepths.Len() - 2
			prevDepth := stackDepths.Get(prevDepthIndex)

			if op != ops.Else && blockSig != wasm.BlockTypeEmpty {
				stackDepths.Set(prevDepthIndex, prevDepth+1)
				disas.checkMaxDepth(int(stackDepths.Get(prevDepthIndex)))
			}

			//logger.Printf("setting new stack for %s block (%d)", disas.Code[blockStartIndex].Op.Name, blockStartIndex)
			blockPolymorphicOps = blockPolymorphicOps[:len(blockPolymorphicOps)-1]

			stackDepths.Pop()
			if op == ops.Else {
				stackDepths.Push(stackDepths.Top())
				blockPolymorphicOps = append(blockPolymorphicOps, []int{})
			}
		case ops.Block, ops.Loop, ops.If:
			sig := instr.Immediates[0].(wasm.BlockType)
			//logger.Printf("if, depth is %d", stackDepths.Top())
			stackDepths.Push(stackDepths.Top())
			blockPolymorphicOps = append(blockPolymorphicOps, []int{})
			instr.Block = &BlockInfo{
				Start:     true,
				Signature: sig,
			}
		case ops.Br, ops.BrIf:
			depth := instr.Immediates[0].(uint32)
			if int(depth) == blockIndices.Len() {
				instr.IsReturn = true
			} else {
				curDepth := stackDepths.Top()
				// whenever we take a branch, the stack is unwound
				// to the height of stack of its *parent* block, which
				// is why we subtract 2 instead of 1.
				// prevDepth holds the height of the stack when
				// the block that we branch to started.
				prevDepth := stackDepths.Get(stackDepths.Len() - 2 - int(depth))
				elemsDiscard := int(curDepth) - int(prevDepth)
				if elemsDiscard < 0 {
					return nil, ErrStackUnderflow
				}

				// No need to subtract 2 here, we are getting the block
				// we need to branch to.
				// No need Discard one. and PreserveTop.
				if elemsDiscard > 1 {
					index := blockIndices.Get(blockIndices.Len() - 1 - int(depth))
					instr.NewStack = &StackInfo{
						StackTopDiff: int64(elemsDiscard),
						PreserveTop:  disas.Code[index].Block.Signature != wasm.BlockTypeEmpty,
					}
				}
			}
			if op == ops.Br {
				pushPolymorphicOp(blockPolymorphicOps, curIndex)
			}

		case ops.BrTable:
			stackDepths.SetTop(stackDepths.Top() - 1)
			targetCount := instr.Immediates[0].(uint32)
			for i := uint32(0); i < targetCount; i++ {
				entry := instr.Immediates[i+1].(uint32)

				var info StackInfo
				if int(entry) == blockIndices.Len() {
					info.IsReturn = true
				} else {
					curDepth := stackDepths.Top()
					branchDepth := stackDepths.Get(stackDepths.Len() - 2 - int(entry))
					elemsDiscard := int(curDepth) - int(branchDepth)
					//logger.Printf("Curdepth %d branchDepth %d discard %d", curDepth, branchDepth, elemsDiscard)

					if elemsDiscard < 0 {
						return nil, ErrStackUnderflow
					}
					index := blockIndices.Get(blockIndices.Len() - 1 - int(entry))
					info.StackTopDiff = int64(elemsDiscard)
					info.PreserveTop = disas.Code[index].Block.Signature != wasm.BlockTypeEmpty
				}
				instr.Branches = append(instr.Branches, info)
			}
			defaultTarget := instr.Immediates[targetCount+1].(uint32)

			var info StackInfo
			if int(defaultTarget) == blockIndices.Len() {
				info.IsReturn = true
			} else {

				curDepth := stackDepths.Top()
				branchDepth := stackDepths.Get(stackDepths.Len() - 2 - int(defaultTarget))
				elemsDiscard := int(curDepth) - int(branchDepth)

				if elemsDiscard < 0 {
					return nil, ErrStackUnderflow
				}
				index := blockIndices.Get(blockIndices.Len() - 1 - int(defaultTarget))
				info.StackTopDiff = int64(elemsDiscard)
				info.PreserveTop = disas.Code[index].Block.Signature != wasm.BlockTypeEmpty
			}
			instr.Branches = append(instr.Branches, info)
			pushPolymorphicOp(blockPolymorphicOps, curIndex)
		case ops.Call, ops.CallIndirect:
			index := instr.Immediates[0].(uint32)
			var sig *wasm.FunctionSig
			top := int(stackDepths.Top())

			switch op {
			case ops.CallIndirect:
				if module.Types == nil {
					return nil, errors.New("missing types section")
				}
				sig = &module.Types.Entries[index]
				top--
			default:
				sig, err = module.GetFunctionSig(index)
				if err != nil {
					return nil, err
				}
			}

			top -= len(sig.ParamTypes)
			top += len(sig.ReturnTypes)
			stackDepths.SetTop(uint64(top))
			disas.checkMaxDepth(top)
		case ops.GetLocal, ops.SetLocal, ops.TeeLocal, ops.GetGlobal, ops.SetGlobal:
			top := stackDepths.Top()
			switch op {
			case ops.GetLocal, ops.GetGlobal:
				top++
				stackDepths.SetTop(top)
				disas.checkMaxDepth(int(top))
			case ops.SetLocal, ops.SetGlobal:
				top--
				stackDepths.SetTop(top)
			case ops.TeeLocal:
				// stack remains unchanged for tee_local
			}
		}

		disas.Code = append(disas.Code, instr)
		curIndex++
	}

	if logging {
		for _, instr := range disas.Code {
			logger.Printf("%v %v", instr.Op.Name, instr.NewStack)
		}
	}

	return disas, nil
}

// Disassemble disassembles a given function body into a set of instructions. It won't check operations for validity.
func Disassemble(code []byte) ([]Instr, error) {
	reader := bytes.NewReader(code)
	var out []Instr
	for {
		op, err := reader.ReadByte()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		opStr, err := ops.New(op)
		if err != nil {
			return nil, err
		}
		instr := Instr{
			Op: opStr,
		}

		switch op {
		case ops.Block, ops.Loop, ops.If:
			sig, err := wasm.ReadByte(reader)
			if err != nil {
				return nil, err
			}
			instr.Immediates = append(instr.Immediates, wasm.BlockType(sig))
		case ops.Br, ops.BrIf:
			depth, err := leb128.ReadVarUint32(reader)
			if err != nil {
				return nil, err
			}
			instr.Immediates = append(instr.Immediates, depth)
		case ops.BrTable:
			targetCount, err := leb128.ReadVarUint32(reader)
			if err != nil {
				return nil, err
			}
			instr.Immediates = append(instr.Immediates, targetCount)
			for i := uint32(0); i < targetCount; i++ {
				entry, err := leb128.ReadVarUint32(reader)
				if err != nil {
					return nil, err
				}
				instr.Immediates = append(instr.Immediates, entry)
			}

			defaultTarget, err := leb128.ReadVarUint32(reader)
			if err != nil {
				return nil, err
			}
			instr.Immediates = append(instr.Immediates, defaultTarget)
		case ops.Call, ops.CallIndirect:
			index, err := leb128.ReadVarUint32(reader)
			if err != nil {
				return nil, err
			}
			instr.Immediates = append(instr.Immediates, index)
			if op == ops.CallIndirect {
				idx, err := wasm.ReadByte(reader)
				if err != nil {
					return nil, err
				}
				if idx != 0x00 {
					return nil, errors.New("disasm: table index in call_indirect must be 0")
				}
				instr.Immediates = append(instr.Immediates, uint32(idx))
			}
		case ops.GetLocal, ops.SetLocal, ops.TeeLocal, ops.GetGlobal, ops.SetGlobal:
			index, err := leb128.ReadVarUint32(reader)
			if err != nil {
				return nil, err
			}
			instr.Immediates = append(instr.Immediates, index)
		case ops.I32Const:
			i, err := leb128.ReadVarint32(reader)
			if err != nil {
				return nil, err
			}
			instr.Immediates = append(instr.Immediates, i)
		case ops.I64Const:
			i, err := leb128.ReadVarint64(reader)
			if err != nil {
				return nil, err
			}
			instr.Immediates = append(instr.Immediates, i)
		case ops.F32Const:
			var b [4]byte
			if _, err := io.ReadFull(reader, b[:]); err != nil {
				return nil, err
			}
			i := binary.LittleEndian.Uint32(b[:])
			instr.Immediates = append(instr.Immediates, math.Float32frombits(i))
		case ops.F64Const:
			var b [8]byte
			if _, err := io.ReadFull(reader, b[:]); err != nil {
				return nil, err
			}
			i := binary.LittleEndian.Uint64(b[:])
			instr.Immediates = append(instr.Immediates, math.Float64frombits(i))
		case ops.I32Load, ops.I64Load, ops.F32Load, ops.F64Load, ops.I32Load8s, ops.I32Load8u, ops.I32Load16s, ops.I32Load16u, ops.I64Load8s, ops.I64Load8u, ops.I64Load16s, ops.I64Load16u, ops.I64Load32s, ops.I64Load32u, ops.I32Store, ops.I64Store, ops.F32Store, ops.F64Store, ops.I32Store8, ops.I32Store16, ops.I64Store8, ops.I64Store16, ops.I64Store32:
			// read memory_immediate
			align, err := leb128.ReadVarUint32(reader)
			if err != nil {
				return nil, err
			}
			instr.Immediates = append(instr.Immediates, align)

			offset, err := leb128.ReadVarUint32(reader)
			if err != nil {
				return nil, err
			}
			instr.Immediates = append(instr.Immediates, offset)
		case ops.CurrentMemory, ops.GrowMemory:
			idx, err := wasm.ReadByte(reader)
			if err != nil {
				return nil, err
			}
			if idx != 0x00 {
				return nil, errors.New("disasm: memory index must be 0")
			}
			instr.Immediates = append(instr.Immediates, uint8(idx))
		}
		out = append(out, instr)
	}
	return out, nil
}

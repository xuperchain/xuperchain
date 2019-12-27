// Copyright 2019 The go-interpreter Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package compile

import (
	ops "github.com/xuperchain/wagon/wasm/operators"
)

// ScanFunc scans the given function information, emitting selections of
// bytecode which could be compiled into function code.
func (s *scanner) ScanFunc(bytecode []byte, meta *BytecodeMetadata) ([]CompilationCandidate, error) {
	var finishedCandidates []CompilationCandidate
	inProgress := CompilationCandidate{}

	for i, inst := range meta.Instructions {
		// Except for the first instruction, we cant emit a native section
		// where other parts of code try and call into us halfway. Maybe we
		// can support that in the future.
		_, hasInboundTarget := meta.InboundTargets[int64(inst.Start)]
		isInsideBranchTarget := hasInboundTarget && inst.Start > 0 && inProgress.Metrics.AllOps > 0

		if !s.supportedOpcodes[inst.Op] || isInsideBranchTarget {
			// See if the candidate can be emitted.
			if inProgress.Metrics.AllOps > 2 {
				finishedCandidates = append(finishedCandidates, inProgress)
			}
			inProgress.reset()
			continue
		}

		// Still a supported run.

		if inProgress.Metrics.AllOps == 0 {
			// First instruction of the candidate - setup structure.
			inProgress.Start = uint(inst.Start)
			inProgress.StartInstruction = i
		}
		inProgress.EndInstruction = i + 1
		inProgress.End = uint(inst.Start) + uint(inst.Size)

		// TODO: Add to this table as backends support more opcodes.
		switch inst.Op {
		case ops.I64Load, ops.I32Load, ops.F64Load, ops.F32Load:
			fakeBE := &AMD64Backend{}
			memSize, _ := fakeBE.paramsForMemoryOp(inst.Op)
			inProgress.Metrics.MemoryReads += memSize
			inProgress.Metrics.StackWrites++
		case ops.I64Store, ops.I32Store, ops.F64Store, ops.F32Store:
			fakeBE := &AMD64Backend{}
			memSize, _ := fakeBE.paramsForMemoryOp(inst.Op)
			inProgress.Metrics.MemoryWrites += memSize
			inProgress.Metrics.StackReads += 2
		case ops.I64Const, ops.I32Const, ops.GetLocal, ops.GetGlobal:
			inProgress.Metrics.IntegerOps++
			inProgress.Metrics.StackWrites++
		case ops.F64Const, ops.F32Const:
			inProgress.Metrics.FloatOps++
			inProgress.Metrics.StackWrites++
		case ops.SetLocal, ops.SetGlobal:
			inProgress.Metrics.IntegerOps++
			inProgress.Metrics.StackReads++
		case ops.I64Eqz:
			inProgress.Metrics.IntegerOps++
			inProgress.Metrics.StackReads++
			inProgress.Metrics.StackWrites++

		case ops.I64Eq, ops.I64Ne, ops.I64LtU, ops.I64GtU, ops.I64LeU, ops.I64GeU,
			ops.I64Shl, ops.I64ShrU, ops.I64ShrS,
			ops.I64DivU, ops.I32DivU, ops.I64RemU, ops.I32RemU, ops.I64DivS, ops.I32DivS, ops.I64RemS, ops.I32RemS,
			ops.I64Add, ops.I32Add, ops.I64Sub, ops.I32Sub, ops.I64Mul, ops.I32Mul,
			ops.I64And, ops.I32And, ops.I64Or, ops.I32Or, ops.I64Xor, ops.I32Xor:
			inProgress.Metrics.IntegerOps++
			inProgress.Metrics.StackReads += 2
			inProgress.Metrics.StackWrites++

		case ops.F64Add, ops.F32Add, ops.F64Sub, ops.F32Sub, ops.F64Div, ops.F32Div, ops.F64Mul, ops.F32Mul,
			ops.F64Min, ops.F32Min, ops.F64Max, ops.F32Max,
			ops.F64Eq, ops.F64Ne, ops.F64Lt, ops.F64Gt, ops.F64Le, ops.F64Ge,
			ops.F32Eq, ops.F32Ne, ops.F32Lt, ops.F32Gt, ops.F32Le, ops.F32Ge:
			inProgress.Metrics.FloatOps++
			inProgress.Metrics.StackReads += 2
			inProgress.Metrics.StackWrites++

		case ops.F64ConvertUI64, ops.F64ConvertSI64, ops.F32ConvertUI64, ops.F32ConvertSI64,
			ops.F64ConvertUI32, ops.F64ConvertSI32, ops.F32ConvertUI32, ops.F32ConvertSI32:
			inProgress.Metrics.FloatOps++
			inProgress.Metrics.StackReads++
			inProgress.Metrics.StackWrites++

		case ops.Drop:
			inProgress.Metrics.StackReads++
		case ops.Select:
			inProgress.Metrics.StackReads += 3
			inProgress.Metrics.StackWrites++

		case ops.F64ReinterpretI64, ops.F32ReinterpretI32, ops.I64ReinterpretF64, ops.I32ReinterpretF32:
			inProgress.Metrics.FloatOps++
			inProgress.Metrics.IntegerOps++
		}
		inProgress.Metrics.AllOps++
	}

	// End of instructions - emit the inProgress candidate if
	// its at least 3 instructions.
	if inProgress.Metrics.AllOps > 2 {
		finishedCandidates = append(finishedCandidates, inProgress)
	}
	return finishedCandidates, nil
}

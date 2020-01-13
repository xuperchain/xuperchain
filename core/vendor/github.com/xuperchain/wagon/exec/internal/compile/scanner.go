// Copyright 2019 The go-interpreter Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package compile

type scanner struct {
	supportedOpcodes map[byte]bool
}

// InstructionMetadata describes a bytecode instruction.
type InstructionMetadata struct {
	Op byte
	// Start represents the byte offset of this instruction
	// in the function's instruction stream.
	Start int
	// Size is the number of bytes in the instruction stream
	// needed to represent this instruction.
	Size int
}

// CompilationCandidate describes a range of bytecode that can
// be translated to native code.
type CompilationCandidate struct {
	Start            uint    // Bytecode index of the first opcode.
	End              uint    // Bytecode index of the last byte in the instruction.
	StartInstruction int     // InstructionMeta index of the first instruction.
	EndInstruction   int     // InstructionMeta index of the last instruction.
	Metrics          Metrics // Metrics about the instructions between first & last index.
}

func (s *CompilationCandidate) reset() {
	s.Start = 0
	s.End = 0
	s.StartInstruction = 0
	s.EndInstruction = 1
	s.Metrics = Metrics{}
}

// Bounds returns the beginning & end index in the bytecode which
// this candidate would replace. The end index is not inclusive.
func (s *CompilationCandidate) Bounds() (uint, uint) {
	return s.Start, s.End
}

// Metrics describes the heuristics of an instruction sequence.
type Metrics struct {
	MemoryReads, MemoryWrites uint
	StackReads, StackWrites   uint

	AllOps     int
	IntegerOps int
	FloatOps   int
}

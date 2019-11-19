// Copyright 2019 The go-interpreter Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package exec

import (
	"encoding/binary"
	"fmt"
	"runtime"

	"github.com/xuperchain/wagon/exec/internal/compile"
	ops "github.com/xuperchain/wagon/wasm/operators"
)

// Parameters that decide whether a sequence should be compiled.
// TODO: Expose some way for these to be customized at runtime
// via VMOptions.
const (
	// NOTE: must never be less than 5, as room is needed to pack the
	// wagon.nativeExec instruction and its parameter.
	minInstBytes                = 5
	minArithInstructionSequence = 2
)

var supportedNativeArchs []nativeArch

type nativeArch struct {
	Arch, OS string
	make     func(endianness binary.ByteOrder) *nativeCompiler
}

// nativeCompiler represents a backend for native code generation + execution.
type nativeCompiler struct {
	Scanner   sequenceScanner
	Builder   instructionBuilder
	allocator pageAllocator
}

func (c *nativeCompiler) Close() error {
	return c.allocator.Close()
}

// pageAllocator is responsible for the efficient allocation of
// executable, aligned regions of executable memory.
type pageAllocator interface {
	AllocateExec(asm []byte) (compile.NativeCodeUnit, error)
	Close() error
}

// sequenceScanner is responsible for detecting runs of supported opcodes
// that could benefit from compilation into native instructions.
type sequenceScanner interface {
	// ScanFunc returns an ordered, non-overlapping set of
	// sequences to compile into native code.
	ScanFunc(bytecode []byte, meta *compile.BytecodeMetadata) ([]compile.CompilationCandidate, error)
}

// instructionBuilder is responsible for compiling wasm opcodes into
// native instructions.
type instructionBuilder interface {
	// Build compiles the specified bytecode into native instructions.
	Build(candidate compile.CompilationCandidate, code []byte, meta *compile.BytecodeMetadata) ([]byte, error)
}

// NativeCompilationError represents a failure to compile a sequence
// of instructions to native code.
type NativeCompilationError struct {
	Start, End uint
	FuncIndex  int
	Err        error
}

func (e NativeCompilationError) Error() string {
	return fmt.Sprintf("exec: native compilation failed on vm.funcs[%d].code[%d:%d]: %v", e.FuncIndex, e.Start, e.End, e.Err)
}

func nativeBackend() (bool, *nativeCompiler) {
	for _, c := range supportedNativeArchs {
		if c.Arch == runtime.GOARCH && c.OS == runtime.GOOS {
			backend := c.make(endianess)
			return true, backend
		}
	}
	return false, nil
}

func (vm *VM) tryNativeCompile() error {
	if vm.nativeBackend == nil {
		return nil
	}

	for i := range vm.funcs {
		if _, isGoFunc := vm.funcs[i].(goFunction); isGoFunc {
			continue
		}

		fn := vm.funcs[i].(compiledFunction)
		candidates, err := vm.nativeBackend.Scanner.ScanFunc(fn.code, fn.codeMeta)
		if err != nil {
			return fmt.Errorf("exec: AOT scan failed on vm.funcs[%d]: %v", i, err)
		}

		for _, candidate := range candidates {
			if (candidate.Metrics.IntegerOps + candidate.Metrics.FloatOps) < minArithInstructionSequence {
				continue
			}
			lower, upper := candidate.Bounds()
			if (upper - lower) < minInstBytes {
				continue
			}

			asm, err := vm.nativeBackend.Builder.Build(candidate, fn.code, fn.codeMeta)
			if err != nil {
				return NativeCompilationError{
					Err:       err,
					Start:     lower,
					End:       upper,
					FuncIndex: i,
				}
			}
			unit, err := vm.nativeBackend.allocator.AllocateExec(asm)
			if err != nil {
				return fmt.Errorf("exec: allocator.AllocateExec() failed: %v", err)
			}
			fn.asm = append(fn.asm, asmBlock{
				nativeUnit: unit,
				resumePC:   upper,
			})

			// Patch the wasm opcode stream to call into the native section.
			// The number of bytes touched here must always be equal to
			// nativeExecPrologueSize and <= minInstructionSequence.
			fn.code[lower] = ops.WagonNativeExec
			endianess.PutUint32(fn.code[lower+1:], uint32(len(fn.asm)-1))
			// make the remainder of the recompiled instructions
			// unreachable: this should trap the program in the event that
			// a bug in code offsets & candidate sequence detection results in
			// a jump to the middle of re-compiled code.
			// This conservative behaviour is the least likely to result in
			// bugs becoming security issues.
			for i := lower + 5; i < upper-1; i++ {
				fn.code[i] = ops.Unreachable
			}
		}
		vm.funcs[i] = fn
	}

	return nil
}

// nativeCodeInvocation calls into one of the assembled code blocks.
// Assembled code blocks expect the following two pieces of
// information on the stack:
// [fp:fp+pointerSize]: sliceHeader for the stack.
// [fp+pointerSize:fp+pointerSize*2]: sliceHeader for locals variables.
func (vm *VM) nativeCodeInvocation(asmIndex uint32) {
	block := vm.ctx.asm[asmIndex]
	finishSignal := block.nativeUnit.Invoke(&vm.ctx.stack, &vm.ctx.locals, &vm.globals, &vm.memory)

	switch finishSignal.CompletionStatus() {
	case compile.CompletionOK:
	case compile.CompletionFatalInternalError:
		panic("fatal error in native execution")
	case compile.CompletionBadBounds:
		panic("exec: out of bounds memory access")
	case compile.CompletionDivideZero:
		panic("runtime error: integer divide by zero")
	}
	vm.ctx.pc = int64(block.resumePC)
}

// CompileStats returns statistics about native compilation performed on
// the VM.
func (vm *VM) CompileStats() NativeCompileStats {
	out := NativeCompileStats{
		Ops: map[byte]*OpStats{},
	}

	for i := range vm.funcs {
		if _, isGoFunc := vm.funcs[i].(*goFunction); isGoFunc {
			continue
		}

		fn := vm.funcs[i].(compiledFunction)
		out.NumCompiledBlocks += len(fn.asm)

		for _, inst := range fn.codeMeta.Instructions {
			if _, exists := out.Ops[inst.Op]; !exists {
				out.Ops[inst.Op] = &OpStats{}
			}

			// Instructions which are native-compiled are re-written to the
			// ops.WagonNativeExec opcode, so a mismatch indicates native compilation.
			if fn.code[inst.Start] == inst.Op {
				out.Ops[inst.Op].Interpreted++
			} else {
				out.Ops[inst.Op].Compiled++
			}
		}
	}

	return out
}

type OpStats struct {
	Interpreted int
	Compiled    int
}

// NativeCompileStats encapsulates statistics about any native
// compilation performed on the VM.
type NativeCompileStats struct {
	Ops               map[byte]*OpStats
	NumCompiledBlocks int
}

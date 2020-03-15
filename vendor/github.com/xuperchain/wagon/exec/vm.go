// Copyright 2017 The go-interpreter Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package exec provides functions for executing WebAssembly bytecode.
package exec

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"reflect"

	"github.com/edsrzf/mmap-go"
	"github.com/xuperchain/wagon/disasm"
	"github.com/xuperchain/wagon/exec/internal/compile"
	"github.com/xuperchain/wagon/wasm"
	ops "github.com/xuperchain/wagon/wasm/operators"
)

var (
	// ErrMultipleLinearMemories is returned by (*VM).NewVM when the module
	// has more then one entries in the linear memory space.
	ErrMultipleLinearMemories = errors.New("exec: more than one linear memories in module")
	// ErrInvalidArgumentCount is returned by (*VM).ExecCode when an invalid
	// number of arguments to the WebAssembly function are passed to it.
	ErrInvalidArgumentCount = errors.New("exec: invalid number of arguments to function")
)

// InvalidReturnTypeError is returned by (*VM).ExecCode when the module
// specifies an invalid return type value for the executed function.
type InvalidReturnTypeError int8

func (e InvalidReturnTypeError) Error() string {
	return fmt.Sprintf("Function has invalid return value_type: %d", int8(e))
}

// InvalidFunctionIndexError is returned by (*VM).ExecCode when the function
// index provided is invalid.
type InvalidFunctionIndexError int64

func (e InvalidFunctionIndexError) Error() string {
	return fmt.Sprintf("Invalid index to function index space: %d", int64(e))
}

type context struct {
	stack   []uint64
	locals  []uint64
	code    []byte
	asm     []asmBlock
	pc      int64
	curFunc int64
}

// VM is the execution context for executing WebAssembly bytecode.
type VM struct {
	ctx context
	cfg config

	module  *wasm.Module
	globals []uint64
	memory  []byte
	funcs   []function

	funcTable [256]func()

	// RecoverPanic controls whether the `ExecCode` method
	// recovers from a panic and returns it as an error
	// instead.
	// A panic can occur either when executing an invalid VM
	// or encountering an invalid instruction, e.g. `unreachable`.
	RecoverPanic bool

	abort bool // Flag for host functions to terminate execution

	nativeBackend *nativeCompiler

	gasMapper disasm.GasMapper
	gasLimit  int64
	UserData  interface{}
	GasUsed   int64
	StaticTop int64
}

// As per the WebAssembly spec: https://github.com/WebAssembly/design/blob/27ac254c854994103c24834a994be16f74f54186/Semantics.md#linear-memory
const wasmPageSize = 65536 // (64 KB)

var endianess = binary.LittleEndian

type config struct {
	EnableAOT   bool
	GasMapper   disasm.GasMapper
	GasLimit    int64
	LazyCompile bool
	CacheStore  FuncCacheStore
}

// VMOption describes a customization that can be applied to the VM.
type VMOption func(c *config)

// EnableAOT enables ahead-of-time compilation of supported opcodes
// into runs of native instructions, if wagon supports native compilation
// for the current architecture.
func EnableAOT(v bool) VMOption {
	return func(c *config) {
		c.EnableAOT = v
	}
}

// WithGasMapper add gas mapper to vm which enable gas statistics
func WithGasMapper(gasMapper disasm.GasMapper) VMOption {
	return func(c *config) {
		c.GasMapper = gasMapper
	}
}

// WithGasLimit add gas limit to vm
func WithGasLimit(limit int64) VMOption {
	return func(c *config) {
		c.GasLimit = limit
	}
}

// WithLazyCompile compile function when needed and cached result.
func WithLazyCompile(v bool) VMOption {
	return func(c *config) {
		c.LazyCompile = v
	}
}

// WithCacheStore set the store to cache compiled function.
func WithCacheStore(s FuncCacheStore) VMOption {
	return func(c *config) {
		c.CacheStore = s
	}
}

// NewVM creates a new VM from a given module and options. If the module defines
// a start function, it will be executed.
func NewVM(module *wasm.Module, opts ...VMOption) (*VM, error) {
	var vm VM
	var options config
	for _, opt := range opts {
		opt(&options)
	}
	// for test purpose
	if !options.EnableAOT && os.Getenv("WAGON_LAZY_COMPILE") == "on" {
		options.LazyCompile = true
	}
	if options.LazyCompile && options.EnableAOT {
		return nil, errors.New("LazyCompile and EnableAOT can't be true at the same time")
	}
	vm.cfg = options
	vm.gasMapper = options.GasMapper
	vm.gasLimit = options.GasLimit

	var err error
	if module.Memory != nil && len(module.Memory.Entries) != 0 {
		if len(module.Memory.Entries) > 1 {
			return nil, ErrMultipleLinearMemories
		}
		size := uint(module.Memory.Entries[0].Limits.Initial) * wasmPageSize
		if size != 0 {
			vm.memory, err = mmap.MapRegion(nil, int(size), mmap.RDWR, mmap.ANON, 0)
			if err != nil {
				return nil, err
			}
		}
		// vm.memory = make([]byte, uint(module.Memory.Entries[0].Limits.Initial)*wasmPageSize)
		err = vm.initMemory(module)
		if err != nil {
			return nil, err
		}
	}

	vm.funcs = make([]function, len(module.FunctionIndexSpace))
	vm.globals = make([]uint64, len(module.GlobalIndexSpace))
	vm.newFuncTable()
	vm.module = module

	nNatives := 0
	for i, fn := range module.FunctionIndexSpace {
		// Skip native methods as they need not be
		// disassembled; simply add them at the end
		// of the `funcs` array as is, as specified
		// in the spec. See the "host functions"
		// section of:
		// https://webassembly.github.io/spec/core/exec/modules.html#allocation
		if fn.IsHost() {
			vm.funcs[i] = goFunction{
				typ: fn.Host.Type(),
				val: fn.Host,
			}
			nNatives++
			continue
		}
		if !options.LazyCompile {
			ifn, err := vm.compileFunction(fn)
			if err != nil {
				return nil, err
			}
			vm.funcs[i] = ifn
		}
	}

	if err := vm.resetGlobals(); err != nil {
		return nil, err
	}

	if module.Start != nil {
		_, err := vm.ExecCode(int64(module.Start.Index))
		if err != nil {
			return nil, err
		}
	}

	if options.EnableAOT {
		supportedBackend, backend := nativeBackend()
		if supportedBackend {
			vm.nativeBackend = backend
			if err := vm.tryNativeCompile(); err != nil {
				return nil, err
			}
		}
	}

	return &vm, nil
}

func (vm *VM) initMemory(m *wasm.Module) error {
	if m.Data == nil || len(m.Data.Entries) == 0 {
		return nil
	}
	// each module can only have a single linear memory in the MVP

	maxoff := uint64(0)
	for _, entry := range m.Data.Entries {
		if entry.Index != 0 {
			return wasm.InvalidLinearMemoryIndexError(entry.Index)
		}
		val, err := m.ExecInitExpr(entry.Offset)
		if err != nil {
			return err
		}
		off, ok := val.(int32)
		if !ok {
			return wasm.InvalidValueTypeInitExprError{
				Wanted: reflect.Int32,
				Got:    reflect.TypeOf(val).Kind(),
			}
		}
		offset := uint32(off)

		memory := vm.memory
		dataEnd := uint64(offset) + uint64(len(entry.Data))
		if dataEnd > uint64(len(memory)) {
			return fmt.Errorf("data entry out of memory, offset:%d", dataEnd)
		} else {
			if maxoff < dataEnd {
				maxoff = dataEnd
			}
			copy(memory[offset:], entry.Data)
		}
	}
	vm.StaticTop = int64(maxoff)
	return nil
}

func (vm *VM) getFunc(idx int) (function, error) {
	ifn := vm.funcs[idx]
	if ifn != nil {
		return ifn, nil
	}

	fn := vm.module.FunctionIndexSpace[idx]
	// native function does not need compile
	if fn.IsHost() {
		return vm.funcs[idx], nil
	}

	// found from cache
	if vm.cfg.CacheStore != nil {
		v, ok := vm.cfg.CacheStore.Get(fn.Body.Hash)
		if ok {
			ifn = v.(function)
			vm.funcs[idx] = ifn
			return ifn, nil
		}
	}

	ifn, err := vm.compileFunction(fn)
	if err != nil {
		return nil, err
	}
	if vm.cfg.CacheStore != nil {
		vm.cfg.CacheStore.Put(fn.Body.Hash, ifn)
	}
	vm.funcs[idx] = ifn
	return ifn, nil
}

func (vm *VM) compileFunction(fn wasm.Function) (function, error) {
	disassembly, err := disasm.NewDisassemblyWithGas(fn, vm.module, vm.gasMapper)
	if err != nil {
		return nil, err
	}

	totalLocalVars := 0
	totalLocalVars += len(fn.Sig.ParamTypes)
	for _, entry := range fn.Body.Locals {
		totalLocalVars += int(entry.Count)
	}
	code, meta := compile.Compile(disassembly.Code)
	compiled := compiledFunction{
		codeMeta:       meta,
		code:           code,
		branchTables:   meta.BranchTables,
		maxDepth:       disassembly.MaxDepth,
		totalLocalVars: totalLocalVars,
		args:           len(fn.Sig.ParamTypes),
		returns:        len(fn.Sig.ReturnTypes) != 0,
	}
	return compiled, nil
}

func (vm *VM) resetGlobals() error {
	for i, global := range vm.module.GlobalIndexSpace {
		val, err := vm.module.ExecInitExpr(global.Init)
		if err != nil {
			return err
		}
		switch v := val.(type) {
		case int32:
			vm.globals[i] = uint64(v)
		case int64:
			vm.globals[i] = uint64(v)
		case float32:
			vm.globals[i] = uint64(math.Float32bits(v))
		case float64:
			vm.globals[i] = uint64(math.Float64bits(v))
		}
	}

	return nil
}

// Memory returns the linear memory space for the VM.
func (vm *VM) Memory() []byte {
	return vm.memory
}

func (vm *VM) pushBool(v bool) {
	if v {
		vm.pushUint64(1)
	} else {
		vm.pushUint64(0)
	}
}

func (vm *VM) fetchBool() bool {
	return vm.fetchInt8() != 0
}

func (vm *VM) fetchInt8() int8 {
	i := int8(vm.ctx.code[vm.ctx.pc])
	vm.ctx.pc++
	return i
}

func (vm *VM) fetchUint32() uint32 {
	v := endianess.Uint32(vm.ctx.code[vm.ctx.pc:])
	vm.ctx.pc += 4
	return v
}

func (vm *VM) fetchInt32() int32 {
	return int32(vm.fetchUint32())
}

func (vm *VM) fetchFloat32() float32 {
	return math.Float32frombits(vm.fetchUint32())
}

func (vm *VM) fetchUint64() uint64 {
	v := endianess.Uint64(vm.ctx.code[vm.ctx.pc:])
	vm.ctx.pc += 8
	return v
}

func (vm *VM) fetchInt64() int64 {
	return int64(vm.fetchUint64())
}

func (vm *VM) fetchFloat64() float64 {
	return math.Float64frombits(vm.fetchUint64())
}

func (vm *VM) popUint64() uint64 {
	i := vm.ctx.stack[len(vm.ctx.stack)-1]
	vm.ctx.stack = vm.ctx.stack[:len(vm.ctx.stack)-1]
	return i
}

func (vm *VM) popInt64() int64 {
	return int64(vm.popUint64())
}

func (vm *VM) popFloat64() float64 {
	return math.Float64frombits(vm.popUint64())
}

func (vm *VM) popUint32() uint32 {
	return uint32(vm.popUint64())
}

func (vm *VM) popInt32() int32 {
	return int32(vm.popUint32())
}

func (vm *VM) popFloat32() float32 {
	return math.Float32frombits(vm.popUint32())
}

func (vm *VM) pushUint64(i uint64) {
	if debugStackDepth {
		if len(vm.ctx.stack) >= cap(vm.ctx.stack) {
			panic("stack exceeding max depth: " + fmt.Sprintf("len=%d,cap=%d", len(vm.ctx.stack), cap(vm.ctx.stack)))
		}
	}
	vm.ctx.stack = append(vm.ctx.stack, i)
}

func (vm *VM) pushInt64(i int64) {
	vm.pushUint64(uint64(i))
}

func (vm *VM) pushFloat64(f float64) {
	vm.pushUint64(math.Float64bits(f))
}

func (vm *VM) pushUint32(i uint32) {
	vm.pushUint64(uint64(i))
}

func (vm *VM) pushInt32(i int32) {
	vm.pushUint64(uint64(i))
}

func (vm *VM) pushFloat32(f float32) {
	vm.pushUint32(math.Float32bits(f))
}

// ExecCode calls the function with the given index and arguments.
// fnIndex should be a valid index into the function index space of
// the VM's module.
func (vm *VM) ExecCode(fnIndex int64, args ...uint64) (rtrn interface{}, err error) {
	// If used as a library, client code should set vm.RecoverPanic to true
	// in order to have an error returned.
	if vm.RecoverPanic {
		defer func() {
			if r := recover(); r != nil {
				switch e := r.(type) {
				case error:
					err = e
				default:
					err = fmt.Errorf("exec: %v", e)
				}
			}
		}()
	}
	if int(fnIndex) > len(vm.funcs) {
		return nil, InvalidFunctionIndexError(fnIndex)
	}
	if len(vm.module.GetFunction(int(fnIndex)).Sig.ParamTypes) != len(args) {
		return nil, ErrInvalidArgumentCount
	}
	fn, err := vm.getFunc(int(fnIndex))
	if err != nil {
		return nil, err
	}
	compiled, ok := fn.(compiledFunction)
	if !ok {
		panic(fmt.Sprintf("exec: function at index %d is not a compiled function", fnIndex))
	}
	savedCtx := vm.ctx
	vm.ctx = context{}
	defer func() {
		vm.ctx = savedCtx
	}()

	depth := compiled.maxDepth + 1
	if cap(vm.ctx.stack) < depth {
		vm.ctx.stack = make([]uint64, 0, depth)
	} else {
		vm.ctx.stack = vm.ctx.stack[:0]
	}

	vm.ctx.locals = make([]uint64, compiled.totalLocalVars)
	vm.ctx.pc = 0
	vm.ctx.code = compiled.code
	vm.ctx.asm = compiled.asm
	vm.ctx.curFunc = fnIndex

	for i, arg := range args {
		vm.ctx.locals[i] = arg
	}

	res := vm.execCode(compiled)
	if compiled.returns {
		rtrnType := vm.module.GetFunction(int(fnIndex)).Sig.ReturnTypes[0]
		switch rtrnType {
		case wasm.ValueTypeI32:
			rtrn = uint32(res)
		case wasm.ValueTypeI64:
			rtrn = uint64(res)
		case wasm.ValueTypeF32:
			rtrn = math.Float32frombits(uint32(res))
		case wasm.ValueTypeF64:
			rtrn = math.Float64frombits(res)
		default:
			return nil, InvalidReturnTypeError(rtrnType)
		}
	}
	return rtrn, nil
}

func (vm *VM) checkGas() {
	used := vm.fetchInt64()
	if vm.GasUsed+used > vm.gasLimit {
		// FIXME
		panic("out of gas")
	}
	vm.GasUsed += used
}

func (vm *VM) execCode(compiled compiledFunction) uint64 {
outer:
	for int(vm.ctx.pc) < len(vm.ctx.code) && !vm.abort {
		op := vm.ctx.code[vm.ctx.pc]
		vm.ctx.pc++
		switch op {
		case ops.Return:
			break outer
		case compile.OpJmp:
			vm.ctx.pc = vm.fetchInt64()
			continue
		case compile.OpJmpZ:
			target := vm.fetchInt64()
			if vm.popUint32() == 0 {
				vm.ctx.pc = target
				continue
			}
		case compile.OpJmpNz:
			target := vm.fetchInt64()
			preserveTop := vm.fetchBool()
			discard := vm.fetchInt64()
			if vm.popUint32() != 0 {
				vm.ctx.pc = target
				var top uint64
				if preserveTop {
					top = vm.ctx.stack[len(vm.ctx.stack)-1]
				}
				vm.ctx.stack = vm.ctx.stack[:len(vm.ctx.stack)-int(discard)]
				if preserveTop {
					vm.pushUint64(top)
				}
				continue
			}
		case ops.BrTable:
			index := vm.fetchInt64()
			label := vm.popInt32()
			fn, err := vm.getFunc(int(vm.ctx.curFunc))
			if err != nil {
				panic(err)
			}
			cf, ok := fn.(compiledFunction)
			if !ok {
				panic(fmt.Sprintf("exec: function at index %d is not a compiled function", vm.ctx.curFunc))
			}
			table := cf.branchTables[index]
			var target compile.Target
			if label >= 0 && label < int32(len(table.Targets)) {
				target = table.Targets[int32(label)]
			} else {
				target = table.DefaultTarget
			}

			if target.Return {
				break outer
			}
			vm.ctx.pc = target.Addr
			var top uint64
			if target.PreserveTop {
				top = vm.ctx.stack[len(vm.ctx.stack)-1]
			}
			vm.ctx.stack = vm.ctx.stack[:len(vm.ctx.stack)-int(target.Discard)]
			if target.PreserveTop {
				vm.pushUint64(top)
			}
			continue
		case compile.OpDiscard:
			place := vm.fetchInt64()
			vm.ctx.stack = vm.ctx.stack[:len(vm.ctx.stack)-int(place)]
		case compile.OpDiscardPreserveTop:
			top := vm.ctx.stack[len(vm.ctx.stack)-1]
			place := vm.fetchInt64()
			vm.ctx.stack = vm.ctx.stack[:len(vm.ctx.stack)-int(place)]
			vm.pushUint64(top)

		case ops.WagonNativeExec:
			i := vm.fetchUint32()
			vm.nativeCodeInvocation(i)
		default:
			vm.funcTable[op]()
		}
	}

	if compiled.returns && !vm.abort {
		return vm.ctx.stack[len(vm.ctx.stack)-1]
	}
	return 0
}

// Restart readies the VM for another run.
func (vm *VM) Restart() {
	vm.resetGlobals()
	vm.ctx.locals = make([]uint64, 0)
	vm.abort = false
}

func (vm *VM) releaseMemory() {
	if vm.memory != nil {
		mem := mmap.MMap(vm.memory)
		mem.Unmap()
	}
}

// Close frees any resources managed by the VM.
func (vm *VM) Close() error {
	vm.releaseMemory()
	vm.abort = true // prevents further use.
	if vm.nativeBackend != nil {
		if err := vm.nativeBackend.Close(); err != nil {
			return err
		}
	}
	return nil
}

// Process is a proxy passed to host functions in order to access
// things such as memory and control.
type Process struct {
	vm *VM
}

// NewProcess creates a VM interface object for host functions
func NewProcess(vm *VM) *Process {
	return &Process{vm: vm}
}

// ReadAt implements the ReaderAt interface: it copies into p
// the content of memory at offset off.
func (proc *Process) ReadAt(p []byte, off int64) (int, error) {
	mem := proc.vm.Memory()

	var length int
	if len(mem) < len(p)+int(off) {
		length = len(mem) - int(off)
	} else {
		length = len(p)
	}

	copy(p, mem[off:off+int64(length)])

	var err error
	if length < len(p) {
		err = io.ErrShortBuffer
	}

	return length, err
}

// WriteAt implements the WriterAt interface: it writes the content of p
// into the VM memory at offset off.
func (proc *Process) WriteAt(p []byte, off int64) (int, error) {
	mem := proc.vm.Memory()

	var length int
	if len(mem) < len(p)+int(off) {
		length = len(mem) - int(off)
	} else {
		length = len(p)
	}

	copy(mem[off:], p[:length])

	var err error
	if length < len(p) {
		err = io.ErrShortWrite
	}

	return length, err
}

// MemSize returns the current allocated memory size in bytes.
func (proc *Process) MemSize() int {
	return len(proc.vm.Memory())
}

// Terminate stops the execution of the current module.
func (proc *Process) Terminate() {
	proc.vm.abort = true
}

func (proc *Process) VM() *VM {
	return proc.vm
}

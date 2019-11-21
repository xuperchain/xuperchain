package exec

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"reflect"

	"github.com/xuperchain/wagon/exec"
	"github.com/xuperchain/wagon/wasm"
	"github.com/xuperchain/wagon/wasm/leb128"
)

var funcTypes = []interface{}{
	(func(*exec.Process) uint32)(nil),
	(func(*exec.Process, uint32) uint32)(nil),
	(func(*exec.Process, uint32, uint32) uint32)(nil),
	(func(*exec.Process, uint32, uint32, uint32) uint32)(nil),
	(func(*exec.Process, uint32, uint32, uint32, uint32) uint32)(nil),
	(func(*exec.Process, uint32, uint32, uint32, uint32, uint32) uint32)(nil),
	(func(*exec.Process, uint32, uint32, uint32, uint32, uint32, uint32) uint32)(nil),
	(func(*exec.Process, uint32, uint32, uint32, uint32, uint32, uint32, uint32) uint32)(nil),
}

func makeExportFunc(sig wasm.FunctionSig, fun interface{}) (*wasm.Function, error) {
	paramLen := len(sig.ParamTypes)
	if paramLen >= len(funcTypes) {
		return nil, errors.New("bad function type")
	}
	ftype := reflect.TypeOf(funcTypes[paramLen])
	body := reflect.MakeFunc(ftype, func(args []reflect.Value) []reflect.Value {
		proc := args[0].Interface().(*exec.Process)
		ctx := proc.VM().UserData.(*wagonContext)
		params := make([]uint32, len(args)-1)
		for i := 1; i < len(args); i++ {
			params[i-1] = uint32(args[i].Uint())
		}
		ret, _ := applyFuncCall(ctx, fun, params)
		return []reflect.Value{reflect.ValueOf(ret)}
	})

	return &wasm.Function{
		Sig:  &sig,
		Host: body,
		Body: new(wasm.FunctionBody),
	}, nil
}

func makeExportGlobal(sig *wasm.GlobalVar, v int64) (*wasm.GlobalEntry, error) {
	buf := new(bytes.Buffer)
	switch sig.Type {
	case wasm.ValueTypeI32:
		buf.WriteByte(0x41)
		leb128.WriteVarUint32(buf, uint32(v))
	case wasm.ValueTypeI64:
		buf.WriteByte(0x42)
		leb128.WriteVarint64(buf, v)
	case wasm.ValueTypeF32:
		buf.WriteByte(0x43)
		binary.Write(buf, binary.LittleEndian, uint32(v))
	case wasm.ValueTypeF64:
		buf.WriteByte(0x44)
		binary.Write(buf, binary.LittleEndian, uint64(v))
	}

	return &wasm.GlobalEntry{
		Type: *sig,
		Init: buf.Bytes(),
	}, nil
}

func makeWagonModule(resolver Resolver) wasm.ResolveModuleFunc {
	return func(module string, main *wasm.Module) (*wasm.Module, error) {
		export := wasm.NewModule()
		export.Export.Entries = map[string]wasm.ExportEntry{}
		for _, importEntry := range main.Import.Entries {
			field := importEntry.FieldName

			switch importEntry.Type.Kind() {
			case wasm.ExternalFunction:
				ifunc, ok := resolver.ResolveFunc(module, field)
				if !ok {
					return nil, fmt.Errorf("%s.%s not found", module, field)
				}

				index := importEntry.Type.(wasm.FuncImport).Type
				sig := main.Types.Entries[index]
				fun, err := makeExportFunc(sig, ifunc)
				if err != nil {
					return nil, err
				}
				export.Types.Entries = append(export.Types.Entries, sig)
				export.FunctionIndexSpace = append(export.FunctionIndexSpace, *fun)
				export.Export.Entries[field] = wasm.ExportEntry{
					FieldStr: field,
					Kind:     wasm.ExternalFunction,
					Index:    uint32(len(export.FunctionIndexSpace) - 1),
				}
			case wasm.ExternalGlobal:
				v, ok := resolver.ResolveGlobal(module, field)
				if !ok {
					return nil, fmt.Errorf("%s.%s not found", module, field)
				}
				sig := importEntry.Type.(wasm.GlobalVarImport).Type
				global, err := makeExportGlobal(&sig, v)
				if err != nil {
					return nil, err
				}
				export.GlobalIndexSpace = append(export.GlobalIndexSpace, *global)
				export.Export.Entries[field] = wasm.ExportEntry{
					FieldStr: field,
					Kind:     wasm.ExternalGlobal,
					Index:    uint32(len(export.GlobalIndexSpace) - 1),
				}

			case wasm.ExternalTable:
				export.TableIndexSpace = [][]uint32{nil}
				export.Export.Entries[field] = wasm.ExportEntry{
					FieldStr: field,
					Kind:     wasm.ExternalTable,
					Index:    0,
				}
			case wasm.ExternalMemory:
				export.LinearMemoryIndexSpace = [][]byte{nil}
				export.Export.Entries[field] = wasm.ExportEntry{
					FieldStr: field,
					Kind:     wasm.ExternalMemory,
					Index:    0,
				}
			}
		}
		return export, nil
	}
}

// InterpCode is the Code interface of interpreter mode
type InterpCode struct {
	module *wasm.Module
}

// NewInterpCode instance a Code based on the wasm code and resolver
func NewInterpCode(wasmCode []byte, resolver Resolver) (code *InterpCode, err error) {
	defer CaptureTrap(&err)
	importModuleFunc := makeWagonModule(resolver)
	module, err := wasm.LoadModule(bytes.NewBuffer(wasmCode), importModuleFunc)
	if err != nil {
		return nil, err
	}
	code = &InterpCode{
		module: module,
	}
	return
}

// NewContext instances a new context
func (code *InterpCode) NewContext(cfg *ContextConfig) (ictx Context, err error) {
	defer CaptureTrap(&err)
	vm, err := exec.NewVM(code.module,
		exec.WithLazyCompile(true),
		exec.WithGasMapper(new(GasMapper)),
		exec.WithGasLimit(cfg.GasLimit))
	if err != nil {
		return nil, err
	}
	vm.RecoverPanic = true
	ctx := &wagonContext{
		module:   code.module,
		vm:       vm,
		userData: make(map[string]interface{}),
	}
	vm.UserData = ctx
	ictx = ctx
	return
}

// Release releases the resources
func (code *InterpCode) Release() {
}

type wagonContext struct {
	module   *wasm.Module
	vm       *exec.VM
	userData map[string]interface{}
}

func (c *wagonContext) Exec(name string, param []int64) (ret int64, err error) {
	defer CaptureTrap(&err)

	entry, ok := c.module.Export.Entries[name]
	if !ok {
		return 0, &ErrFuncNotFound{Name: name}
	}
	idx := entry.Index
	args := make([]uint64, len(param))
	for i, v := range param {
		args[i] = uint64(v)
	}
	iret, err := c.vm.ExecCode(int64(idx), args...)
	if err != nil {
		return 0, err
	}
	if iret == nil {
		return 0, nil
	}
	switch v := iret.(type) {
	case uint32:
		return int64(v), nil
	case uint64:
		return int64(v), nil
	case float32:
		return int64(math.Float32bits(v)), nil
	case float64:
		return int64(math.Float64bits(v)), nil
	default:
		return 0, fmt.Errorf("bad type: %v:%T", iret, iret)
	}
}

func (c *wagonContext) GasUsed() int64 {
	return c.vm.GasUsed
}

func (c *wagonContext) ResetGasUsed() {
	c.vm.GasUsed = 0
}

func (c *wagonContext) Memory() []byte {
	return c.vm.Memory()
}

func (c *wagonContext) StaticTop() uint32 {
	return uint32(c.vm.StaticTop)
}

func (c *wagonContext) Release() {
	c.vm.Close()
}

// SetUserData store key-value pair to Context which can be retrieved by GetUserData
func (c *wagonContext) SetUserData(key string, value interface{}) {
	c.userData[key] = value
}

// GetUserData retrieves user data stored by SetUserData
func (c *wagonContext) GetUserData(key string) interface{} {
	return c.userData[key]
}

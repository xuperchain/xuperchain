package gowasm

import (
	"reflect"

	"github.com/xuperchain/xuperunion/xvm/exec"
)

type resolver struct {
}

// NewResolver returns exec.Resolver which resolvers symbols needed by go runtime
func NewResolver() exec.Resolver {
	return &resolver{}
}
func (r *resolver) resolveFunc(module, name string) (interface{}, bool) {
	fullname := module + "::" + name
	switch fullname {
	case "go::debug":
		return (*Runtime).debug, true
	case "go::runtime.wasmExit":
		return (*Runtime).wasmExit, true
	case "go::runtime.wasmWrite":
		return (*Runtime).wasmWrite, true
	case "go::runtime.nanotime":
		return (*Runtime).nanotime, true
	case "go::runtime.walltime":
		return (*Runtime).walltime, true
	case "go::runtime.getRandomData":
		return (*Runtime).getRandomData, true
	case "go::runtime.scheduleCallback": // for go.11
		return (*Runtime).scheduleCallback, true
	case "go::runtime.clearScheduledCallback": // for go.11
		return (*Runtime).clearScheduledCallback, true
	case "go::runtime.scheduleTimeoutEvent":
		return (*Runtime).scheduleCallback, true
	case "go::runtime.clearTimeoutEvent":
		return (*Runtime).clearScheduledCallback, true
	case "go::syscall/js.valueGet":
		return (*Runtime).syscallJsValueGet, true
	case "go::syscall/js.valueSet":
		return (*Runtime).syscallJsValueSet, true
	case "go::syscall/js.valueNew":
		return (*Runtime).syscallJsValueNew, true
	case "go::syscall/js.valuePrepareString":
		return (*Runtime).syscallJsValuePrepareString, true
	case "go::syscall/js.valueCall":
		return (*Runtime).syscallJsValueCall, true
	case "go::syscall/js.valueInvoke":
		return (*Runtime).syscallJsValueInvoke, true
	case "go::syscall/js.stringVal":
		return (*Runtime).syscallJsStringVal, true
	case "go::syscall/js.valueLoadString":
		return (*Runtime).syscallJsValueLoadString, true
	case "go::syscall/js.valueLength":
		return (*Runtime).syscallJsValueLength, true
	case "go::syscall/js.valueIndex":
		return (*Runtime).syscallJsValueIndex, true
	case "go::syscall/js.valueSetIndex":
		return (*Runtime).syscallJsValueSetIndex, true
	case "go::syscall/js.valueInstanceOf":
		return (*Runtime).syscallJsValueInstanceOf, true
	}
	return nil, false
}

func (r *resolver) ResolveFunc(module, name string) (interface{}, bool) {
	ifunc, ok := r.resolveFunc(module, name)
	if !ok {
		return nil, false
	}
	Type, Value := reflect.TypeOf(ifunc), reflect.ValueOf(ifunc)
	realFunc := func(ctx *exec.Context, sp uint32) uint32 {
		rt := ctx.GetUserData(goRuntimeKey).(*Runtime)
		mem := ctx.Memory()
		dec := NewDecoder(mem, sp+8)
		args := []reflect.Value{reflect.ValueOf(rt)}
		for i := 1; i < Type.NumIn(); i++ {
			argtype := Type.In(i)
			ref := reflect.New(argtype)
			dec.Decode(ref)
			args = append(args, ref.Elem())
		}
		rets := Value.Call(args)
		enc := NewEncoder(mem, dec.Offset())
		for i := 0; i < len(rets); i++ {
			ret := rets[i]
			enc.Encode(ret)
		}

		return 0
	}
	return realFunc, true
}

func (r *resolver) ResolveGlobal(module, name string) (float64, bool) {
	return 0, false
}

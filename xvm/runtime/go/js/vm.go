package js

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"unsafe"
)

// A PropertyGetter can get property from GetProperty method
type PropertyGetter interface {
	GetProperty(property string) (interface{}, bool)
}

// VM  simulates the js runtime
type VM struct {
	cfg     *VMConfig
	valueid Ref
	values  map[Ref]*Value
	Log     *log.Logger
	//refs    map[reflect.Value]Ref
}

// VMConfig is the config of VM object
type VMConfig struct {
	// the wasm Memory
	Memory *Memory

	Global *Global
}

// NewVM instance a VM object
func NewVM(config *VMConfig) *VM {
	vm := &VM{
		cfg:     config,
		valueid: ValueGo + 1,
		values:  make(map[Ref]*Value),
	}
	RegisterBuiltins(vm.cfg.Global)
	vm.initDefaultValue()
	return vm
}

func (vm *VM) initDefaultValue() {
	vm.values[ValueNaN] = valueNaN
	vm.values[ValueZero] = valueZero
	vm.values[ValueNull] = valueNull
	vm.values[ValueTrue] = valueTrue
	vm.values[ValueFalse] = valueFalse
	vm.values[ValueMemory] = &Value{
		name:  "Memory",
		ref:   ValueMemory,
		value: vm.cfg.Memory,
	}

	goruntime := &Value{
		name: "Go",
		ref:  ValueGo,
		value: map[string]interface{}{
			"_makeFuncWrapper": func([]interface{}) interface{} {
				return nil
			},
		},
	}
	vm.values[ValueGo] = goruntime
	vm.cfg.Global.Register("Go", goruntime)

	vm.values[ValueGlobal] = &Value{
		name:  "Global",
		ref:   ValueGlobal,
		value: vm.cfg.Global,
	}
}

const (
	tagString = 1
	tagSymbol = 2
	tagFunc   = 3
	tagObject = 4
)

func floatValue(f float64) Ref {
	if f != f {
		return ValueNaN
	}
	if f == 0 {
		return ValueZero
	}
	return *(*Ref)(unsafe.Pointer(&f))
}

func (vm *VM) storeValue(name string, x interface{}) Ref {
	if x == nil {
		return ValueNull
	}
	switch xx := x.(type) {
	case int8, int16, int32, int64, int:
		return floatValue(float64(reflect.ValueOf(x).Int()))
	case uint8, uint16, uint32, uint64, uint:
		return floatValue(float64(reflect.ValueOf(x).Uint()))
	case float32, float64:
		return floatValue(reflect.ValueOf(x).Float())
	case bool:
		if xx {
			return ValueTrue
		}
		return ValueFalse
	}

	var tag int64
	v := reflect.ValueOf(x)
	t := v.Type()
	switch t.Kind() {
	case reflect.String:
		tag = tagString
	case reflect.Func:
		tag = tagFunc
	default:
		tag = tagObject
	}
	vm.valueid++
	ref := vm.valueid | Ref(tag<<32)
	vm.values[ref] = &Value{
		name:  name,
		value: x,
		ref:   ref,
	}
	return ref
}

func (vm *VM) loadValue(ref Ref) (*Value, bool) {
	if int64(ref) == 0 {
		return valueUndefined, true
	}
	n, ok := ref.Number()
	if ok {
		return &Value{
			name:  "number",
			value: n,
			ref:   ref,
		}, true
	}
	v, ok := vm.values[ref]
	return v, ok
}

// Property return the property of a js object
// if not found, undefined will be returned
func (vm *VM) Property(ref Ref, name string) Ref {
	if ref == ValueUndefined {
		Throw("get property %s on undefined", name)
	}
	parent, ok := vm.values[ref]
	if !ok {
		ThrowException(ExceptionRefNotFound(ref))
	}
	v, ok := vm.property(parent.value, name)
	if !ok {
		Throw("property %s not found on %s", name, parent.Name())
	}
	if value, ok := v.(*Value); ok {
		return value.ref
	}
	fullname := fmt.Sprintf("%s.%s", parent.name, name)
	return vm.storeValue(fullname, v)
}

func (vm *VM) property(object interface{}, name string) (interface{}, bool) {
	p := reflect.ValueOf(object)
	name = strings.Title(name)
	// Map
	if p.Kind() == reflect.Map {
		g := p.MapIndex(reflect.ValueOf(name))
		if g.IsValid() {
			return g.Interface(), true
		}
		return nil, false
	}

	if (p.Kind() == reflect.Ptr && p.Elem().Kind() == reflect.Struct) || p.Kind() == reflect.Struct {
		// Method
		prop := p.MethodByName(name)
		if prop.IsValid() {
			return prop.Interface(), true
		}

		var pp reflect.Value
		// FieldByName must not be a ptr
		if p.Kind() == reflect.Ptr {
			pp = p.Elem()
		} else {
			pp = p
		}
		prop = pp.FieldByName(name)
		if prop.IsValid() {
			return prop.Interface(), true
		}
	}

	// Getter interface
	if g, ok := p.Interface().(PropertyGetter); ok {
		prop, ok := g.GetProperty(name)
		return prop, ok
	}

	return nil, false
}

// Exception wraps an *Exception to Ref
func (vm *VM) Exception(e *Exception) Ref {
	return vm.storeValue("Exception", e)
}

func (vm *VM) call(name string, ifunc interface{}, args []Ref) Ref {
	f, ok := ifunc.(func([]interface{}) interface{})
	if !ok {
		Throw("%s is not a js function", name)
	}
	ret := f(vm.parseArgs(args))
	return vm.storeValue(name, ret)
}

// New treat ref as the construct function, args as arguments to instance a js object
func (vm *VM) New(ref Ref, args []Ref) Ref {
	if ref == ValueUndefined {
		Throw("call new on undefined")
	}
	v, ok := vm.loadValue(ref)
	if !ok {
		ThrowException(ExceptionRefNotFound(ref))
	}
	return vm.call(v.name, v.value, args)
}

// Call call ref's method using method as method name
func (vm *VM) Call(ref Ref, method string, args []Ref) Ref {
	if ref == ValueUndefined {
		Throw("call method %s on undefined", method)
	}
	v, ok := vm.loadValue(ref)
	if !ok {
		ThrowException(ExceptionRefNotFound(ref))
	}
	name := fmt.Sprintf("%s.%s", v.name, method)
	method = strings.Title(method)
	prop, ok := vm.property(v.value, method)
	if !ok {
		Throw("property %s on %s not found", method, v.Name())
	}
	return vm.call(name, prop, args)
}

// Invoke call ref as a js function
func (vm *VM) Invoke(ref Ref, args []Ref) Ref {
	if ref == ValueUndefined {
		Throw("call invoke on undefined")
	}
	v, ok := vm.loadValue(ref)
	if !ok {
		ThrowException(ExceptionRefNotFound(ref))
	}
	return vm.call(v.name, v.value, args)
}

// Store wrap a Go object to Ref
func (vm *VM) Store(x interface{}) Ref {
	return vm.storeValue("store", x)
}

// Value return the stored value from ref
func (vm *VM) Value(ref Ref) *Value {
	v, ok := vm.loadValue(ref)
	if !ok {
		return nil
	}
	return v
}

func (vm *VM) parseArgs(args []Ref) []interface{} {
	var ret []interface{}
	for _, arg := range args {
		v, ok := vm.loadValue(arg)
		if !ok {
			Throw("bad ref:%s", arg)
		}
		ret = append(ret, v.value)
	}
	return ret
}

// DebugStr return the debug string of ref
func (vm *VM) DebugStr(ref Ref) string {
	v, ok := vm.loadValue(ref)
	if !ok {
		return "<undefined>"
	}
	return fmt.Sprintf("<%s,%s,%T>", v.name, v.ref, v.value)
}

// CatchException will recover panic and store the value to e only if the recovered type is *Exception
// otherwise panic will go on
func (vm *VM) CatchException(e *Ref) {
	ret := recover()
	if ret == nil {
		return
	}
	exp, ok := ret.(*Exception)
	if !ok {
		panic(ret)
	}
	*e = vm.Exception(exp)
}

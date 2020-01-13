package gowasm

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"unsafe"
)

// Decoder decode Go types from wasm memory
type Decoder struct {
	mem []byte
	buf *bytes.Buffer
}

// NewDecoder instances a new Decoder from given memory and offset
func NewDecoder(mem []byte, offset uint32) *Decoder {
	return &Decoder{
		mem: mem,
		buf: bytes.NewBuffer(mem[offset:]),
	}
}

func (r *Decoder) readSlice(v reflect.Value, t reflect.Type) {
	var ptr, length, capacity int64
	binary.Read(r.buf, binary.LittleEndian, &ptr)
	binary.Read(r.buf, binary.LittleEndian, &length)
	binary.Read(r.buf, binary.LittleEndian, &capacity)
	if t.Elem().Kind() == reflect.Uint8 {
		v.SetBytes(r.mem[ptr : ptr+length])
		return
	}
	s := (*reflect.SliceHeader)(unsafe.Pointer(v.Addr().Pointer()))
	s.Data = uintptr(unsafe.Pointer(&r.mem[ptr]))
	s.Len = int(length)
	s.Cap = int(capacity)
	// v.Set(reflect.MakeSlice(t, int(length), int(capacity)))
	// buf := bytes.NewBuffer(r.mem[ptr:])
	// for i := 0; i < int(length); i++ {
	// 	elem := v.Index(i)
	// 	binary.Read(buf, binary.LittleEndian, elem.Addr().Interface())
	// }
}

func (r *Decoder) readString() string {
	var ptr, length int64
	binary.Read(r.buf, binary.LittleEndian, &ptr)
	binary.Read(r.buf, binary.LittleEndian, &length)
	return string(r.mem[ptr : ptr+length])
}

// Decode decode go type from memory, ref must be a pointer type
func (r *Decoder) Decode(ref reflect.Value) {
	elem := ref.Elem()
	tp := elem.Type()
	switch tp.Kind() {
	case reflect.String:
		elem.SetString(r.readString())
	case reflect.Slice:
		r.readSlice(elem, tp)
	case reflect.Int32, reflect.Int64, reflect.Float64:
		binary.Read(r.buf, binary.LittleEndian, ref.Interface())
	default:
		panic("bad arg type:" + tp.String())
	}
}

// Offset returns total decoded memory length
func (r *Decoder) Offset() uint32 {
	return uint32(len(r.mem) - r.buf.Len())
}

// Encoder encodes go type to wasm memory
type Encoder struct {
	buf *bytes.Buffer
}

// NewEncoder instances a new Encoder
func NewEncoder(mem []byte, offset uint32) *Encoder {
	return &Encoder{
		buf: bytes.NewBuffer(mem[offset:offset]),
	}
}

// Encode encode go type to wasm memory
func (e *Encoder) Encode(v reflect.Value) {
	t := v.Type()
	switch t.Kind() {
	case reflect.Bool:
		vv := uint8(0)
		if v.Bool() {
			vv = 1
		}
		binary.Write(e.buf, binary.LittleEndian, vv)
	case reflect.Int32, reflect.Int64, reflect.Float64:
		binary.Write(e.buf, binary.LittleEndian, v.Interface())
	default:
		panic("bad return type:" + t.String())
	}
}

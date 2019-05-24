package exec

import (
	"encoding/binary"
	"fmt"
)

var (
	trapNilMemory = NewTrap("code has no memory")
)

// TrapInvalidAddress is the trap raised when encounter an invalid address
type TrapInvalidAddress uint32

// Reason implements Trap interface
func (t TrapInvalidAddress) Reason() string {
	return fmt.Sprintf("invalid address:0x%x", uint32(t))
}

// Codec helps encoding and decoding data between wasm code and go code
type Codec struct {
	mem []byte
}

// NewCodec instances a Codec, if memory of ctx is nil, trapNilMemory will be raised
func NewCodec(ctx *Context) Codec {
	mem := ctx.Memory()
	if mem == nil {
		Throw(trapNilMemory)
	}

	return Codec{
		mem: mem,
	}
}

// Bytes returns memory region starting from addr, limiting by length
func (c Codec) Bytes(addr, length uint32) []byte {
	if addr+length >= uint32(len(c.mem)) {
		Throw(TrapInvalidAddress(addr + length))
	}
	return c.mem[addr : addr+length]
}

// Uint32 decodes memory[addr:addr+4] to uint32
func (c Codec) Uint32(addr uint32) uint32 {
	buf := c.Bytes(addr, 4)
	return binary.LittleEndian.Uint32(buf)
}

// Uint64 decodes memory[addr:addr+8] to uint64
func (c Codec) Uint64(addr uint32) uint64 {
	buf := c.Bytes(addr, 8)
	return binary.LittleEndian.Uint64(buf)
}

// GoBytes decodes Go []byte start from sp
func (c Codec) GoBytes(sp uint32) []byte {
	addr := c.Uint64(sp)
	length := c.Uint64(sp + 8)
	return c.Bytes(uint32(addr), uint32(length))
}

// GoString decodes Go string start from sp
func (c Codec) GoString(sp uint32) string {
	return string(c.GoBytes(sp))
}

// String decodes memory[addr:addr+length] to string
func (c Codec) String(addr, length uint32) string {
	return string(c.Bytes(addr, length))
}

// CString decodes a '\x00' terminated c style string
func (c Codec) CString(addr uint32) string {
	if addr == 0 {
		Throw(TrapInvalidAddress(addr))
	}
	mem := c.mem
	var i = int(addr)
	for ; i < len(mem) && mem[i] != '\x00'; i++ {
	}
	return string(mem[addr:i])
}

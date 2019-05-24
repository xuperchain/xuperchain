package js

// Array simulates Array function
func Array(args []interface{}) interface{} {
	return args
}

// Uint8Array simulates Uint8Array function
func Uint8Array(args []interface{}) interface{} {
	mem, ok := args[0].([]byte)
	if !ok {
		ThrowException(ExceptionInvalidArgument)
	}
	offset, ok := args[1].(int64)
	if !ok {
		ThrowException(ExceptionInvalidArgument)
	}
	length, ok := args[2].(int64)
	if !ok {
		ThrowException(ExceptionInvalidArgument)
	}
	if offset >= int64(len(mem)) || offset+length > int64(len(mem)) {
		ThrowException(ExceptionInvalidArgument)
	}

	return mem[offset : offset+length]
}

// Memory simulates the Memory object in wasm_exec.js
type Memory struct {
	memfunc func() []byte
}

// NewMemory instance a Memory
func NewMemory(f func() []byte) *Memory {
	return &Memory{
		memfunc: f,
	}
}

// GetProperty implements the PropertyGetter interface
func (m *Memory) GetProperty(name string) (interface{}, bool) {
	switch name {
	case "Buffer":
		return m.memfunc(), true
	default:
		return nil, false
	}
}

// RegisterBuiltins register js builtins to Global object
func RegisterBuiltins(g *Global) {
	g.Register("Array", Array)
	g.Register("Uint8Array", Uint8Array)
	g.Register("Date", Date)
}

package debug

import (
	"io"

	"github.com/xuperchain/xuperunion/xvm/exec"
)

const (
	debugWriterKey = "debugWriter"
)

// SetWriter set debug writer to Context
func SetWriter(ctx *exec.Context, w io.Writer) {
	ctx.SetUserData(debugWriterKey, w)
}

// GetDebugWriter get debug writer
func GetWriter(ctx *exec.Context) io.Writer {
	value := ctx.GetUserData(debugWriterKey)
	if value == nil {
		return nil
	}
	w, ok := value.(io.Writer)
	if !ok {
		return nil
	}
	return w
}

// Write write debug message
// if SetWriter is not set, message will be ignored
func Write(ctx *exec.Context, p []byte) {
	w := GetWriter(ctx)
	if w == nil {
		return
	}
	w.Write(p)
}

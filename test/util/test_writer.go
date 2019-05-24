package util

import (
	"bytes"
	"io"
	"testing"
)

// TestLogWriter 用于生成一个输出到*testing.T的io.Writer供logger使用
type TestLogWriter struct {
	buf bytes.Buffer
	t   *testing.T
}

// NewTestLogWriter create TestLogWriter instance
func NewTestLogWriter(t *testing.T) io.Writer {
	return &TestLogWriter{
		t: t,
	}
}

// Write write data to log
func (t *TestLogWriter) Write(p []byte) (int, error) {
	idx := bytes.IndexByte(p, '\n')
	if idx == -1 {
		return t.buf.Write(p)
	}
	t.buf.Write(p[:idx])
	t.t.Log(t.buf.String())
	t.buf.Reset()
	t.buf.Write(p[idx+1:])
	return len(p), nil
}

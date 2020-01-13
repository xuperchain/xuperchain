package xvm

import (
	"bytes"
	"io"

	log15 "github.com/xuperchain/log15"
)

// debugWriter implements a io.Writer which writes messages as lines to log.Logger
type debugWriter struct {
	buf    bytes.Buffer
	logger log15.Logger
}

func newDebugWriter(logger log15.Logger) io.Writer {
	return &debugWriter{
		logger: logger,
	}
}

func (w *debugWriter) Write(p []byte) (int, error) {
	idx := bytes.IndexByte(p, '\n')
	if idx == -1 {
		w.write(p)
		return len(p), nil
	}
	w.write(p[:idx])
	w.flush()
	w.write(p[idx+1:])

	return len(p), nil
}

func (w *debugWriter) write(p []byte) {
	w.buf.Write(p)
	if w.buf.Len() >= 1024 {
		w.flush()
	}
}

func (w *debugWriter) flush() {
	w.logger.Debug(w.buf.String())
	w.buf.Reset()
}

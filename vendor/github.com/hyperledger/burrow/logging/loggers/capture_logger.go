package loggers

import (
	"sync"

	"github.com/eapache/channels"
	"github.com/go-kit/kit/log"
	"github.com/hyperledger/burrow/logging/structure"
)

type CaptureLogger struct {
	bufferLogger *ChannelLogger
	outputLogger log.Logger
	passthrough  bool
	sync.RWMutex
}

var _ log.Logger = (*CaptureLogger)(nil)

// Capture logger captures output sent to it in a buffer retaining
// a reference to its output logger (the logger whose input it is capturing).
// It can optionally pass logs through to the output logger.
// Because it holds a reference to its output it can also be used to coordinate
// Flushing of the buffer to the output logger in special circumstances.
func NewCaptureLogger(outputLogger log.Logger, bufferCap channels.BufferCap, passthrough bool) *CaptureLogger {
	return &CaptureLogger{
		bufferLogger: NewChannelLogger(bufferCap),
		outputLogger: outputLogger,
		passthrough:  passthrough,
	}
}

func (cl *CaptureLogger) Log(keyvals ...interface{}) error {
	switch structure.Signal(keyvals) {
	case structure.SyncSignal:
		err := cl.Flush()
		if err != nil {
			return err
		}
		// Pass the signal along
		return cl.outputLogger.Log(keyvals...)
	}
	err := cl.bufferLogger.Log(keyvals...)
	if err != nil {
		return err
	}
	if cl.Passthrough() {
		err = cl.outputLogger.Log(keyvals...)
		if err != nil {
			return err
		}
	}
	return nil
}

// Sets whether the CaptureLogger is forwarding log lines sent to it through
// to its output logger. Concurrently safe.
func (cl *CaptureLogger) SetPassthrough(passthrough bool) {
	cl.Lock()
	defer cl.Unlock()
	cl.passthrough = passthrough
}

// Gets whether the CaptureLogger is forwarding log lines sent to through to its
// OutputLogger. Concurrently Safe.
func (cl *CaptureLogger) Passthrough() bool {
	cl.RLock()
	defer cl.RUnlock()
	return cl.passthrough
}

// Flushes every log line available in the buffer at the time of calling
// to the OutputLogger and returns. Does not block indefinitely.
//
// Note: will remove log lines from buffer so they will not be produced on any
// subsequent flush of buffer
func (cl *CaptureLogger) Flush() error {
	return cl.bufferLogger.Flush(cl.outputLogger)
}

// Flushes every log line available in the buffer at the time of calling
// to a slice and returns it. Does not block indefinitely.
//
// Note: will remove log lines from buffer so they will not be produced on any
// subsequent flush of buffer
func (cl *CaptureLogger) FlushLogLines() [][]interface{} {
	return cl.bufferLogger.FlushLogLines()
}

// The OutputLogger whose input this CaptureLogger is capturing
func (cl *CaptureLogger) OutputLogger() log.Logger {
	return cl.outputLogger
}

// The BufferLogger where the input into these CaptureLogger is stored in a ring
// buffer of log lines.
func (cl *CaptureLogger) BufferLogger() *ChannelLogger {
	return cl.bufferLogger
}

// Copyright Monax Industries Limited
// SPDX-License-Identifier: Apache-2.0

package loggers

import (
	"sync"

	"github.com/eapache/channels"
	"github.com/go-kit/kit/log"
	"github.com/hyperledger/burrow/logging/errors"
)

const (
	DefaultLoggingRingBufferCap channels.BufferCap = 100
)

type ChannelLogger struct {
	ch channels.Channel
	sync.RWMutex
}

var _ log.Logger = (*ChannelLogger)(nil)

// Creates a Logger that uses a uses a non-blocking ring buffered channel.
// This logger provides a common abstraction for both a buffered, flushable
// logging cache. And a non-blocking conduit to transmit logs via
// DrainForever (or NonBlockingLogger).
func NewChannelLogger(loggingRingBufferCap channels.BufferCap) *ChannelLogger {
	return &ChannelLogger{
		ch: channels.NewRingChannel(loggingRingBufferCap),
	}
}

func (cl *ChannelLogger) Log(keyvals ...interface{}) error {
	// In case channel is being reset
	cl.RLock()
	defer cl.RUnlock()
	cl.ch.In() <- keyvals
	// We don't have a way to pass on any logging errors, but that's okay: Log is
	// a maximal interface and the error return type is only there for special
	// cases.
	return nil
}

// Get the current occupancy level of the ring buffer
func (cl *ChannelLogger) BufferLength() int {
	return cl.ch.Len()
}

// Get the cap off the internal ring buffer
func (cl *ChannelLogger) BufferCap() channels.BufferCap {
	return cl.ch.Cap()
}

// Read a log line by waiting until one is available and returning it
func (cl *ChannelLogger) WaitReadLogLine() []interface{} {
	logLine, ok := <-cl.ch.Out()
	return readLogLine(logLine, ok)
}

// Tries to read a log line from the channel buffer or returns nil if none is
// immediately available
func (cl *ChannelLogger) ReadLogLine() []interface{} {
	select {
	case logLine, ok := <-cl.ch.Out():
		return readLogLine(logLine, ok)
	default:
		return nil
	}
}

func readLogLine(logLine interface{}, ok bool) []interface{} {
	if !ok {
		// Channel closed
		return nil
	}
	// We are passing slices of interfaces down this channel (go-kit log's Log
	// interface type), a panic is the right thing to do if this type assertion
	// fails.
	return logLine.([]interface{})
}

// Enters an infinite loop that will drain any log lines from the passed logger.
// You may pass in a channel
//
// Exits if the channel is closed.
func (cl *ChannelLogger) DrainForever(logger log.Logger, errCh channels.Channel) {
	// logLine could be nil if channel was closed while waiting for next line
	for logLine := cl.WaitReadLogLine(); logLine != nil; logLine = cl.WaitReadLogLine() {
		err := logger.Log(logLine...)
		if err != nil && errCh != nil {
			errCh.In() <- err
		}
	}
}

// Drains everything that is available at the time of calling
func (cl *ChannelLogger) Flush(logger log.Logger) error {
	// Grab the buffer at the here rather than within loop condition so that we
	// do not drain the buffer forever
	cl.Lock()
	defer cl.Unlock()
	bufferLength := cl.BufferLength()
	var errs []error
	for i := 0; i < bufferLength; i++ {
		logLine := cl.WaitReadLogLine()
		if logLine != nil {
			err := logger.Log(logLine...)
			if err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errors.CombineErrors(errs)
}

// Drains the next contiguous segment of loglines up to the buffer cap waiting
// for at least one line
func (cl *ChannelLogger) FlushLogLines() [][]interface{} {
	logLines := make([][]interface{}, 0, cl.ch.Len())
	cl.Flush(log.LoggerFunc(func(keyvals ...interface{}) error {
		logLines = append(logLines, keyvals)
		return nil
	}))
	return logLines
}

// Close the existing channel halting goroutines that are draining the channel
// and create a new channel to buffer into. Should not cause any log lines
// arriving concurrently to be lost, but any that have not been drained from
// old channel may be.
func (cl *ChannelLogger) Reset() {
	cl.RWMutex.Lock()
	defer cl.RWMutex.Unlock()
	cl.ch.Close()
	cl.ch = channels.NewRingChannel(cl.ch.Cap())
}

// Returns a Logger that wraps the outputLogger passed and does not block on
// calls to Log and a channel of any errors from the underlying logger
func NonBlockingLogger(outputLogger log.Logger) (*ChannelLogger, channels.Channel) {
	cl := NewChannelLogger(DefaultLoggingRingBufferCap)
	errCh := channels.NewRingChannel(cl.BufferCap())
	go cl.DrainForever(outputLogger, errCh)
	return cl, errCh
}

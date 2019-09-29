// Package mux provides stream multiplexing interfaces for libp2p.
//
// For a conceptual overview of stream multiplexing in libp2p, see
// https://docs.libp2p.io/concepts/stream-multiplexing/
package mux

import (
	"errors"
	"io"
	"net"
	"time"
)

// ErrReset is returned when reading or writing on a reset stream.
var ErrReset = errors.New("stream reset")

// Stream is a bidirectional io pipe within a connection.
type MuxedStream interface {
	io.Reader
	io.Writer

	// Close closes the stream for writing. Reading will still work (that
	// is, the remote side can still write).
	io.Closer

	// Reset closes both ends of the stream. Use this to tell the remote
	// side to hang up and go away.
	Reset() error

	SetDeadline(time.Time) error
	SetReadDeadline(time.Time) error
	SetWriteDeadline(time.Time) error
}

// NoopHandler do nothing. Resets streams as soon as they are opened.
var NoopHandler = func(s MuxedStream) { s.Reset() }

// MuxedConn represents a connection to a remote peer that has been
// extended to support stream multiplexing.
//
// A MuxedConn allows a single net.Conn connection to carry many logically
// independent bidirectional streams of binary data.
//
// Together with network.ConnSecurity, MuxedConn is a component of the
// transport.CapableConn interface, which represents a "raw" network
// connection that has been "upgraded" to support the libp2p capabilities
// of secure communication and stream multiplexing.
type MuxedConn interface {
	// Close closes the stream muxer and the the underlying net.Conn.
	io.Closer

	// IsClosed returns whether a connection is fully closed, so it can
	// be garbage collected.
	IsClosed() bool

	// OpenStream creates a new stream.
	OpenStream() (MuxedStream, error)

	// AcceptStream accepts a stream opened by the other side.
	AcceptStream() (MuxedStream, error)
}

// Multiplexer wraps a net.Conn with a stream multiplexing
// implementation and returns a MuxedConn that supports opening
// multiple streams over the underlying net.Conn
type Multiplexer interface {

	// NewConn constructs a new connection
	NewConn(c net.Conn, isServer bool) (MuxedConn, error)
}

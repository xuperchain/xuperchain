package sm_yamux

import (
	"io/ioutil"
	"net"

	mux "github.com/libp2p/go-libp2p-core/mux"
	yamux "github.com/libp2p/go-yamux"
)

// Conn is a connection to a remote peer.
type conn yamux.Session

func (c *conn) yamuxSession() *yamux.Session {
	return (*yamux.Session)(c)
}

func (c *conn) Close() error {
	return c.yamuxSession().Close()
}

func (c *conn) IsClosed() bool {
	return c.yamuxSession().IsClosed()
}

// OpenStream creates a new stream.
func (c *conn) OpenStream() (mux.MuxedStream, error) {
	s, err := c.yamuxSession().OpenStream()
	if err != nil {
		return nil, err
	}

	return s, nil
}

// AcceptStream accepts a stream opened by the other side.
func (c *conn) AcceptStream() (mux.MuxedStream, error) {
	s, err := c.yamuxSession().AcceptStream()
	return s, err
}

// Transport is a go-peerstream transport that constructs
// yamux-backed connections.
type Transport yamux.Config

var DefaultTransport *Transport

func init() {
	config := yamux.DefaultConfig()
	// We've bumped this to 16MiB as this critically limits throughput.
	//
	// 1MiB means a best case of 10MiB/s (83.89Mbps) on a connection with
	// 100ms latency. The default gave us 2.4MiB *best case* which was
	// totally unacceptable.
	config.MaxStreamWindowSize = uint32(16 * 1024 * 1024)
	// don't spam
	config.LogOutput = ioutil.Discard
	// We always run over a security transport that buffers internally
	// (i.e., uses a block cipher).
	config.ReadBufSize = 0
	DefaultTransport = (*Transport)(config)
}

func (t *Transport) NewConn(nc net.Conn, isServer bool) (mux.MuxedConn, error) {
	var s *yamux.Session
	var err error
	if isServer {
		s, err = yamux.Server(nc, t.Config())
	} else {
		s, err = yamux.Client(nc, t.Config())
	}
	return (*conn)(s), err
}

func (t *Transport) Config() *yamux.Config {
	return (*yamux.Config)(t)
}

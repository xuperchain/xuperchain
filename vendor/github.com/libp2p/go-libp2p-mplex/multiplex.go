package peerstream_multiplex

import (
	"net"

	"github.com/libp2p/go-libp2p-core/mux"

	mp "github.com/libp2p/go-mplex"
)

type conn struct {
	*mp.Multiplex
}

func (c *conn) Close() error {
	return c.Multiplex.Close()
}

func (c *conn) IsClosed() bool {
	return c.Multiplex.IsClosed()
}

// OpenStream creates a new stream.
func (c *conn) OpenStream() (mux.MuxedStream, error) {
	return c.Multiplex.NewStream()
}

// AcceptStream accepts a stream opened by the other side.
func (c *conn) AcceptStream() (mux.MuxedStream, error) {
	return c.Multiplex.Accept()
}

// Transport is a go-peerstream transport that constructs
// multiplex-backed connections.
type Transport struct{}

// DefaultTransport has default settings for multiplex
var DefaultTransport = &Transport{}

func (t *Transport) NewConn(nc net.Conn, isServer bool) (mux.MuxedConn, error) {
	return &conn{mp.NewMultiplex(nc, isServer)}, nil
}

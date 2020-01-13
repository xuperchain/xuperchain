package network

import (
	"github.com/libp2p/go-libp2p-core/mux"
	"github.com/libp2p/go-libp2p-core/protocol"
)

// Stream represents a bidirectional channel between two agents in
// a libp2p network. "agent" is as granular as desired, potentially
// being a "request -> reply" pair, or whole protocols.
//
// Streams are backed by a multiplexer underneath the hood.
type Stream interface {
	mux.MuxedStream

	Protocol() protocol.ID
	SetProtocol(id protocol.ID)

	// Stat returns metadata pertaining to this stream.
	Stat() Stat

	// Conn returns the connection this stream is part of.
	Conn() Conn
}

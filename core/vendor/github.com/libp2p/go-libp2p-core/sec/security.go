// Package sec provides secure connection and transport interfaces for libp2p.
package sec

import (
	"context"
	"net"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
)

// SecureConn is an authenticated, encrypted connection.
type SecureConn interface {
	net.Conn
	network.ConnSecurity
}

// A SecureTransport turns inbound and outbound unauthenticated,
// plain-text, native connections into authenticated, encrypted connections.
type SecureTransport interface {
	// SecureInbound secures an inbound connection.
	SecureInbound(ctx context.Context, insecure net.Conn) (SecureConn, error)

	// SecureOutbound secures an outbound connection.
	SecureOutbound(ctx context.Context, insecure net.Conn, p peer.ID) (SecureConn, error)
}

// Package pnet provides interfaces for private networking in libp2p.
package pnet

import "net"

// Protector interface is a way for private network implementation to be transparent in
// libp2p. It is created by implementation and use by libp2p-conn to secure connections
// so they can be only established with selected number of peers.
type Protector interface {
	// Wraps passed connection to protect it
	Protect(net.Conn) (net.Conn, error)

	// Returns key fingerprint that is safe to expose
	Fingerprint() []byte
}

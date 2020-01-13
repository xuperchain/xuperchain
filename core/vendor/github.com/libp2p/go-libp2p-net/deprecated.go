// Deprecated: use github.com/libp2p/go-libp2p-core/network instead.
package net

import (
	"context"
	"time"

	"github.com/libp2p/go-libp2p-core/helpers"
	core "github.com/libp2p/go-libp2p-core/network"
)

// Deprecated: use github.com/libp2p/go-libp2p-core/network.MessageSizeMax instead.
const MessageSizeMax = core.MessageSizeMax

// Deprecated: use github.com/libp2p/go-libp2p-core/network.Stream instead.
type Stream = core.Stream

// Deprecated: use github.com/libp2p/go-libp2p-core/network.Direction instead.
type Direction = core.Direction

const (
	// Deprecated: use github.com/libp2p/go-libp2p-core/network.DirectionUnknown instead.
	DirUnknown = core.DirUnknown
	// Deprecated: use github.com/libp2p/go-libp2p-core/network.DirInbound instead.
	DirInbound = core.DirInbound
	// Deprecated: use github.com/libp2p/go-libp2p-core/network.DirOutbound instead.
	DirOutbound = core.DirOutbound
)

// Deprecated: use github.com/libp2p/go-libp2p-core/network.Stat instead.
type Stat = core.Stat

// Deprecated: use github.com/libp2p/go-libp2p-core/network.StreamHandler instead.
type StreamHandler = core.StreamHandler

// Deprecated: use github.com/libp2p/go-libp2p-core/network.ConnSecurity instead.
type ConnSecurity = core.ConnSecurity

// Deprecated: use github.com/libp2p/go-libp2p-core/network.ConnMultiaddrs instead.
type ConnMultiaddrs = core.ConnMultiaddrs

// Deprecated: use github.com/libp2p/go-libp2p-core/network.Conn instead.
type Conn = core.Conn

// Deprecated: use github.com/libp2p/go-libp2p-core/network.ConnHandler instead.
type ConnHandler = core.ConnHandler

// Deprecated: use github.com/libp2p/go-libp2p-core/network.Network instead.
type Network = core.Network

// Deprecated: use github.com/libp2p/go-libp2p-core/network.ErrNoRemoteAddrs instead.
var ErrNoRemoteAddrs = core.ErrNoRemoteAddrs

// Deprecated: use github.com/libp2p/go-libp2p-core/network.ErrNoConn instead.
var ErrNoConn = core.ErrNoConn

// Deprecated: use github.com/libp2p/go-libp2p-core/network.Dialer instead.
type Dialer = core.Dialer

// Deprecated: use github.com/libp2p/go-libp2p-core/network.Connectedness instead.
type Connectedness = core.Connectedness

const (
	// Deprecated: use github.com/libp2p/go-libp2p-core/network.NotConnected instead.
	NotConnected = core.NotConnected

	// Deprecated: use github.com/libp2p/go-libp2p-core/network.Connected instead.
	Connected = core.Connected

	// Deprecated: use github.com/libp2p/go-libp2p-core/network.CanConnect instead.
	CanConnect = core.CanConnect

	// Deprecated: use github.com/libp2p/go-libp2p-core/network.CannotConnect instead.
	CannotConnect = core.CannotConnect
)

// Deprecated: use github.com/libp2p/go-libp2p-core/network.Notifiee instead.
type Notifiee = core.Notifiee

// Deprecated: use github.com/libp2p/go-libp2p-core/network.NotifyBundle instead.
type NotifyBundle = core.NotifyBundle

// Deprecated: use github.com/libp2p/go-libp2p-core/network.WithNoDial instead.
func WithNoDial(ctx context.Context, reason string) context.Context {
	return core.WithNoDial(ctx, reason)
}

// Deprecated: use github.com/libp2p/go-libp2p-core/network.GetNoDial instead.
func GetNoDial(ctx context.Context) (nodial bool, reason string) {
	return core.GetNoDial(ctx)
}

// Deprecated: use github.com/libp2p/go-libp2p-core/helpers.EOFTimeout instead.
var EOFTimeout = helpers.EOFTimeout

// Deprecated: use github.com/libp2p/go-libp2p-core/helpers.ErrExpectedEOF instead.
var ErrExpectedEOF = helpers.ErrExpectedEOF

// Deprecated: use github.com/libp2p/go-libp2p-core/helpers.FullClose instead.
func FullClose(s core.Stream) error {
	return helpers.FullClose(s)
}

// Deprecated: use github.com/libp2p/go-libp2p-core/helpers.AwaitEOF instead.
func AwaitEOF(s core.Stream) error {
	return helpers.AwaitEOF(s)
}

// Deprecated: use github.com/libp2p/go-libp2p-core/network.DialPeerTimeout instead.
// Warning: it's impossible to alias a var in go. Writes to this var would have no longer
// have any effect, so it has been commented out to induce breakage for added safety.
// var DialPeerTimeout = core.DialPeerTimeout

// Deprecated: use github.com/libp2p/go-libp2p-core/network.GetDialPeerTimeout instead.
func GetDialPeerTimeout(ctx context.Context) time.Duration {
	return core.GetDialPeerTimeout(ctx)
}

// Deprecated: use github.com/libp2p/go-libp2p-core/network.WithDialPeerTimeout instead.
func WithDialPeerTimeout(ctx context.Context, timeout time.Duration) context.Context {
	return core.WithDialPeerTimeout(ctx, timeout)
}

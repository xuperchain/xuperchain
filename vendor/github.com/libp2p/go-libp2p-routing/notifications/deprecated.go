// Deprecated: use github.com/libp2p/go-libp2p-core/routing instead.
package notifications

import (
	"context"

	core "github.com/libp2p/go-libp2p-core/routing"
)

// Deprecated: use github.com/libp2p/go-libp2p-core/routing/QueryEventType instead.
type QueryEventType = core.QueryEventType

// Deprecated: use github.com/libp2p/go-libp2p-core/routing/QueryEventBufferSize instead.
// Warning: it's impossible to alias a var in go, so reads and writes to this variable may be inaccurate
// or not have the intended effect.
var QueryEventBufferSize = core.QueryEventBufferSize

const (
	// Deprecated: use github.com/libp2p/go-libp2p-core/routing/SendingQuery instead.
	SendingQuery = core.SendingQuery
	// Deprecated: use github.com/libp2p/go-libp2p-core/routing/PeerResponse instead.
	PeerResponse = core.PeerResponse
	// Deprecated: use github.com/libp2p/go-libp2p-core/routing/FinalPeer instead.
	FinalPeer = core.FinalPeer
	// Deprecated: use github.com/libp2p/go-libp2p-core/routing/QueryError instead.
	QueryError = core.QueryError
	// Deprecated: use github.com/libp2p/go-libp2p-core/routing/Provider instead.
	Provider = core.Provider
	// Deprecated: use github.com/libp2p/go-libp2p-core/routing/Value instead.
	Value = core.Value
	// Deprecated: use github.com/libp2p/go-libp2p-core/routing/AddingPeer instead.
	AddingPeer = core.AddingPeer
	// Deprecated: use github.com/libp2p/go-libp2p-core/routing/DialingPeer instead.
	DialingPeer = core.DialingPeer
)

// Deprecated: use github.com/libp2p/go-libp2p-core/routing/QueryEvent instead.
type QueryEvent = core.QueryEvent

// Deprecated: use github.com/libp2p/go-libp2p-core/routing/RegisterForQueryEvents instead.
func RegisterForQueryEvents(ctx context.Context) (context.Context, <-chan *core.QueryEvent) {
	return core.RegisterForQueryEvents(ctx)
}

// Deprecated: use github.com/libp2p/go-libp2p-core/routing/PublishQueryEvent instead.
func PublishQueryEvent(ctx context.Context, ev *core.QueryEvent) {
	core.PublishQueryEvent(ctx, ev)
}

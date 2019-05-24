package discovery

import (
	"context"
	"time"

	pstore "github.com/libp2p/go-libp2p-peerstore"
)

// Advertiser is an interface for advertising services
type Advertiser interface {
	// Advertise advertises a service
	Advertise(ctx context.Context, ns string, opts ...Option) (time.Duration, error)
}

// Discoverer is an interface for peer discovery
type Discoverer interface {
	// FindPeers discovers peers providing a service
	FindPeers(ctx context.Context, ns string, opts ...Option) (<-chan pstore.PeerInfo, error)
}

// Discovery is an interface that combines service advertisement and peer discovery
type Discovery interface {
	Advertiser
	Discoverer
}

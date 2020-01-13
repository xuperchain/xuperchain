package relay

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/transport"
	ma "github.com/multiformats/go-multiaddr"
)

func (d *RelayTransport) Dial(ctx context.Context, a ma.Multiaddr, p peer.ID) (transport.CapableConn, error) {
	c, err := d.Relay().Dial(ctx, a, p)
	if err != nil {
		return nil, err
	}
	c.tagHop()
	return d.upgrader.UpgradeOutbound(ctx, d, c, p)
}

func (r *Relay) Dial(ctx context.Context, a ma.Multiaddr, p peer.ID) (*Conn, error) {
	// split /a/p2p-circuit/b into (/a, /p2p-circuit/b)
	relayaddr, destaddr := ma.SplitFunc(a, func(c ma.Component) bool {
		return c.Protocol().Code == P_CIRCUIT
	})

	// If the address contained no /p2p-circuit part, the second part is nil.
	if destaddr == nil {
		return nil, fmt.Errorf("%s is not a relay address", a)
	}

	// Strip the /p2p-circuit prefix from the destaddr.
	_, destaddr = ma.SplitFirst(destaddr)

	dinfo := &peer.AddrInfo{ID: p, Addrs: []ma.Multiaddr{}}
	if destaddr != nil {
		dinfo.Addrs = append(dinfo.Addrs, destaddr)
	}

	if relayaddr == nil {
		// unspecific relay address, try dialing using known hop relays
		return r.tryDialRelays(ctx, *dinfo)
	}

	var rinfo *peer.AddrInfo
	rinfo, err := peer.AddrInfoFromP2pAddr(relayaddr)
	if err != nil {
		return nil, fmt.Errorf("error parsing multiaddr '%s': %s", relayaddr.String(), err)
	}

	return r.DialPeer(ctx, *rinfo, *dinfo)
}

func (r *Relay) tryDialRelays(ctx context.Context, dinfo peer.AddrInfo) (*Conn, error) {
	var relays []peer.ID
	r.mx.Lock()
	for p := range r.relays {
		relays = append(relays, p)
	}
	r.mx.Unlock()

	// shuffle list of relays, avoid overloading a specific relay
	for i := range relays {
		j := rand.Intn(i + 1)
		relays[i], relays[j] = relays[j], relays[i]
	}

	for _, relay := range relays {
		if len(r.host.Network().ConnsToPeer(relay)) == 0 {
			continue
		}

		rctx, cancel := context.WithTimeout(ctx, HopConnectTimeout)
		c, err := r.DialPeer(rctx, peer.AddrInfo{ID: relay}, dinfo)
		cancel()

		if err == nil {
			return c, nil
		}

		log.Debugf("error opening relay connection through %s: %s", dinfo.ID, err.Error())
	}

	return nil, fmt.Errorf("Failed to dial through %d known relay hosts", len(relays))
}

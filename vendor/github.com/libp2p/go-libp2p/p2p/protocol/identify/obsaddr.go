package identify

import (
	"context"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peerstore"

	ma "github.com/multiformats/go-multiaddr"
)

const ActivationThresh = 4

var GCInterval = 10 * time.Minute

type observation struct {
	seenTime      time.Time
	connDirection network.Direction
}

// ObservedAddr is an entry for an address reported by our peers.
// We only use addresses that:
// - have been observed at least 4 times in last 1h. (counter symmetric nats)
// - have been observed at least once recently (1h), because our position in the
//   network, or network port mapppings, may have changed.
type ObservedAddr struct {
	Addr     ma.Multiaddr
	SeenBy   map[string]observation // peer(observer) address -> observation info
	LastSeen time.Time
}

func (oa *ObservedAddr) activated(ttl time.Duration) bool {
	// We only activate if in the TTL other peers observed the same address
	// of ours at least 4 times.
	return len(oa.SeenBy) >= ActivationThresh
}

type newObservation struct {
	observed, local, observer ma.Multiaddr
	direction                 network.Direction
}

// ObservedAddrSet keeps track of a set of ObservedAddrs
// the zero-value is ready to be used.
type ObservedAddrSet struct {
	sync.RWMutex // guards whole datastruct.

	// local(internal) address -> list of observed(external) addresses
	addrs map[string][]*ObservedAddr
	ttl   time.Duration

	// this is the worker channel
	wch chan newObservation
}

func NewObservedAddrSet(ctx context.Context) *ObservedAddrSet {
	oas := &ObservedAddrSet{
		addrs: make(map[string][]*ObservedAddr),
		ttl:   peerstore.OwnObservedAddrTTL,
		wch:   make(chan newObservation, 16),
	}
	go oas.worker(ctx)
	return oas
}

// AddrsFor return all activated observed addresses associated with the given
// (resolved) listen address.
func (oas *ObservedAddrSet) AddrsFor(addr ma.Multiaddr) (addrs []ma.Multiaddr) {
	oas.RLock()
	defer oas.RUnlock()

	if len(oas.addrs) == 0 {
		return nil
	}

	key := string(addr.Bytes())
	observedAddrs, ok := oas.addrs[key]
	if !ok {
		return
	}

	now := time.Now()
	for _, a := range observedAddrs {
		if now.Sub(a.LastSeen) <= oas.ttl && a.activated(oas.ttl) {
			addrs = append(addrs, a.Addr)
		}
	}

	return addrs
}

// Addrs return all activated observed addresses
func (oas *ObservedAddrSet) Addrs() (addrs []ma.Multiaddr) {
	oas.RLock()
	defer oas.RUnlock()

	if len(oas.addrs) == 0 {
		return nil
	}

	now := time.Now()
	for _, observedAddrs := range oas.addrs {
		for _, a := range observedAddrs {
			if now.Sub(a.LastSeen) <= oas.ttl && a.activated(oas.ttl) {
				addrs = append(addrs, a.Addr)
			}
		}
	}
	return addrs
}

func (oas *ObservedAddrSet) Add(observed, local, observer ma.Multiaddr,
	direction network.Direction) {
	select {
	case oas.wch <- newObservation{observed: observed, local: local, observer: observer, direction: direction}:
	default:
		log.Debugf("dropping address observation of %s; buffer full", observed)
	}
}

func (oas *ObservedAddrSet) worker(ctx context.Context) {
	ticker := time.NewTicker(GCInterval)
	defer ticker.Stop()

	for {
		select {
		case obs := <-oas.wch:
			oas.doAdd(obs.observed, obs.local, obs.observer, obs.direction)

		case <-ticker.C:
			oas.gc()

		case <-ctx.Done():
			return
		}
	}
}

func (oas *ObservedAddrSet) gc() {
	oas.Lock()
	defer oas.Unlock()

	now := time.Now()
	for local, observedAddrs := range oas.addrs {
		// TODO we can do this without allocating by compacting the array in place
		filteredAddrs := make([]*ObservedAddr, 0, len(observedAddrs))

		for _, a := range observedAddrs {
			// clean up SeenBy set
			for k, ob := range a.SeenBy {
				if now.Sub(ob.seenTime) > oas.ttl*ActivationThresh {
					delete(a.SeenBy, k)
				}
			}

			// leave only alive observed addresses
			if now.Sub(a.LastSeen) <= oas.ttl {
				filteredAddrs = append(filteredAddrs, a)
			}
		}
		if len(filteredAddrs) > 0 {
			oas.addrs[local] = filteredAddrs
		} else {
			delete(oas.addrs, local)
		}
	}
}

func (oas *ObservedAddrSet) doAdd(observed, local, observer ma.Multiaddr,
	direction network.Direction) {

	now := time.Now()
	observerString := observerGroup(observer)
	localString := string(local.Bytes())
	ob := observation{
		seenTime:      now,
		connDirection: direction,
	}

	oas.Lock()
	defer oas.Unlock()

	observedAddrs := oas.addrs[localString]
	// check if observed address seen yet, if so, update it
	for i, previousObserved := range observedAddrs {
		if previousObserved.Addr.Equal(observed) {
			observedAddrs[i].SeenBy[observerString] = ob
			observedAddrs[i].LastSeen = now
			return
		}
	}
	// observed address not seen yet, append it
	oas.addrs[localString] = append(oas.addrs[localString], &ObservedAddr{
		Addr: observed,
		SeenBy: map[string]observation{
			observerString: ob,
		},
		LastSeen: now,
	})
}

// observerGroup is a function that determines what part of
// a multiaddr counts as a different observer. for example,
// two ipfs nodes at the same IP/TCP transport would get
// the exact same NAT mapping; they would count as the
// same observer. This may protect against NATs who assign
// different ports to addresses at different IP hosts, but
// not TCP ports.
//
// Here, we use the root multiaddr address. This is mostly
// IP addresses. In practice, this is what we want.
func observerGroup(m ma.Multiaddr) string {
	//TODO: If IPv6 rolls out we should mark /64 routing zones as one group
	first, _ := ma.SplitFirst(m)
	return string(first.Bytes())
}

func (oas *ObservedAddrSet) SetTTL(ttl time.Duration) {
	oas.Lock()
	defer oas.Unlock()
	oas.ttl = ttl
}

func (oas *ObservedAddrSet) TTL() time.Duration {
	oas.RLock()
	defer oas.RUnlock()
	return oas.ttl
}

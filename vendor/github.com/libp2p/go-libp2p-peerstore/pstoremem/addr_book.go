package pstoremem

import (
	"context"
	"sort"
	"sync"
	"time"

	logging "github.com/ipfs/go-log"
	peer "github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"

	pstore "github.com/libp2p/go-libp2p-core/peerstore"
	addr "github.com/libp2p/go-libp2p-peerstore/addr"
)

var log = logging.Logger("peerstore")

type expiringAddr struct {
	Addr    ma.Multiaddr
	TTL     time.Duration
	Expires time.Time
}

func (e *expiringAddr) ExpiredBy(t time.Time) bool {
	return t.After(e.Expires)
}

type addrSegments [256]*addrSegment

type addrSegment struct {
	sync.RWMutex

	// Use pointers to save memory. Maps always leave some fraction of their
	// space unused. storing the *values* directly in the map will
	// drastically increase the space waste. In our case, by 6x.
	addrs map[peer.ID]map[string]*expiringAddr
}

func (s *addrSegments) get(p peer.ID) *addrSegment {
	return s[byte(p[len(p)-1])]
}

// memoryAddrBook manages addresses.
type memoryAddrBook struct {
	segments addrSegments

	ctx    context.Context
	cancel func()

	subManager *AddrSubManager
}

var _ pstore.AddrBook = (*memoryAddrBook)(nil)

func NewAddrBook() pstore.AddrBook {
	ctx, cancel := context.WithCancel(context.Background())

	ab := &memoryAddrBook{
		segments: func() (ret addrSegments) {
			for i, _ := range ret {
				ret[i] = &addrSegment{addrs: make(map[peer.ID]map[string]*expiringAddr)}
			}
			return ret
		}(),
		subManager: NewAddrSubManager(),
		ctx:        ctx,
		cancel:     cancel,
	}

	go ab.background()
	return ab
}

// background periodically schedules a gc
func (mab *memoryAddrBook) background() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			mab.gc()

		case <-mab.ctx.Done():
			return
		}
	}
}

func (mab *memoryAddrBook) Close() error {
	mab.cancel()
	return nil
}

// gc garbage collects the in-memory address book.
func (mab *memoryAddrBook) gc() {
	now := time.Now()
	for _, s := range mab.segments {
		s.Lock()
		for p, amap := range s.addrs {
			for k, addr := range amap {
				if addr.ExpiredBy(now) {
					delete(amap, k)
				}
			}
			if len(amap) == 0 {
				delete(s.addrs, p)
			}
		}
		s.Unlock()
	}

}

func (mab *memoryAddrBook) PeersWithAddrs() peer.IDSlice {
	var pids peer.IDSlice
	for _, s := range mab.segments {
		s.RLock()
		for pid, _ := range s.addrs {
			pids = append(pids, pid)
		}
		s.RUnlock()
	}
	return pids
}

// AddAddr calls AddAddrs(p, []ma.Multiaddr{addr}, ttl)
func (mab *memoryAddrBook) AddAddr(p peer.ID, addr ma.Multiaddr, ttl time.Duration) {
	mab.AddAddrs(p, []ma.Multiaddr{addr}, ttl)
}

// AddAddrs gives memoryAddrBook addresses to use, with a given ttl
// (time-to-live), after which the address is no longer valid.
// This function never reduces the TTL or expiration of an address.
func (mab *memoryAddrBook) AddAddrs(p peer.ID, addrs []ma.Multiaddr, ttl time.Duration) {
	// if ttl is zero, exit. nothing to do.
	if ttl <= 0 {
		return
	}

	s := mab.segments.get(p)
	s.Lock()
	defer s.Unlock()

	amap := s.addrs[p]
	if amap == nil {
		amap = make(map[string]*expiringAddr, len(addrs))
		s.addrs[p] = amap
	}
	exp := time.Now().Add(ttl)
	for _, addr := range addrs {
		if addr == nil {
			log.Warningf("was passed nil multiaddr for %s", p)
			continue
		}
		asBytes := addr.Bytes()
		a, found := amap[string(asBytes)] // won't allocate.
		if !found {
			// not found, save and announce it.
			amap[string(asBytes)] = &expiringAddr{Addr: addr, Expires: exp, TTL: ttl}
			mab.subManager.BroadcastAddr(p, addr)
		} else {
			// Update expiration/TTL independently.
			// We never want to reduce either.
			if ttl > a.TTL {
				a.TTL = ttl
			}
			if exp.After(a.Expires) {
				a.Expires = exp
			}
		}
	}
}

// SetAddr calls mgr.SetAddrs(p, addr, ttl)
func (mab *memoryAddrBook) SetAddr(p peer.ID, addr ma.Multiaddr, ttl time.Duration) {
	mab.SetAddrs(p, []ma.Multiaddr{addr}, ttl)
}

// SetAddrs sets the ttl on addresses. This clears any TTL there previously.
// This is used when we receive the best estimate of the validity of an address.
func (mab *memoryAddrBook) SetAddrs(p peer.ID, addrs []ma.Multiaddr, ttl time.Duration) {
	s := mab.segments.get(p)
	s.Lock()
	defer s.Unlock()

	amap := s.addrs[p]
	if amap == nil {
		amap = make(map[string]*expiringAddr, len(addrs))
		s.addrs[p] = amap
	}

	exp := time.Now().Add(ttl)
	for _, addr := range addrs {
		if addr == nil {
			log.Warningf("was passed nil multiaddr for %s", p)
			continue
		}

		// re-set all of them for new ttl.
		aBytes := addr.Bytes()
		if ttl > 0 {
			amap[string(aBytes)] = &expiringAddr{Addr: addr, Expires: exp, TTL: ttl}
			mab.subManager.BroadcastAddr(p, addr)
		} else {
			delete(amap, string(aBytes))
		}
	}
}

// UpdateAddrs updates the addresses associated with the given peer that have
// the given oldTTL to have the given newTTL.
func (mab *memoryAddrBook) UpdateAddrs(p peer.ID, oldTTL time.Duration, newTTL time.Duration) {
	s := mab.segments.get(p)
	s.Lock()
	defer s.Unlock()

	amap, found := s.addrs[p]
	if !found {
		return
	}

	exp := time.Now().Add(newTTL)
	for k, addr := range amap {
		if oldTTL == addr.TTL {
			addr.TTL = newTTL
			addr.Expires = exp
			amap[k] = addr
		}
	}
}

// Addresses returns all known (and valid) addresses for a given
func (mab *memoryAddrBook) Addrs(p peer.ID) []ma.Multiaddr {
	s := mab.segments.get(p)
	s.RLock()
	defer s.RUnlock()

	amap, found := s.addrs[p]
	if !found {
		return nil
	}

	now := time.Now()
	good := make([]ma.Multiaddr, 0, len(amap))
	for _, m := range amap {
		if !m.ExpiredBy(now) {
			good = append(good, m.Addr)
		}
	}

	return good
}

// ClearAddrs removes all previously stored addresses
func (mab *memoryAddrBook) ClearAddrs(p peer.ID) {
	s := mab.segments.get(p)
	s.Lock()
	defer s.Unlock()

	delete(s.addrs, p)
}

// AddrStream returns a channel on which all new addresses discovered for a
// given peer ID will be published.
func (mab *memoryAddrBook) AddrStream(ctx context.Context, p peer.ID) <-chan ma.Multiaddr {
	s := mab.segments.get(p)
	s.RLock()
	defer s.RUnlock()

	baseaddrslice := s.addrs[p]
	initial := make([]ma.Multiaddr, 0, len(baseaddrslice))
	for _, a := range baseaddrslice {
		initial = append(initial, a.Addr)
	}

	return mab.subManager.AddrStream(ctx, p, initial)
}

type addrSub struct {
	pubch  chan ma.Multiaddr
	lk     sync.Mutex
	buffer []ma.Multiaddr
	ctx    context.Context
}

func (s *addrSub) pubAddr(a ma.Multiaddr) {
	select {
	case s.pubch <- a:
	case <-s.ctx.Done():
	}
}

// An abstracted, pub-sub manager for address streams. Extracted from
// memoryAddrBook in order to support additional implementations.
type AddrSubManager struct {
	mu   sync.RWMutex
	subs map[peer.ID][]*addrSub
}

// NewAddrSubManager initializes an AddrSubManager.
func NewAddrSubManager() *AddrSubManager {
	return &AddrSubManager{
		subs: make(map[peer.ID][]*addrSub),
	}
}

// Used internally by the address stream coroutine to remove a subscription
// from the manager.
func (mgr *AddrSubManager) removeSub(p peer.ID, s *addrSub) {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	subs := mgr.subs[p]
	if len(subs) == 1 {
		if subs[0] != s {
			return
		}
		delete(mgr.subs, p)
		return
	}

	for i, v := range subs {
		if v == s {
			subs[i] = subs[len(subs)-1]
			subs[len(subs)-1] = nil
			mgr.subs[p] = subs[:len(subs)-1]
			return
		}
	}
}

// BroadcastAddr broadcasts a new address to all subscribed streams.
func (mgr *AddrSubManager) BroadcastAddr(p peer.ID, addr ma.Multiaddr) {
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	if subs, ok := mgr.subs[p]; ok {
		for _, sub := range subs {
			sub.pubAddr(addr)
		}
	}
}

// AddrStream creates a new subscription for a given peer ID, pre-populating the
// channel with any addresses we might already have on file.
func (mgr *AddrSubManager) AddrStream(ctx context.Context, p peer.ID, initial []ma.Multiaddr) <-chan ma.Multiaddr {
	sub := &addrSub{pubch: make(chan ma.Multiaddr), ctx: ctx}
	out := make(chan ma.Multiaddr)

	mgr.mu.Lock()
	if _, ok := mgr.subs[p]; ok {
		mgr.subs[p] = append(mgr.subs[p], sub)
	} else {
		mgr.subs[p] = []*addrSub{sub}
	}
	mgr.mu.Unlock()

	sort.Sort(addr.AddrList(initial))

	go func(buffer []ma.Multiaddr) {
		defer close(out)

		sent := make(map[string]bool, len(buffer))
		var outch chan ma.Multiaddr

		for _, a := range buffer {
			sent[string(a.Bytes())] = true
		}

		var next ma.Multiaddr
		if len(buffer) > 0 {
			next = buffer[0]
			buffer = buffer[1:]
			outch = out
		}

		for {
			select {
			case outch <- next:
				if len(buffer) > 0 {
					next = buffer[0]
					buffer = buffer[1:]
				} else {
					outch = nil
					next = nil
				}
			case naddr := <-sub.pubch:
				if sent[string(naddr.Bytes())] {
					continue
				}

				sent[string(naddr.Bytes())] = true
				if next == nil {
					next = naddr
					outch = out
				} else {
					buffer = append(buffer, naddr)
				}
			case <-ctx.Done():
				mgr.removeSub(p, sub)
				return
			}
		}

	}(initial)

	return out
}

package autonat

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"time"

	host "github.com/libp2p/go-libp2p-host"
	inet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
)

// NATStatus is the state of NAT as detected by the ambient service.
type NATStatus int

const (
	// NAT status is unknown; this means that the ambient service has not been
	// able to decide the presence of NAT in the most recent attempt to test
	// dial through known autonat peers.  initial state.
	NATStatusUnknown NATStatus = iota
	// NAT status is publicly dialable
	NATStatusPublic
	// NAT status is private network
	NATStatusPrivate
)

var (
	AutoNATBootDelay       = 15 * time.Second
	AutoNATRetryInterval   = 90 * time.Second
	AutoNATRefreshInterval = 15 * time.Minute
	AutoNATRequestTimeout  = 60 * time.Second
)

// AutoNAT is the interface for ambient NAT autodiscovery
type AutoNAT interface {
	// Status returns the current NAT status
	Status() NATStatus
	// PublicAddr returns the public dial address when NAT status is public and an
	// error otherwise
	PublicAddr() (ma.Multiaddr, error)
}

// AmbientAutoNAT is the implementation of ambient NAT autodiscovery
type AmbientAutoNAT struct {
	ctx  context.Context
	host host.Host

	getAddrs GetAddrs

	mx     sync.Mutex
	peers  map[peer.ID][]ma.Multiaddr
	status NATStatus
	addr   ma.Multiaddr
	// Reflects the confidence on of the NATStatus being private, as a single
	// dialback may fail for reasons unrelated to NAT.
	// If it is <3, then multiple autoNAT peers may be contacted for dialback
	// If only a single autoNAT peer is known, then the confidence increases
	// for each failure until it reaches 3.
	confidence int
}

// NewAutoNAT creates a new ambient NAT autodiscovery instance attached to a host
// If getAddrs is nil, h.Addrs will be used
func NewAutoNAT(ctx context.Context, h host.Host, getAddrs GetAddrs) AutoNAT {
	if getAddrs == nil {
		getAddrs = h.Addrs
	}

	as := &AmbientAutoNAT{
		ctx:      ctx,
		host:     h,
		getAddrs: getAddrs,
		peers:    make(map[peer.ID][]ma.Multiaddr),
		status:   NATStatusUnknown,
	}

	h.Network().Notify(as)
	go as.background()

	return as
}

func (as *AmbientAutoNAT) Status() NATStatus {
	return as.status
}

func (as *AmbientAutoNAT) PublicAddr() (ma.Multiaddr, error) {
	as.mx.Lock()
	defer as.mx.Unlock()

	if as.status != NATStatusPublic {
		return nil, errors.New("NAT Status is not public")
	}

	return as.addr, nil
}

func (as *AmbientAutoNAT) background() {
	// wait a bit for the node to come online and establish some connections
	// before starting autodetection
	select {
	case <-time.After(AutoNATBootDelay):
	case <-as.ctx.Done():
		return
	}

	for {
		as.autodetect()

		delay := AutoNATRefreshInterval
		if as.status == NATStatusUnknown {
			delay = AutoNATRetryInterval
		}

		select {
		case <-time.After(delay):
		case <-as.ctx.Done():
			return
		}
	}
}

func (as *AmbientAutoNAT) autodetect() {
	peers := as.getPeers()

	if len(peers) == 0 {
		log.Debugf("skipping NAT auto detection; no autonat peers")
		return
	}

	cli := NewAutoNATClient(as.host, as.getAddrs)
	failures := 0

	for _, pi := range peers {
		ctx, cancel := context.WithTimeout(as.ctx, AutoNATRequestTimeout)
		as.host.Peerstore().AddAddrs(pi.ID, pi.Addrs, pstore.TempAddrTTL)
		a, err := cli.DialBack(ctx, pi.ID)
		cancel()

		switch {
		case err == nil:
			log.Debugf("NAT status is public; address through %s: %s", pi.ID.Pretty(), a.String())
			as.mx.Lock()
			as.addr = a
			as.status = NATStatusPublic
			as.confidence = 0
			as.mx.Unlock()
			return

		case IsDialError(err):
			log.Debugf("dial error through %s: %s", pi.ID.Pretty(), err.Error())
			failures++
			if failures >= 3 || as.confidence >= 3 { // 3 times is enemy action
				log.Debugf("NAT status is private")
				as.mx.Lock()
				as.status = NATStatusPrivate
				as.confidence = 3
				as.mx.Unlock()
				return
			}

		default:
			log.Debugf("Error dialing through %s: %s", pi.ID.Pretty(), err.Error())
		}
	}

	as.mx.Lock()
	if failures > 0 {
		as.status = NATStatusPrivate
		as.confidence++
		log.Debugf("NAT status is private")
	} else {
		as.status = NATStatusUnknown
		as.confidence = 0
		log.Debugf("NAT status is unknown")
	}
	as.mx.Unlock()
}

func (as *AmbientAutoNAT) getPeers() []pstore.PeerInfo {
	as.mx.Lock()
	defer as.mx.Unlock()

	if len(as.peers) == 0 {
		return nil
	}

	var connected, others []pstore.PeerInfo

	for p, addrs := range as.peers {
		if as.host.Network().Connectedness(p) == inet.Connected {
			connected = append(connected, pstore.PeerInfo{ID: p, Addrs: addrs})
		} else {
			others = append(others, pstore.PeerInfo{ID: p, Addrs: addrs})
		}
	}

	shufflePeers(connected)

	if len(connected) < 3 {
		shufflePeers(others)
		return append(connected, others...)
	} else {
		return connected
	}
}

func shufflePeers(peers []pstore.PeerInfo) {
	for i := range peers {
		j := rand.Intn(i + 1)
		peers[i], peers[j] = peers[j], peers[i]
	}
}

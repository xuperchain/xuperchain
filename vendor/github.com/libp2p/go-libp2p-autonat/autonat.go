package autonat

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
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
	AutoNATRequestTimeout  = 30 * time.Second
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
	as.mx.Lock()
	defer as.mx.Unlock()
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
	ctx, cancel := context.WithTimeout(as.ctx, AutoNATRequestTimeout)
	defer cancel()

	var result struct {
		sync.Mutex
		private int
		public  int
		pubaddr ma.Multiaddr
	}

	probe := 3 - as.confidence
	if probe == 0 {
		probe = 1
	}
	if probe > len(peers) {
		probe = len(peers)
	}

	var wg sync.WaitGroup

	for _, pi := range peers[:probe] {
		wg.Add(1)
		go func(pi peer.AddrInfo) {
			defer wg.Done()

			as.host.Peerstore().AddAddrs(pi.ID, pi.Addrs, peerstore.TempAddrTTL)
			a, err := cli.DialBack(ctx, pi.ID)

			switch {
			case err == nil:
				log.Debugf("Dialback through %s successful; public address is %s", pi.ID.Pretty(), a.String())
				result.Lock()
				result.public++
				result.pubaddr = a
				result.Unlock()

			case IsDialError(err):
				log.Debugf("Dialback through %s failed", pi.ID.Pretty())
				result.Lock()
				result.private++
				result.Unlock()

			default:
				log.Debugf("Dialback error through %s: %s", pi.ID.Pretty(), err)
			}
		}(pi)
	}

	wg.Wait()

	as.mx.Lock()
	if result.public > 0 {
		log.Debugf("NAT status is public")
		if as.status == NATStatusPrivate {
			// we are flipping our NATStatus, so confidence drops to 0
			as.confidence = 0
		} else if as.confidence < 3 {
			as.confidence++
		}
		as.status = NATStatusPublic
		as.addr = result.pubaddr
	} else if result.private > 0 {
		log.Debugf("NAT status is private")
		if as.status == NATStatusPublic {
			// we are flipping our NATStatus, so confidence drops to 0
			as.confidence = 0
		} else if as.confidence < 3 {
			as.confidence++
		}
		as.status = NATStatusPrivate
		as.addr = nil
	} else if as.confidence > 0 {
		// don't just flip to unknown, reduce confidence first
		as.confidence--
	} else {
		log.Debugf("NAT status is unknown")
		as.status = NATStatusUnknown
		as.addr = nil
	}
	as.mx.Unlock()
}

func (as *AmbientAutoNAT) getPeers() []peer.AddrInfo {
	as.mx.Lock()
	defer as.mx.Unlock()

	if len(as.peers) == 0 {
		return nil
	}

	var connected, others []peer.AddrInfo

	for p, addrs := range as.peers {
		if as.host.Network().Connectedness(p) == network.Connected {
			connected = append(connected, peer.AddrInfo{ID: p, Addrs: addrs})
		} else {
			others = append(others, peer.AddrInfo{ID: p, Addrs: addrs})
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

func shufflePeers(peers []peer.AddrInfo) {
	for i := range peers {
		j := rand.Intn(i + 1)
		peers[i], peers[j] = peers[j], peers[i]
	}
}

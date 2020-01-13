package autonat

import (
	"time"

	"github.com/libp2p/go-libp2p-core/network"

	ma "github.com/multiformats/go-multiaddr"
)

var _ network.Notifiee = (*AmbientAutoNAT)(nil)

var AutoNATIdentifyDelay = 5 * time.Second

func (as *AmbientAutoNAT) Listen(net network.Network, a ma.Multiaddr)         {}
func (as *AmbientAutoNAT) ListenClose(net network.Network, a ma.Multiaddr)    {}
func (as *AmbientAutoNAT) OpenedStream(net network.Network, s network.Stream) {}
func (as *AmbientAutoNAT) ClosedStream(net network.Network, s network.Stream) {}

func (as *AmbientAutoNAT) Connected(net network.Network, c network.Conn) {
	p := c.RemotePeer()

	go func() {
		// add some delay for identify
		time.Sleep(AutoNATIdentifyDelay)

		protos, err := as.host.Peerstore().SupportsProtocols(p, AutoNATProto)
		if err != nil {
			log.Debugf("error retrieving supported protocols for peer %s: %s", p, err)
			return
		}

		if len(protos) > 0 {
			log.Infof("Discovered AutoNAT peer %s", p.Pretty())
			as.mx.Lock()
			as.peers[p] = as.host.Peerstore().Addrs(p)
			as.mx.Unlock()
		}
	}()
}

func (as *AmbientAutoNAT) Disconnected(net network.Network, c network.Conn) {}

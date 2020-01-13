package dht

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"

	u "github.com/ipfs/go-ipfs-util"
	"github.com/multiformats/go-multiaddr"
	_ "github.com/multiformats/go-multiaddr-dns"
)

var DefaultBootstrapPeers []multiaddr.Multiaddr

func init() {
	for _, s := range []string{
		"/dnsaddr/bootstrap.libp2p.io/ipfs/QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN",
		"/dnsaddr/bootstrap.libp2p.io/ipfs/QmQCU2EcMqAqQPR2i9bChDtGNJchTbq5TbXJJ16u19uLTa",
		"/dnsaddr/bootstrap.libp2p.io/ipfs/QmbLHAnMoJPWSCR5Zhtx6BHJX9KiKNN6tpvbUcqanj75Nb",
		"/dnsaddr/bootstrap.libp2p.io/ipfs/QmcZf59bWwK5XFi76CZX8cbJ4BhTzzA3gU1ZjYZcYW3dwt",
		"/ip4/104.131.131.82/tcp/4001/ipfs/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",            // mars.i.ipfs.io
		"/ip4/104.236.179.241/tcp/4001/ipfs/QmSoLPppuBtQSGwKDZT2M73ULpjvfd3aZ6ha4oFGL1KrGM",           // pluto.i.ipfs.io
		"/ip4/128.199.219.111/tcp/4001/ipfs/QmSoLSafTMBsPKadTEgaXctDQVcqN88CNLHXMkTNwMKPnu",           // saturn.i.ipfs.io
		"/ip4/104.236.76.40/tcp/4001/ipfs/QmSoLV4Bbm51jM9C4gDYZQ9Cy3U6aXMJDAbzgu2fzaDs64",             // venus.i.ipfs.io
		"/ip4/178.62.158.247/tcp/4001/ipfs/QmSoLer265NRgSp2LA3dPaeykiS1J6DifTC88f5uVQKNAd",            // earth.i.ipfs.io
		"/ip6/2604:a880:1:20::203:d001/tcp/4001/ipfs/QmSoLPppuBtQSGwKDZT2M73ULpjvfd3aZ6ha4oFGL1KrGM",  // pluto.i.ipfs.io
		"/ip6/2400:6180:0:d0::151:6001/tcp/4001/ipfs/QmSoLSafTMBsPKadTEgaXctDQVcqN88CNLHXMkTNwMKPnu",  // saturn.i.ipfs.io
		"/ip6/2604:a880:800:10::4a:5001/tcp/4001/ipfs/QmSoLV4Bbm51jM9C4gDYZQ9Cy3U6aXMJDAbzgu2fzaDs64", // venus.i.ipfs.io
		"/ip6/2a03:b0c0:0:1010::23:1001/tcp/4001/ipfs/QmSoLer265NRgSp2LA3dPaeykiS1J6DifTC88f5uVQKNAd", // earth.i.ipfs.io
	} {
		ma, err := multiaddr.NewMultiaddr(s)
		if err != nil {
			panic(err)
		}
		DefaultBootstrapPeers = append(DefaultBootstrapPeers, ma)
	}
}

// BootstrapConfig specifies parameters used bootstrapping the DHT.
//
// Note there is a tradeoff between the bootstrap period and the
// number of queries. We could support a higher period with less
// queries.
type BootstrapConfig struct {
	Queries int           // how many queries to run per period
	Period  time.Duration // how often to run periodic bootstrap.
	Timeout time.Duration // how long to wait for a bootstrap query to run
}

var DefaultBootstrapConfig = BootstrapConfig{
	// For now, this is set to 1 query.
	// We are currently more interested in ensuring we have a properly formed
	// DHT than making sure our dht minimizes traffic. Once we are more certain
	// of our implementation's robustness, we should lower this down to 8 or 4.
	Queries: 1,

	// For now, this is set to 5 minutes, which is a medium period. We are
	// We are currently more interested in ensuring we have a properly formed
	// DHT than making sure our dht minimizes traffic.
	Period: time.Duration(5 * time.Minute),

	Timeout: time.Duration(10 * time.Second),
}

// A method in the IpfsRouting interface. It calls BootstrapWithConfig with
// the default bootstrap config.
func (dht *IpfsDHT) Bootstrap(ctx context.Context) error {
	return dht.BootstrapWithConfig(ctx, DefaultBootstrapConfig)
}

// Runs cfg.Queries bootstrap queries every cfg.Period.
func (dht *IpfsDHT) BootstrapWithConfig(ctx context.Context, cfg BootstrapConfig) error {
	// Because this method is not synchronous, we have to duplicate sanity
	// checks on the config so that callers aren't oblivious.
	if cfg.Queries <= 0 {
		return fmt.Errorf("invalid number of queries: %d", cfg.Queries)
	}
	go func() {
		for {
			err := dht.runBootstrap(ctx, cfg)
			if err != nil {
				logger.Warningf("error bootstrapping: %s", err)
			}
			select {
			case <-time.After(cfg.Period):
			case <-ctx.Done():
				return
			}
		}
	}()
	return nil
}

// This is a synchronous bootstrap. cfg.Queries queries will run each with a
// timeout of cfg.Timeout. cfg.Period is not used.
func (dht *IpfsDHT) BootstrapOnce(ctx context.Context, cfg BootstrapConfig) error {
	if cfg.Queries <= 0 {
		return fmt.Errorf("invalid number of queries: %d", cfg.Queries)
	}
	return dht.runBootstrap(ctx, cfg)
}

func newRandomPeerId() peer.ID {
	id := make([]byte, 32) // SHA256 is the default. TODO: Use a more canonical way to generate random IDs.
	rand.Read(id)
	id = u.Hash(id) // TODO: Feed this directly into the multihash instead of hashing it.
	return peer.ID(id)
}

// Traverse the DHT toward the given ID.
func (dht *IpfsDHT) walk(ctx context.Context, target peer.ID) (peer.AddrInfo, error) {
	// TODO: Extract the query action (traversal logic?) inside FindPeer,
	// don't actually call through the FindPeer machinery, which can return
	// things out of the peer store etc.
	return dht.FindPeer(ctx, target)
}

// Traverse the DHT toward a random ID.
func (dht *IpfsDHT) randomWalk(ctx context.Context) error {
	id := newRandomPeerId()
	p, err := dht.walk(ctx, id)
	switch err {
	case routing.ErrNotFound:
		return nil
	case nil:
		// We found a peer from a randomly generated ID. This should be very
		// unlikely.
		logger.Warningf("random walk toward %s actually found peer: %s", id, p)
		return nil
	default:
		return err
	}
}

// Traverse the DHT toward the self ID
func (dht *IpfsDHT) selfWalk(ctx context.Context) error {
	_, err := dht.walk(ctx, dht.self)
	if err == routing.ErrNotFound {
		return nil
	}
	return err
}

// runBootstrap builds up list of peers by requesting random peer IDs
func (dht *IpfsDHT) runBootstrap(ctx context.Context, cfg BootstrapConfig) error {
	doQuery := func(n int, target string, f func(context.Context) error) error {
		logger.Infof("starting bootstrap query (%d/%d) to %s (routing table size was %d)",
			n, cfg.Queries, target, dht.routingTable.Size())
		defer func() {
			logger.Infof("finished bootstrap query (%d/%d) to %s (routing table size is now %d)",
				n, cfg.Queries, target, dht.routingTable.Size())
		}()
		queryCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
		defer cancel()
		err := f(queryCtx)
		if err == context.DeadlineExceeded && queryCtx.Err() == context.DeadlineExceeded && ctx.Err() == nil {
			return nil
		}
		return err
	}

	// Do all but one of the bootstrap queries as random walks.
	for i := 0; i < cfg.Queries; i++ {
		err := doQuery(i, "random ID", dht.randomWalk)
		if err != nil {
			return err
		}
	}

	// Find self to distribute peer info to our neighbors.
	return doQuery(cfg.Queries, fmt.Sprintf("self: %s", dht.self), dht.selfWalk)
}

func (dht *IpfsDHT) BootstrapRandom(ctx context.Context) error {
	return dht.randomWalk(ctx)
}

func (dht *IpfsDHT) BootstrapSelf(ctx context.Context) error {
	return dht.selfWalk(ctx)
}

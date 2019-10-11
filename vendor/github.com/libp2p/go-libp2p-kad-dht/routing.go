package dht

import (
	"bytes"
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-core/routing"

	cid "github.com/ipfs/go-cid"
	u "github.com/ipfs/go-ipfs-util"
	logging "github.com/ipfs/go-log"
	pb "github.com/libp2p/go-libp2p-kad-dht/pb"
	kb "github.com/libp2p/go-libp2p-kbucket"
	record "github.com/libp2p/go-libp2p-record"
)

// asyncQueryBuffer is the size of buffered channels in async queries. This
// buffer allows multiple queries to execute simultaneously, return their
// results and continue querying closer peers. Note that different query
// results will wait for the channel to drain.
var asyncQueryBuffer = 10

// This file implements the Routing interface for the IpfsDHT struct.

// Basic Put/Get

// PutValue adds value corresponding to given Key.
// This is the top level "Store" operation of the DHT
func (dht *IpfsDHT) PutValue(ctx context.Context, key string, value []byte, opts ...routing.Option) (err error) {
	eip := logger.EventBegin(ctx, "PutValue")
	defer func() {
		eip.Append(loggableKey(key))
		if err != nil {
			eip.SetError(err)
		}
		eip.Done()
	}()
	logger.Debugf("PutValue %s", key)

	// don't even allow local users to put bad values.
	if err := dht.Validator.Validate(key, value); err != nil {
		return err
	}

	old, err := dht.getLocal(key)
	if err != nil {
		// Means something is wrong with the datastore.
		return err
	}

	// Check if we have an old value that's not the same as the new one.
	if old != nil && !bytes.Equal(old.GetValue(), value) {
		// Check to see if the new one is better.
		i, err := dht.Validator.Select(key, [][]byte{value, old.GetValue()})
		if err != nil {
			return err
		}
		if i != 0 {
			return fmt.Errorf("can't replace a newer value with an older value")
		}
	}

	rec := record.MakePutRecord(key, value)
	rec.TimeReceived = u.FormatRFC3339(time.Now())
	err = dht.putLocal(key, rec)
	if err != nil {
		return err
	}

	pchan, err := dht.GetClosestPeers(ctx, key)
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	for p := range pchan {
		wg.Add(1)
		go func(p peer.ID) {
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()
			defer wg.Done()
			routing.PublishQueryEvent(ctx, &routing.QueryEvent{
				Type: routing.Value,
				ID:   p,
			})

			err := dht.putValueToPeer(ctx, p, rec)
			if err != nil {
				logger.Debugf("failed putting value to peer: %s", err)
			}
		}(p)
	}
	wg.Wait()
	return nil
}

// RecvdVal stores a value and the peer from which we got the value.
type RecvdVal struct {
	Val  []byte
	From peer.ID
}

// GetValue searches for the value corresponding to given Key.
func (dht *IpfsDHT) GetValue(ctx context.Context, key string, opts ...routing.Option) (_ []byte, err error) {
	eip := logger.EventBegin(ctx, "GetValue")
	defer func() {
		eip.Append(loggableKey(key))
		if err != nil {
			eip.SetError(err)
		}
		eip.Done()
	}()

	// apply defaultQuorum if relevant
	var cfg routing.Options
	if err := cfg.Apply(opts...); err != nil {
		return nil, err
	}
	opts = append(opts, Quorum(getQuorum(&cfg, defaultQuorum)))

	responses, err := dht.SearchValue(ctx, key, opts...)
	if err != nil {
		return nil, err
	}
	var best []byte

	for r := range responses {
		best = r
	}

	if ctx.Err() != nil {
		return best, ctx.Err()
	}

	if best == nil {
		return nil, routing.ErrNotFound
	}
	logger.Debugf("GetValue %v %v", key, best)
	return best, nil
}

func (dht *IpfsDHT) SearchValue(ctx context.Context, key string, opts ...routing.Option) (<-chan []byte, error) {
	var cfg routing.Options
	if err := cfg.Apply(opts...); err != nil {
		return nil, err
	}

	responsesNeeded := 0
	if !cfg.Offline {
		responsesNeeded = getQuorum(&cfg, -1)
	}

	valCh, err := dht.getValues(ctx, key, responsesNeeded)
	if err != nil {
		return nil, err
	}

	out := make(chan []byte)
	go func() {
		defer close(out)

		maxVals := responsesNeeded
		if maxVals < 0 {
			maxVals = defaultQuorum * 4 // we want some upper bound on how
			// much correctional entries we will send
		}

		// vals is used collect entries we got so far and send corrections to peers
		// when we exit this function
		vals := make([]RecvdVal, 0, maxVals)
		var best *RecvdVal

		defer func() {
			if len(vals) <= 1 || best == nil {
				return
			}
			fixupRec := record.MakePutRecord(key, best.Val)
			for _, v := range vals {
				// if someone sent us a different 'less-valid' record, lets correct them
				if !bytes.Equal(v.Val, best.Val) {
					go func(v RecvdVal) {
						if v.From == dht.self {
							err := dht.putLocal(key, fixupRec)
							if err != nil {
								logger.Error("Error correcting local dht entry:", err)
							}
							return
						}
						ctx, cancel := context.WithTimeout(dht.Context(), time.Second*30)
						defer cancel()
						err := dht.putValueToPeer(ctx, v.From, fixupRec)
						if err != nil {
							logger.Debug("Error correcting DHT entry: ", err)
						}
					}(v)
				}
			}
		}()

		for {
			select {
			case v, ok := <-valCh:
				if !ok {
					return
				}

				if len(vals) < maxVals {
					vals = append(vals, v)
				}

				if v.Val == nil {
					continue
				}
				// Select best value
				if best != nil {
					if bytes.Equal(best.Val, v.Val) {
						continue
					}
					sel, err := dht.Validator.Select(key, [][]byte{best.Val, v.Val})
					if err != nil {
						logger.Warning("Failed to select dht key: ", err)
						continue
					}
					if sel != 1 {
						continue
					}
				}
				best = &v
				select {
				case out <- v.Val:
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return out, nil
}

// GetValues gets nvals values corresponding to the given key.
func (dht *IpfsDHT) GetValues(ctx context.Context, key string, nvals int) (_ []RecvdVal, err error) {
	eip := logger.EventBegin(ctx, "GetValues")

	eip.Append(loggableKey(key))
	defer eip.Done()

	valCh, err := dht.getValues(ctx, key, nvals)
	if err != nil {
		eip.SetError(err)
		return nil, err
	}

	out := make([]RecvdVal, 0, nvals)
	for val := range valCh {
		out = append(out, val)
	}

	return out, ctx.Err()
}

func (dht *IpfsDHT) getValues(ctx context.Context, key string, nvals int) (<-chan RecvdVal, error) {
	vals := make(chan RecvdVal, 1)

	done := func(err error) (<-chan RecvdVal, error) {
		defer close(vals)
		return vals, err
	}

	// If we have it local, don't bother doing an RPC!
	lrec, err := dht.getLocal(key)
	if err != nil {
		// something is wrong with the datastore.
		return done(err)
	}
	if lrec != nil {
		// TODO: this is tricky, we don't always want to trust our own value
		// what if the authoritative source updated it?
		logger.Debug("have it locally")
		vals <- RecvdVal{
			Val:  lrec.GetValue(),
			From: dht.self,
		}

		if nvals == 0 || nvals == 1 {
			return done(nil)
		}

		nvals--
	} else if nvals == 0 {
		return done(routing.ErrNotFound)
	}

	// get closest peers in the routing table
	rtp := dht.routingTable.NearestPeers(kb.ConvertKey(key), AlphaValue)
	logger.Debugf("peers in rt: %d %s", len(rtp), rtp)
	if len(rtp) == 0 {
		logger.Warning("No peers from routing table!")
		return done(kb.ErrLookupFailure)
	}

	var valslock sync.Mutex
	var got int

	// setup the Query
	parent := ctx
	query := dht.newQuery(key, func(ctx context.Context, p peer.ID) (*dhtQueryResult, error) {
		routing.PublishQueryEvent(parent, &routing.QueryEvent{
			Type: routing.SendingQuery,
			ID:   p,
		})

		rec, peers, err := dht.getValueOrPeers(ctx, p, key)
		switch err {
		case routing.ErrNotFound:
			// in this case, they responded with nothing,
			// still send a notification so listeners can know the
			// request has completed 'successfully'
			routing.PublishQueryEvent(parent, &routing.QueryEvent{
				Type: routing.PeerResponse,
				ID:   p,
			})
			return nil, err
		default:
			return nil, err

		case nil, errInvalidRecord:
			// in either of these cases, we want to keep going
		}

		res := &dhtQueryResult{closerPeers: peers}

		if rec.GetValue() != nil || err == errInvalidRecord {
			rv := RecvdVal{
				Val:  rec.GetValue(),
				From: p,
			}
			valslock.Lock()
			select {
			case vals <- rv:
			case <-ctx.Done():
				valslock.Unlock()
				return nil, ctx.Err()
			}
			got++

			// If we have collected enough records, we're done
			if nvals == got {
				res.success = true
			}
			valslock.Unlock()
		}

		routing.PublishQueryEvent(parent, &routing.QueryEvent{
			Type:      routing.PeerResponse,
			ID:        p,
			Responses: peers,
		})

		return res, nil
	})

	go func() {
		reqCtx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()

		_, err = query.Run(reqCtx, rtp)

		// We do have some values but we either ran out of peers to query or
		// searched for a whole minute.
		//
		// We'll just call this a success.
		if got > 0 && (err == routing.ErrNotFound || reqCtx.Err() == context.DeadlineExceeded) {
			err = nil
		}
		done(err)
	}()

	return vals, nil
}

// Provider abstraction for indirect stores.
// Some DHTs store values directly, while an indirect store stores pointers to
// locations of the value, similarly to Coral and Mainline DHT.

// Provide makes this node announce that it can provide a value for the given key
func (dht *IpfsDHT) Provide(ctx context.Context, key cid.Cid, brdcst bool) (err error) {
	eip := logger.EventBegin(ctx, "Provide", key, logging.LoggableMap{"broadcast": brdcst})
	defer func() {
		if err != nil {
			eip.SetError(err)
		}
		eip.Done()
	}()

	// add self locally
	dht.providers.AddProvider(ctx, key, dht.self)
	if !brdcst {
		return nil
	}

	closerCtx := ctx
	if deadline, ok := ctx.Deadline(); ok {
		now := time.Now()
		timeout := deadline.Sub(now)

		if timeout < 0 {
			// timed out
			return context.DeadlineExceeded
		} else if timeout < 10*time.Second {
			// Reserve 10% for the final put.
			deadline = deadline.Add(-timeout / 10)
		} else {
			// Otherwise, reserve a second (we'll already be
			// connected so this should be fast).
			deadline = deadline.Add(-time.Second)
		}
		var cancel context.CancelFunc
		closerCtx, cancel = context.WithDeadline(ctx, deadline)
		defer cancel()
	}

	peers, err := dht.GetClosestPeers(closerCtx, key.KeyString())
	if err != nil {
		return err
	}

	mes, err := dht.makeProvRecord(key)
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	for p := range peers {
		wg.Add(1)
		go func(p peer.ID) {
			defer wg.Done()
			logger.Debugf("putProvider(%s, %s)", key, p)
			err := dht.sendMessage(ctx, p, mes)
			if err != nil {
				logger.Debug(err)
			}
		}(p)
	}
	wg.Wait()
	return nil
}
func (dht *IpfsDHT) makeProvRecord(skey cid.Cid) (*pb.Message, error) {
	pi := peer.AddrInfo{
		ID:    dht.self,
		Addrs: dht.host.Addrs(),
	}

	// // only share WAN-friendly addresses ??
	// pi.Addrs = addrutil.WANShareableAddrs(pi.Addrs)
	if len(pi.Addrs) < 1 {
		return nil, fmt.Errorf("no known addresses for self. cannot put provider.")
	}

	pmes := pb.NewMessage(pb.Message_ADD_PROVIDER, skey.Bytes(), 0)
	pmes.ProviderPeers = pb.RawPeerInfosToPBPeers([]peer.AddrInfo{pi})
	return pmes, nil
}

// FindProviders searches until the context expires.
func (dht *IpfsDHT) FindProviders(ctx context.Context, c cid.Cid) ([]peer.AddrInfo, error) {
	var providers []peer.AddrInfo
	for p := range dht.FindProvidersAsync(ctx, c, KValue) {
		providers = append(providers, p)
	}
	return providers, nil
}

// FindProvidersAsync is the same thing as FindProviders, but returns a channel.
// Peers will be returned on the channel as soon as they are found, even before
// the search query completes.
func (dht *IpfsDHT) FindProvidersAsync(ctx context.Context, key cid.Cid, count int) <-chan peer.AddrInfo {
	logger.Event(ctx, "findProviders", key)
	peerOut := make(chan peer.AddrInfo, count)
	go dht.findProvidersAsyncRoutine(ctx, key, count, peerOut)
	return peerOut
}

func (dht *IpfsDHT) findProvidersAsyncRoutine(ctx context.Context, key cid.Cid, count int, peerOut chan peer.AddrInfo) {
	defer logger.EventBegin(ctx, "findProvidersAsync", key).Done()
	defer close(peerOut)

	ps := peer.NewLimitedSet(count)
	provs := dht.providers.GetProviders(ctx, key)
	for _, p := range provs {
		// NOTE: Assuming that this list of peers is unique
		if ps.TryAdd(p) {
			pi := dht.peerstore.PeerInfo(p)
			select {
			case peerOut <- pi:
			case <-ctx.Done():
				return
			}
		}

		// If we have enough peers locally, don't bother with remote RPC
		// TODO: is this a DOS vector?
		if ps.Size() >= count {
			return
		}
	}

	peers := dht.routingTable.NearestPeers(kb.ConvertKey(key.KeyString()), AlphaValue)
	if len(peers) == 0 {
		routing.PublishQueryEvent(ctx, &routing.QueryEvent{
			Type:  routing.QueryError,
			Extra: kb.ErrLookupFailure.Error(),
		})
		return
	}

	// setup the Query
	parent := ctx
	query := dht.newQuery(key.KeyString(), func(ctx context.Context, p peer.ID) (*dhtQueryResult, error) {
		routing.PublishQueryEvent(parent, &routing.QueryEvent{
			Type: routing.SendingQuery,
			ID:   p,
		})
		pmes, err := dht.findProvidersSingle(ctx, p, key)
		if err != nil {
			return nil, err
		}

		logger.Debugf("%d provider entries", len(pmes.GetProviderPeers()))
		provs := pb.PBPeersToPeerInfos(pmes.GetProviderPeers())
		logger.Debugf("%d provider entries decoded", len(provs))

		// Add unique providers from request, up to 'count'
		for _, prov := range provs {
			if prov.ID != dht.self {
				dht.peerstore.AddAddrs(prov.ID, prov.Addrs, peerstore.TempAddrTTL)
			}
			logger.Debugf("got provider: %s", prov)
			if ps.TryAdd(prov.ID) {
				logger.Debugf("using provider: %s", prov)
				select {
				case peerOut <- *prov:
				case <-ctx.Done():
					logger.Debug("context timed out sending more providers")
					return nil, ctx.Err()
				}
			}
			if ps.Size() >= count {
				logger.Debugf("got enough providers (%d/%d)", ps.Size(), count)
				return &dhtQueryResult{success: true}, nil
			}
		}

		// Give closer peers back to the query to be queried
		closer := pmes.GetCloserPeers()
		clpeers := pb.PBPeersToPeerInfos(closer)
		logger.Debugf("got closer peers: %d %s", len(clpeers), clpeers)

		routing.PublishQueryEvent(parent, &routing.QueryEvent{
			Type:      routing.PeerResponse,
			ID:        p,
			Responses: clpeers,
		})
		return &dhtQueryResult{closerPeers: clpeers}, nil
	})

	_, err := query.Run(ctx, peers)
	if err != nil {
		logger.Debugf("Query error: %s", err)
		// Special handling for issue: https://github.com/ipfs/go-ipfs/issues/3032
		if fmt.Sprint(err) == "<nil>" {
			logger.Error("reproduced bug 3032:")
			logger.Errorf("Errors type information: %#v", err)
			logger.Errorf("go version: %s", runtime.Version())
			logger.Error("please report this information to: https://github.com/ipfs/go-ipfs/issues/3032")

			// replace problematic error with something that won't crash the daemon
			err = fmt.Errorf("<nil>")
		}
		routing.PublishQueryEvent(ctx, &routing.QueryEvent{
			Type:  routing.QueryError,
			Extra: err.Error(),
		})
	}
}

// FindPeer searches for a peer with given ID.
func (dht *IpfsDHT) FindPeer(ctx context.Context, id peer.ID) (_ peer.AddrInfo, err error) {
	eip := logger.EventBegin(ctx, "FindPeer", id)
	defer func() {
		if err != nil {
			eip.SetError(err)
		}
		eip.Done()
	}()

	// Check if were already connected to them
	if pi := dht.FindLocal(id); pi.ID != "" {
		return pi, nil
	}

	peers := dht.routingTable.NearestPeers(kb.ConvertPeerID(id), AlphaValue)
	if len(peers) == 0 {
		return peer.AddrInfo{}, kb.ErrLookupFailure
	}

	// Sanity...
	for _, p := range peers {
		if p == id {
			logger.Debug("found target peer in list of closest peers...")
			return dht.peerstore.PeerInfo(p), nil
		}
	}

	// setup the Query
	parent := ctx
	query := dht.newQuery(string(id), func(ctx context.Context, p peer.ID) (*dhtQueryResult, error) {
		routing.PublishQueryEvent(parent, &routing.QueryEvent{
			Type: routing.SendingQuery,
			ID:   p,
		})

		pmes, err := dht.findPeerSingle(ctx, p, id)
		if err != nil {
			return nil, err
		}

		closer := pmes.GetCloserPeers()
		clpeerInfos := pb.PBPeersToPeerInfos(closer)

		// see if we got the peer here
		for _, npi := range clpeerInfos {
			if npi.ID == id {
				return &dhtQueryResult{
					peer:    npi,
					success: true,
				}, nil
			}
		}

		routing.PublishQueryEvent(parent, &routing.QueryEvent{
			Type:      routing.PeerResponse,
			ID:        p,
			Responses: clpeerInfos,
		})

		return &dhtQueryResult{closerPeers: clpeerInfos}, nil
	})

	// run it!
	result, err := query.Run(ctx, peers)
	if err != nil {
		return peer.AddrInfo{}, err
	}

	logger.Debugf("FindPeer %v %v", id, result.success)
	if result.peer.ID == "" {
		return peer.AddrInfo{}, routing.ErrNotFound
	}

	return *result.peer, nil
}

// FindPeersConnectedToPeer searches for peers directly connected to a given peer.
func (dht *IpfsDHT) FindPeersConnectedToPeer(ctx context.Context, id peer.ID) (<-chan *peer.AddrInfo, error) {

	peerchan := make(chan *peer.AddrInfo, asyncQueryBuffer)
	peersSeen := make(map[peer.ID]struct{})
	var peersSeenMx sync.Mutex

	peers := dht.routingTable.NearestPeers(kb.ConvertPeerID(id), AlphaValue)
	if len(peers) == 0 {
		return nil, kb.ErrLookupFailure
	}

	// setup the Query
	query := dht.newQuery(string(id), func(ctx context.Context, p peer.ID) (*dhtQueryResult, error) {

		pmes, err := dht.findPeerSingle(ctx, p, id)
		if err != nil {
			return nil, err
		}

		var clpeers []*peer.AddrInfo
		closer := pmes.GetCloserPeers()
		for _, pbp := range closer {
			pi := pb.PBPeerToPeerInfo(pbp)

			// skip peers already seen
			peersSeenMx.Lock()
			if _, found := peersSeen[pi.ID]; found {
				peersSeenMx.Unlock()
				continue
			}
			peersSeen[pi.ID] = struct{}{}
			peersSeenMx.Unlock()

			// if peer is connected, send it to our client.
			if pb.Connectedness(pbp.Connection) == network.Connected {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case peerchan <- pi:
				}
			}

			// if peer is the peer we're looking for, don't bother querying it.
			// TODO maybe query it?
			if pb.Connectedness(pbp.Connection) != network.Connected {
				clpeers = append(clpeers, pi)
			}
		}

		return &dhtQueryResult{closerPeers: clpeers}, nil
	})

	// run it! run it asynchronously to gen peers as results are found.
	// this does no error checking
	go func() {
		if _, err := query.Run(ctx, peers); err != nil {
			logger.Debug(err)
		}

		// close the peerchan channel when done.
		close(peerchan)
	}()

	return peerchan, nil
}

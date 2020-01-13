package dht

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/libp2p/go-libp2p-core/routing"

	"go.opencensus.io/tag"
	"golang.org/x/xerrors"

	"github.com/libp2p/go-libp2p-kad-dht/metrics"
	opts "github.com/libp2p/go-libp2p-kad-dht/opts"
	pb "github.com/libp2p/go-libp2p-kad-dht/pb"
	providers "github.com/libp2p/go-libp2p-kad-dht/providers"

	proto "github.com/gogo/protobuf/proto"
	cid "github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log"
	goprocess "github.com/jbenet/goprocess"
	goprocessctx "github.com/jbenet/goprocess/context"
	kb "github.com/libp2p/go-libp2p-kbucket"
	record "github.com/libp2p/go-libp2p-record"
	recpb "github.com/libp2p/go-libp2p-record/pb"
	base32 "github.com/whyrusleeping/base32"
)

var logger = logging.Logger("dht")

// NumBootstrapQueries defines the number of random dht queries to do to
// collect members of the routing table.
const NumBootstrapQueries = 5

// IpfsDHT is an implementation of Kademlia with S/Kademlia modifications.
// It is used to implement the base Routing module.
type IpfsDHT struct {
	host      host.Host           // the network services we need
	self      peer.ID             // Local peer (yourself)
	peerstore peerstore.Peerstore // Peer Registry

	datastore ds.Datastore // Local data

	routingTable *kb.RoutingTable // Array of routing tables for differently distanced nodes
	providers    *providers.ProviderManager

	birth time.Time // When this peer started up

	Validator record.Validator

	ctx  context.Context
	proc goprocess.Process

	strmap map[peer.ID]*messageSender
	smlk   sync.Mutex

	plk sync.Mutex

	stripedPutLocks [256]sync.Mutex

	protocols []protocol.ID // DHT protocols
}

// Assert that IPFS assumptions about interfaces aren't broken. These aren't a
// guarantee, but we can use them to aid refactoring.
var (
	_ routing.ContentRouting = (*IpfsDHT)(nil)
	_ routing.Routing        = (*IpfsDHT)(nil)
	_ routing.PeerRouting    = (*IpfsDHT)(nil)
	_ routing.PubKeyFetcher  = (*IpfsDHT)(nil)
	_ routing.ValueStore     = (*IpfsDHT)(nil)
)

// New creates a new DHT with the specified host and options.
func New(ctx context.Context, h host.Host, options ...opts.Option) (*IpfsDHT, error) {
	var cfg opts.Options
	if err := cfg.Apply(append([]opts.Option{opts.Defaults}, options...)...); err != nil {
		return nil, err
	}
	dht := makeDHT(ctx, h, cfg.Datastore, cfg.Protocols)

	// register for network notifs.
	dht.host.Network().Notify((*netNotifiee)(dht))

	dht.proc = goprocessctx.WithContextAndTeardown(ctx, func() error {
		// remove ourselves from network notifs.
		dht.host.Network().StopNotify((*netNotifiee)(dht))
		return nil
	})

	dht.proc.AddChild(dht.providers.Process())
	dht.Validator = cfg.Validator

	if !cfg.Client {
		for _, p := range cfg.Protocols {
			h.SetStreamHandler(p, dht.handleNewStream)
		}
	}
	return dht, nil
}

// NewDHT creates a new DHT object with the given peer as the 'local' host.
// IpfsDHT's initialized with this function will respond to DHT requests,
// whereas IpfsDHT's initialized with NewDHTClient will not.
func NewDHT(ctx context.Context, h host.Host, dstore ds.Batching) *IpfsDHT {
	dht, err := New(ctx, h, opts.Datastore(dstore))
	if err != nil {
		panic(err)
	}
	return dht
}

// NewDHTClient creates a new DHT object with the given peer as the 'local'
// host. IpfsDHT clients initialized with this function will not respond to DHT
// requests. If you need a peer to respond to DHT requests, use NewDHT instead.
// NewDHTClient creates a new DHT object with the given peer as the 'local' host
func NewDHTClient(ctx context.Context, h host.Host, dstore ds.Batching) *IpfsDHT {
	dht, err := New(ctx, h, opts.Datastore(dstore), opts.Client(true))
	if err != nil {
		panic(err)
	}
	return dht
}

func makeDHT(ctx context.Context, h host.Host, dstore ds.Batching, protocols []protocol.ID) *IpfsDHT {
	rt := kb.NewRoutingTable(KValue, kb.ConvertPeerID(h.ID()), time.Minute, h.Peerstore())

	cmgr := h.ConnManager()
	rt.PeerAdded = func(p peer.ID) {
		cmgr.TagPeer(p, "kbucket", 5)
	}
	rt.PeerRemoved = func(p peer.ID) {
		cmgr.UntagPeer(p, "kbucket")
	}

	dht := &IpfsDHT{
		datastore:    dstore,
		self:         h.ID(),
		peerstore:    h.Peerstore(),
		host:         h,
		strmap:       make(map[peer.ID]*messageSender),
		ctx:          ctx,
		providers:    providers.NewProviderManager(ctx, h.ID(), dstore),
		birth:        time.Now(),
		routingTable: rt,
		protocols:    protocols,
	}

	dht.ctx = dht.newContextWithLocalTags(ctx)

	return dht
}

// putValueToPeer stores the given key/value pair at the peer 'p'
func (dht *IpfsDHT) putValueToPeer(ctx context.Context, p peer.ID, rec *recpb.Record) error {

	pmes := pb.NewMessage(pb.Message_PUT_VALUE, rec.Key, 0)
	pmes.Record = rec
	rpmes, err := dht.sendRequest(ctx, p, pmes)
	if err != nil {
		logger.Debugf("putValueToPeer: %v. (peer: %s, key: %s)", err, p.Pretty(), loggableKey(string(rec.Key)))
		return err
	}

	if !bytes.Equal(rpmes.GetRecord().Value, pmes.GetRecord().Value) {
		logger.Warningf("putValueToPeer: value not put correctly. (%v != %v)", pmes, rpmes)
		return errors.New("value not put correctly")
	}

	return nil
}

var errInvalidRecord = errors.New("received invalid record")

// getValueOrPeers queries a particular peer p for the value for
// key. It returns either the value or a list of closer peers.
// NOTE: It will update the dht's peerstore with any new addresses
// it finds for the given peer.
func (dht *IpfsDHT) getValueOrPeers(ctx context.Context, p peer.ID, key string) (*recpb.Record, []*peer.AddrInfo, error) {

	pmes, err := dht.getValueSingle(ctx, p, key)
	if err != nil {
		return nil, nil, err
	}

	// Perhaps we were given closer peers
	peers := pb.PBPeersToPeerInfos(pmes.GetCloserPeers())

	if record := pmes.GetRecord(); record != nil {
		// Success! We were given the value
		logger.Debug("getValueOrPeers: got value")

		// make sure record is valid.
		err = dht.Validator.Validate(string(record.GetKey()), record.GetValue())
		if err != nil {
			logger.Info("Received invalid record! (discarded)")
			// return a sentinal to signify an invalid record was received
			err = errInvalidRecord
			record = new(recpb.Record)
		}
		return record, peers, err
	}

	if len(peers) > 0 {
		logger.Debug("getValueOrPeers: peers")
		return nil, peers, nil
	}

	logger.Warning("getValueOrPeers: routing.ErrNotFound")
	return nil, nil, routing.ErrNotFound
}

// getValueSingle simply performs the get value RPC with the given parameters
func (dht *IpfsDHT) getValueSingle(ctx context.Context, p peer.ID, key string) (*pb.Message, error) {
	meta := logging.LoggableMap{
		"key":  key,
		"peer": p,
	}

	eip := logger.EventBegin(ctx, "getValueSingle", meta)
	defer eip.Done()

	pmes := pb.NewMessage(pb.Message_GET_VALUE, []byte(key), 0)
	resp, err := dht.sendRequest(ctx, p, pmes)
	switch err {
	case nil:
		return resp, nil
	case ErrReadTimeout:
		logger.Warningf("getValueSingle: read timeout %s %s", p.Pretty(), key)
		fallthrough
	default:
		eip.SetError(err)
		return nil, err
	}
}

// getLocal attempts to retrieve the value from the datastore
func (dht *IpfsDHT) getLocal(key string) (*recpb.Record, error) {
	logger.Debugf("getLocal %s", key)
	rec, err := dht.getRecordFromDatastore(mkDsKey(key))
	if err != nil {
		logger.Warningf("getLocal: %s", err)
		return nil, err
	}

	// Double check the key. Can't hurt.
	if rec != nil && string(rec.GetKey()) != key {
		logger.Errorf("BUG getLocal: found a DHT record that didn't match it's key: %s != %s", rec.GetKey(), key)
		return nil, nil

	}
	return rec, nil
}

// putLocal stores the key value pair in the datastore
func (dht *IpfsDHT) putLocal(key string, rec *recpb.Record) error {
	logger.Debugf("putLocal: %v %v", key, rec)
	data, err := proto.Marshal(rec)
	if err != nil {
		logger.Warningf("putLocal: %s", err)
		return err
	}

	return dht.datastore.Put(mkDsKey(key), data)
}

// Update signals the routingTable to Update its last-seen status
// on the given peer.
func (dht *IpfsDHT) Update(ctx context.Context, p peer.ID) {
	logger.Event(ctx, "updatePeer", p)
	dht.routingTable.Update(p)
}

// FindLocal looks for a peer with a given ID connected to this dht and returns the peer and the table it was found in.
func (dht *IpfsDHT) FindLocal(id peer.ID) peer.AddrInfo {
	switch dht.host.Network().Connectedness(id) {
	case network.Connected, network.CanConnect:
		return dht.peerstore.PeerInfo(id)
	default:
		return peer.AddrInfo{}
	}
}

// findPeerSingle asks peer 'p' if they know where the peer with id 'id' is
func (dht *IpfsDHT) findPeerSingle(ctx context.Context, p peer.ID, id peer.ID) (*pb.Message, error) {
	eip := logger.EventBegin(ctx, "findPeerSingle",
		logging.LoggableMap{
			"peer":   p,
			"target": id,
		})
	defer eip.Done()

	pmes := pb.NewMessage(pb.Message_FIND_NODE, []byte(id), 0)
	resp, err := dht.sendRequest(ctx, p, pmes)
	switch err {
	case nil:
		return resp, nil
	case ErrReadTimeout:
		logger.Warningf("read timeout: %s %s", p.Pretty(), id)
		fallthrough
	default:
		eip.SetError(err)
		return nil, err
	}
}

func (dht *IpfsDHT) findProvidersSingle(ctx context.Context, p peer.ID, key cid.Cid) (*pb.Message, error) {
	eip := logger.EventBegin(ctx, "findProvidersSingle", p, key)
	defer eip.Done()

	pmes := pb.NewMessage(pb.Message_GET_PROVIDERS, key.Bytes(), 0)
	resp, err := dht.sendRequest(ctx, p, pmes)
	switch err {
	case nil:
		return resp, nil
	case ErrReadTimeout:
		logger.Warningf("read timeout: %s %s", p.Pretty(), key)
		fallthrough
	default:
		eip.SetError(err)
		return nil, err
	}
}

// nearestPeersToQuery returns the routing tables closest peers.
func (dht *IpfsDHT) nearestPeersToQuery(pmes *pb.Message, count int) []peer.ID {
	closer := dht.routingTable.NearestPeers(kb.ConvertKey(string(pmes.GetKey())), count)
	return closer
}

// betterPeersToQuery returns nearestPeersToQuery, but if and only if closer than self.
func (dht *IpfsDHT) betterPeersToQuery(pmes *pb.Message, p peer.ID, count int) []peer.ID {
	closer := dht.nearestPeersToQuery(pmes, count)

	// no node? nil
	if closer == nil {
		logger.Warning("betterPeersToQuery: no closer peers to send:", p)
		return nil
	}

	filtered := make([]peer.ID, 0, len(closer))
	for _, clp := range closer {

		// == to self? thats bad
		if clp == dht.self {
			logger.Error("BUG betterPeersToQuery: attempted to return self! this shouldn't happen...")
			return nil
		}
		// Dont send a peer back themselves
		if clp == p {
			continue
		}

		filtered = append(filtered, clp)
	}

	// ok seems like closer nodes
	return filtered
}

// Context return dht's context
func (dht *IpfsDHT) Context() context.Context {
	return dht.ctx
}

// Process return dht's process
func (dht *IpfsDHT) Process() goprocess.Process {
	return dht.proc
}

// RoutingTable return dht's routingTable
func (dht *IpfsDHT) RoutingTable() *kb.RoutingTable {
	return dht.routingTable
}

// Close calls Process Close
func (dht *IpfsDHT) Close() error {
	return dht.proc.Close()
}

func (dht *IpfsDHT) protocolStrs() []string {
	pstrs := make([]string, len(dht.protocols))
	for idx, proto := range dht.protocols {
		pstrs[idx] = string(proto)
	}

	return pstrs
}

func mkDsKey(s string) ds.Key {
	return ds.NewKey(base32.RawStdEncoding.EncodeToString([]byte(s)))
}

func (dht *IpfsDHT) PeerID() peer.ID {
	return dht.self
}

func (dht *IpfsDHT) PeerKey() []byte {
	return kb.ConvertPeerID(dht.self)
}

func (dht *IpfsDHT) Host() host.Host {
	return dht.host
}

func (dht *IpfsDHT) Ping(ctx context.Context, p peer.ID) error {
	req := pb.NewMessage(pb.Message_PING, nil, 0)
	resp, err := dht.sendRequest(ctx, p, req)
	if err != nil {
		return xerrors.Errorf("sending request: %w", err)
	}
	if resp.Type != pb.Message_PING {
		return xerrors.Errorf("got unexpected response type: %v", resp.Type)
	}
	return nil
}

// newContextWithLocalTags returns a new context.Context with the InstanceID and
// PeerID keys populated. It will also take any extra tags that need adding to
// the context as tag.Mutators.
func (dht *IpfsDHT) newContextWithLocalTags(ctx context.Context, extraTags ...tag.Mutator) context.Context {
	extraTags = append(
		extraTags,
		tag.Upsert(metrics.KeyPeerID, dht.self.Pretty()),
		tag.Upsert(metrics.KeyInstanceID, fmt.Sprintf("%p", dht)),
	)
	ctx, _ = tag.New(
		ctx,
		extraTags...,
	) // ignoring error as it is unrelated to the actual function of this code.
	return ctx
}

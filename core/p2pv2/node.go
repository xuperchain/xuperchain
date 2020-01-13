package p2pv2

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	//"path/filepath"
	"strings"

	iaddr "github.com/ipfs/go-ipfs-addr"
	"github.com/libp2p/go-libp2p"
	circuit "github.com/libp2p/go-libp2p-circuit"
	crypto "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	net "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	log "github.com/xuperchain/log15"

	"github.com/xuperchain/xuperunion/common/config"
	p2pPb "github.com/xuperchain/xuperunion/p2pv2/pb"

	"github.com/xuperchain/xuperunion/kv/kvdb"
)

// define the common config
const (
	XuperProtocolID    = "/xuper/2.0.0" // protocol version
	P2PMultiAddrPrefix = "p2pMulti_"
)

var (
	// MaxBroadCastPeers define the maximum number of common peers to broadcast messages
	MaxBroadCastPeers = 20
	// MaxBroadCastCorePeers define the maximum number of core peers to broadcast messages
	MaxBroadCastCorePeers = 10
)

// define errors
var (
	ErrGenerateOpts     = errors.New("generate host opts error")
	ErrCreateHost       = errors.New("create host error")
	ErrCreateKadDht     = errors.New("create kad dht error")
	ErrCreateStreamPool = errors.New("create stream pool error")
	ErrCreateBootStrap  = errors.New("create bootstrap error pool error")
	ErrConnectBootStrap = errors.New("error to connect to all bootstrap")
	ErrConnectCorePeers = errors.New("error to connect to all core peers")
	ErrInvalidParams    = errors.New("invalid params")
)

type corePeerInfo struct {
	Distance  int
	PeerAddr  string
	PeerInfo  *pstore.PeerInfo
	XuperAddr string
}

type corePeersRoute struct {
	CurrentPeers []*corePeerInfo
	NextPeers    []*corePeerInfo
}

// Node is the node in the network
type Node struct {
	id          peer.ID
	privKey     crypto.PrivKey
	log         log.Logger
	host        host.Host
	kdht        *dht.IpfsDHT
	strPool     *StreamPool
	ctx         context.Context
	srv         *P2PServerV2
	quitCh      chan bool
	addrs       map[string]*XchainAddrInfo
	coreRoute   map[string]*corePeersRoute
	staticNodes map[string][]peer.ID
	routeLock   sync.RWMutex
	// StreamLimit
	streamLimit *StreamLimit
	// ldb persist peers info and get peers info
	ldb kvdb.Database
	// isStorePeers determine whether open isStorePeers
	isStorePeers bool
	p2pDataPath  string
}

// NewNode define the node of the xuper, it will set streamHandler for this node.
func NewNode(cfg config.P2PConfig, log log.Logger) (*Node, error) {
	ctx := context.Background()
	opts, err := genHostOption(cfg)
	if err != nil {
		log.Error("genHostOption error!", "error", err.Error())
		return nil, ErrGenerateOpts
	}

	ho, err := libp2p.New(ctx, opts...)
	if err != nil {
		log.Error("Create libp2p host error!", "error", err.Error())
		return nil, ErrCreateHost
	}
	no := &Node{
		id:        ho.ID(),
		log:       log,
		ctx:       ctx,
		host:      ho,
		quitCh:    make(chan bool, 1),
		addrs:     map[string]*XchainAddrInfo{},
		coreRoute: make(map[string]*corePeersRoute),
		// new StreamLimit
		streamLimit:  &StreamLimit{},
		isStorePeers: cfg.IsStorePeers,
		p2pDataPath:  cfg.P2PDataPath,
	}
	if no.isStorePeers {
		no.ldb, err = newBaseDB(no.p2pDataPath)
		if err != nil {
			return nil, err
		}
	}

	// set broadcast peers limitation
	MaxBroadCastPeers = cfg.MaxBroadcastPeers
	MaxBroadCastCorePeers = cfg.MaxBroadcastCorePeers

	// initialize StreamLimit, set limit size
	no.streamLimit.Init(cfg.StreamIPLimitSize, log)

	if no.kdht, err = dht.New(ctx, ho); err != nil {
		return nil, ErrCreateKadDht
	}

	if no.strPool, err = NewStreamPool(cfg.MaxStreamLimits, no, log); err != nil {
		return nil, ErrCreateStreamPool
	}

	if !cfg.IsHidden {
		if err = no.kdht.Bootstrap(ctx); err != nil {
			return nil, ErrCreateBootStrap
		}
	}

	// connect to peers stored last time recently
	// connect to bootNodes
	peers := []string{}
	if no.isStorePeers {
		peers, err = no.getPeersFromDisk()
		if err != nil {
			no.log.Warn("getPeersFromDisk error", "err", err)
		}
	}
	if len(cfg.BootNodes) > 0 {
		peers = append(peers, cfg.BootNodes...)
	}
	for _, ps := range cfg.StaticNodes {
		peers = append(peers, ps...)
	}

	succNum := no.ConnectToPeersByAddr(peers)
	if succNum == 0 && len(cfg.BootNodes) != 0 {
		return nil, ErrConnectBootStrap
	}

	// setup static nodes
	setStaticNodes(cfg, no)
	return no, nil
}

func setStaticNodes(cfg config.P2PConfig, node *Node) error {
	staticNodes := map[string][]peer.ID{}
	for bcname, peers := range cfg.StaticNodes {
		ps := []peer.ID{}
		for _, peer := range peers {
			id, err := GetIDFromAddr(peer)
			if err != nil {
				continue
			}
			ps = append(ps, id)
		}
		staticNodes[bcname] = ps
	}
	node.staticNodes = staticNodes
	return nil
}

func genHostOption(cfg config.P2PConfig) ([]libp2p.Option, error) {
	muAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", cfg.Port))
	opts := []libp2p.Option{
		libp2p.ListenAddrs(muAddr),
	}
	if cfg.IsNat {
		opts = append(opts, libp2p.NATPortMap())
	}
	if cfg.IsSecure {
		opts = append(opts, libp2p.DefaultSecurity)
	}
	opts = append(opts, libp2p.EnableRelay(circuit.OptHop))

	priv, err := GetKeyPairFromPath(cfg.KeyPath)
	if err != nil {
		return nil, err
	}
	opts = append(opts, libp2p.Identity(priv))
	return opts, nil
}

// Start start the node
func (no *Node) Start() {
	no.log.Trace("Start node")
	no.host.SetStreamHandler(XuperProtocolID, no.handlerNewStream)
	t := time.NewTicker(time.Duration(time.Second * 30))
	defer t.Stop()
	for {
		select {
		case <-no.quitCh:
			no.Stop()
			return
		case <-t.C:
			no.log.Trace("RoutingTable", "size", no.kdht.RoutingTable().Size())
			no.kdht.RoutingTable().Print()
			if no.isStorePeers {
				ret := no.persistPeersToDisk()
				if !ret {
					log.Warn("persistPeersToDisk failed")
				}
			}
		}
	}
}

// Stop stop the node
func (no *Node) Stop() {
	fmt.Println("Stop node")
	no.strPool.quitCh <- true
	no.kdht.Close()
	no.host.Close()
}

// handlerNewStream parse message type and process message by handlerForMsgType
func (no *Node) handlerNewStream(s net.Stream) {
	no.strPool.Add(s)
}

// NodeID return the node ID
func (no *Node) NodeID() peer.ID {
	return no.id
}

// Context return the node context
func (no *Node) Context() context.Context {
	return no.ctx
}

// SetServer set the p2p server of the node
func (no *Node) SetServer(srv *P2PServerV2) {
	no.srv = srv
}

// SendMessage send message to given peers
func (no *Node) SendMessage(ctx context.Context, msg *p2pPb.XuperMessage, peers []peer.ID) error {
	return no.strPool.SendMessage(ctx, msg, peers)
}

// SendMessageWithResponse send message to given peers, expecting response from peers
func (no *Node) SendMessageWithResponse(ctx context.Context, msg *p2pPb.XuperMessage, peers []peer.ID, percentage float32) ([]*p2pPb.XuperMessage, error) {
	return no.strPool.SendMessageWithResponse(ctx, msg, peers, percentage)
}

// ListPeers return the list of peer ID in routing table
func (no *Node) ListPeers() []peer.ID {
	// get peer list from kdht routing table
	rt := no.kdht.RoutingTable()
	peers := rt.ListPeers()
	idmap := make(map[string]bool)
	for _, pid := range peers {
		idmap[pid.Pretty()] = true
	}

	no.routeLock.RLock()
	defer no.routeLock.RUnlock()
	for _, coreRoute := range no.coreRoute {
		for _, cpi := range coreRoute.CurrentPeers {
			pid := cpi.PeerInfo.ID
			if _, ok := idmap[pid.Pretty()]; ok {
				continue
			}
			peers = append(peers, pid)
			idmap[pid.Pretty()] = true
		}
		for _, cpi := range coreRoute.NextPeers {
			pid := cpi.PeerInfo.ID
			if _, ok := idmap[pid.Pretty()]; ok {
				continue
			}
			peers = append(peers, pid)
			idmap[pid.Pretty()] = true
		}
	}
	return peers
}

// UpdateCorePeers update core peers' info and keep connection to core peers
func (no *Node) UpdateCorePeers(cp *CorePeersInfo) error {
	if cp == nil {
		return ErrInvalidParams
	}
	no.routeLock.Lock()
	defer no.routeLock.Unlock()
	var oldInfo *corePeersRoute
	if _, ok := no.coreRoute[cp.Name]; ok {
		oldInfo = no.coreRoute[cp.Name]
	}

	// update connections
	newInfo, err := no.updateCoreConnection(oldInfo, cp)
	if err != nil {
		return err
	}

	// update routing table
	no.coreRoute[cp.Name] = newInfo
	return nil
}

// updateCoreConnection update direct connections to core peers.
// this function remove out-of-date core peers and create connections to new peers
func (no *Node) updateCoreConnection(oldInfo *corePeersRoute,
	newInfo *CorePeersInfo) (*corePeersRoute, error) {
	newCurrentPeers := make([]*corePeerInfo, 0)
	newNextPeers := make([]*corePeerInfo, 0)
	allPeers := make([]*pstore.PeerInfo, 0)
	processedPeers := make(map[string]bool)
	if oldInfo != nil {
		for _, pr := range oldInfo.CurrentPeers {
			processedPeers[pr.PeerAddr] = true
		}
		for _, pr := range oldInfo.NextPeers {
			processedPeers[pr.PeerAddr] = true
		}
	}
	for _, paddr := range newInfo.CurrentPeerIDs {
		if paddr == "" || processedPeers[paddr] {
			continue
		}
		newPeer, err := no.getRoutePeerFromAddr(paddr)
		if err != nil {
			no.log.Warn("parse peer address failed, ignore this one", "peer", paddr, "error", err)
			continue
		}
		newCurrentPeers = append(newCurrentPeers, newPeer)
		allPeers = append(allPeers, newPeer.PeerInfo)
	}
	for _, paddr := range newInfo.NextPeerIDs {
		if paddr == "" || processedPeers[paddr] {
			continue
		}
		newPeer, err := no.getRoutePeerFromAddr(paddr)
		if err != nil {
			no.log.Warn("parse peer address failed, ignore this one", "peer", paddr, "error", err)
			continue
		}
		newNextPeers = append(newNextPeers, newPeer)
		allPeers = append(allPeers, newPeer.PeerInfo)
	}

	// connect new peers
	succNum := no.createPeerStream(allPeers)
	if len(allPeers) != 0 && succNum == 0 {
		return nil, ErrConnectCorePeers
	}
	newRoute := &corePeersRoute{
		CurrentPeers: newCurrentPeers,
		NextPeers:    newNextPeers,
	}

	return newRoute, nil
}

func (no *Node) getRoutePeerFromAddr(peerAddr string) (*corePeerInfo, error) {
	addr, err := iaddr.ParseString(peerAddr)
	if err != nil {
		no.log.Error("Parse peer address error!", "peerAddr", peerAddr, "error", err.Error())
		return nil, err
	}
	peerinfo, err := pstore.InfoFromP2pAddr(addr.Multiaddr())
	if err != nil {
		no.log.Error("Get peer node info error!", "peerAddr", peerAddr, "error", err.Error())
		return nil, err
	}
	cpi := &corePeerInfo{
		PeerAddr: peerAddr,
		PeerInfo: peerinfo,
		Distance: 0, // TODO: calc the distance between peers
	}
	return cpi, nil
}

// ConnectToPeersByAddr provide connection support using peer address(netURL)
func (no *Node) ConnectToPeersByAddr(addrs []string) int {
	peers := make([]*pstore.PeerInfo, 0)
	for _, addr := range addrs {
		pi, err := no.getRoutePeerFromAddr(addr)
		if err != nil {
			continue
		}
		peers = append(peers, pi.PeerInfo)
	}
	return no.connectToPeers(peers)
}

// connectToPeers connect to given peers, return the connected number of peers
func (no *Node) connectToPeers(ppi []*pstore.PeerInfo) int {
	// empty slice, do nothing
	ppiSize := len(ppi)
	if ppiSize <= 0 {
		return 0
	}

	// connect to bootNodes
	succNum := 0
	retryCount := 5
	for retryCount > 0 {
		for _, pi := range ppi {
			if err := no.host.Connect(no.ctx, *pi); err != nil {
				no.log.Error("Connection with peer node error!", "error", err.Error())
			} else {
				succNum++
				no.log.Info("Connection established with peer node, ", "nodeInfo", *pi)
			}
		}
		if succNum > 0 {
			break
		}
		// only retry if all connection failed
		retryCount--
		num := rand.Int63n(10)
		time.Sleep(time.Duration(num) * time.Second)
	}
	return succNum
}

// createPeerStream create stream to given peers, return the connected number of peers
func (no *Node) createPeerStream(ppi []*pstore.PeerInfo) int {
	succNum := 0
	maxSleepMS := 1000
	rand.Seed(time.Now().Unix())
	for _, pi := range ppi {
		retries := 3
		for retries > 0 {
			_, err := no.strPool.streamForPeer(pi.ID)
			if err == nil {
				succNum++
				break
			}
			no.log.Warn("create stream for peer failed", "peer", pi.ID.Pretty(), "error", err)
			time.Sleep(time.Duration(rand.Intn(maxSleepMS)) * time.Millisecond)
			retries--
		}
	}
	return succNum
}

// persistPeersToDisk persist peers connecting to each other to disk
func (no *Node) persistPeersToDisk() bool {
	batch := no.ldb.NewBatch()
	prefix := no.GetP2PMultiAddrPrefix()
	it := no.ldb.NewIteratorWithPrefix([]byte(prefix))
	defer it.Release()
	// delete history records before
	for it.Next() {
		batch.Delete(it.Key())
	}
	if it.Error() != nil {
		return false
	}
	peers := no.streamLimit.GetStreams()
	// persist recent records after
	for _, peer := range peers {
		batch.Put([]byte(prefix+peer), []byte("true"))
	}
	writeErr := batch.Write()
	if writeErr != nil {
		log.Warn("p2p module, persistPeersToDisk error", "err", writeErr)
		return false
	}
	return true
}

// getPeersFromDisk get peers from disk
func (no *Node) getPeersFromDisk() ([]string, error) {
	peers := []string{}
	prefix := no.GetP2PMultiAddrPrefix()
	it := no.ldb.NewIteratorWithPrefix([]byte(prefix))
	defer it.Release()
	for it.Next() {
		key := string(it.Key())
		key = strings.TrimPrefix(key, prefix)
		peers = append(peers, key)
	}
	if it.Error() != nil {
		return nil, it.Error()
	}
	return peers, nil
}

func newBaseDB(dbPath string) (kvdb.Database, error) {
	// new kv instance
	kvParam := &kvdb.KVParameter{
		DBPath:                dbPath,
		KVEngineType:          "default",
		MemCacheSize:          128,
		FileHandlersCacheSize: 512,
		OtherPaths:            []string{},
	}
	return kvdb.NewKVDBInstance(kvParam)
}

// GetP2PMultiAddrPrefix return P2PMultiAddrPrefix
func (no *Node) GetP2PMultiAddrPrefix() string {
	return P2PMultiAddrPrefix
}

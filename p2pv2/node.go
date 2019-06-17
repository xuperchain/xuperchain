package p2pv2

import (
	"context"
	"fmt"
	"time"

	iaddr "github.com/ipfs/go-ipfs-addr"
	"github.com/libp2p/go-libp2p"
	circuit "github.com/libp2p/go-libp2p-circuit"
	crypto "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
	net "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	dht "github.com/xuperchain/go-libp2p-kad-dht"
	log "github.com/xuperchain/log15"

	"github.com/xuperchain/xuperunion/common/config"
	p2pPb "github.com/xuperchain/xuperunion/p2pv2/pb"
)

// define the common config
const (
	XuperProtocolID   = "/xuper/2.0.0" // protocol version
	MaxBroadCastPeers = 20             // the maximum peers to broadcast messages
)

// define errors
var (
	ErrGenerateOpts     = errors.New("generate host opts error")
	ErrCreateHost       = errors.New("create host error")
	ErrCreateKadDht     = errors.New("create kad dht error")
	ErrCreateStreamPool = errors.New("create stream pool error")
	ErrCreateBootStrap  = errors.New("create bootstrap error pool error")
	ErrConnectBootStrap = errors.New("error to connect to all bootstrap")
)

// Node is the node in the network
type Node struct {
	id      peer.ID
	privKey crypto.PrivKey
	log     log.Logger
	host    host.Host
	kdht    *dht.IpfsDHT
	strPool *StreamPool
	ctx     context.Context
	srv     *P2PServerV2
	quitCh  chan bool
	addrs   map[string]*XchainAddrInfo
	// StreamLimit
	streamLimit *StreamLimit
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
		id:     ho.ID(),
		log:    log,
		ctx:    ctx,
		host:   ho,
		quitCh: make(chan bool, 1),
		addrs:  map[string]*XchainAddrInfo{},
		// new StreamLimit
		streamLimit: &StreamLimit{},
	}
	// initialize StreamLimit, set limit size
	no.streamLimit.Init(cfg.StreamIPLimitSize, log)
	ho.SetStreamHandler(XuperProtocolID, no.handlerNewStream)

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

	// connect to bootNodes
	succNum := 0
	retryCount := 5
	for retryCount > 0 {
		for _, peerAddr := range cfg.BootNodes {
			addr, err := iaddr.ParseString(peerAddr)
			if err != nil {
				log.Error("Parse boot node address error!", "bootnode", peerAddr, "error", err.Error())
				continue
			}
			peerinfo, err := pstore.InfoFromP2pAddr(addr.Multiaddr())
			if err != nil {
				log.Error("Get boot node info error!", "bootnode", peerAddr, "error", err.Error())
				continue
			}

			if err := no.host.Connect(ctx, *peerinfo); err != nil {
				log.Error("Connection with bootstrap node error!", "error", err.Error())
			} else {
				succNum++
				log.Info("Connection established with bootstrap node, ", "nodeInfo", *peerinfo)
			}
		}
		if len(cfg.BootNodes) != 0 && succNum == 0 {
			retryCount--
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}
	if len(cfg.BootNodes) != 0 && succNum == 0 {
		return nil, ErrConnectBootStrap
	}
	return no, nil
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
func (no *Node) SendMessageWithResponse(ctx context.Context, msg *p2pPb.XuperMessage, peers []peer.ID, withBreak bool) ([]*p2pPb.XuperMessage, error) {
	return no.strPool.SendMessageWithResponse(ctx, msg, peers, withBreak)
}

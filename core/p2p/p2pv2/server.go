// Package p2pv2 is the v2 of XuperChain p2p network.
package p2pv2

import (
	"context"
	"fmt"
	"os"
	"strings"

	peer "github.com/libp2p/go-libp2p-peer"
	"github.com/pkg/errors"
	log "github.com/xuperchain/log15"

	"github.com/xuperchain/xuperchain/core/common/config"
	p2p_base "github.com/xuperchain/xuperchain/core/p2p/base"
	p2pPb "github.com/xuperchain/xuperchain/core/p2p/pb"
)

// define errors
var (
	ErrValidateConfig   = errors.New("config not valid")
	ErrCreateNode       = errors.New("create node error")
	ErrCreateHandlerMap = errors.New("create handlerMap error")
)

// make sure p2pv2 implemented the P2PServer interface
var _ p2p_base.P2PServer = (*P2PServerV2)(nil)

// P2PServerV2 is the v2 of XuperChain p2p server. An implement of P2PServer interface.
type P2PServerV2 struct {
	log log.Logger
	// config is the p2p v2 设置
	config     config.P2PConfig
	node       *Node
	handlerMap *p2p_base.HandlerMap
	quitCh     chan bool
}

// NewP2PServerV2 create P2PServerV2 instance
func NewP2PServerV2() *P2PServerV2 {
	return &P2PServerV2{}
}

// Init initialize p2p server using given config
func (p *P2PServerV2) Init(cfg config.P2PConfig, lg log.Logger, extra map[string]interface{}) error {
	if lg == nil {
		lg = log.New("module", "p2pv2")
		lg.SetHandler(log.StreamHandler(os.Stderr, log.LogfmtFormat()))
	}

	no, err := NewNode(cfg, lg)
	if err != nil {
		lg.Trace("NewP2PServerV2 create node error", "error", err)
		return ErrCreateNode
	}

	hm, err := p2p_base.NewHandlerMap(lg)
	if err != nil {
		lg.Trace("NewP2PServerV2 new handler map error", "errors", err)
		return ErrCreateHandlerMap
	}

	// set p2p server members
	p.log = lg
	p.config = cfg
	p.node = no
	p.handlerMap = hm
	p.quitCh = make(chan bool, 1)

	no.SetServer(p)

	if err := p.registerSubscriber(); err != nil {
		return err
	}

	go p.Start()
	return nil
}

// Start start P2P server V2
func (p *P2PServerV2) Start() {
	p.log.Info("Start p2pv2 server!")
	go p.node.Start()
	go p.handlerMap.Start()
	for {
		select {
		case <-p.quitCh:
			p.Stop()
			p.log.Info("P2pv2 server have stopped!")
			return
		}
	}
}

// Stop stop P2P server V2
func (p *P2PServerV2) Stop() {
	p.log.Info("Stop p2pv2 server!")
	p.node.Stop()
	p.handlerMap.Stop()
}

// SendMessage send message to peers using given filter strategy
func (p *P2PServerV2) SendMessage(ctx context.Context, msg *p2pPb.XuperMessage,
	opts ...p2p_base.MessageOption) error {
	msgOpts := p2p_base.GetMessageOption(opts)
	filter := p.getFilter(msgOpts)
	peers, _ := filter.Filter()
	// 还要做一层过滤(使用白名单过滤)
	peersRes := []peer.ID{}
	whiteList := msgOpts.WhiteList
	if len(whiteList) > 0 {
		for _, v := range peers.([]peer.ID) {
			if _, exist := whiteList[v.Pretty()]; exist {
				peersRes = append(peersRes, v)
			}
		}
	} else {
		peersRes = peers.([]peer.ID)
	}
	// 是否需要经过压缩,针对由本节点产生的消息以及grpc获取的信息
	if needCompress := p.getCompress(msgOpts); needCompress {
		// 更新MsgInfo & Header.enableCompress
		// 重新计算CheckSum
		enableCompress := msg.Header.EnableCompress
		// msg原本没有被压缩
		if !enableCompress {
			msg = p2p_base.Compress(msg)
		}
	}
	p.log.Trace("Server SendMessage", "logid", msg.GetHeader().GetLogid(), "msgType", msg.GetHeader().GetType(), "checksum", msg.GetHeader().GetDataCheckSum())
	return p.node.SendMessage(ctx, msg, peersRes)
}

// SendMessageWithResponse send message to peers using given filter strategy, expect response from peers
// 客户端再使用该方法请求带返回的消息时，最好带上log_id, 否则会导致收消息时收到不匹配的消息而影响后续的处理
func (p *P2PServerV2) SendMessageWithResponse(ctx context.Context, msg *p2pPb.XuperMessage,
	opts ...p2p_base.MessageOption) ([]*p2pPb.XuperMessage, error) {
	msgOpts := p2p_base.GetMessageOption(opts)
	filter := p.getFilter(msgOpts)
	peers, _ := filter.Filter()
	peersRes := []peer.ID{}
	// 做一层过滤(基于白名单过滤)
	whiteList := msgOpts.WhiteList
	if len(whiteList) > 0 {
		for _, v := range peers.([]peer.ID) {
			if _, exist := whiteList[v.Pretty()]; exist {
				peersRes = append(peersRes, v)
			}
		}
	} else {
		peersRes = peers.([]peer.ID)
	}
	percentage := msgOpts.Percentage
	p.log.Trace("Server SendMessage with response", "logid", msg.GetHeader().GetLogid(),
		"msgType", msg.GetHeader().GetType(), "checksum", msg.GetHeader().GetDataCheckSum(), "peers", peers)
	return p.node.SendMessageWithResponse(ctx, msg, peersRes, percentage)
}

// NewSubscriber create a subscriber instance
func (p *P2PServerV2) NewSubscriber(msgCh chan *p2pPb.XuperMessage, msgType p2pPb.XuperMessage_MessageType, handler p2p_base.XuperHandler, msgFrom string, log log.Logger) p2p_base.Subscriber {
	return NewMsgSubscriber(msgCh, msgType, handler, msgFrom, log)
}

// Register register message subscribers to handle messages
func (p *P2PServerV2) Register(sub p2p_base.Subscriber) (p2p_base.Subscriber, error) {
	return p.handlerMap.Register(sub)
}

// UnRegister remove message subscribers
func (p *P2PServerV2) UnRegister(sub p2p_base.Subscriber) error {
	return p.handlerMap.UnRegister(sub)
}

// GetNetURL return net url of the xuper node
// url = /ip4/127.0.0.1/tcp/<port>/p2p/<peer.Id>
func (p *P2PServerV2) GetNetURL() string {
	return fmt.Sprintf("/ip4/127.0.0.1/tcp/%v/p2p/%s", p.config.Port, p.node.id.Pretty())
}

func (p *P2PServerV2) getCompress(opts *p2p_base.MsgOptions) bool {
	if opts == nil {
		return false
	}
	return opts.Compress
}

func (p *P2PServerV2) getFilter(opts *p2p_base.MsgOptions) p2p_base.PeersFilter {
	// All filtering strategies will invalid if
	if len(p.node.GetStaticNodes(opts.Bcname)) != 0 {
		return &StaticNodeStrategy{node: p.node, bcname: opts.Bcname}
	}
	fs := opts.Filters
	bcname := opts.Bcname
	peerids := make([]peer.ID, 0)
	tpaLen := len(opts.TargetPeerAddrs)
	tpiLen := len(opts.TargetPeerIDs)
	if len(fs) == 0 && tpaLen == 0 && tpiLen == 0 {
		return &BucketsFilter{node: p.node}
	}
	pfs := make([]p2p_base.PeersFilter, 0)
	for _, f := range fs {
		var filter p2p_base.PeersFilter
		switch f {
		case p2p_base.NearestBucketStrategy:
			filter = &NearestBucketFilter{node: p.node}
		case p2p_base.BucketsStrategy:
			filter = &BucketsFilter{node: p.node}
		case p2p_base.BucketsWithFactorStrategy:
			filter = &BucketsFilterWithFactor{node: p.node}
		case p2p_base.CorePeersStrategy:
			filter = &CorePeersFilter{node: p.node, name: bcname}
		default:
			filter = &BucketsFilter{node: p.node}
		}
		pfs = append(pfs, filter)
	}
	// process target peer addresses
	if tpaLen > 0 {
		// connect to extra target peers async
		go p.node.ConnectToPeersByAddr(opts.TargetPeerAddrs)
		// get corresponding peer ids
		for _, addr := range opts.TargetPeerAddrs {
			pid, err := p2p_base.GetIDFromAddr(addr)
			if err != nil {
				p.log.Warn("getFilter parse peer address failed", "paddr", addr, "error", err)
				continue
			}
			peerids = append(peerids, pid)
		}
	}

	// process target peer IDs
	if tpiLen > 0 {
		for _, tpid := range opts.TargetPeerIDs {
			peerid, err := peer.IDB58Decode(tpid)
			if err != nil {
				p.log.Warn("getFilter parse peer ID failed", "pid", tpid, "error", err)
				continue
			}
			peerids = append(peerids, peerid)
		}
	}
	return NewMultiStrategy(pfs, peerids)
}

// GetPeerUrls 查询所连接节点的信息
func (p *P2PServerV2) GetPeerUrls() []string {
	urls := []string{}

	// 获取路由表中节点的信息
	//rt := p.node.kdht.RoutingTable()
	//peers := rt.ListPeers()
	peers := p.node.ListPeers()
	for _, v := range peers {
		if s, err := p.node.strPool.FindStream(v); err == nil {
			if s.gp == "" {
				s.getRPCPort()
			}
			addrSli := strings.Split(s.addr.String(), "/")
			if len(addrSli) < 3 {
				continue
			}
			url := addrSli[2] + s.gp
			urls = append(urls, url)
		}
	}
	return urls
}

func (p *P2PServerV2) GetPeerIDAndUrls() map[string]string {
	id2Url := map[string]string{}
	peers := p.node.ListPeers()
	for _, v := range peers {
		if s, err := p.node.strPool.FindStream(v); err == nil {
			if s.gp == "" {
				s.getRPCPort()
			}
			peerID := string(v.Pretty())
			ipStr := s.addr.String() + "/p2p/" + peerID
			id2Url[peerID] = ipStr
		}
	}

	return id2Url
}

// SetCorePeers set core peers' info to P2P server
func (p *P2PServerV2) SetCorePeers(cp *p2p_base.CorePeersInfo) error {
	if cp == nil {
		return ErrInvalidParams
	}
	return p.node.UpdateCorePeers(cp)
}

// SetXchainAddr Set xchain address info from core
func (p *P2PServerV2) SetXchainAddr(bcname string, info *p2p_base.XchainAddrInfo) {
	p.node.SetXchainAddr(bcname, info)
}

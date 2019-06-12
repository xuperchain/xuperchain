// Package p2pv2 is the v2 of XuperChain p2p network.
package p2pv2

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	log "github.com/xuperchain/log15"

	"github.com/xuperchain/xuperunion/common/config"
	p2pPb "github.com/xuperchain/xuperunion/p2pv2/pb"
)

// define errors
var (
	ErrValidateConfig   = errors.New("config not valid")
	ErrCreateNode       = errors.New("create node error")
	ErrCreateHandlerMap = errors.New("create handlerMap error")
)

// P2PServerV2 is the v2 of XuperChain p2p server. An implement of P2PServer interface.
type P2PServerV2 struct {
	log log.Logger
	// config is the p2p v2 设置
	config     config.P2PConfig
	node       *Node
	handlerMap *HandlerMap
	quitCh     chan bool
}

// NewP2PServerV2 create P2PServerV2 instance
func NewP2PServerV2(cfg config.P2PConfig, lg log.Logger) (*P2PServerV2, error) {

	if lg == nil {
		lg = log.New("module", "p2pv2")
		lg.SetHandler(log.StreamHandler(os.Stderr, log.LogfmtFormat()))
	}

	no, err := NewNode(cfg, lg)
	if err != nil {
		lg.Trace("NewP2PServerV2 create node error", "error", err)
		return nil, ErrCreateNode
	}

	hm, err := NewHandlerMap(lg)
	if err != nil {
		lg.Trace("NewP2PServerV2 new handler map error", "errors", err)
		return nil, ErrCreateHandlerMap
	}

	p2pSrv := &P2PServerV2{
		log:        lg,
		config:     cfg,
		node:       no,
		handlerMap: hm,
		quitCh:     make(chan bool, 1),
	}

	no.SetServer(p2pSrv)

	if err := p2pSrv.registerSubscriber(); err != nil {
		return nil, err
	}

	go p2pSrv.Start()
	return p2pSrv, nil
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
	p.node.quitCh <- true
	p.handlerMap.quitCh <- true
}

// SendMessage send message to peers using given filter strategy
func (p *P2PServerV2) SendMessage(ctx context.Context, msg *p2pPb.XuperMessage, fs FilterStrategy) error {
	filter := p.getFilter(fs)
	peers, _ := filter.Filter()
	p.log.Trace("Server SendMessage", "logid", msg.GetHeader().GetLogid(), "msgType", msg.GetHeader().GetType(), "checksum", msg.GetHeader().GetDataCheckSum())
	return p.node.SendMessage(ctx, msg, peers)
}

// SendMessageWithResponse send message to peers using given filter strategy, expect response from peers
// 客户端再使用该方法请求带返回的消息时，最好带上log_id, 否则会导致收消息时收到不匹配的消息而影响后续的处理
func (p *P2PServerV2) SendMessageWithResponse(ctx context.Context, msg *p2pPb.XuperMessage, fs FilterStrategy, withBreak bool) ([]*p2pPb.XuperMessage,
	error) {
	filter := p.getFilter(fs)
	peers, _ := filter.Filter()
	p.log.Trace("Server SendMessage with response", "logid", msg.GetHeader().GetLogid(), "msgType", msg.GetHeader().GetType(), "checksum", msg.GetHeader().GetDataCheckSum())
	return p.node.SendMessageWithResponse(ctx, msg, peers, withBreak)
}

// Register register message subscribers to handle messages
func (p *P2PServerV2) Register(sub *Subscriber) (*Subscriber, error) {
	return p.handlerMap.Register(sub)
}

// UnRegister remove message subscribers
func (p *P2PServerV2) UnRegister(sub *Subscriber) error {
	return p.handlerMap.UnRegister(sub)
}

// GetNetURL return net url of the xuper node
// url = /ip4/127.0.0.1/tcp/<port>/p2p/<peer.Id>
func (p *P2PServerV2) GetNetURL() string {
	return fmt.Sprintf("/ip4/127.0.0.1/tcp/%v/p2p/%s", p.config.Port, p.node.id.Pretty())
}

func (p *P2PServerV2) getFilter(fs FilterStrategy) PeersFilter {
	switch fs {
	case NearestBucketStrategy:
		return &NearestBucketFilter{node: p.node}
	case BucketsStrategy:
		return &BucketsFilter{node: p.node}
	case BucketsWithFactorStrategy:
		return &BucketsFilterWithFactor{node: p.node}
	default:
		return &BucketsFilter{node: p.node}
	}
}

// GetPeerUrls 查询所连接节点的信息
func (p *P2PServerV2) GetPeerUrls() []string {
	urls := []string{}

	// 获取路由表中节点的信息
	rt := p.node.kdht.RoutingTable()
	peers := rt.ListPeers()
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

// SetXchainAddr Set xchain address info from core
func (p *P2PServerV2) SetXchainAddr(bcname string, info *XchainAddrInfo) {
	if _, ok := p.node.addrs[bcname]; !ok {
		info.PeerID = p.node.id.Pretty()
		p.node.addrs[bcname] = info
	}
}

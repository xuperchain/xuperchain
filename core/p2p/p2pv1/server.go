// Package p2pV1 is the v1 of XuperChain p2p network.
package p2pv1

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/pkg/errors"
	log "github.com/xuperchain/log15"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"

	"github.com/xuperchain/xuperchain/core/common/config"
	p2p_base "github.com/xuperchain/xuperchain/core/p2p/base"
	p2pPb "github.com/xuperchain/xuperchain/core/p2p/pb"
)

// make sure p2pv1 implemented the P2PServer interface
var _ p2p_base.P2PServer = (*P2PServerV1)(nil)

// define errors
var (
	ErrValidateConfig   = errors.New("config not valid")
	ErrCreateHandlerMap = errors.New("create handlerMap error")
	ErrInvalidParams    = errors.New("invalid params")
	ErrConnPool         = errors.New("create conn pool error")
)

// make sure p2pv1 implemented the P2PServer interface
var _ p2p_base.P2PServer = (*P2PServerV1)(nil)

// P2PServerV1 is the v1 of XuperChain p2p server. An implement of P2PServer interface.
type P2PServerV1 struct {
	log        log.Logger
	config     config.P2PConfig
	id         string
	handlerMap *p2p_base.HandlerMap
	connPool   *ConnPool
	// key: "bcname", value: "id"
	staticNodes  map[string][]string
	bootNodes    []string
	dynamicNodes []string
	msgChan      chan *p2pPb.XuperMessage
	quitCh       chan bool
	lock         sync.Mutex
	localAddr    map[string]*p2p_base.XchainAddrInfo
}

// NewP2PServerV1 create P2PServerV1 instance
func NewP2PServerV1() *P2PServerV1 {
	return &P2PServerV1{}
}

// Init initialize p2p server using given config
func (p *P2PServerV1) Init(cfg config.P2PConfig, lg log.Logger, extra map[string]interface{}) error {
	if lg == nil {
		lg = log.New("module", "p2pv1")
		lg.SetHandler(log.StreamHandler(os.Stderr, log.LogfmtFormat()))
	}

	hm, err := p2p_base.NewHandlerMap(lg)
	if err != nil {
		p.log.Error("Init P2PServerV1 NewHandlerMap error", "error", err)
		return ErrCreateHandlerMap
	}
	cp, err := NewConnPool(lg, cfg)
	if err != nil {
		p.log.Error("Init P2PServerV1 NewConnPool error", "error", err)
		return ErrConnPool
	}
	// set p2p server members
	p.log = lg
	p.config = cfg
	p.handlerMap = hm
	p.connPool = cp
	p.quitCh = make(chan bool, 1)
	p.msgChan = make(chan *p2pPb.XuperMessage, 5000)
	p.localAddr = map[string]*p2p_base.XchainAddrInfo{}

	peerids := []string{}
	hasPeerMap := map[string]bool{}
	p.bootNodes = cfg.BootNodes
	if len(cfg.StaticNodes) > 0 {
		// connect to all static nodes
		for _, peers := range cfg.StaticNodes {
			for _, peer := range peers {
				// peer address connected before
				if _, ok := hasPeerMap[peer]; ok {
					continue
				}

				conn, err := NewConn(lg, peer, cfg.CertPath, cfg.ServiceName, cfg.IsUseCert, (int)(cfg.MaxMessageSize)<<20, p.config.Timeout)
				if err != nil {
					p.log.Warn("p2p connect to peer failed", "peer", peer, "error", err)
					continue
				}
				p.connPool.Add(conn)

				hasPeerMap[peer] = true
				peerids = append(peerids, peer)
			}
		}

		p.staticNodes = cfg.StaticNodes
		// "xuper" blockchain is super set of all blockchains
		if len(p.staticNodes["xuper"]) < len(peerids) {
			p.staticNodes["xuper"] = peerids
		}
	}
	p.connectToBootNodes()
	p.registerSubscribe()
	go p.Start()
	go p.handleMsg()
	return nil
}

// ConnectToBootNodes connect to bootnode
func (p *P2PServerV1) connectToBootNodes() error {
	msg, err := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion3, "", "", p2pPb.XuperMessage_NEW_NODE, []byte{}, p2pPb.XuperMessage_NONE)
	if err != nil {
		p.log.Error("ConnectToBootNodes NewXuperMessage error", "error", err)
		return err
	}
	msg.Header.From = strconv.Itoa(int(p.config.Port))
	p.ConnectToPeersByAddr(p.bootNodes)
	opts := []p2p_base.MessageOption{
		p2p_base.WithTargetPeerAddrs(p.bootNodes),
	}
	go p.SendMessage(context.Background(), msg, opts...)
	return nil
}

// Start start P2P server V1
func (p *P2PServerV1) Start() {
	p.log.Info("Start p2pv1 server!")
	go p.handlerMap.Start()
	p.startServer()
	for {
		select {
		case <-p.quitCh:
			p.Stop()
			p.log.Info("p2pv1 server have stopped!")
			return
		}
	}
}

// Stop stop P2P server V1
func (p *P2PServerV1) Stop() {
	p.log.Info("Stop p2pv1 server!")
	p.handlerMap.Stop()
}

// SendMessage send message to peers using given filter strategy
func (p *P2PServerV1) SendMessage(ctx context.Context, msg *p2pPb.XuperMessage,
	opts ...p2p_base.MessageOption) error {
	msgOpts := p2p_base.GetMessageOption(opts)
	filter := p.getFilter(msgOpts)
	peers, _ := filter.Filter()
	p.log.Info("testlog SendMessage", "peers", peers)
	peerids, ok := peers.([]string)
	if !ok {
		p.log.Warn("p2p filter get peers failed, ignore this message",
			"logid", msg.GetHeader().GetLogid())
		return errors.New("p2p SendMessage: filter returned error data")
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
	return p.sendMessage(ctx, msg, peerids)
}

// SendMessageWithResponse send message to peers using given filter strategy, expect response from peers
// 客户端再使用该方法请求带返回的消息时，最好带上log_id, 否则会导致收消息时收到不匹配的消息而影响后续的处理
func (p *P2PServerV1) SendMessageWithResponse(ctx context.Context, msg *p2pPb.XuperMessage,
	opts ...p2p_base.MessageOption) ([]*p2pPb.XuperMessage, error) {
	msgOpts := p2p_base.GetMessageOption(opts)
	filter := p.getFilter(msgOpts)
	peers, _ := filter.Filter()
	p.log.Info("testlog SendMessageWithResponse", "peers", peers)
	percentage := msgOpts.Percentage
	peerids, ok := peers.([]string)
	if !ok {
		p.log.Warn("p2p filter get peers failed, ignore this message",
			"logid", msg.GetHeader().GetLogid())
		return nil, errors.New("p2p SendMessageWithRes: filter returned error data")
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

	return p.sendMessageWithRes(ctx, msg, peerids, percentage)
}

// NewSubscriber create a subscriber instance
func (p *P2PServerV1) NewSubscriber(msgCh chan *p2pPb.XuperMessage, msgType p2pPb.XuperMessage_MessageType,
	handler p2p_base.XuperHandler, msgFrom string, log log.Logger) p2p_base.Subscriber {
	return NewMsgSubscriber(msgCh, msgType, handler, msgFrom, p.log)
}

// Register register message subscribers to handle messages
func (p *P2PServerV1) Register(sub p2p_base.Subscriber) (p2p_base.Subscriber, error) {
	return p.handlerMap.Register(sub)
}

// UnRegister remove message subscribers
func (p *P2PServerV1) UnRegister(sub p2p_base.Subscriber) error {
	return p.handlerMap.UnRegister(sub)
}

// GetNetURL return net url of the xuper node
// url = /ip4/127.0.0.1/tcp/<port>/p2p/<peer.Id>
func (p *P2PServerV1) GetNetURL() string {
	return fmt.Sprintf("/ip4/127.0.0.1/tcp/%v", p.config.Port)
}

func (p *P2PServerV1) sendMessage(ctx context.Context, msg *p2pPb.XuperMessage, peerids []string) error {
	// send message to all peers
	p.log.Trace("Server SendMessage", "logid", msg.GetHeader().GetLogid(),
		"msgType", msg.GetHeader().GetType(), "checksum", msg.GetHeader().GetDataCheckSum())
	wg := sync.WaitGroup{}
	for _, peerid := range peerids {
		// find connection in connPool
		conn, err := p.connPool.Find(peerid)
		if err != nil {
			p.log.Warn("p2p connPool find conn failed", "logid", msg.GetHeader().GetLogid(),
				"peerid", peerid, "error", err)
			continue
		}
		// send message async
		wg.Add(1)
		go func(conn *Conn) {
			defer wg.Done()
			err = conn.SendMessage(ctx, msg)
			if err != nil {
				p.log.Error("SendMessage to peer error", "logid", msg.GetHeader().GetLogid(),
					"peerid", conn.id, "error", err)
			}
		}(conn)
	}
	wg.Wait()
	return nil
}

func (p *P2PServerV1) sendMessageWithRes(ctx context.Context, msg *p2pPb.XuperMessage,
	peerids []string, percent float32) ([]*p2pPb.XuperMessage, error) {
	// send message to all peers
	p.log.Trace("Server sendMessageWithRes", "logid", msg.GetHeader().GetLogid(),
		"msgType", msg.GetHeader().GetType(), "checksum", msg.GetHeader().GetDataCheckSum())
	conns := []*Conn{}
	for _, peerid := range peerids {
		// find connection in connPool
		conn, err := p.connPool.Find(peerid)
		if err != nil {
			p.log.Warn("p2p connPool find conn failed", "logid", msg.GetHeader().GetLogid(),
				"peerid", peerid, "error", err)
			continue
		}
		conns = append(conns, conn)
	}

	wg := sync.WaitGroup{}
	msgChan := make(chan *p2pPb.XuperMessage, len(conns))
	for _, conn := range conns {
		// send message async
		wg.Add(1)
		go func(conn *Conn) {
			defer wg.Done()
			res, err := conn.SendMessageWithResponse(ctx, msg)
			if err != nil {
				p.log.Error("sendMessageWithRes to peer error", "logid", msg.GetHeader().GetLogid(),
					"peerid", conn.id, "error", err)
			}
			msgChan <- res
		}(conn)
	}
	wg.Wait()

	res := []*p2pPb.XuperMessage{}
	lenCh := len(msgChan)
	if lenCh <= 0 {
		p.log.Warn("SendMessageWithResponse response error: lenCh is nil")
		return res, errors.New("Request get no results")
	}

	i := 0
	for r := range msgChan {
		if len(res) > int(float32(len(conns))*percent) {
			break
		}
		if p2p_base.VerifyDataCheckSum(r) {
			res = append(res, r)
		}
		if i >= lenCh-1 {
			break
		}
		i++
	}
	return res, nil
}

func (p *P2PServerV1) getCompress(opts *p2p_base.MsgOptions) bool {
	if opts == nil {
		return false
	}
	return opts.Compress
}

func (p *P2PServerV1) getFilter(opts *p2p_base.MsgOptions) p2p_base.PeersFilter {
	fs := opts.Filters
	bcname := opts.Bcname
	peerids := make([]string, 0)
	pfs := make([]p2p_base.PeersFilter, 0)
	// TODO: support other filters in feature
	for _, f := range fs {
		var filter p2p_base.PeersFilter
		switch f {
		default:
			filter = &StaticNodeStrategy{isBroadCast: p.config.IsBroadCast, pSer: p, bcname: bcname}
		}
		pfs = append(pfs, filter)
	}
	// process target peer addresses
	// connect to extra target peers async
	go p.ConnectToPeersByAddr(opts.TargetPeerAddrs)
	// get corresponding peer ids
	peerids = append(peerids, opts.TargetPeerAddrs...)
	return NewMultiStrategy(pfs, peerids)
}

// ConnectToPeersByAddr establish contact with given nodes
func (p *P2PServerV1) ConnectToPeersByAddr(addrs []string) {
	for _, peer := range addrs {
		// peer address connected before
		_, err := p.connPool.Find(peer)
		if err != nil {
			p.log.Error("ConnectToPeersByAddr error", "addr", peer, "error", err)
		}
	}
}

// GetPeerUrls 查询所连接节点的信息
func (p *P2PServerV1) GetPeerUrls() []string {
	res := []string{}
	conns, err := p.connPool.GetConns()
	if err != nil {
		p.log.Warn("p2p get peer urls failed", "error", err)
		return res
	}
	for _, conn := range conns {
		id := conn.GetConnID()
		res = append(res, id)
	}
	return res
}

func (p *P2PServerV1) GetPeerIDAndUrls() map[string]string {
	id2Url := map[string]string{}
	return id2Url
}

// SetCorePeers set core peers' info to P2P server
func (p *P2PServerV1) SetCorePeers(cp *p2p_base.CorePeersInfo) error {
	// TODO: p2pv1 only support static nodes at this time, do not support core peers
	return nil
}

// SetXchainAddr Set xchain address info from core
func (p *P2PServerV1) SetXchainAddr(bcname string, info *p2p_base.XchainAddrInfo) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.localAddr[bcname] = info
}

// startServer start p2p server
func (p *P2PServerV1) startServer() {
	options := append([]grpc.ServerOption{}, grpc.MaxRecvMsgSize(int(p.config.MaxMessageSize)<<20),
		grpc.MaxSendMsgSize(int(p.config.MaxMessageSize)<<20))
	if p.config.IsUseCert {
		creds, err := genCreds(p.config.CertPath, p.config.ServiceName)
		if err != nil {
			panic(err)
		}
		options = append(options, grpc.Creds(creds))
	}

	l, err := net.Listen("tcp", ":"+strconv.Itoa((int)(p.config.Port)))
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer(options...)
	p2pPb.RegisterP2PServiceServer(s, p)
	reflection.Register(s)
	log.Trace("start p2p rpc server")
	go func() {
		err := s.Serve(l)
		if err != nil {
			panic(err)
		}
	}()
}

// SendP2PMessage implement the SendP2PMessageServer
func (p *P2PServerV1) SendP2PMessage(str p2pPb.P2PService_SendP2PMessageServer) error {
	in, err := str.Recv()
	p.log.Trace("SendP2PMessage Recv msg", "logid", in.GetHeader().GetLogid(), "type", in.GetHeader().GetType())
	if err == io.EOF {
		p.log.Warn("SendP2PMessage Recv msg error", "error", "io.EOF")
		return nil
	}
	if err != nil {
		p.log.Warn("SendP2PMessage Recv msg error")
		return err
	}

	if !strings.Contains(in.Header.From, ":") {
		ip, _ := getRemoteIP(str.Context())
		in.Header.From = ip + ":" + in.Header.From
	}
	p.handlerMap.HandleMessage(str, in)
	return nil
}

func getRemoteIP(ctx context.Context) (string, error) {
	pr, ok := peer.FromContext(ctx)
	if ok && pr.Addr != net.Addr(nil) {
		return strings.Split(pr.Addr.String(), ":")[0], nil
	}
	return "", errors.New("Get node addr error")
}

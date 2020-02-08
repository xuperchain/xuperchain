// Package p2pV1 is the v1 of XuperChain p2p network.
package p2pv1

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"sync"

	"github.com/pkg/errors"
	log "github.com/xuperchain/log15"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
	staticNodes map[string][]string
	quitCh      chan bool
	lock        sync.Mutex
	localAddr   map[string]*p2p_base.XchainAddrInfo
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
		return ErrCreateHandlerMap
	}
	cp, err := NewConnPool(lg, cfg)
	if err != nil {
		return ErrConnPool
	}
	// set p2p server members
	p.log = lg
	p.config = cfg
	p.handlerMap = hm
	p.connPool = cp
	p.quitCh = make(chan bool, 1)

	peerids := []string{}
	hasPeerMap := map[string]bool{}

	if len(cfg.StaticNodes) > 0 {
		// connect to all static nodes
		for _, peers := range cfg.StaticNodes {
			for _, peer := range peers {
				// peer address connected before
				if _, ok := hasPeerMap[peer]; ok {
					continue
				}

				conn, err := NewConn(lg, peer, cfg.CertPath, cfg.ServiceName, (int)(cfg.MaxMessageSize))
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

	go p.Start()
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
				p.log.Error("SendMessage to peer error", "logid", msg.GetHeader().GetLogid(),
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
	// All filtering strategies will invalid if
	// TODO: support TargetPeerAddrs and TargetPeerIDs options
	return &StaticNodeStrategy{pSer: p, bcname: opts.Bcname}
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
	certPath := p.config.CertPath
	bs, err := ioutil.ReadFile(certPath + "/cacert.pem")
	if err != nil {
		panic(err)
	}
	certPool := x509.NewCertPool()
	ok := certPool.AppendCertsFromPEM(bs)
	if !ok {
		panic(errors.New("AppendCertsFromPEM error"))
	}

	certificate, err := tls.LoadX509KeyPair(certPath+"/cert.pem", certPath+"/private.key")
	if err != nil {
		panic(err)
	}
	creds := credentials.NewTLS(
		&tls.Config{
			ServerName:   p.config.ServiceName,
			Certificates: []tls.Certificate{certificate},
			RootCAs:      certPool,
			ClientCAs:    certPool,
			ClientAuth:   tls.RequireAndVerifyClientCert,
		})

	l, err := net.Listen("tcp", ":"+string(p.config.Port))
	if err != nil {
		panic(err)
	}

	// TODO: zq add other option by config
	options := append([]grpc.ServerOption{}, grpc.Creds(creds))
	s := grpc.NewServer(options...)
	p2pPb.RegisterP2PServiceServer(s, p)
	reflection.Register(s)
	log.Trace("start tls rpc server")
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
	if err == io.EOF {
		return nil
	}
	if err != nil {
		p.log.Warn("SendP2PMessage Recv msg error")
		return err
	}
	p.handlerMap.HandleMessage(str, in)
	return nil
}

// Package p2pV1 is the v1 of XuperChain p2p network.
package p2pv1

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
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
	cp, err := NewConnPool(lg)
	if err != nil {
		return ErrConnPool
	}
	// set p2p server members
	p.log = lg
	p.config = cfg
	p.handlerMap = hm
	p.connPool = cp
	p.quitCh = make(chan bool, 1)
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
	// msgOpts := p2p_base.GetMessageOption(opts)
	// filter := p.getFilter(msgOpts)
	// peers, _ := filter.Filter()
	// // 是否需要经过压缩,针对由本节点产生的消息以及grpc获取的信息
	// if needCompress := p.getCompress(msgOpts); needCompress {
	// 	// 更新MsgInfo & Header.enableCompress
	// 	// 重新计算CheckSum
	// 	enableCompress := msg.Header.EnableCompress
	// 	// msg原本没有被压缩
	// 	if !enableCompress {
	// 		msg = p2p_base.Compress(msg)
	// 	}
	// }
	// p.log.Trace("Server SendMessage", "logid", msg.GetHeader().GetLogid(), "msgType", msg.GetHeader().GetType(), "checksum", msg.GetHeader().GetDataCheckSum())
	return nil
}

// SendMessageWithResponse send message to peers using given filter strategy, expect response from peers
// 客户端再使用该方法请求带返回的消息时，最好带上log_id, 否则会导致收消息时收到不匹配的消息而影响后续的处理
func (p *P2PServerV1) SendMessageWithResponse(ctx context.Context, msg *p2pPb.XuperMessage,
	opts ...p2p_base.MessageOption) ([]*p2pPb.XuperMessage, error) {
	// msgOpts := p2p_base.GetMessageOption(opts)
	// filter := p.getFilter(msgOpts)
	// peers, _ := filter.Filter()
	// percentage := msgOpts.Percentage
	// p.log.Trace("Server SendMessage with response", "logid", msg.GetHeader().GetLogid(),
	// 	"msgType", msg.GetHeader().GetType(), "checksum", msg.GetHeader().GetDataCheckSum(), "peers", peers)
	return nil, nil
}

// NewSubscriber create a subscriber instance
func (p *P2PServerV1) NewSubscriber(msgCh chan *p2pPb.XuperMessage, msgType p2pPb.XuperMessage_MessageType, handler p2p_base.XuperHandler, msgFrom string) p2p_base.Subscriber {
	return NewMsgSubscriber(msgCh, msgType, handler, msgFrom)
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

func (p *P2PServerV1) getCompress(opts *p2p_base.MsgOptions) bool {
	if opts == nil {
		return false
	}
	return opts.Compress
}

func (p *P2PServerV1) getFilter(opts *p2p_base.MsgOptions) p2p_base.PeersFilter {
	// All filtering strategies will invalid if
	return &StaticNodeStrategy{pSer: p, bcname: opts.Bcname}
}

func (p *P2PServerV1) connectToPeers([]string) {

}

// GetPeerUrls 查询所连接节点的信息
func (p *P2PServerV1) GetPeerUrls() []string {
	urls := []string{}
	return urls
}

// SetCorePeers set core peers' info to P2P server
func (p *P2PServerV1) SetCorePeers(cp *p2p_base.CorePeersInfo) error {
	return nil
}

// SetXchainAddr Set xchain address info from core
func (p *P2PServerV1) SetXchainAddr(bcname string, info *p2p_base.XchainAddrInfo) {
}

// startServer start p2p server
func (p *P2PServerV1) startServer() {
	certPath := p.config.CertPath
	bs, err := ioutil.ReadFile(certPath + "/cert.crt")
	if err != nil {
		panic(err)
	}
	certPool := x509.NewCertPool()
	ok := certPool.AppendCertsFromPEM(bs)
	if !ok {
		panic(err)
	}

	certificate, err := tls.LoadX509KeyPair(certPath+"/key.pem", certPath+"/private.key")
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

// SendP2PMessage
func (p *P2PServerV1) SendP2PMessage(str p2pPb.P2PService_SendP2PMessageServer) error {
	return nil
}

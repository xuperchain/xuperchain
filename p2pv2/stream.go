package p2pv2

import (
	"bufio"
	"context"
	"errors"
	"io"
	"sync"
	"time"

	ggio "github.com/gogo/protobuf/io"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	ma "github.com/multiformats/go-multiaddr"

	p2pPb "github.com/xuperchain/xuperunion/p2pv2/pb"
)

// define common errors
var (
	ErrTimeout     = errors.New("request time out")
	ErrNullResult  = errors.New("request result is null")
	ErrStrNotValid = errors.New("stream not valid")
)

// Stream is the IO wrapper for underly P2P connection
type Stream struct {
	p       peer.ID
	addr    ma.Multiaddr
	s       net.Stream
	rc      ggio.ReadCloser
	w       *bufio.Writer
	wc      ggio.WriteCloser
	lk      *sync.Mutex
	node    *Node
	isvalid bool
	// Grpc port
	gp string
}

// NewStream create Stream instance
func NewStream(s net.Stream, no *Node) *Stream {
	w := bufio.NewWriter(s)
	wc := ggio.NewDelimitedWriter(w)
	maxMsgSize := (int(no.srv.config.MaxMessageSize) << 20)
	stream := &Stream{
		p:       s.Conn().RemotePeer(),
		addr:    s.Conn().RemoteMultiaddr(),
		s:       s,
		rc:      ggio.NewDelimitedReader(s, maxMsgSize),
		w:       w,
		wc:      wc,
		lk:      new(sync.Mutex),
		node:    no,
		isvalid: true,
	}
	stream.Start()
	return stream
}

// Start used to start
func (s *Stream) Start() {
	go s.readData()
	s.getRPCPort()
}

// Close close the connected IO stream
func (s *Stream) Close() {
	s.reset()
}

func (s *Stream) valid() bool {
	return s.isvalid
}

func (s *Stream) reset() {
	if s.valid() {
		s.lk.Lock()
		defer s.lk.Unlock()
		if s.s != nil {
			s.s.Reset()
		}
		s.s = nil
		s.isvalid = false
	}
	s.node.strPool.DelStream(s)
	s.node.strPool.streamLimit.DelStream(s.addr.String())
}

// readData loop to read data from stream
func (s *Stream) readData() {
	for {
		msg := new(p2pPb.XuperMessage)
		err := s.rc.ReadMsg(msg)
		switch err {
		case io.EOF:
			s.node.log.Trace("Stream readData", "error", "io.EOF")
			s.reset()
			return
		case nil:
		default:
			s.node.log.Trace("Stream readData error to reset", "error", err)
			s.reset()
			return
		}
		err = s.handlerNewMessage(msg)
		if err != nil {
			s.reset()
			return
		}
		msg = nil
	}
}

// handlerNewMessage handler new message from a peer
func (s *Stream) handlerNewMessage(msg *p2pPb.XuperMessage) error {
	if s.node.srv == nil {
		s.node.log.Warn("Stream not ready, omit", "msg", msg)
		return nil
	}
	return s.node.srv.handlerMap.HandleMessage(s, msg)
}

// getRPCPort 刚建立链接的时候获取对方的GPRC端口
func (s *Stream) getRPCPort() {
	msg, err := p2pPb.NewXuperMessage(p2pPb.XuperMsgVersion2, "", "", p2pPb.XuperMessage_GET_RPC_PORT, []byte{}, p2pPb.XuperMessage_NONE)
	if err != nil {
		return
	}
	res, err := s.SendMessageWithResponse(context.Background(), msg)
	if err != nil {
		s.node.log.Warn("getRPCPort error", "err", err, "res", res)
		return
	}
	port := string(res.GetData().GetMsgInfo())
	s.gp = port
}

// SendMessage will send a message to a peer
func (s *Stream) SendMessage(ctx context.Context, msg *p2pPb.XuperMessage) error {
	s.node.log.Trace("Stream SendMessage", "logid", msg.GetHeader().GetLogid(), "msgType", msg.GetHeader().GetType(), "checksum", msg.GetHeader().GetDataCheckSum(), "to", s.p.Pretty())
	if err := s.writeData(msg); err != nil {
		s.node.log.Trace("Stream SendMessage writeData error", "err", err.Error())
		return err
	}
	return nil
}

func (s *Stream) writeData(msg *p2pPb.XuperMessage) error {
	if !s.valid() {
		return ErrStrNotValid
	}
	s.lk.Lock()
	defer s.lk.Unlock()
	msg.Header.From = s.node.NodeID().Pretty()
	if err := s.wc.WriteMsg(msg); err != nil {
		s.reset()
		return err
	}
	return s.w.Flush()
}

// SendMessageWithResponse will send a message to a peer and wait for response
func (s *Stream) SendMessageWithResponse(ctx context.Context, msg *p2pPb.XuperMessage) (*p2pPb.XuperMessage, error) {
	s.node.log.Trace("Stream SendMessageWithResponse", "logid", msg.GetHeader().GetLogid(), "msgType", msg.GetHeader().GetType(), "checksum", msg.GetHeader().GetDataCheckSum(), "to", s.p.Pretty())
	//  todo: zq 外层的这个循环是为了将来加重试
	for {
		// 注册临时的消息订阅着
		resType := p2pPb.GetResMsgType(msg.GetHeader().GetType())
		resCh := make(chan *p2pPb.XuperMessage, 100)
		responseCh := make(chan *p2pPb.XuperMessage, 1)
		errCh := make(chan error, 1)
		sub := NewSubscriber(resCh, resType, nil, s.p.Pretty())
		sub, err := s.node.srv.Register(sub)
		if err != nil {
			s.node.log.Trace("sendMessageWithResponse register error", "error", err)
			return nil, err
		}

		// 程序结束需要注销该订阅者
		defer s.node.srv.UnRegister(sub)
		go func() {
			res, err := s.ctxWaitRes(ctx, msg, resCh)
			if res != nil {
				responseCh <- res
			}
			if err != nil {
				errCh <- err
			}
		}()

		// 开始写消息
		s.node.log.Trace("sendMessageWithResponse start to write msg", "logid", msg.GetHeader().GetLogid(), "msgType", msg.GetHeader().GetType(), "checksum", msg.GetHeader().GetDataCheckSum(), "to", s.p.Pretty())
		if err := s.writeData(msg); err != nil {
			s.node.log.Warn("sendMessageWithResponse", "err", err)
			return nil, err
		}

		// 等待返回
		select {
		case res := <-responseCh:
			return res, nil
		case err := <-errCh:
			return nil, err
		}
	}
}

// ctxWaitRes wait res with timeout
func (s *Stream) ctxWaitRes(ctx context.Context, msg *p2pPb.XuperMessage, resCh chan *p2pPb.XuperMessage) (*p2pPb.XuperMessage, error) {
	timeout := s.node.srv.config.Timeout
	t := time.NewTimer(time.Duration(timeout) * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-t.C:
			return nil, ErrTimeout
		case res := <-resCh:
			if p2pPb.VerifyMsgMatch(msg, res, s.p.Pretty()) {
				s.node.log.Trace("ctxWaitRes get res done", "type", res.GetHeader().GetType(), "logid", res.GetHeader().GetLogid(), "checksum", res.GetHeader().GetDataCheckSum(), "res.from", res.GetHeader().GetFrom(), "pid", s.p.Pretty())
				return res, nil
			}
			s.node.log.Trace("ctxWaitRes get res continue", "type", res.GetHeader().GetType(), "logid", res.GetHeader().GetLogid(), "checksum", res.GetHeader().GetDataCheckSum(), "res.from", res.GetHeader().GetFrom(), "pid", s.p.Pretty())
			continue
		}
	}
}

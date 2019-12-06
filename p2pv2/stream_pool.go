package p2pv2

import (
	"context"
	"errors"
	"sync"

	net "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"

	log "github.com/xuperchain/log15"
	"github.com/xuperchain/xuperunion/common"
	p2pPb "github.com/xuperchain/xuperunion/p2pv2/pb"
)

// define common errors
var (
	ErrStreamNotFound = errors.New("stream not found")
	ErrStreamPoolFull = errors.New("stream pool is full")
	ErrAddStream      = errors.New("error to add stream")
	ErrRequest        = errors.New("error request from network")
	ErrAuth           = errors.New("error invalid auth request")
)

// StreamPool manage all the stream
type StreamPool struct {
	log log.Logger
	// key: peer id, value: Stream
	streams        *common.LRUCache
	maxStreamLimit int32
	no             *Node
	quitCh         chan bool
}

// NewStreamPool create StreamPool instance
func NewStreamPool(maxStreamLimit int32, no *Node, log log.Logger) (*StreamPool, error) {
	return &StreamPool{
		log:            log,
		streams:        common.NewLRUCache(int(maxStreamLimit)),
		quitCh:         make(chan bool, 1),
		maxStreamLimit: maxStreamLimit,
		no:             no,
	}, nil
}

// Start start the stream pool
func (sp *StreamPool) Start() {
	for {
		select {
		case <-sp.quitCh:
			sp.Stop()
			return
		}
	}
}

// Stop will stop all streams
func (sp *StreamPool) Stop() {
	//sw.quitCh <- true
}

// Add used to add a new net stream into pool
func (sp *StreamPool) Add(s net.Stream) *Stream {
	// filter by StreamLimit first
	addrStr := s.Conn().RemoteMultiaddr().String()
	peerID := s.Conn().RemotePeer()
	if ok := sp.no.streamLimit.AddStream(addrStr, peerID); !ok {
		s.Reset()
		return nil
	}
	stream := NewStream(s, sp.no)
	if err := sp.AddStream(stream); err != nil {
		stream.Close()
		sp.DelStream(stream)
		sp.no.kdht.RoutingTable().Remove(stream.p)
		sp.log.Warn("New stream is deleted", "error", err)
		return nil
	}
	return stream
}

// AddStream used to add a new P2P stream into pool
func (sp *StreamPool) AddStream(stream *Stream) error {
	if int32(sp.streams.Len()) > sp.maxStreamLimit {
		return ErrStreamPoolFull
	}

	if sp.no.srv.config.IsAuthentication {
		err := sp.Authenticate(stream)
		if err != nil {
			return err
		}
	}

	if v, ok := sp.streams.Get(stream.p.Pretty()); ok {
		val, _ := v.(*Stream)
		sp.streams.Del(val.p.Pretty())
		if val.s != nil {
			val.s.Close()
		}
	}
	sp.streams.Add(stream.p.Pretty(), stream)
	return nil
}

// DelStream delete a stream
func (sp *StreamPool) DelStream(stream *Stream) error {
	if v, ok := sp.streams.Get(stream.p.Pretty()); ok {
		val, _ := v.(*Stream)
		sp.streams.Del(val.p.Pretty())
	}
	sp.no.streamLimit.DelStream(stream.addr.String())
	return nil
}

// FindStream get the stream between given peer ID
func (sp *StreamPool) FindStream(peer peer.ID) (*Stream, error) {
	if v, ok := sp.streams.Get(peer.Pretty()); ok {
		val, _ := v.(*Stream)
		if val != nil {
			return val, nil
		}
	}
	return nil, ErrStreamNotFound
}

// SendMessage send message to given peer ID
func (sp *StreamPool) SendMessage(ctx context.Context, msg *p2pPb.XuperMessage, peers []peer.ID) error {
	wg := sync.WaitGroup{}
	for _, p := range peers {
		wg.Add(1)
		go func(p peer.ID) {
			defer wg.Done()
			err := sp.sendMessage(ctx, msg, p)
			if err != nil {
				sp.log.Error("StreamPool SendMessage error", "peer", p.Pretty())
			}
		}(p)
	}
	wg.Wait()
	return nil
}

func (sp *StreamPool) sendMessage(ctx context.Context, msg *p2pPb.XuperMessage, p peer.ID) error {
	str, err := sp.streamForPeer(p)
	if err != nil {
		sp.log.Error("StreamPool sendMessage error!", "error", err.Error())
		return err
	}
	if err := str.SendMessage(ctx, msg); err != nil {
		if common.NormalizedKVError(err) == common.ErrP2PError {
			// delete the stream when error happens
			sp.DelStream(str)
		}
		sp.log.Error("StreamPool stream SendMessage error!", "error", err)
		return err
	}
	return nil
}

// streamForPeer will probe and return a stream
func (sp *StreamPool) streamForPeer(p peer.ID) (*Stream, error) {
	if v, ok := sp.streams.Get(p.Pretty()); ok {
		s, _ := v.(*Stream)
		if s.valid() {
			if sp.no.srv.config.IsAuthentication && !s.auth() {
				sp.log.Warn("stream failed to be authenticated")
				return nil, ErrAuth
			}
			return s, nil
		}
	}

	s, err := sp.no.host.NewStream(sp.no.ctx, p, XuperProtocolID)
	if err != nil {
		return nil, err
	}

	str := sp.Add(s)
	if str == nil {
		return nil, ErrAddStream
	}
	return str, nil
}

// SendMessageWithResponse will send message to peers with response
// withBreak means whether request wait for all response
func (sp *StreamPool) SendMessageWithResponse(ctx context.Context, msg *p2pPb.XuperMessage, peers []peer.ID, percentage float32) ([]*p2pPb.XuperMessage, error) {
	ch := make(chan *p2pPb.XuperMessage, len(peers))
	defer close(ch)
	wg := &sync.WaitGroup{}
	for _, v := range peers {
		wg.Add(1)
		go sp.sendMessageWithResponse(ctx, msg, v, wg, ch)
	}
	wg.Wait()
	res := []*p2pPb.XuperMessage{}
	lenCh := len(ch)
	if lenCh <= 0 {
		sp.log.Warn("StreamPool SendMessageWithResponse lenCh is nil")
		return res, ErrRequest
	}

	i := 0
	for r := range ch {
		if len(res) > int(float32(len(peers))*percentage) {
			break
		}
		if p2pPb.VerifyDataCheckSum(r) {
			res = append(res, r)
		}
		if i >= lenCh-1 {
			break
		}
		i++
	}
	sp.log.Trace("StreamPool SendMessageWithResponse done")
	return res, nil
}

func (sp *StreamPool) sendMessageWithResponse(ctx context.Context, msg *p2pPb.XuperMessage, p peer.ID, wg *sync.WaitGroup, ch chan *p2pPb.XuperMessage) {
	defer wg.Done()
	str, err := sp.streamForPeer(p)
	if err != nil {
		sp.log.Warn("StreamPool sendMessageWithResponse streamForPeer error!", "error", err.Error())
		return
	}
	res, err := str.SendMessageWithResponse(ctx, msg)
	if err != nil {
		if common.NormalizedKVError(err) == common.ErrP2PError {
			sp.DelStream(str)
		}
		sp.log.Warn("StreamPool sendMessageWithResponse SendMessageWithResponse error!", "error", err.Error())
		return
	}
	ch <- res
}

// Authenticate it's used for identity authentication
func (sp *StreamPool) Authenticate(stream *Stream) error {
	err := stream.Authenticate()
	return err
}

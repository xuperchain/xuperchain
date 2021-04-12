package server

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"
	mathRand "math/rand"
	"net"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc/peer"

	"github.com/xuperchain/xuperchain/core/common/config"
	xchaincore "github.com/xuperchain/xuperchain/core/core"
	"github.com/xuperchain/xuperchain/core/event"
	"github.com/xuperchain/xuperchain/core/pb"
)

var deadline = 5 * time.Minute

// eventService implements the interface of pb.EventService
type eventService struct {
	cfg    *config.EventConfig
	router *event.Router

	mutex       sync.Mutex
	connCounter map[string]int
}

func newEventService(cfg *config.EventConfig, chainmg *xchaincore.XChainMG) *eventService {
	return &eventService{
		cfg:         cfg,
		router:      event.NewRouter(chainmg),
		connCounter: make(map[string]int),
	}
}

func (e *eventService)GetLogs(ctx context.Context,req *pb.SubscribeRequest) (*pb.Logs, error){
	if !e.cfg.Enable {
		return nil,errors.New("event service disabled")
	}
	encfunc, iter, err := e.router.Subscribe(req.GetType(), req.GetFilter())
	if err != nil {
		return nil,err
	}
	events := []*pb.Event{}
	for iter.Next() {					// 过滤的条件在Next()
		payload := iter.Data()
		buf, _ := encfunc(payload)
		event := &pb.Event{
			Payload: buf,
		}
		events = append(events, event)
	}
	iter.Close()

	if iter.Error() != nil {
		return nil,iter.Error()
	}
	logs := &pb.Logs{Events:events}
	return logs,nil
}

func generateID()string{
	var buf = make([]byte,8)
	var seed int64
	if _,err := rand.Read(buf);err != nil {
		seed = int64(binary.BigEndian.Uint64(buf))
	} else {
		seed = int64(time.Now().Nanosecond())
	}
	rng := mathRand.New(mathRand.NewSource(seed))
	mu := sync.Mutex{}
	mu.Lock()
	bz := make([]byte, 16)
	rng.Read(bz)

	id := hex.EncodeToString(bz)
	id = strings.TrimLeft(id, "0")
	if id == "" {
		id = "0" // ID's are RPC quantities, no leading zero's and 0 is 0x0.
	}
	return "0x"+id
}

// Subscribe start an event subscribe
func (e *eventService) Subscribe(req *pb.SubscribeRequest, stream pb.EventService_SubscribeServer) error {
	if !e.cfg.Enable {
		return errors.New("event service disabled")
	}

	// check same ip limit
	remoteIP, err := e.connPermit(stream.Context())
	if err != nil {
		return err
	}
	defer e.releaseConn(remoteIP)

	encfunc, iter, err := e.router.Subscribe(req.GetType(), req.GetFilter())
	if err != nil {
		return err
	}
	for iter.Next() {
		payload := iter.Data()
		buf, _ := encfunc(payload)
		event := &pb.Event{
			Payload: buf,
		}
		err := stream.Send(event)
		if err != nil {
			break
		}
	}
	iter.Close()

	if iter.Error() != nil {
		return iter.Error()
	}
	return nil
}

func (e *eventService) connPermit(ctx context.Context) (string, error) {
	peer, ok := peer.FromContext(ctx)
	if !ok {
		return "", errors.New("get remote address error")
	}
	remoteIP, _, err := net.SplitHostPort(peer.Addr.String())
	if err != nil {
		return "", err
	}

	if e.cfg.AddrMaxConn == 0 {
		return remoteIP, nil
	}

	e.mutex.Lock()
	defer e.mutex.Unlock()
	cnt, ok := e.connCounter[remoteIP]
	if !ok {
		e.connCounter[remoteIP] = 1
		return remoteIP, nil
	}
	if cnt >= e.cfg.AddrMaxConn {
		return "", errors.New("maximum connections exceeded")
	}
	e.connCounter[remoteIP]++
	return remoteIP, nil
}

func (e *eventService) releaseConn(addr string) {
	if e.cfg.AddrMaxConn == 0 {
		return
	}

	e.mutex.Lock()
	defer e.mutex.Unlock()
	if e.connCounter[addr] <= 1 {
		delete(e.connCounter, addr)
		return
	}
	e.connCounter[addr]--
}

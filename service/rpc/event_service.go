package rpc

import (
	"errors"
	"net"
	"sync"

	"golang.org/x/net/context"
	"google.golang.org/grpc/peer"

	sconf "github.com/xuperchain/xuperchain/common/config"
	"github.com/xuperchain/xuperchain/common/xupospb/pb"
	acom "github.com/xuperchain/xuperchain/service/common"
	ecom "github.com/xuperchain/xupercore/kernel/engines/xuperos/common"
	"github.com/xuperchain/xupercore/kernel/engines/xuperos/event"
)

// eventService implements the interface of pb.EventService
type eventService struct {
	cfg    *sconf.ServConf
	router *event.Router

	mutex       sync.Mutex
	connCounter map[string]int
}

func newEventService(cfg *sconf.ServConf, engine ecom.Engine) *eventService {
	return &eventService{
		cfg:         cfg,
		router:      event.NewRouter(engine),
		connCounter: make(map[string]int),
	}
}

// Subscribe start an event subscribe
func (e *eventService) Subscribe(req *pb.SubscribeRequest, stream pb.EventService_SubscribeServer) error {
	if !e.cfg.EnableEvent {
		return errors.New("event service disabled")
	}

	// check same ip limit
	remoteIP, err := e.connPermit(stream.Context())
	if err != nil {
		return err
	}
	defer e.releaseConn(remoteIP)

	encfunc, iter, err := e.router.Subscribe(acom.ConvertEventSubType(req.GetType()), req.GetFilter())
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

	if e.cfg.EventAddrMaxConn == 0 {
		return remoteIP, nil
	}

	e.mutex.Lock()
	defer e.mutex.Unlock()
	cnt, ok := e.connCounter[remoteIP]
	if !ok {
		e.connCounter[remoteIP] = 1
		return remoteIP, nil
	}
	if cnt >= e.cfg.EventAddrMaxConn {
		return "", errors.New("maximum connections exceeded")
	}
	e.connCounter[remoteIP]++
	return remoteIP, nil
}

func (e *eventService) releaseConn(addr string) {
	if e.cfg.EventAddrMaxConn == 0 {
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

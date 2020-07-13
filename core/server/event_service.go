package server

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc/peer"

	xchaincore "github.com/xuperchain/xuperchain/core/core"
	"github.com/xuperchain/xuperchain/core/event"
	"github.com/xuperchain/xuperchain/core/pb"
)

// ip amount limit per ip
const totalLimitPerIP = 5

// eventService implements the interface of pb.EventService
type eventService struct {
	router *event.Router

	mutex     sync.Mutex
	srcIP2Cnt sync.Map
}

func newEventService(chainmg *xchaincore.XChainMG) *eventService {
	return &eventService{
		router: event.NewRouter(chainmg),
	}
}

// Subscribe start an event subscribe
func (e *eventService) Subscribe(req *pb.SubscribeRequest, stream pb.EventService_SubscribeServer) error {
	// check same ip limit
	valid, remoteIP := e.isValid(stream.Context())
	if !valid {
		return errors.New("Subscribe failed")
	}
	defer e.sub(remoteIP)

	iter, err := e.router.Subscribe(req.GetType(), req.GetFilter())
	if err != nil {
		return err
	}
	for iter.Next() {
		block := iter.Data().(*pb.FilteredBlock)
		buf, _ := proto.Marshal(block)
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

func (e *eventService) isValid(ctx context.Context) (bool, string) {
	peer, ok := peer.FromContext(ctx)
	if !ok {
		fmt.Println("get remote ip error")
		return false, ""
	}
	if peer.Addr == net.Addr(nil) {
		fmt.Println("peer's Addr is nil")
		return false, ""
	}
	remoteIP := strings.Split(peer.Addr.String(), ":")[0]
	val, exist := e.srcIP2Cnt.Load(remoteIP)

	e.mutex.Lock()
	defer e.mutex.Unlock()
	if exist {
		currCnt := val.(int)
		if currCnt >= totalLimitPerIP {
			fmt.Println("same ip up to limit")
			return false, remoteIP
		}
		e.srcIP2Cnt.Store(remoteIP, currCnt+1)
	} else {
		e.srcIP2Cnt.Store(remoteIP, 1)
	}

	return true, remoteIP
}

func (e *eventService) sub(ip string) {
	if val, exist := e.srcIP2Cnt.Load(ip); exist {
		currCnt := val.(int)
		if currCnt > 0 {
			e.srcIP2Cnt.Store(ip, currCnt-1)
		}
	}
}

package server

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc/peer"

	"github.com/xuperchain/xuperchain/core/crypto/hash"
	"github.com/xuperchain/xuperchain/core/event"
	"github.com/xuperchain/xuperchain/core/pb"
)

// ip amount limit per ip
const totalLimitPerIP = 5

// PubsubService rpc object
type PubsubService struct {
	EventService *event.EventService
	// same ip limitation
	srcIP2Cnt sync.Map
	mutex     sync.Mutex
}

// Unsubscribe unsubscribe an event by eventID
func (p *PubsubService) Unsubscribe(ctx context.Context, arg *pb.UnsubscribeRequest) (*pb.UnsubscribeResponse, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.EventService.Unsubscribe(arg.GetId())
	return &pb.UnsubscribeResponse{}, nil
}

// Subscribe start an event subscribe
func (p *PubsubService) Subscribe(arg *pb.EventRequest, stream pb.PubsubService_SubscribeServer) error {
	// check same ip limit
	valid, remoteIP := p.isValid(stream.Context())
	if !valid {
		return errors.New("Subscribe failed")
	}
	defer p.sub(remoteIP)

	randSeed := time.Now().UnixNano()
	eventID := fmt.Sprintf("%x", hash.DoubleSha256([]byte(strconv.FormatInt(randSeed, 10))))

	if err := stream.Send(&pb.Event{
		Id:   eventID,
		Type: pb.EventType_SUBSCRIBE_RESPONSE,
	}); err != nil {
		p.EventService.Unsubscribe(eventID)
		return err
	}

	return p.EventService.Subscribe(stream, eventID, arg.GetType(), arg.GetPayload())
}

func (p *PubsubService) isValid(ctx context.Context) (bool, string) {
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
	val, exist := p.srcIP2Cnt.Load(remoteIP)

	p.mutex.Lock()
	defer p.mutex.Unlock()
	if exist {
		currCnt := val.(int)
		if currCnt >= totalLimitPerIP {
			fmt.Println("same ip up to limit")
			return false, remoteIP
		}
		p.srcIP2Cnt.Store(remoteIP, currCnt+1)
	} else {
		p.srcIP2Cnt.Store(remoteIP, 1)
	}

	return true, remoteIP
}

func (p *PubsubService) sub(ip string) {
	if val, exist := p.srcIP2Cnt.Load(ip); exist {
		currCnt := val.(int)
		if currCnt > 0 {
			p.srcIP2Cnt.Store(ip, currCnt-1)
		}
	}
}

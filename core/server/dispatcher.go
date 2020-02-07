package server

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc/peer"

	"github.com/xuperchain/xuperchain/core/crypto/hash"
	"github.com/xuperchain/xuperchain/core/pb"
)

// ip amount limit per ip
const TOTAL_LIMIT_PER_IP = 5

// PubsubService meta data for pubsub
type PubsubService struct {
	pub     *Publisher
	msgChan chan *pb.Event
	// same ip limitation
	srcIP2Cnt       *sync.Map
	eventID2MsgChan *sync.Map
	mutex           *sync.Mutex
}

// Start recv msg from producer and dispatch to consumer
func (p *PubsubService) Start() {
	p.Init()
	for {
		select {
		case msg := <-p.msgChan:
			p.pub.Publish(msg)
		}
	}
}

// Init initialize meta data
func (p *PubsubService) Init() {
	p.srcIP2Cnt = &sync.Map{}
	p.eventID2MsgChan = &sync.Map{}
	p.mutex = &sync.Mutex{}
}

// clean clean meta data after a stream is invalid or calling Unsubscribe
func (p *PubsubService) clean(eventID string) {
	// if Evict equals false, which means eventID has been unsubscribed already
	p.mutex.Lock()
	defer p.mutex.Unlock()
	eventChan, exist := p.eventID2MsgChan.Load(eventID)
	if !exist {
		return
	}
	if p.pub.Evict(eventChan.(chan *pb.Event)) {
		p.eventID2MsgChan.Delete(eventID)
	}
}

// Unsubscribe unsubscribe an event by eventID
func (p *PubsubService) Unsubscribe(ctx context.Context, arg *pb.UnsubscribeRequest) (*pb.UnsubscribeResponse, error) {
	p.clean(arg.GetId())
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
		p.clean(eventID)
		return err
	}

	needContent := false

	ch := p.pub.SubscribeTopic(func(v *pb.Event) bool {
		eventType := arg.GetType()
		switch eventType {
		case pb.EventType_TRANSACTION:
			request := &pb.TransactionEventRequest{}
			proto.Unmarshal(arg.GetPayload(), request)
			metaData := v.GetTxStatus()

			requestBcname := request.GetBcname()
			requestInitiator := request.GetInitiator()
			requestAuthRequire := request.GetAuthRequire()
			needContent = request.GetNeedContent()
			eventBcname := metaData.GetBcname()
			eventInitiator := metaData.GetInitiator()
			eventAuthRequire := metaData.GetAuthRequire()

			if requestBcname == eventBcname &&
				((eventInitiator != "" && requestInitiator == eventInitiator) ||
					StringContains(requestAuthRequire, eventAuthRequire)) {
				return true
			}
			return false
		case pb.EventType_ACCOUNT:
			request := &pb.AccountEventRequest{}
			proto.Unmarshal(arg.GetPayload(), request)
			metaData := v.GetAccountStatus()

			requestBcname := request.GetBcname()
			requestFromAddr := request.GetFromAddr()
			requestToAddr := request.GetToAddr()
			needContent = request.GetNeedContent()
			eventBcname := metaData.GetBcname()
			eventFromAddr := metaData.GetFromAddr()
			eventToAddr := metaData.GetToAddr()

			if requestBcname == eventBcname &&
				(StringContains(requestFromAddr, eventFromAddr) ||
					StringContains(requestToAddr, eventToAddr)) {
				return true
			}
			return false
		case pb.EventType_BLOCK:
			request := &pb.BlockEventRequest{}
			proto.Unmarshal(arg.GetPayload(), request)
			metaData := v.GetBlockStatus()
			needContent = request.GetNeedContent()
			if request.GetBcname() == metaData.GetBcname() &&
				request.GetProposer() == metaData.GetProposer() &&
				request.GetStartHeight() <= metaData.GetHeight() &&
				request.GetEndHeight() >= metaData.GetHeight() {
				return true
			}
			return false
		}
		return true
	})
	p.eventID2MsgChan.Store(eventID, ch)

	for {
		msg, ok := <-ch
		if !ok {
			// return directly
			p.clean(eventID)
			return errors.New("the event has expired")
		}
		localMsg := *msg
		localMsg.Id = eventID
		if needContent == false {
			localMsg.Payload = nil
		}
		if err := stream.Send(&localMsg); err != nil {
			// clean work
			// stream is invalid probably
			// clean metadata about stream
			p.clean(eventID)
			return err
		}
	}

	return nil
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
		if currCnt >= TOTAL_LIMIT_PER_IP {
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

func StringContains(target string, arr []string) bool {
	if arr == nil || len(arr) == 0 {
		return false
	}
	for i := 0; i < len(arr); i++ {
		if arr[i] == target {
			return true
		}
	}

	return false
}

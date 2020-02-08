package event

import (
	"errors"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"

	"github.com/xuperchain/xuperchain/core/pb"
)

const (
	messageCapacityLimit = 100000
)

// EventService meta data for event service
type EventService struct {
	pub             *Publisher
	msgChan         chan *pb.Event
	eventID2MsgChan sync.Map
	serviceSwitch   bool
}

func (e *EventService) Init(serviceSwitch bool) {
	e.pub = NewPublisher(100*time.Millisecond, 10)
	e.msgChan = make(chan *pb.Event, messageCapacityLimit)
	e.serviceSwitch = serviceSwitch
}

// Start recv msg from producer and dispatch to consumer
func (e *EventService) Start() {
	for {
		select {
		case msg := <-e.msgChan:
			e.pub.Publish(msg)
		}
	}
}

// clean clean meta data after a stream is invalid or calling Unsubscribe
func (e *EventService) clean(eventID string) {
	// if Evict equals false, which means eventID has been unsubscribed already
	eventChan, exist := e.eventID2MsgChan.Load(eventID)
	if !exist {
		return
	}
	if e.pub.Evict(eventChan.(chan *pb.Event)) {
		e.eventID2MsgChan.Delete(eventID)
	}
}

func (e *EventService) Publish(msg *pb.Event) {
	if msg != nil && e.serviceSwitch {
		e.msgChan <- msg
	}
	return
}

// Unsubscribe unsubscribe an event by eventID
func (e *EventService) Unsubscribe(eventID string) {
	e.clean(eventID)
	return
}

// Subscribe start an event subscribe
func (e *EventService) Subscribe(stream pb.PubsubService_SubscribeServer,
	eventID string, eventType pb.EventType, payload []byte) error {
	needContent := false
	ch := e.pub.SubscribeTopic(func(v *pb.Event) bool {
		res := false
		switch eventType {
		case pb.EventType_TRANSACTION:
			res, needContent = filterForTransactionEvent(payload, v)
			return res
		case pb.EventType_ACCOUNT:
			res, needContent = filterForAccountEvent(payload, v)
			return res
		case pb.EventType_BLOCK:
			res, needContent = filterForBlockEvent(payload, v)
			return res
		}
		return false
	})
	e.eventID2MsgChan.Store(eventID, ch)

	for {
		msg, ok := <-ch
		if !ok {
			// return directly
			e.clean(eventID)
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
			e.clean(eventID)
			return err
		}
	}

	return nil
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

func filterForTransactionEvent(payload []byte, v *pb.Event) (bool, bool) {
	request := &pb.TransactionEventRequest{}
	proto.Unmarshal(payload, request)
	metaData := v.GetTxStatus()

	requestInitiator := request.GetInitiator()
	eventInitiator := metaData.GetInitiator()
	needContent := request.GetNeedContent()
	if request.GetBcname() == metaData.GetBcname() &&
		((eventInitiator != "" && requestInitiator == eventInitiator) ||
			StringContains(request.GetAuthRequire(), metaData.GetAuthRequire())) {
		return true, needContent
	}

	return false, needContent
}

func filterForBlockEvent(payload []byte, v *pb.Event) (bool, bool) {
	request := &pb.BlockEventRequest{}
	proto.Unmarshal(payload, request)
	metaData := v.GetBlockStatus()
	needContent := request.GetNeedContent()

	if request.GetBcname() == metaData.GetBcname() &&
		request.GetProposer() == metaData.GetProposer() &&
		request.GetStartHeight() <= metaData.GetHeight() &&
		request.GetEndHeight() >= metaData.GetHeight() {
		return true, needContent
	}

	return false, needContent
}

func filterForAccountEvent(payload []byte, v *pb.Event) (bool, bool) {
	request := &pb.AccountEventRequest{}
	proto.Unmarshal(payload, request)
	metaData := v.GetAccountStatus()
	needContent := request.GetNeedContent()

	if request.GetBcname() == metaData.GetBcname() &&
		(StringContains(request.GetFromAddr(), metaData.GetFromAddr()) ||
			StringContains(request.GetToAddr(), metaData.GetToAddr())) {
		return true, needContent
	}

	return false, needContent
}

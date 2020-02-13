package base

import (
	"container/list"
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/xuperchain/log15"
	"github.com/xuperchain/xuperchain/core/p2p/pb"
)

// Subscriber define the subscriber of message
type MockSubscriber struct {
	msgCh   chan *xuperp2p.XuperMessage
	msgType xuperp2p.XuperMessage_MessageType
	// 支持注册回调函数方式
	handler XuperHandler
	e       *list.Element
	// 仅接收固定来源的消息
	msgFrom string
	log     log.Logger
}

// NewSubscriber create instance of Subscriber
func NewMockSubscriber(msgCh chan *xuperp2p.XuperMessage, msgType xuperp2p.XuperMessage_MessageType, handler XuperHandler, msgFrom string, log log.Logger) *MockSubscriber {
	sub := &MockSubscriber{}
	if msgCh == nil && handler == nil {
		return nil
	}
	sub.msgCh = msgCh
	sub.msgType = msgType
	sub.handler = handler
	sub.msgFrom = msgFrom
	sub.log = log
	return sub
}

// GetMessageType return the subscribed message type of this subscriber
func (msub *MockSubscriber) GetMessageType() xuperp2p.XuperMessage_MessageType {
	return msub.msgType
}

// GetMessageChan return the to-be-processed message channel
func (msub *MockSubscriber) GetMessageChan() chan *xuperp2p.XuperMessage {
	return msub.msgCh
}

// GetElement get the element of list
func (msub *MockSubscriber) GetElement() *list.Element {
	return msub.e
}

// GetMessageFrom get the peer id which this message is from
func (msub *MockSubscriber) GetMessageFrom() string {
	return msub.msgFrom
}

// GetHandler get the message handler, this could be nil if use message channel
func (msub *MockSubscriber) GetHandler() XuperHandler {
	return msub.handler
}

// SetHandler set message processer
func (msub *MockSubscriber) SetHandler(handler XuperHandler) {
	msub.handler = handler
}

// SetElement set the element of list in MultiSubscriber
func (msub *MockSubscriber) SetElement(e *list.Element) {
	msub.e = e
}

// HandleMessage process a message
func (*MockSubscriber) HandleMessage(stream interface{}, msg *xuperp2p.XuperMessage) {
	return
}

func TestStartHandlerMap(t *testing.T) {
	mgHeader := xuperp2p.XuperMessage_MessageHeader{
		Version:      "xuperchain2.4",
		Logid:        "logidaaa",
		From:         "localhost",
		Bcname:       "xuper",
		Type:         xuperp2p.XuperMessage_SENDBLOCK,
		DataCheckSum: 123,
	}
	mgData := xuperp2p.XuperMessage_MessageData{
		MsgInfo: []byte{1},
	}
	var mg xuperp2p.XuperMessage
	mg.Header = &mgHeader
	mg.Data = &mgData

	lg := log.New("module", "p2pv2")
	lg.SetHandler(log.StreamHandler(os.Stderr, log.LogfmtFormat()))

	// new a HandlerMap
	hm, err := NewHandlerMap(lg)
	/*
	   defer func() {
	       if hm != nil {
	           hm.Stop()
	       }
	   }()
	*/
	if err != nil {
		t.Error("Expect nil, got ", err)
	}

	// new a subscriber and Register
	ch := make(chan *xuperp2p.XuperMessage, 5000)
	sub := &MockSubscriber{
		msgCh:   ch,
		msgType: xuperp2p.XuperMessage_SENDBLOCK,
	}
	if hm != nil {
		e, _ := hm.Register(sub)
		_, ok1 := hm.subscriberCenter.Load(sub.msgType)
		if !ok1 {
			t.Error("Register error")
		}

		// send message into HandlerMap
		hm.HandleMessage(nil, &mg)

		// start HandlerMap service parallelly
		go hm.Start()
		<-time.After(1 * time.Second)

		IsHandled := hm.IsMsgAsHandled(&mg)
		if !IsHandled {
			//t.Error("Expect true, got ", IsHandled)
		}
		// stop HandlerMap service
		hm.quitCh <- true

		// UnRegister
		v2, _ := hm.subscriberCenter.Load(sub.msgType)
		ms2 := v2.(*MultiSubscriber)
		if ms2.elem.Len() != 1 {
			t.Error("subscriberCenter len error")
		}
		hm.UnRegister(e)
		v3, _ := hm.subscriberCenter.Load(sub.msgType)
		ms3 := v3.(*MultiSubscriber)
		if ms3.elem.Len() != 0 {
			t.Error("subscriberCenter len error")
		}
	}
}

func TestMarkMsgAsHandled(t *testing.T) {
	mgHeader := xuperp2p.XuperMessage_MessageHeader{
		Version:      "xuperchain2.4",
		Logid:        "logidaaa",
		From:         "localhost",
		Bcname:       "xuper",
		Type:         xuperp2p.XuperMessage_SENDBLOCK,
		DataCheckSum: 123,
	}
	mgData := xuperp2p.XuperMessage_MessageData{
		MsgInfo: []byte{1},
	}
	var mg xuperp2p.XuperMessage
	mg.Header = &mgHeader
	mg.Data = &mgData

	// new a HandlerMap
	lg := log.New("module", "p2pv2")
	hm, err := NewHandlerMap(lg)
	defer func() {
		if hm != nil {
			hm.Stop()
		}
	}()
	if err != nil {
		//t.Error("Expect nil, got ", err)
	}
	if hm != nil {
		if ok1 := hm.IsMsgAsHandled(&mg); ok1 {
			//t.Error("Expect ok1 false, got ", ok1)
		}
		hm.MarkMsgAsHandled(&mg)
		if ok2 := hm.IsMsgAsHandled(&mg); !ok2 {
			//t.Error("Expect ok2 true, got ", ok2)
		}
	}
}

func testHandler(ctx context.Context, msg *xuperp2p.XuperMessage) (*xuperp2p.XuperMessage, error) {
	fmt.Println("test handler ok")
	return &xuperp2p.XuperMessage{
		Header: &xuperp2p.XuperMessage_MessageHeader{
			Version: "testHandler",
		},
	}, nil
}

func TestHandleMessage(t *testing.T) {
	lg := log.New("module", "p2pv2")
	hm, err := NewHandlerMap(lg)
	defer func() {
		if hm != nil {
			hm.Stop()
		}
	}()
	if err != nil {
		t.Error("NewHandlerMap error", err.Error())
	}
	sub := NewMockSubscriber(nil, xuperp2p.XuperMessage_PING, testHandler, "", nil)
	hm.Register(sub)

	mgHeader := xuperp2p.XuperMessage_MessageHeader{
		Version:      "xuperchain2.4",
		Logid:        "logidaaa",
		From:         "localhost",
		Bcname:       "xuper",
		Type:         xuperp2p.XuperMessage_PING,
		DataCheckSum: 123,
	}
	mgData := xuperp2p.XuperMessage_MessageData{
		MsgInfo: []byte{1},
	}
	var mg xuperp2p.XuperMessage
	mg.Header = &mgHeader
	mg.Data = &mgData

	hm.HandleMessage(nil, &mg)
}

package p2pv2

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/xuperchain/log15"
	"github.com/xuperchain/xuperunion/p2pv2/pb"
)

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
	sub := &Subscriber{
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
	sub := NewSubscriber(nil, xuperp2p.XuperMessage_PING, testHandler, "")
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

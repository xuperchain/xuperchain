package p2pv2

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/xuperchain/log15"
	"github.com/xuperchain/xuperunion/p2pv2/pb"
)

// define default message config
const (
	MsgChanSize         = 50000
	MsgHandledCacheSize = 50000
)

// define errors
var (
	ErrSubscribe       = errors.New("subscribe error")
	ErrAlreadyRegisted = errors.New("subscriber already registered")
	ErrUnregister      = errors.New("unregister subscriber error")
)

type xuperHandler func(context.Context, *xuperp2p.XuperMessage) (*xuperp2p.XuperMessage, error)

// HandlerMap the message handler manager
// keeps the message and handler mapping and recently handled messages
type HandlerMap struct {
	lg log.Logger
	// key: xuperp2p.XuperMessage_MessageType, value: *MultiSubscriber
	subscriberCenter *sync.Map
	msgHandled       *cache.Cache
	quitCh           chan bool
}

// NewHandlerMap create instance of HandlerMap
func NewHandlerMap(log log.Logger) (*HandlerMap, error) {
	return &HandlerMap{
		lg:               log,
		subscriberCenter: new(sync.Map),
		msgHandled:       cache.New(time.Duration(3)*time.Second, 1*time.Second),
		quitCh:           make(chan bool, 1),
	}, nil
}

// Start start message handling
func (hm *HandlerMap) Start() {
	for {
		select {
		case <-hm.quitCh:
			hm.Stop()
		}
	}
}

// Stop stop message handling
func (hm *HandlerMap) Stop() {
	// todo
	hm.lg.Trace("Stop HandlerMap")
}

// Register used to register subscriber to handlerMap.
func (hm *HandlerMap) Register(sub *Subscriber) (*Subscriber, error) {
	if sub == nil {
		return nil, ErrSubscribe
	}
	v, ok := hm.subscriberCenter.Load(sub.msgType)
	if !ok {
		ms := newMultiSubscriber()
		hm.subscriberCenter.Store(sub.msgType, ms)
		return ms.register(sub)
	}
	ms, ok := v.(*MultiSubscriber)
	if !ok {
		return nil, ErrSubscribe
	}
	return ms.register(sub)

}

// UnRegister used to un register subscriber from handlerMap.
func (hm *HandlerMap) UnRegister(sub *Subscriber) error {
	if sub.e == nil {
		return ErrUnregister
	}

	sub, ok := (sub.e.Value).(*Subscriber)
	if !ok {
		return ErrUnregister
	}

	v, ok := hm.subscriberCenter.Load(sub.msgType)
	if !ok {
		return ErrUnregister
	}
	ms, ok := v.(*MultiSubscriber)
	if !ok {
		return ErrUnregister
	}
	return ms.unRegister(sub)
}

// MarkMsgAsHandled used to mark message has been dealt with.
func (hm *HandlerMap) MarkMsgAsHandled(msg *xuperp2p.XuperMessage) {
	hm.lg.Trace("MarkMsgAsHandled ", "logid", msg.GetHeader().GetLogid(), "msgType", msg.GetHeader().GetType(), "checksum", msg.GetHeader().GetDataCheckSum())
	msgHash := msg.GetHeader().GetDataCheckSum()
	key := msg.GetHeader().GetLogid() + "_" + strconv.FormatUint(uint64(msgHash), 10)
	hm.lg.Trace("MarkMsgAsHandled", "key", key)
	hm.msgHandled.Set(key, true, time.Duration(3)*time.Second)
}

// IsMsgAsHandled used to check whether the msg has been dealt with.
func (hm *HandlerMap) IsMsgAsHandled(msg *xuperp2p.XuperMessage) bool {
	msgHash := msg.GetHeader().GetDataCheckSum()
	key := msg.GetHeader().GetLogid() + "_" + strconv.FormatUint(uint64(msgHash), 10)
	hm.lg.Trace("IsMsgAsHandled", "key", key)
	_, ok := hm.msgHandled.Get(key)
	return ok
}

// HandleMessage handle new messages with registered handlers
func (hm *HandlerMap) HandleMessage(s *Stream, msg *xuperp2p.XuperMessage) error {
	if s == nil {
		hm.lg.Warn("handlerMap stream can not be null")
		return nil
	}
	if msg.GetHeader() == nil || msg.GetData() == nil {
		hm.lg.Warn("HandlerMap receive msg is null!")
		return nil
	}
	hm.lg.Trace("HandlerMap receive msg", "logid", msg.GetHeader().GetLogid(), "msgType", msg.GetHeader().GetType(), "checksum", msg.GetHeader().GetDataCheckSum(), "peer", s.p.Pretty())
	if ok := hm.IsMsgAsHandled(msg); ok {
		hm.lg.Trace("HandlerMap receive is handled", "logid", msg.GetHeader().GetLogid(), "msgType", msg.GetHeader().GetType(), "checksum", msg.GetHeader().GetDataCheckSum(), "peer", s.p.Pretty())
		return nil
	}
	msgType := msg.GetHeader().GetType()
	v, ok := hm.subscriberCenter.Load(msgType)
	if !ok {
		hm.lg.Warn("HandlerMap load subscribeCenter not found!", "msgType", msgType)
		return nil
	}

	if ms, ok := v.(*MultiSubscriber); ok {
		// 如果注册了回调方法，则调用回调方法, 如果注册了channel,则进行通知
		go ms.handleMessage(s, msg)
		hm.MarkMsgAsHandled(msg)
		return nil
	}
	hm.lg.Warn("HandleMessage get MultiSubscriber error")
	return nil
}

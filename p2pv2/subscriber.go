package p2pv2

import (
	"container/list"
	"context"
	"fmt"
	"sync"

	xuperp2p "github.com/xuperchain/xuperunion/p2pv2/pb"
)

// Subscriber define the subscriber of message
type Subscriber struct {
	msgCh   chan *xuperp2p.XuperMessage
	msgType xuperp2p.XuperMessage_MessageType
	// 支持注册回调函数方式
	handler xuperHandler
	e       *list.Element
	// 仅接收固定来源的消息
	msgFrom string
}

// NewSubscriber create instance of Subscriber
func NewSubscriber(msgCh chan *xuperp2p.XuperMessage, msgType xuperp2p.XuperMessage_MessageType, handler xuperHandler, msgFrom string) *Subscriber {
	sub := &Subscriber{}
	if msgCh == nil && handler == nil {
		return nil
	}
	sub.msgCh = msgCh
	sub.msgType = msgType
	sub.handler = handler
	sub.msgFrom = msgFrom
	return sub
}

// handleMessage process subscribed message
func (sub *Subscriber) handleMessage(s *Stream, msg *xuperp2p.XuperMessage) {
	if !s.valid() {
		return
	}

	if msg.Header.Type != xuperp2p.XuperMessage_GET_AUTHENTICATION_RES &&
		msg.Header.Type != xuperp2p.XuperMessage_GET_AUTHENTICATION {
		if s.node.srv.config.IsAuthentication && !s.auth() {
			s.node.log.Trace("Stream not authenticated")
			return
		}
	}

	if sub.handler != nil {
		go func(sub *Subscriber, s *Stream, msg *xuperp2p.XuperMessage) {
			ctx := context.WithValue(context.Background(), "Stream", s)
			res, err := sub.handler(ctx, msg)
			if err != nil {
				fmt.Println("subscriber handleMessage error", "err", err)
			}
			if err := s.writeData(res); err != nil {
				fmt.Println("subscriber handleMessage to write msg error", "err", err)
			}
		}(sub, s, msg)
		return
	}
	if sub.msgCh == nil {
		return
	}
	select {
	case sub.msgCh <- msg:
	default:
	}
}

// MultiSubscriber wrap a list of Subscriber of same message
type MultiSubscriber struct {
	// elem 存监听同一消息类型的多个Subscriber
	elem *list.List
	lk   *sync.Mutex
}

func newMultiSubscriber() *MultiSubscriber {
	return &MultiSubscriber{
		elem: list.New(),
		lk:   new(sync.Mutex),
	}
}

func (ms *MultiSubscriber) register(sub *Subscriber) (*Subscriber, error) {
	ms.lk.Lock()
	defer ms.lk.Unlock()
	e := ms.elem.PushBack(sub)
	sub.e = e
	return sub, nil
}

func (ms *MultiSubscriber) unRegister(sub *Subscriber) error {
	ms.lk.Lock()
	defer ms.lk.Unlock()
	if sub.e == nil {
		return nil
	}
	ms.elem.Remove(sub.e)
	sub = nil
	return nil
}

func (ms *MultiSubscriber) handleMessage(s *Stream, msg *xuperp2p.XuperMessage) {
	ms.lk.Lock()
	defer ms.lk.Unlock()
	for e := ms.elem.Front(); e != nil; e = e.Next() {
		if sub, ok := e.Value.(*Subscriber); !ok {
			continue
		} else {
			if sub.msgFrom == "" || (sub.msgFrom == msg.GetHeader().GetFrom()) {
				sub.handleMessage(s, msg)
			}
		}
	}
}

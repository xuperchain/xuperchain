package base

import (
	"container/list"
	"sync"

	xuperp2p "github.com/xuperchain/xuperchain/core/p2p/pb"
)

// Subscriber is the interface for p2p message subscriber
type Subscriber interface {
	// GetMessageType return the subscribed message type of this subscriber
	GetMessageType() xuperp2p.XuperMessage_MessageType
	// GetMessageChan return the to-be-processed message channel
	GetMessageChan() chan *xuperp2p.XuperMessage
	// GetElement get the element of list
	GetElement() *list.Element
	// GetMessageFrom get the peer id which this message is from
	GetMessageFrom() string
	// GetHandler get the message handler, this could be nil if use message channel
	GetHandler() XuperHandler

	// SetHandler set message processer
	SetHandler(XuperHandler)
	// SetElement set the element of list in MultiSubscriber
	SetElement(*list.Element)

	// HandleMessage process a message
	HandleMessage(stream interface{}, msg *xuperp2p.XuperMessage)
}

// MultiSubscriber wrap a list of Subscriber of same message
type MultiSubscriber struct {
	// elem 存监听同一消息类型的多个Subscriber
	elem *list.List
	lk   *sync.Mutex
}

// NewMultiSubscriber init MultiSubscriber
func NewMultiSubscriber() *MultiSubscriber {
	return &MultiSubscriber{
		elem: list.New(),
		lk:   new(sync.Mutex),
	}
}

func (ms *MultiSubscriber) register(sub Subscriber) (Subscriber, error) {
	ms.lk.Lock()
	defer ms.lk.Unlock()
	e := ms.elem.PushBack(sub)
	sub.SetElement(e)
	return sub, nil
}

func (ms *MultiSubscriber) unRegister(sub Subscriber) error {
	ms.lk.Lock()
	defer ms.lk.Unlock()
	if sub.GetElement() == nil {
		return nil
	}
	ms.elem.Remove(sub.GetElement())
	sub = nil
	return nil
}

func (ms *MultiSubscriber) handleMessage(stream interface{}, msg *xuperp2p.XuperMessage) {
	ms.lk.Lock()
	defer ms.lk.Unlock()
	for e := ms.elem.Front(); e != nil; e = e.Next() {
		if sub, ok := e.Value.(Subscriber); !ok {
			continue
		} else {
			if sub.GetMessageFrom() == "" || (sub.GetMessageFrom() == msg.GetHeader().GetFrom()) {
				sub.HandleMessage(stream, msg)
			}
		}
	}
}

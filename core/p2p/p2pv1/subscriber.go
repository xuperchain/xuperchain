package p2pv1

import (
	"container/list"

	p2p_base "github.com/xuperchain/xuperchain/core/p2p/base"
	xuperp2p "github.com/xuperchain/xuperchain/core/p2p/pb"
)

// MsgSubscriber define the subscriber of message
type MsgSubscriber struct {
	msgCh   chan *xuperp2p.XuperMessage
	msgType xuperp2p.XuperMessage_MessageType
	// 支持注册回调函数方式
	handler p2p_base.XuperHandler
	e       *list.Element
	// 仅接收固定来源的消息
	msgFrom string
}

// NewMsgSubscriber create instance of Subscriber
func NewMsgSubscriber(msgCh chan *xuperp2p.XuperMessage, msgType xuperp2p.XuperMessage_MessageType, handler p2p_base.XuperHandler, msgFrom string) *MsgSubscriber {
	sub := &MsgSubscriber{}
	if msgCh == nil && handler == nil {
		return nil
	}
	sub.msgCh = msgCh
	sub.msgType = msgType
	sub.handler = handler
	sub.msgFrom = msgFrom
	return sub
}

// GetMessageType return the subscribed message type of this subscriber
func (sub *MsgSubscriber) GetMessageType() xuperp2p.XuperMessage_MessageType {
	return sub.msgType
}

// GetMessageChan return the to-be-processed message channel
func (sub *MsgSubscriber) GetMessageChan() chan *xuperp2p.XuperMessage {
	return sub.msgCh
}

// GetElement get the element of list
func (sub *MsgSubscriber) GetElement() *list.Element {
	return sub.e
}

// GetMessageFrom get the peer id which this message is from
func (sub *MsgSubscriber) GetMessageFrom() string {
	return sub.msgFrom
}

// GetHandler get the message handler, this could be nil if use message channel
func (sub *MsgSubscriber) GetHandler() p2p_base.XuperHandler {
	return sub.handler
}

// SetHandler set message processer
func (sub *MsgSubscriber) SetHandler(handler p2p_base.XuperHandler) {
	sub.handler = handler
}

// SetElement set the element of list in MultiSubscriber
func (sub *MsgSubscriber) SetElement(e *list.Element) {
	sub.e = e
}

// HandleMessage process a message
// TODO
func (sub *MsgSubscriber) HandleMessage(conn interface{}, msg *xuperp2p.XuperMessage) {
	// s, ok := conn.(*Conn)
	// if !ok {
	// 	fmt.Println("invalid message stream")
	// 	return
	// }
	// if s == nil {
	// 	fmt.Println("message stream cannot be nil")
	// 	return
	// }
	// if !s.valid() {
	// 	return
	// }

	// if sub.handler != nil {
	// 	go func(sub *MsgSubscriber, s *Conn, msg *xuperp2p.XuperMessage) {
	// 		ctx := context.WithValue(context.Background(), "Stream", s)
	// 		res, err := sub.handler(ctx, msg)
	// 		if err != nil {
	// 			fmt.Println("subscriber handleMessage error", "err", err)
	// 		}
	// 		if err := s.writeData(res); err != nil {
	// 			fmt.Println("subscriber handleMessage to write msg error", "err", err)
	// 		}
	// 	}(sub, s, msg)
	// 	return
	// }
	// if sub.msgCh == nil {
	// 	return
	// }
	// select {
	// case sub.msgCh <- msg:
	// default:
	// }
}

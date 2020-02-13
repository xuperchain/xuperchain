package p2pv2

import (
	"container/list"
	"context"

	log "github.com/xuperchain/log15"

	p2p_base "github.com/xuperchain/xuperchain/core/p2p/base"
	xuperp2p "github.com/xuperchain/xuperchain/core/p2p/pb"
)

// make sure MsgSubscriber implemented the Subscriber interface
var _ p2p_base.Subscriber = (*MsgSubscriber)(nil)

// MsgSubscriber define the subscriber of message
type MsgSubscriber struct {
	msgCh   chan *xuperp2p.XuperMessage
	msgType xuperp2p.XuperMessage_MessageType
	// 支持注册回调函数方式
	handler p2p_base.XuperHandler
	e       *list.Element
	// 仅接收固定来源的消息
	msgFrom string
	log     log.Logger
}

// NewMsgSubscriber create instance of Subscriber
func NewMsgSubscriber(msgCh chan *xuperp2p.XuperMessage, msgType xuperp2p.XuperMessage_MessageType, handler p2p_base.XuperHandler, msgFrom string, log log.Logger) *MsgSubscriber {
	sub := &MsgSubscriber{}
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
func (sub *MsgSubscriber) HandleMessage(msgStream interface{}, msg *xuperp2p.XuperMessage) {
	s, ok := msgStream.(*Stream)
	if !ok {
		sub.log.Warn("p2p HandleMessage: invalid message stream")
		return
	}
	if s == nil {
		sub.log.Warn("p2p HandleMessage: message stream cannot be nil")
		return
	}
	if !s.valid() {
		return
	}

	if sub.handler != nil {
		go func(sub *MsgSubscriber, s *Stream, msg *xuperp2p.XuperMessage) {
			ctx := context.WithValue(context.Background(), "Stream", s)
			if msg.Header.Type != xuperp2p.XuperMessage_GET_AUTHENTICATION_RES &&
				msg.Header.Type != xuperp2p.XuperMessage_GET_AUTHENTICATION {
				if s.node.srv.config.IsAuthentication && !s.auth() {
					sub.log.Trace("Stream not authenticated")
					resType := p2p_base.GetResMsgType(msg.GetHeader().GetType())
					res, _ := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion2, "", msg.GetHeader().GetLogid(),
						resType, []byte(""), xuperp2p.XuperMessage_GET_AUTHENTICATION_NOT_PASS)
					if err := s.writeData(res); err != nil {
						sub.log.Warn("Stream not authenticated to write msg error", "err", err)
					}
					return
				}
			}

			res, err := sub.handler(ctx, msg)
			if err != nil {
				sub.log.Warn("subscriber handleMessage error", "err", err)
			}
			if err := s.writeData(res); err != nil {
				sub.log.Warn("subscriber handleMessage to write msg error", "err", err)
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

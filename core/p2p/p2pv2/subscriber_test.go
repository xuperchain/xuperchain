package p2pv2

import (
	"testing"

	log "github.com/xuperchain/log15"

	_ "github.com/xuperchain/xuperchain/core/p2p/base"
	"github.com/xuperchain/xuperchain/core/p2p/pb"
)

func TestNewSubscriber(t *testing.T) {
	resch := make(chan *xuperp2p.XuperMessage, 1)
	lg := log.New("module", "p2pv2")
	sub := NewMsgSubscriber(resch, xuperp2p.XuperMessage_PING, nil, "", lg)
	if sub.GetMessageType() != xuperp2p.XuperMessage_PING {
		t.Error("message type not the same")
		return
	}

	if sub.GetMessageChan() != resch {
		t.Error("message channel not the same")
		return
	}

	if sub.GetHandler() != nil {
		t.Error("message handler not the same")
		return
	}

	sub.HandleMessage(nil, nil)
}

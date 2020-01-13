package p2pv2

import (
	"testing"

	"github.com/xuperchain/xuperchain/core/p2pv2/pb"
)

func TestNewSubscriber(t *testing.T) {
	ms := newMultiSubscriber()
	resch := make(chan *xuperp2p.XuperMessage, 1)
	sub := NewSubscriber(resch, xuperp2p.XuperMessage_PING, nil, "")
	sub, _ = ms.register(sub)
	if ms.elem.Len() != 1 {
		t.Error("register sub error")
	}
	ms.unRegister(sub)
	if ms.elem.Len() != 0 {
		t.Error("unRegister sub error")
	}
}

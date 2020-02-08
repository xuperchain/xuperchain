package base

import (
	"testing"

	xuperp2p "github.com/xuperchain/xuperchain/core/p2p/pb"
)

func TestNewSubscriber(t *testing.T) {
	ms := NewMultiSubscriber()
	resch := make(chan *xuperp2p.XuperMessage, 1)
	sub := NewMockSubscriber(resch, xuperp2p.XuperMessage_PING, nil, "", nil)
	newsub, _ := ms.register(sub)
	if ms.elem.Len() != 1 {
		t.Error("register sub error")
	}
	ms.unRegister(newsub)
	if ms.elem.Len() != 0 {
		t.Error("unRegister sub error")
	}
}

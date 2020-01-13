package p2pv2

import (
	"fmt"
	"testing"
	"time"

	"github.com/xuperchain/xuperchain/core/common/config"
	xuperp2p "github.com/xuperchain/xuperchain/core/p2pv2/pb"
)

func TestNewP2PServerV2(t *testing.T) {
	testCases := map[string]struct {
		in config.P2PConfig
	}{
		"testNewServer": {
			in: config.P2PConfig{
				Port:            47103,
				KeyPath:         "./data/netkeys/",
				IsNat:           true,
				IsSecure:        true,
				IsHidden:        false,
				BootNodes:       []string{},
				MaxStreamLimits: 32,
			},
		},
	}

	srv, err := NewP2PServerV2(testCases["testNewServer"].in, nil)
	if err != nil {
		t.Log(err.Error())
	}
	if srv != nil {
		fmt.Println(srv.GetNetURL())

		ch := make(chan *xuperp2p.XuperMessage, 5000)
		time.Sleep(1 * time.Second)

		sub := &Subscriber{
			msgCh:   ch,
			msgType: xuperp2p.XuperMessage_PING,
		}
		e, _ := srv.Register(sub)
		_, ok := srv.handlerMap.subscriberCenter.Load(xuperp2p.XuperMessage_PING)
		if !ok {
			t.Error("Register sub error")
		}

		srv.UnRegister(e)
	}
}

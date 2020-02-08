package p2pv2

import (
	"fmt"
	"testing"
	"time"

	log "github.com/xuperchain/log15"
	"github.com/xuperchain/xuperchain/core/common/config"
)

func TestStreamPoolBasic(t *testing.T) {
	t.Skip()
	cfg1 := config.P2PConfig{
		Port:            20016,
		KeyPath:         "./data/netkeys/",
		IsNat:           true,
		IsSecure:        true,
		IsHidden:        false,
		MaxStreamLimits: 20,
	}
	lg := log.New("module", "p2pv2")
	node, err := NewNode(cfg1, lg)
	defer func() {
		if node != nil {
			node.Stop()
		}
	}()
	if err != nil {
		t.Error("create node error ", err.Error())
	}
	if node != nil {
		streamPool, err := NewStreamPool(cfg1.MaxStreamLimits, node, lg)
		if err != nil {
			fmt.Println("new NewStreamPool error ", err.Error())
		}
		if streamPool != nil {
			// start stream pool
			go streamPool.Start()
			// stop stream pool
			streamPool.quitCh <- true
			<-time.After(1 * time.Second)

			// test for Add
			// creat a new net.Stream
			// streamPool.Add()

			// test for SendMessageWithResponse
			// so far, SendMessageWithResponse is not implemented
			streamPool.SendMessageWithResponse(nil, nil, nil, 1)
		}
	}
}

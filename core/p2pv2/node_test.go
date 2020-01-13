package p2pv2

import (
	"fmt"
	"testing"

	log "github.com/xuperchain/log15"
	"github.com/xuperchain/xuperchain/core/common/config"
)

func TestNewNode(t *testing.T) {
	t.Skip()
	t.Log("Test two node communicate")
	cfg1 := config.P2PConfig{
		Port:            45101,
		KeyPath:         "./data/netkeys/",
		IsNat:           true,
		IsSecure:        true,
		IsHidden:        false,
		MaxStreamLimits: 20,
	}
	lg := log.New("module", "p2pv2")
	node1, err := NewNode(cfg1, lg)
	defer func() {
		if node1 != nil {
			node1.Stop()
		}
	}()
	if err != nil {
		t.Error("create node1 error!", err)
	}

	cfg2 := config.P2PConfig{
		Port:            45102,
		KeyPath:         "./data/netkeys/",
		IsNat:           true,
		IsSecure:        true,
		IsHidden:        false,
		MaxStreamLimits: 20,
	}
	node2, err := NewNode(cfg2, lg)
	if err != nil {
		t.Error("create node2 error!", err)
	}
	defer func() {
		if node2 != nil {
			node2.Stop()
		}
	}()
	t.Log(node1)
	t.Log(node2)
	fmt.Println(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d/p2p/%s", cfg1.Port, node1.host.ID().Pretty()))
	fmt.Println(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d/p2p/%s", cfg2.Port, node2.host.ID().Pretty()))
}

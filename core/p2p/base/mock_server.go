package base

import (
	"context"

	log "github.com/xuperchain/log15"

	"github.com/xuperchain/xuperchain/core/common/config"
	p2pPb "github.com/xuperchain/xuperchain/core/p2p/pb"
)

// make sure mockp2pserver implemented the P2PServer interface
var _ P2PServer = (*MockP2pServer)(nil)

// MockP2pServer is mock struct of P2PServer interface
// Used in unit tests
type MockP2pServer struct {
}

// Init initialize the Mock p2p server
func (mp *MockP2pServer) Init(cfg config.P2PConfig, lg log.Logger, extra map[string]interface{}) error {
	return nil
}

// Start implements the start interface
func (mp *MockP2pServer) Start() {

}

// Stop implements the Stop interface
func (mp *MockP2pServer) Stop() {

}

// NewSubscriber create a subscriber instance
func (mp *MockP2pServer) NewSubscriber(chan *p2pPb.XuperMessage, p2pPb.XuperMessage_MessageType, XuperHandler, string, log.Logger) Subscriber {
	return nil
}

// Register implements the Register interface
func (mp *MockP2pServer) Register(sub Subscriber) (Subscriber, error) {
	return nil, nil
}

// UnRegister implements the UnRegister interface
func (mp *MockP2pServer) UnRegister(sub Subscriber) error {
	return nil
}

// SendMessage implements the SendMessage interface
func (mp *MockP2pServer) SendMessage(context context.Context, msg *p2pPb.XuperMessage, opts ...MessageOption) error {
	return nil
}

// SendMessageWithResponse implements the Register interface
func (mp *MockP2pServer) SendMessageWithResponse(context.Context,
	*p2pPb.XuperMessage, ...MessageOption) ([]*p2pPb.XuperMessage, error) {
	return nil, nil

}

// GetNetURL implements the GetNetURL interface
func (mp *MockP2pServer) GetNetURL() string {
	return ""
}

// GetPeerUrls implements the GetPeerUrls interface
func (mp *MockP2pServer) GetPeerUrls() []string {
	return nil
}

func (mp *MockP2pServer) GetPeerIDAndUrls() map[string]string {
	return nil
}

// SetCorePeers implements the SetCorePeers interface
func (mp *MockP2pServer) SetCorePeers(corePeers *CorePeersInfo) error {
	return nil
}

// SetXchainAddr implements the SetXchainAddr interface
func (mp *MockP2pServer) SetXchainAddr(bcname string, info *XchainAddrInfo) {

}

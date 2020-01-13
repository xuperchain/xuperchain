package p2pv2

import (
	"context"

	p2pPb "github.com/xuperchain/xuperunion/p2pv2/pb"
)

// MockP2pServer is mock struct of P2PServer interface
// Used in unit tests
type MockP2pServer struct {
}

// Start implements the start interface
func (mp *MockP2pServer) Start() {

}

// Stop implements the Stop interface
func (mp *MockP2pServer) Stop() {

}

// Register implements the Register interface
func (mp *MockP2pServer) Register(sub *Subscriber) (*Subscriber, error) {
	return nil, nil
}

// UnRegister implements the UnRegister interface
func (mp *MockP2pServer) UnRegister(sub *Subscriber) error {
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

// SetCorePeers implements the SetCorePeers interface
func (mp *MockP2pServer) SetCorePeers(corePeers *CorePeersInfo) error {
	return nil
}

// SetXchainAddr implements the SetXchainAddr interface
func (mp *MockP2pServer) SetXchainAddr(bcname string, info *XchainAddrInfo) {

}

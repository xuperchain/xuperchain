package base

import (
	"context"

	log "github.com/xuperchain/log15"

	"github.com/xuperchain/xuperchain/core/common/config"
	p2pPb "github.com/xuperchain/xuperchain/core/p2p/pb"
)

// P2PServer is the p2p server interface of Xuper
type P2PServer interface {
	Start()
	Stop()

	// Initialize the p2p server with given config
	Init(cfg config.P2PConfig, log log.Logger, extra map[string]interface{}) error

	// NewSubscriber create a subscriber instance
	NewSubscriber(chan *p2pPb.XuperMessage, p2pPb.XuperMessage_MessageType, XuperHandler, string, log.Logger) Subscriber
	// 注册订阅者，支持多个用户订阅同一类消息
	Register(sub Subscriber) (Subscriber, error)
	// 注销订阅者，需要根据当时注册时返回的Subscriber实例删除
	UnRegister(sub Subscriber) error

	SendMessage(context.Context, *p2pPb.XuperMessage, ...MessageOption) error

	SendMessageWithResponse(context.Context, *p2pPb.XuperMessage, ...MessageOption) ([]*p2pPb.XuperMessage, error)

	GetNetURL() string
	// 查询所连接节点的信息
	GetPeerUrls() []string
	GetPeerIDAndUrls() map[string]string

	// SetCorePeers set core peers' info to P2P server
	SetCorePeers(cp *CorePeersInfo) error

	// SetXchainAddr Set xchain address from xchaincore
	SetXchainAddr(bcname string, info *XchainAddrInfo)
}

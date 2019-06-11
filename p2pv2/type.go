package p2pv2

import (
	"context"

	p2pPb "github.com/xuperchain/xuperunion/p2pv2/pb"
)

// P2PServer is the p2p server interface of Xuper
type P2PServer interface {
	Start()
	Stop()

	// 注册订阅者，支持多个用户订阅同一类消息
	Register(sub *Subscriber) (*Subscriber, error)
	// 注销订阅者，需要根据当时注册时返回的Subscriber实例删除
	UnRegister(sub *Subscriber) error

	SendMessage(context.Context, *p2pPb.XuperMessage, FilterStrategy) error
	// todo: 将请求的参数改为Option的方式
	SendMessageWithResponse(context.Context, *p2pPb.XuperMessage, FilterStrategy, bool) ([]*p2pPb.XuperMessage, error)

	GetNetURL() string
	// 查询所连接节点的信息
	GetPeerUrls() []string

	// SetXchainAddr Set xchain address from xchaincore
	SetXchainAddr(bcname string, info *XchainAddrInfo)
}

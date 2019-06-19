package p2pv2

import (
	"context"

	p2pPb "github.com/xuperchain/xuperunion/p2pv2/pb"
)

// CorePeersInfo defines the peers' info for core nodes
// By setting this info, we can keep some core peers always connected directly
// It's useful for keeping DPoS key network security and for some BFT-like consensus
type CorePeersInfo struct {
	Name           string   // distinguished name of the node routing
	CurrentTermNum int64    // the current term number
	CurrentPeerIDs []string // current core peer IDs
	NextPeerIDs    []string // upcoming core peer IDs
}

type msgOptions struct {
	filters         []FilterStrategy
	bcname          string
	targetPeerAddrs []string
	isBreak         bool
}

// MessageOption define single option function
type MessageOption func(*msgOptions)

// WithBreak set is break option
func WithBreak() MessageOption {
	return func(o *msgOptions) {
		o.isBreak = true
	}
}

// WithFilters add filter strategies to message option
func WithFilters(filter []FilterStrategy) MessageOption {
	return func(o *msgOptions) {
		o.filters = filter
	}
}

// WithBcName add bcname to message option
func WithBcName(bcname string) MessageOption {
	return func(o *msgOptions) {
		o.bcname = bcname
	}
}

// WithTargetPeerAddrs add target peer addresses to message option
func WithTargetPeerAddrs(peerAddrs []string) MessageOption {
	return func(o *msgOptions) {
		o.targetPeerAddrs = peerAddrs
	}
}

// getMessageOption create MessageOptions with given options
func getMessageOption(opts []MessageOption) *msgOptions {
	msgOpts := &msgOptions{
		isBreak: false,
		filters: []FilterStrategy{DefaultStrategy},
	}
	for _, f := range opts {
		f(msgOpts)
	}
	return msgOpts
}

// P2PServer is the p2p server interface of Xuper
type P2PServer interface {
	Start()
	Stop()

	// 注册订阅者，支持多个用户订阅同一类消息
	Register(sub *Subscriber) (*Subscriber, error)
	// 注销订阅者，需要根据当时注册时返回的Subscriber实例删除
	UnRegister(sub *Subscriber) error

	SendMessage(context.Context, *p2pPb.XuperMessage, ...MessageOption) error
	// todo: 将请求的参数改为Option的方式
	SendMessageWithResponse(context.Context, *p2pPb.XuperMessage, ...MessageOption) ([]*p2pPb.XuperMessage, error)

	GetNetURL() string
	// 查询所连接节点的信息
	GetPeerUrls() []string

	// SetCorePeers set core peers' info to P2P server
	SetCorePeers(cp *CorePeersInfo) error

	// SetXchainAddr Set xchain address from xchaincore
	SetXchainAddr(bcname string, info *XchainAddrInfo)
}

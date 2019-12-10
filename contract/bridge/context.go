package bridge

import (
	"sync"

	"github.com/xuperchain/xuperunion/contract"
	pb "github.com/xuperchain/xuperunion/contractsdk/go/pb"
	"github.com/xuperchain/xuperunion/xmodel"
)

// Context 保存了合约执行的内核状态，
// 所有的系统调用产生的状态保存在这里
type Context struct {
	ID int64
	// 合约名字
	ContractName string

	ResourceLimits contract.Limits

	Cache *xmodel.XMCache

	Args map[string][]byte

	Method string

	Initiator string

	AuthRequire []string

	CanInitialize bool

	Core contract.ChainCore

	TransferAmount string

	Instance Instance

	// resource used by sub contract call
	SubResourceUsed contract.Limits

	// Contract being called
	// set by bridge to check recursive contract call
	ContractSet map[string]bool

	// Write by contract
	Output *pb.Response
}

// DiskUsed returns the bytes written to xmodel
func (c *Context) DiskUsed() int64 {
	size := int64(0)
	_, wset, _ := c.Cache.GetRWSets()
	for _, w := range wset {
		size += int64(len(w.GetKey()))
		size += int64(len(w.GetValue()))
	}
	return size
}

// ExceedDiskLimit check whether disk usage exceeds limit
func (c *Context) ExceedDiskLimit() bool {
	size := c.DiskUsed()
	return size > c.ResourceLimits.Disk
}

// ResourceUsed returns the resource used by context
func (c *Context) ResourceUsed() contract.Limits {
	var total contract.Limits
	total.Add(c.Instance.ResourceUsed()).Add(c.SubResourceUsed)
	total.Disk += c.DiskUsed()
	return total
}

// ContextManager 用于管理产生和销毁Context
type ContextManager struct {
	// 保护如下两个变量
	// 合约进行系统调用以及合约执行会并发访问ctxs
	ctxlock sync.Mutex
	ctxid   int64
	ctxs    map[int64]*Context
}

// NewContextManager instances a new ContextManager
func NewContextManager() *ContextManager {
	return &ContextManager{
		ctxs: make(map[int64]*Context),
	}
}

// Context 根据context的id返回当前运行当前合约的上下文
func (n *ContextManager) Context(id int64) (*Context, bool) {
	n.ctxlock.Lock()
	defer n.ctxlock.Unlock()
	ctx, ok := n.ctxs[id]
	return ctx, ok
}

// MakeContext allocates a Context with unique context id
func (n *ContextManager) MakeContext() *Context {
	n.ctxlock.Lock()
	defer n.ctxlock.Unlock()
	n.ctxid++
	ctx := new(Context)
	ctx.ID = n.ctxid
	n.ctxs[ctx.ID] = ctx
	return ctx
}

// DestroyContext 一定要在合约执行完毕（成功或失败）进行销毁
func (n *ContextManager) DestroyContext(ctx *Context) {
	n.ctxlock.Lock()
	defer n.ctxlock.Unlock()
	delete(n.ctxs, ctx.ID)
}

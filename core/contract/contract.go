package contract

import (
	"encoding/json"
	"fmt"

	"github.com/xuperchain/xuperchain/core/kv/kvdb"
	"github.com/xuperchain/xuperchain/core/ledger"
	"github.com/xuperchain/xuperchain/core/pb"
	"github.com/xuperchain/xuperchain/core/xmodel"
)

// KernelModuleName is the name of kernel contract
const KernelModuleName = "kernel"

// ConsensusModueName is the name of consensus contract
const ConsensusModueName = "consensus"

// UtxoMetaRegister in avoid to being refered in a cycle way
type UtxoMetaRegister interface {
	GetMaxBlockSize() int64
	UpdateMaxBlockSize(int64, kvdb.Batch) error
	GetReservedContracts() []*pb.InvokeRequest
	UpdateReservedContracts([]*pb.InvokeRequest, kvdb.Batch) error
	GetForbiddenContract() *pb.InvokeRequest
	UpdateForbiddenContract(*pb.InvokeRequest, kvdb.Batch) error
	GetNewAccountResourceAmount() int64
	UpdateNewAccountResourceAmount(int64, kvdb.Batch) error
	QueryTx(txid []byte) (*pb.Transaction, error)
	GetXModel() *xmodel.XModel
	// Get irreversible slide window
	GetIrreversibleSlideWindow() int64
	// Update irreversible slide window
	UpdateIrreversibleSlideWindow(nextIrreversibleSlideWindow int64, batch kvdb.Batch) error
	// Get gas price
	GetGasPrice() *pb.GasPrice
	UpdateGasPrice(*pb.GasPrice, kvdb.Batch) error
}

// TxDesc is the description to running a contract
type TxDesc struct {
	Module string                 `json:"module"`
	Method string                 `json:"method"`
	Args   map[string]interface{} `json:"args"`
	//纯文本注释
	Tag []byte `json:"tag"`
	//表示当前合约执行到期时间, 只有大于0的时候才有效, 否则应该被认为是不限制到期时间
	Deadline int64           `json:"deadline"`
	Tx       *pb.Transaction `json:"tx"`
	Trigger  *TriggerDesc    `json:"trigger"`
}

// TriggerDesc is the description to trigger a event used by proposal
type TriggerDesc struct {
	TxDesc
	Height  int64  `json:"height"`
	RefTxid []byte `json:"refTxid"` //创建trigger的txid，系统自动回填的
}

// TxContext 合约的上下文，通常生命周期是Block范围
type TxContext struct {
	UtxoBatch kvdb.Batch //如果合约机和UtxoVM共用DB, 可以将修改打包到这个batch确保原子性
	//... 其他的需要UtxoVM与合约机共享的也可以放到这里
	Block     *pb.InternalBlock
	UtxoMeta  UtxoMetaRegister
	LedgerObj *ledger.Ledger
	IsUndo    bool
}

// ContractOutputInterface used to read output of a contract
type ContractOutputInterface interface {
	Decode(data []byte) error
	Encode() ([]byte, error)
	GetGasUsed() uint64
	Digest() ([]byte, error)
}

// ContractInterface is the interface to implement a contract driver
type ContractInterface interface {
	//TX级别的接口
	Run(desc *TxDesc) error
	Rollback(desc *TxDesc) error
	//获取执行合约的结果
	ReadOutput(desc *TxDesc) (ContractOutputInterface, error)

	//block级别的接口
	//区块生成之后，用来更新各个合约的状态
	Finalize(blockid []byte) error
	//用于被设置上下文
	SetContext(context *TxContext) error
	Stop()
}

// ContractExtInterface is used to initialize contract plugin
type ContractExtInterface interface {
	// 使用额外的参数初始化
	Init(extParams map[string]interface{}) error
}

type privContractInterface struct {
	//虚拟机的权限等级，系统级别的虚拟机全部状态为0级别，用户态合约状态为3级别
	priv int
	vm   ContractInterface
}

// SmartContract manage smart contracts
type SmartContract struct {
	handlers map[string]privContractInterface
}

// NewSmartContract instances a new SmartContract instance
func NewSmartContract() *SmartContract {
	return &SmartContract{
		handlers: map[string]privContractInterface{},
	}
}

// Parse 解析智能合约json
func Parse(desc string) (*TxDesc, error) {
	txDesc := &TxDesc{}
	jsErr := json.Unmarshal([]byte(desc), txDesc)
	if jsErr != nil {
		return nil, jsErr
	}
	return txDesc, nil
}

// RegisterHandler 注册module对应的handler
func (s *SmartContract) RegisterHandler(moduleName string, handler ContractInterface, priv int) bool {
	if vm, exist := s.handlers[moduleName]; exist && vm.priv >= priv {
		return false
	}
	s.handlers[moduleName] = privContractInterface{vm: handler, priv: priv}
	return true
}

// Get returns ContractInterface from contract driver name
func (s *SmartContract) Get(name string) (ContractInterface, bool) {
	handler, exist := s.handlers[name]
	if exist {
		return handler.vm, true
	}
	return nil, false
}

// GetAll returns all the contract drivers
func (s *SmartContract) GetAll() map[string]ContractInterface {
	ret := make(map[string]ContractInterface)
	for name, pci := range s.handlers {
		ret[name] = pci.vm
	}
	return ret
}

// Remove remove contract driver
func (s *SmartContract) Remove(name string, priv int) {
	if vm, ok := s.handlers[name]; ok && vm.priv == priv {
		delete(s.handlers, name)
	}
}

// SetContext 设置所有注册合约的上下文。这里必须在run之前设置，后设置会覆盖前面设置的
func (s *SmartContract) SetContext(ctx *TxContext) {
	for _, handler := range s.handlers {
		handler.vm.SetContext(ctx)
	}
}

// Finalize 在一个块的合约执行完毕之后调用。这里必须在run 之后调用，这里有可能提交之前没有提交过的合约结果
func (s *SmartContract) Finalize(blockid []byte) error {
	for _, handler := range s.handlers {
		handler.vm.Finalize(blockid)
	}
	return nil
}

// Run 执行合约
func (s *SmartContract) Run(desc *TxDesc) error {
	if desc.Module == "" {
		return nil //不是合约,跳过
	}
	handler, exist := s.handlers[desc.Module]
	if !exist {
		return fmt.Errorf("this module has no registered handlers, %s, when Run", desc.Module)
	}
	return handler.vm.Run(desc)
}

// Stop stops all the contract drivers
func (s *SmartContract) Stop() {
	for _, handler := range s.handlers {
		handler.vm.Stop()
	}
}

// Rollback 回滚合约
func (s *SmartContract) Rollback(desc *TxDesc) error {
	if desc.Module == "" {
		return nil
	}
	handler, exist := s.handlers[desc.Module]
	if !exist {
		return fmt.Errorf("this module has no registered handlers, %s, when Rollback", desc.Module)
	}
	return handler.vm.Rollback(desc)
}

// Package utxo is the key part of XuperChain, this module keeps all Unspent Transaction Outputs.
//
// For a transaction, the UTXO checks the tokens used in reference transactions are unspent, and
// reject the transaction if the initiator doesn't have enough tokens.
// UTXO also checks the signature and permission of transaction members.
package utxo

import (
	"bytes"
	"container/list"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	kverr "github.com/syndtr/goleveldb/leveldb/errors"
	log "github.com/xuperchain/log15"
	"github.com/xuperchain/xuperunion/common"
	"github.com/xuperchain/xuperunion/contract"
	crypto_client "github.com/xuperchain/xuperunion/crypto/client"
	crypto_base "github.com/xuperchain/xuperunion/crypto/client/base"
	"github.com/xuperchain/xuperunion/global"
	"github.com/xuperchain/xuperunion/kv/kvdb"
	ledger_pkg "github.com/xuperchain/xuperunion/ledger"
	"github.com/xuperchain/xuperunion/pb"
	pm "github.com/xuperchain/xuperunion/permission"
	acl "github.com/xuperchain/xuperunion/permission/acl"
	acli "github.com/xuperchain/xuperunion/permission/acl/impl"
	"github.com/xuperchain/xuperunion/pluginmgr"
	"github.com/xuperchain/xuperunion/utxo/txhash"
	"github.com/xuperchain/xuperunion/vat"
	"github.com/xuperchain/xuperunion/xmodel"
)

// 常用VM执行错误码
var (
	ErrDoubleSpent             = errors.New("utxo can not be spent more than once")
	ErrAlreadyInUnconfirmed    = errors.New("this transaction is in unconfirmed state")
	ErrAlreadyConfirmed        = errors.New("this transaction is already confirmed")
	ErrNoEnoughUTXO            = errors.New("no enough money(UTXO) to start this transaction")
	ErrUTXONotFound            = errors.New("this utxo can not be found")
	ErrInputOutputNotEqual     = errors.New("input's amount is not equal to output's")
	ErrTxNotFound              = errors.New("this tx can not be found in unconfirmed-table")
	ErrPreBlockMissMatch       = errors.New("play block failed because pre-hash != latest_block")
	ErrUnexpected              = errors.New("this is a unexpected error")
	ErrNegativeAmount          = errors.New("amount in transaction can not be negative number")
	ErrUTXOFrozen              = errors.New("utxo is still frozen")
	ErrTxSizeLimitExceeded     = errors.New("tx size limit exceeded")
	ErrOverloaded              = errors.New("this node is busy, try again later")
	ErrInvalidAutogenTx        = errors.New("found invalid autogen-tx")
	ErrUnmatchedExtension      = errors.New("found unmatched extension")
	ErrUnsupportedContract     = errors.New("found unspported contract module")
	ErrUTXODuplicated          = errors.New("found duplicated utxo in same tx")
	ErrDestroyProofAlreadyUsed = errors.New("the destroy proof has been used before")
	ErrInvalidWithdrawAmount   = errors.New("withdraw amount is invalid")
	ErrServiceRefused          = errors.New("Service refused")
	ErrRWSetInvalid            = errors.New("RWSet of transaction invalid")
	ErrRWAclNotEnough          = errors.New("ACL not enough")

	ErrGasNotEnough   = errors.New("Gas not enough")
	ErrInvalidAccount = errors.New("Invalid account")
	ErrVersionInvalid = errors.New("Invalid tx version")
	ErrInvalidTxExt   = errors.New("Invalid tx ext")
)

// package constants
const (
	UTXOLockExpiredSecond     = 60
	LatestBlockKey            = "pointer"
	UTXOCacheSize             = 1000
	OfflineTxChanBuffer       = 100000
	TxVersion                 = 1
	BetaTxVersion             = 1
	StableTxVersion           = 1
	RootTxVersion             = 0
	FeePlaceholder            = "$"
	UTXOTotalKey              = "xtotal"
	UTXOContractExecutionTime = 500
	TxWaitTimeout             = 5
)

// UtxoVM UTXO VM
type UtxoVM struct {
	ldb               kvdb.Database
	mutex             *sync.RWMutex // utxo leveldb表读写锁
	mutexMem          *sync.Mutex   // 内存锁定状态互斥锁
	lockKeys          map[string]int64
	lockKeyList       *list.List // 按锁定的先后顺序，方便过期清理
	lockExpireTime    int        // 临时锁定的最长时间
	utxoCache         *UtxoCache
	xlog              log.Logger
	ledger            *ledger_pkg.Ledger       // 引用的账本对象
	latestBlockid     []byte                   // 当前vm最后一次执行到的blockid
	unconfirmedTable  kvdb.Database            // 未确认交易表
	utxoTable         kvdb.Database            // utxo表
	metaTable         kvdb.Database            // 元数据表，会持久化保存latestBlockid
	withdrawTable     kvdb.Database            // 平行币赎回表, 记录已经赎回的destroy proof
	smartContract     *contract.SmartContract  // 智能合约执行机
	OfflineTxChan     chan *pb.Transaction     // 未确认tx的通知chan
	prevFoundKeyCache *common.LRUCache         // 上一次找到的可用utxo key，用于加速GenerateTx
	utxoTotal         *big.Int                 // 总资产
	cryptoClient      crypto_base.CryptoClient // 加密实例
	model3            *xmodel.XModel           // XuperModel实例，处理extutxo
	vmMgr3            *contract.VMManager
	aclMgr            *acli.Manager // ACL manager for read/write acl table

	minerPublicKey  string
	minerPrivateKey string
	minerAddress    []byte
	failedTxBuf     map[string][]string

	inboundTxChan        chan *InboundTx      // 异步tx chan
	verifiedTxChan       chan *pb.Transaction //已经校验通过的tx
	asyncMode            bool                 // 是否工作在异步模式
	asyncCancel          context.CancelFunc   // 停止后台异步batch写的句柄
	asyncWriterWG        *sync.WaitGroup      // 优雅退出异步writer的信号量
	asyncCond            *sync.Cond           // 用来出块线程优先权的条件变量
	asyncTryBlockGen     bool                 // doMiner线程是否准备出块
	vatHandler           *vat.VATHandler      // Verifiable Autogen Tx 生成器
	balanceCache         *common.LRUCache     //余额cache,加速GetBalance查询
	cacheSize            int                  //记录构造utxo时传入的cachesize
	balanceViewDirty     map[string]bool      //balanceCache 标记dirty: addr->bool
	contractExectionTime int
	unconfirmTxInMem     *sync.Map //未确认Tx表的内存镜像
	defaultTxVersion     int32     // 默认的tx version
}

// InboundTx is tx wrapper
type InboundTx struct {
	tx    *pb.Transaction
	txBuf []byte
}

// RootJSON xuper.json对应的struct，目前先只写了utxovm关注的字段
type RootJSON struct {
	Version   string `json:"version"`
	Consensus struct {
		Miner string `json:"miner"`
	} `json:"consensus"`
	Predistribution []struct {
		Address string `json:"address"`
		Quota   string `json:"quota"`
	} `json:"predistribution"`
}

func genUtxoKey(addr []byte, txid []byte, offset int32) string {
	return fmt.Sprintf("%s_%x_%d", addr, txid, offset)
}

// GenUtxoKeyWithPrefix generate UTXO key with given prefix
func GenUtxoKeyWithPrefix(addr []byte, txid []byte, offset int32) string {
	baseUtxoKey := genUtxoKey(addr, txid, offset)
	return pb.UTXOTablePrefix + baseUtxoKey
}

// checkInputEqualOutput 校验交易的输入输出是否相等
func (uv *UtxoVM) checkInputEqualOutput(tx *pb.Transaction) error {
	inputSum := big.NewInt(0)
	outputSum := big.NewInt(0)
	curLedgerHeight := uv.ledger.GetMeta().TrunkHeight
	utxoDedup := map[string]bool{}
	for _, txInput := range tx.TxInputs {
		addr := txInput.FromAddr
		txid := txInput.RefTxid
		offset := txInput.RefOffset
		utxoKey := genUtxoKey(addr, txid, offset)
		if utxoDedup[utxoKey] {
			uv.xlog.Warn("found duplicated utxo in same tx", "utxoKey", utxoKey, "txid", global.F(tx.Txid))
			return ErrUTXODuplicated
		}
		utxoDedup[utxoKey] = true
		var amountBytes []byte
		var frozenHeight int64
		uv.utxoCache.Lock()
		if l2Cache, exist := uv.utxoCache.All[string(addr)]; exist {
			uItem := l2Cache[pb.UTXOTablePrefix+utxoKey]
			if uItem != nil {
				amountBytes = uItem.Amount.Bytes()
				frozenHeight = uItem.FrozenHeight
			}
		}
		uv.utxoCache.Unlock()
		if amountBytes == nil {
			uBinary, findErr := uv.utxoTable.Get([]byte(utxoKey))
			if findErr != nil {
				if findErr.Error() == kverr.ErrNotFound.Error() {
					uv.xlog.Warn("not found utxo key:", "utxoKey", utxoKey)
					return ErrUTXONotFound
				}
				uv.xlog.Warn("unexpected leveldb error when do checkInputEqualOutput", "findErr", findErr)
				return findErr
			}
			uItem := &UtxoItem{}
			uErr := uItem.Loads(uBinary)
			if uErr != nil {
				return uErr
			}
			amountBytes = uItem.Amount.Bytes()
			frozenHeight = uItem.FrozenHeight
		}
		amount := big.NewInt(0)
		amount.SetBytes(amountBytes)
		if !bytes.Equal(amountBytes, txInput.Amount) {
			txInputAmount := big.NewInt(0)
			txInputAmount.SetBytes(txInput.Amount)
			uv.xlog.Warn("unexpected error, txInput amount missmatch utxo amount",
				"in_utxo", amount, "txInputAmount", txInputAmount, "txid", fmt.Sprintf("%x", tx.Txid), "reftxid", fmt.Sprintf("%x", txid))
			return ErrUnexpected
		}
		if frozenHeight > curLedgerHeight || frozenHeight == -1 {
			uv.xlog.Warn("this utxo still be frozen", "frozenHeight", frozenHeight, "ledgerHeight", curLedgerHeight)
			return ErrUTXOFrozen
		}
		inputSum.Add(inputSum, amount)
	}
	for _, txOutput := range tx.TxOutputs {
		amount := big.NewInt(0)
		amount.SetBytes(txOutput.Amount)
		if amount.Cmp(big.NewInt(0)) < 0 {
			return ErrNegativeAmount
		}
		outputSum.Add(outputSum, amount)
	}
	if inputSum.Cmp(outputSum) == 0 {
		return nil
	}
	if inputSum.Cmp(big.NewInt(0)) == 0 && tx.Coinbase {
		// coinbase交易，输入输出不必相等, 特殊处理
		return nil
	}
	uv.xlog.Warn("input != output", "inputSum", inputSum, "outputSum", outputSum)
	return ErrInputOutputNotEqual
}

// utxo是否处于临时锁定状态
func (uv *UtxoVM) isLocked(utxoKey []byte) bool {
	uv.mutexMem.Lock()
	defer uv.mutexMem.Unlock()
	_, exist := uv.lockKeys[string(utxoKey)]
	return exist
}

// 解锁utxo key
func (uv *UtxoVM) unlockKey(utxoKey []byte) {
	uv.mutexMem.Lock()
	defer uv.mutexMem.Unlock()
	uv.xlog.Trace("    unlock utxo key", "key", string(utxoKey))
	delete(uv.lockKeys, string(utxoKey))
}

// 试图临时锁定utxo, 返回是否锁定成功
func (uv *UtxoVM) tryLockKey(key []byte) bool {
	uv.mutexMem.Lock()
	defer uv.mutexMem.Unlock()
	if _, exist := uv.lockKeys[string(key)]; !exist {
		uv.lockKeys[string(key)] = time.Now().Unix()
		uv.lockKeyList.PushBack(key)
		if !uv.asyncMode {
			uv.xlog.Trace("  lock utxo key", "key", string(key))
		}
		return true
	}

	return false
}

// 清理过期的utxo锁定
func (uv *UtxoVM) clearExpiredLocks() {
	uv.mutexMem.Lock()
	defer uv.mutexMem.Unlock()
	now := time.Now().Unix()
	for {
		topItem := uv.lockKeyList.Front()
		if topItem == nil {
			break
		}
		topKey := topItem.Value.([]byte)
		createTime, exist := uv.lockKeys[string(topKey)]
		if !exist {
			uv.lockKeyList.Remove(topItem)
		} else if createTime+int64(uv.lockExpireTime) <= now {
			uv.lockKeyList.Remove(topItem)
			delete(uv.lockKeys, string(topKey))
		} else {
			break
		}
	}
}

// NewUtxoVM 构建一个UtxoVM对象
//   @param ledger 账本对象
//   @param store path, utxo 数据的保存路径
//   @param xlog , 日志handler
func NewUtxoVM(bcname string, ledger *ledger_pkg.Ledger, storePath string, privateKey, publicKey string,
	address []byte, xlog log.Logger, isBeta bool, kvEngineType string, cryptoType string) (*UtxoVM, error) {
	return MakeUtxoVM(bcname, ledger, storePath, privateKey, publicKey, address, xlog, UTXOCacheSize,
		UTXOLockExpiredSecond, UTXOContractExecutionTime, []string{}, isBeta, kvEngineType, cryptoType)
}

// MakeUtxoVM 这个函数比NewUtxoVM更加可订制化
func MakeUtxoVM(bcname string, ledger *ledger_pkg.Ledger, storePath string, privateKey, publicKey string, address []byte, xlog log.Logger,
	cachesize int, tmplockSeconds, contractExectionTime int, otherPaths []string, iBeta bool, kvEngineType string, cryptoType string) (*UtxoVM, error) {
	dbPath := filepath.Join(storePath, "utxoVM")
	plgMgr, plgErr := pluginmgr.GetPluginMgr()
	if plgErr != nil {
		xlog.Warn("fail to get plugin manager")
		return nil, plgErr
	}
	var baseDB kvdb.Database
	soInst, err := plgMgr.PluginMgr.CreatePluginInstance("kv", kvEngineType)
	if err != nil {
		xlog.Warn("fail to create plugin instance", "kvtype", kvEngineType)
		return nil, err
	}
	baseDB = soInst.(kvdb.Database)
	err = baseDB.Open(dbPath, map[string]interface{}{
		"cache":     ledger_pkg.MemCacheSize,
		"fds":       ledger_pkg.FileHandlersCacheSize,
		"dataPaths": otherPaths,
	})
	if err != nil {
		xlog.Warn("fail to open db", "dbPath", dbPath)
		return nil, err
	}

	if xlog == nil { // 如果外面没传进来log对象的话
		xlog = log.New("module", "utxoVM")
		xlog.SetHandler(log.StreamHandler(os.Stderr, log.LogfmtFormat()))
	}
	if err != nil {
		xlog.Warn("fail to open leveldb", "dbPath", dbPath, "err", err)
		return nil, err
	}

	// create crypto client
	cryptoClient, cryptoErr := crypto_client.CreateCryptoClient(cryptoType)
	if cryptoErr != nil {
		xlog.Warn("fail to create crypto client", "err", cryptoErr)
		return nil, cryptoErr
	}

	// create model3
	model3, mErr := xmodel.NewXuperModel(bcname, ledger, baseDB, xlog)
	if mErr != nil {
		xlog.Warn("failed to init xuper model", "err", mErr)
		return nil, mErr
	}

	// create aclMgr
	aclManager, aErr := acli.NewACLManager(model3)
	if aErr != nil {
		xlog.Warn("failed to init acl manager", "err", aErr)
		return nil, aErr
	}

	// create vmMgr
	vmManager, verr := contract.NewVMManager(xlog)
	if verr != nil {
		return nil, verr
	}
	utxoMutex := &sync.RWMutex{}
	utxoVM := &UtxoVM{
		ldb:                  baseDB,
		mutex:                utxoMutex,
		mutexMem:             &sync.Mutex{},
		lockKeys:             map[string]int64{},
		lockKeyList:          list.New(),
		lockExpireTime:       tmplockSeconds,
		xlog:                 xlog,
		ledger:               ledger,
		unconfirmedTable:     kvdb.NewTable(baseDB, pb.UnconfirmedTablePrefix),
		utxoTable:            kvdb.NewTable(baseDB, pb.UTXOTablePrefix),
		metaTable:            kvdb.NewTable(baseDB, pb.MetaTablePrefix),
		withdrawTable:        kvdb.NewTable(baseDB, pb.WithdrawPrefix),
		utxoCache:            NewUtxoCache(cachesize),
		smartContract:        contract.NewSmartContract(),
		vatHandler:           vat.NewVATHandler(),
		OfflineTxChan:        make(chan *pb.Transaction, OfflineTxChanBuffer),
		prevFoundKeyCache:    common.NewLRUCache(cachesize),
		utxoTotal:            big.NewInt(0),
		minerAddress:         address,
		minerPublicKey:       publicKey,
		minerPrivateKey:      privateKey,
		failedTxBuf:          make(map[string][]string),
		inboundTxChan:        make(chan *InboundTx, AsyncQueueBuffer),
		verifiedTxChan:       make(chan *pb.Transaction, AsyncQueueBuffer),
		asyncMode:            false,
		asyncCancel:          nil,
		asyncWriterWG:        &sync.WaitGroup{},
		asyncCond:            sync.NewCond(utxoMutex),
		asyncTryBlockGen:     false,
		balanceCache:         common.NewLRUCache(cachesize),
		cacheSize:            cachesize,
		balanceViewDirty:     map[string]bool{},
		contractExectionTime: contractExectionTime,
		unconfirmTxInMem:     &sync.Map{},
		cryptoClient:         cryptoClient,
		model3:               model3,
		vmMgr3:               vmManager,
		aclMgr:               aclManager,
	}
	if iBeta {
		utxoVM.defaultTxVersion = BetaTxVersion
	} else {
		utxoVM.defaultTxVersion = StableTxVersion
	}

	latestBlockid, findErr := utxoVM.metaTable.Get([]byte(LatestBlockKey))
	if findErr == nil {
		utxoVM.latestBlockid = latestBlockid
	} else {
		if findErr.Error() != kverr.ErrNotFound.Error() {
			return nil, findErr
		}
	}
	utxoTotalBytes, findTotalErr := utxoVM.metaTable.Get([]byte(UTXOTotalKey))
	if findTotalErr == nil {
		total := big.NewInt(0)
		total.SetBytes(utxoTotalBytes)
		utxoVM.utxoTotal = total
	} else {
		if findTotalErr.Error() != kverr.ErrNotFound.Error() {
			return nil, findTotalErr
		}
		//说明是1.1.1版本，没有utxo total字段, 估算一个
		utxoVM.utxoTotal = ledger.GetEstimatedTotal()
		xlog.Info("utxo total is estimated", "total", utxoVM.utxoTotal)
	}
	loadErr := utxoVM.loadUnconfirmedTxFromDisk()
	if loadErr != nil {
		xlog.Warn("faile to load unconfirmed tx from disk", "loadErr", loadErr)
		return nil, loadErr
	}
	return utxoVM, nil
}

// ClearCache 清空cache, 写盘失败的时候
func (uv *UtxoVM) ClearCache() {
	uv.utxoCache = NewUtxoCache(uv.cacheSize)
	uv.prevFoundKeyCache = common.NewLRUCache(uv.cacheSize)
	uv.clearBalanceCache()
	uv.model3.CleanCache()
	uv.xlog.Warn("clear utxo cache")
}

func (uv *UtxoVM) clearBalanceCache() {
	uv.xlog.Warn("clear balance cache")
	uv.balanceCache = common.NewLRUCache(uv.cacheSize) //清空balanceCache
	uv.balanceViewDirty = map[string]bool{}            //清空cache dirty flag表
	uv.model3.CleanCache()
}

// RegisterVM add new contract VM
func (uv *UtxoVM) RegisterVM(name string, vm contract.ContractInterface, priv int) bool {
	return uv.smartContract.RegisterHandler(name, vm, priv)
}

// UnRegisterVM remove contract VM
func (uv *UtxoVM) UnRegisterVM(name string, priv int) {
	uv.smartContract.Remove(name, priv)
}

// RegisterVM3 add Xuper3 contract VM
func (uv *UtxoVM) RegisterVM3(module string, vm contract.VirtualMachine) error {
	return uv.vmMgr3.RegisterVM(module, vm)
}

// RegisterVAT add VAT
func (uv *UtxoVM) RegisterVAT(name string, vat vat.VATInterface, whiteList map[string]bool) {
	uv.vatHandler.RegisterHandler(name, vat, whiteList)
	uv.xlog.Trace("RegisterVAT", "vathandler", uv.vatHandler)
}

// UnRegisterVAT remove VAT
func (uv *UtxoVM) UnRegisterVAT(name string) {
	uv.vatHandler.Remove(name)
}

func (uv *UtxoVM) updateLatestBlockid(newBlockid []byte, batch kvdb.Batch, reason string) error {
	batch.Put(append([]byte(pb.MetaTablePrefix), []byte(LatestBlockKey)...), newBlockid)
	writeErr := batch.Write()
	if writeErr != nil {
		uv.ClearCache()
		uv.xlog.Warn(reason, "writeErr", writeErr)
		return writeErr
	}
	uv.latestBlockid = newBlockid
	return nil
}

func (uv *UtxoVM) updateUtxoTotal(delta *big.Int, batch kvdb.Batch, inc bool) {
	if inc {
		uv.utxoTotal = uv.utxoTotal.Add(uv.utxoTotal, delta)
	} else {
		uv.utxoTotal = uv.utxoTotal.Sub(uv.utxoTotal, delta)
	}
	batch.Put(append([]byte(pb.MetaTablePrefix), []byte(UTXOTotalKey)...), uv.utxoTotal.Bytes())
}

// GenerateAwardTx 生成系统奖励的交易, 比如矿工挖矿所得
func (uv *UtxoVM) GenerateAwardTx(address []byte, awardAmount string, desc []byte) (*pb.Transaction, error) {
	utxoTx := &pb.Transaction{Version: TxVersion}
	amount := big.NewInt(0)
	amount.SetString(awardAmount, 10) // 10进制转换大整数
	if amount.Cmp(big.NewInt(0)) < 0 {
		return nil, ErrNegativeAmount
	}
	txOutput := &pb.TxOutput{}
	txOutput.ToAddr = []byte(address)
	txOutput.Amount = amount.Bytes()
	utxoTx.TxOutputs = append(utxoTx.TxOutputs, txOutput)
	utxoTx.Desc = desc
	utxoTx.Coinbase = true
	utxoTx.Timestamp = time.Now().UnixNano()
	utxoTx.Txid, _ = txhash.MakeTransactionID(utxoTx)
	return utxoTx, nil
}

// GenerateEmptyTx 生成只有Desc的Tx
func (uv *UtxoVM) GenerateEmptyTx(desc []byte) (*pb.Transaction, error) {
	utxoTx := &pb.Transaction{Version: TxVersion}
	utxoTx.Desc = desc
	utxoTx.Timestamp = time.Now().UnixNano()
	txid, err := txhash.MakeTransactionID(utxoTx)
	utxoTx.Txid = txid
	utxoTx.Autogen = true
	return utxoTx, err
}

// GenerateRootTx 通过json内容生成创世区块的交易
func (uv *UtxoVM) GenerateRootTx(js []byte) (*pb.Transaction, error) {
	jsObj := &RootJSON{}
	jsErr := json.Unmarshal(js, jsObj)
	if jsErr != nil {
		uv.xlog.Warn("failed to parse json", "js", string(js), "jsErr", jsErr)
		return nil, jsErr
	}
	utxoTx := &pb.Transaction{Version: RootTxVersion}
	for _, pd := range jsObj.Predistribution {
		amount := big.NewInt(0)
		amount.SetString(pd.Quota, 10) // 10进制转换大整数
		if amount.Cmp(big.NewInt(0)) < 0 {
			return nil, ErrNegativeAmount
		}
		txOutput := &pb.TxOutput{}
		txOutput.ToAddr = []byte(pd.Address)
		txOutput.Amount = amount.Bytes()
		utxoTx.TxOutputs = append(utxoTx.TxOutputs, txOutput)
	}
	utxoTx.Desc = js
	utxoTx.Coinbase = true
	utxoTx.Txid, _ = txhash.MakeTransactionID(utxoTx)
	return utxoTx, nil
}

//SelectUtxos 选择足够的utxo
//输入: 转账人地址、公钥、金额、是否需要锁定utxo
//输出：选出的utxo、utxo keys、实际构成的金额(可能大于需要的金额)、错误码
func (uv *UtxoVM) SelectUtxos(fromAddr string, fromPubKey string, totalNeed *big.Int, needLock, excludeUnconfirmed bool) ([]*pb.TxInput, [][]byte, *big.Int, error) {
	uv.clearExpiredLocks()
	curLedgerHeight := uv.ledger.GetMeta().TrunkHeight
	willLockKeys := make([][]byte, 0)
	utxoTotal := big.NewInt(0)
	foundEnough := false
	// 先从cache里找找，不够再从leveldb找,因为leveldb prefix scan比较慢
	cacheKeys := map[string]bool{}
	txInputs := []*pb.TxInput{}
	uv.clearExpiredLocks()
	uv.utxoCache.Lock()
	if l2Cache, exist := uv.utxoCache.Available[fromAddr]; exist {
		for uKey, uItem := range l2Cache {
			if uItem.FrozenHeight > curLedgerHeight || uItem.FrozenHeight == -1 {
				uv.xlog.Trace("utxo still frozen, skip it", "uKey", uKey, " fheight", uItem.FrozenHeight)
				continue
			}
			if needLock {
				if uv.tryLockKey([]byte(uKey)) {
					willLockKeys = append(willLockKeys, []byte(uKey))
				} else {
					uv.xlog.Debug("can not lock the utxo key, conflict", "uKey", uKey)
					continue
				}
			} else if uv.isLocked([]byte(uKey)) {
				uv.xlog.Debug("skip locked utxo key", "uKey", uKey)
				continue
			}

			keyTuple := strings.Split(uKey[1:], "_") // [1:] 是为了剔除表名字前缀
			refTxid, _ := hex.DecodeString(keyTuple[len(keyTuple)-2])
			if excludeUnconfirmed { //必须依赖已经上链的tx的UTXO
				isOnChain := uv.ledger.IsTxInTrunk(refTxid)
				if !isOnChain {
					continue
				}
			}
			uv.utxoCache.Use(fromAddr, uKey)
			utxoTotal.Add(utxoTotal, uItem.Amount)
			txInput := &pb.TxInput{}
			txInput.RefTxid = refTxid
			offset, _ := strconv.Atoi(keyTuple[len(keyTuple)-1])
			txInput.RefOffset = int32(offset)
			txInput.FromAddr = []byte(fromAddr)
			txInput.Amount = uItem.Amount.Bytes()
			txInput.FrozenHeight = uItem.FrozenHeight
			txInputs = append(txInputs, txInput)
			cacheKeys[uKey] = true
			if utxoTotal.Cmp(totalNeed) >= 0 {
				foundEnough = true
				break
			}
		}
	}
	uv.utxoCache.Unlock()
	if !foundEnough {
		// 底层key: table_prefix from_addr "_" txid "_" offset
		addrPrefix := pb.UTXOTablePrefix + fromAddr + "_"
		var middleKey []byte
		preFoundUtxoKey, mdOK := uv.prevFoundKeyCache.Get(fromAddr)
		if mdOK {
			middleKey = preFoundUtxoKey.([]byte)
		}
		it := kvdb.NewQuickIterator(uv.ldb, []byte(addrPrefix), middleKey)
		defer it.Release()
		for it.Next() {
			key := append([]byte{}, it.Key()...)
			uBinary := it.Value()
			uItem := &UtxoItem{}
			uErr := uItem.Loads(uBinary)
			if uErr != nil {
				return nil, nil, nil, uErr
			}
			if _, inCache := cacheKeys[string(key)]; inCache {
				continue // cache已经命中了，跳过
			}
			if uItem.FrozenHeight > curLedgerHeight || uItem.FrozenHeight == -1 {
				uv.xlog.Trace("utxo still frozen, skip it", "key", string(key), "fheight", uItem.FrozenHeight)
				continue
			}
			if needLock {
				if uv.tryLockKey(key) {
					willLockKeys = append(willLockKeys, key)
				} else {
					uv.xlog.Debug("can not lock the utxo key, conflict", "key", string(key))
					continue
				}
			} else if uv.isLocked(key) {
				uv.xlog.Debug("skip locked utxo key", "key", string(key))
				continue
			}
			keyTuple := bytes.Split(key[1:], []byte("_")) // [1:] 是为了剔除表名字前缀
			refTxid, _ := hex.DecodeString(string(keyTuple[len(keyTuple)-2]))
			if excludeUnconfirmed { //必须依赖已经上链的tx的UTXO
				isOnChain := uv.ledger.IsTxInTrunk(refTxid)
				if !isOnChain {
					continue
				}
			}
			txInput := &pb.TxInput{}
			txInput.RefTxid = refTxid
			offset, _ := strconv.Atoi(string(keyTuple[len(keyTuple)-1]))
			txInput.RefOffset = int32(offset)
			txInput.FromAddr = []byte(fromAddr)
			txInput.Amount = uItem.Amount.Bytes()
			txInput.FrozenHeight = uItem.FrozenHeight
			txInputs = append(txInputs, txInput)
			utxoTotal.Add(utxoTotal, uItem.Amount) // utxo累加
			// uv.xlog.Debug("select", "utxo_amount", utxo_amount, "txid", fmt.Sprintf("%x", txInput.RefTxid))
			if utxoTotal.Cmp(totalNeed) >= 0 { // 找到了足够的utxo用于支付
				foundEnough = true
				uv.prevFoundKeyCache.Add(fromAddr, key)
				break
			}
		}
		if it.Error() != nil {
			return nil, nil, nil, it.Error()
		}
	}
	if !foundEnough {
		for _, lk := range willLockKeys {
			uv.unlockKey(lk)
		}
		return nil, nil, nil, ErrNoEnoughUTXO // 余额不足啦
	}
	return txInputs, willLockKeys, utxoTotal, nil
}

// PreExec the Xuper3 contract model uses previous execution to generate RWSets
func (uv *UtxoVM) PreExec(req *pb.InvokeRequest, hd *global.XContext) (*pb.InvokeResponse, error) {
	moduleName := req.ModuleName
	vm, err := uv.vmMgr3.GetVM(moduleName)
	if err != nil {
		return nil, err
	}
	modelCache, err := xmodel.NewXModelCache(uv.GetXModel(), true)
	if err != nil {
		return nil, err
	}
	ctx, err := vm.NewContext(req.GetContractName(), modelCache, contract.MaxGasLimit)
	if err != nil {
		return nil, err
	}
	response, err := ctx.Invoke(req.MethodName, req.Args)
	defer ctx.Release()
	if err != nil {
		return nil, err
	}
	inputs, outputs, err := modelCache.GetRWSets()
	if err != nil {
		return nil, err
	}
	rsps := &pb.InvokeResponse{
		Inputs:   xmodel.GetTxInputs(inputs),
		Outputs:  xmodel.GetTxOutputs(outputs),
		Response: response,
		GasUsed:  ctx.GasUsed(),
	}
	return rsps, nil
}

// 加载所有未确认的订单表到内存
// 参数:	dedup : true-删除已经确认tx, false-保留已经确认tx
//  返回：txMap : txid -> Transaction
//        txGraph:  txid ->  [依赖此txid的tx]
func (uv *UtxoVM) sortUnconfirmedTx() (map[string]*pb.Transaction, TxGraph, error) {
	// 构造反向依赖关系图, key是被依赖的交易
	txMap := map[string]*pb.Transaction{}
	txGraph := TxGraph{}
	uv.unconfirmTxInMem.Range(func(k, v interface{}) bool {
		txMap[k.(string)] = v.(*pb.Transaction)
		txGraph[k.(string)] = []string{}
		return true
	})
	var totalDelay int64
	now := time.Now().UnixNano()
	for txID, tx := range txMap {
		totalDelay += (now - tx.Timestamp)
		for _, refTx := range tx.TxInputs {
			refTxID := string(refTx.RefTxid)
			if _, exist := txMap[refTxID]; !exist {
				// 说明引用的tx不是在unconfirm里面
				continue
			}
			txGraph[refTxID] = append(txGraph[refTxID], txID)
		}
		for _, txIn := range tx.TxInputsExt {
			refTxID := string(txIn.RefTxid)
			if _, exist := txMap[refTxID]; !exist {
				continue
			}
			txGraph[refTxID] = append(txGraph[refTxID], txID)
		}
	}
	if len(txMap) > 0 {
		avgDelay := totalDelay / int64(len(txMap)) //平均unconfirm滞留时间
		uv.xlog.Info("average unconfirm delay", "micro-senconds", avgDelay/1e6, "count", len(txMap))
	}
	return txMap, txGraph, nil
}

//从disk还原unconfirm表到内存, 初始化的时候
func (uv *UtxoVM) loadUnconfirmedTxFromDisk() error {
	iter := uv.ldb.NewIteratorWithPrefix([]byte(pb.UnconfirmedTablePrefix))
	defer iter.Release()
	for iter.Next() {
		rawKey := iter.Key()
		txid := string(rawKey[1:])
		uv.xlog.Trace("  load unconfirmed tx from db", "txid", fmt.Sprintf("%x", txid))
		txBuf := iter.Value()
		tx := &pb.Transaction{}
		pbErr := proto.Unmarshal(txBuf, tx)
		if pbErr != nil {
			return pbErr
		}
		uv.unconfirmTxInMem.Store(txid, tx)
	}
	return nil
}

// GetUnconfirmedTx 挖掘一批unconfirmed的交易打包，返回的结果要保证是按照交易执行的先后顺序
// maxSize: 打包交易最大的长度（in byte）, -1 表示不限制
func (uv *UtxoVM) GetUnconfirmedTx(dedup bool) ([]*pb.Transaction, error) {
	if uv.asyncMode {
		dedup = false
	}
	var selectedTxs []*pb.Transaction
	txMap, txGraph, loadErr := uv.sortUnconfirmedTx()
	if loadErr != nil {
		return nil, loadErr
	}
	// 拓扑排序，输出的顺序是被依赖的在前，依赖方在后
	outputTxList, unexpectedCyclic := TopSortDFS(txGraph)
	if len(unexpectedCyclic) > 0 { // 交易之间检测出了环形的依赖关系
		uv.xlog.Warn("transaction conflicted", "unexpectedCyclic", unexpectedCyclic)
		return nil, ErrUnexpected
	}
	for _, txid := range outputTxList {
		if dedup && uv.ledger.IsTxInTrunk([]byte(txid)) {
			continue
		}
		selectedTxs = append(selectedTxs, txMap[txid])
	}
	return selectedTxs, nil
}

// DebugTx print transaction info in log for debug
func (uv *UtxoVM) DebugTx(tx *pb.Transaction) error {
	uv.xlog.Debug("debug tx", "txid", fmt.Sprintf("%x", tx.Txid))
	for offset, txInput := range tx.TxInputs {
		addr := txInput.FromAddr
		txid := txInput.RefTxid
		refOffset := txInput.RefOffset
		amountBytes := txInput.Amount
		amount := big.NewInt(0)
		amount.SetBytes(amountBytes)
		uv.xlog.Debug("txinput", "offset", offset, "addr", string(addr),
			"reftxid", fmt.Sprintf("%x", txid), "refoffset", refOffset, "amount", amount)
	}
	for offset, txOutput := range tx.TxOutputs {
		addr := txOutput.ToAddr
		amount := big.NewInt(0)
		amount.SetBytes(txOutput.Amount)
		uv.xlog.Debug("txoutput", "offset", offset, "addr", string(addr), "amount", amount)
	}
	return nil
}

// addBalance 增加cache中的Balance
func (uv *UtxoVM) addBalance(addr []byte, delta *big.Int) {
	balance, hitCache := uv.balanceCache.Get(string(addr))
	if hitCache {
		iBalance := balance.(*big.Int)
		iBalance.Add(iBalance, delta)
	} else {
		uv.balanceViewDirty[string(addr)] = true
	}
}

// subBalance 减少cache中的Balance
func (uv *UtxoVM) subBalance(addr []byte, delta *big.Int) {
	balance, hitCache := uv.balanceCache.Get(string(addr))
	if hitCache {
		iBalance := balance.(*big.Int)
		iBalance.Sub(iBalance, delta)
	} else {
		uv.balanceViewDirty[string(addr)] = true
	}
}

// payFee 扣除小费给矿工
func (uv *UtxoVM) payFee(tx *pb.Transaction, batch kvdb.Batch, block *pb.InternalBlock) error {
	for offset, txOutput := range tx.TxOutputs {
		addr := txOutput.ToAddr
		if !bytes.Equal(addr, []byte(FeePlaceholder)) {
			continue
		}
		addr = block.Proposer // 占位符替换为矿工
		utxoKey := GenUtxoKeyWithPrefix(addr, tx.Txid, int32(offset))
		uItem := &UtxoItem{}
		uItem.Amount = big.NewInt(0)
		uItem.Amount.SetBytes(txOutput.Amount)
		uItemBinary, uErr := uItem.Dumps()
		if uErr != nil {
			return uErr
		}
		batch.Put([]byte(utxoKey), uItemBinary) // 插入本交易产生的utxo
		uv.addBalance(addr, uItem.Amount)
		uv.utxoCache.Insert(string(addr), utxoKey, uItem)
		uv.xlog.Trace("    insert fee utxo key", "utxoKey", utxoKey, "amount", uItem.Amount.String())
	}
	return nil
}

// undoPayFee 回滚小费
func (uv *UtxoVM) undoPayFee(tx *pb.Transaction, batch kvdb.Batch, block *pb.InternalBlock) error {
	for offset, txOutput := range tx.TxOutputs {
		addr := txOutput.ToAddr
		if !bytes.Equal(addr, []byte(FeePlaceholder)) {
			continue
		}
		addr = block.Proposer
		utxoKey := GenUtxoKeyWithPrefix(addr, tx.Txid, int32(offset))
		batch.Delete([]byte(utxoKey)) // 删除产生的UTXO
		uv.utxoCache.Remove(string(addr), utxoKey)
		uv.subBalance(addr, big.NewInt(0).SetBytes(txOutput.Amount))
		uv.xlog.Info("    undo delete fee utxo key", "utxoKey", utxoKey)
	}
	return nil
}

// doTxInternal 交易执行的核心逻辑
// @tx: 要执行的transaction
// @batch: 对数据的变更写入到batch对象
func (uv *UtxoVM) doTxInternal(tx *pb.Transaction, batch kvdb.Batch) error {
	if !uv.asyncMode {
		uv.xlog.Trace("  start to dotx", "txid", fmt.Sprintf("%x", tx.Txid))
	}
	if err := uv.checkInputEqualOutput(tx); err != nil {
		return err
	}
	err := uv.model3.DoTx(tx, batch)
	if err != nil {
		uv.xlog.Warn("model3.DoTx failed", "err", err)
		return ErrRWSetInvalid
	}
	for _, txInput := range tx.TxInputs {
		addr := txInput.FromAddr
		txid := txInput.RefTxid
		offset := txInput.RefOffset
		utxoKey := GenUtxoKeyWithPrefix(addr, txid, offset)
		batch.Delete([]byte(utxoKey)) // 删除用掉的utxo
		uv.utxoCache.Remove(string(addr), utxoKey)
		uv.subBalance(addr, big.NewInt(0).SetBytes(txInput.Amount))
		if !uv.asyncMode {
			uv.xlog.Trace("    delete utxo key", "utxoKey", utxoKey)
		}
	}
	for offset, txOutput := range tx.TxOutputs {
		addr := txOutput.ToAddr
		if bytes.Equal(addr, []byte(FeePlaceholder)) {
			continue
		}
		utxoKey := GenUtxoKeyWithPrefix(addr, tx.Txid, int32(offset))
		uItem := &UtxoItem{}
		uItem.Amount = big.NewInt(0)
		uItem.Amount.SetBytes(txOutput.Amount)
		uItem.FrozenHeight = txOutput.FrozenHeight
		uItemBinary, uErr := uItem.Dumps()
		if uErr != nil {
			return uErr
		}
		batch.Put([]byte(utxoKey), uItemBinary) // 插入本交易产生的utxo
		uv.utxoCache.Insert(string(addr), utxoKey, uItem)
		uv.addBalance(addr, uItem.Amount)
		if !uv.asyncMode {
			uv.xlog.Trace("    insert utxo key", "utxoKey", utxoKey, "amount", uItem.Amount.String())
		}
		if tx.Coinbase {
			// coinbase交易（包括创始块和挖矿奖励)会增加系统的总资产
			uv.updateUtxoTotal(uItem.Amount, batch, true)
		}
	}
	return nil
}

// undoTxInternal 交易回滚的核心逻辑
// @tx: 要执行的transaction
// @batch: 对数据的变更写入到batch对象
// @tx_in_block:  true说明这个tx是来自区块, false说明是回滚unconfirm表的交易
func (uv *UtxoVM) undoTxInternal(tx *pb.Transaction, batch kvdb.Batch) error {
	err := uv.model3.UndoTx(tx, batch)
	if err != nil {
		uv.xlog.Warn("model3.UndoTx failed", "err", err)
		return ErrRWSetInvalid
	}
	for _, txInput := range tx.TxInputs {
		addr := txInput.FromAddr
		txid := txInput.RefTxid
		offset := txInput.RefOffset
		amount := txInput.Amount
		utxoKey := GenUtxoKeyWithPrefix(addr, txid, offset)
		uItem := &UtxoItem{}
		uItem.Amount = big.NewInt(0)
		uItem.Amount.SetBytes(amount)
		uItem.FrozenHeight = txInput.FrozenHeight
		uv.utxoCache.Insert(string(addr), utxoKey, uItem)
		uBinary, uErr := uItem.Dumps()
		if uErr != nil {
			return uErr
		}
		batch.Put([]byte(utxoKey), uBinary) // 退还用掉的UTXO
		uv.unlockKey([]byte(utxoKey))
		uv.addBalance(addr, uItem.Amount)
		uv.xlog.Trace("    undo insert utxo key", "utxoKey", utxoKey)
	}
	for offset, txOutput := range tx.TxOutputs {
		addr := txOutput.ToAddr
		if bytes.Equal(addr, []byte(FeePlaceholder)) {
			continue
		}
		utxoKey := GenUtxoKeyWithPrefix(addr, tx.Txid, int32(offset))
		batch.Delete([]byte(utxoKey)) // 删除产生的UTXO
		uv.utxoCache.Remove(string(addr), utxoKey)
		uv.subBalance(addr, big.NewInt(0).SetBytes(txOutput.Amount))
		uv.xlog.Trace("    undo delete utxo key", "utxoKey", utxoKey)
		if tx.Coinbase {
			// coinbase交易（包括创始块和挖矿奖励), 回滚会导致系统总资产缩水
			delta := big.NewInt(0)
			delta.SetBytes(txOutput.Amount)
			uv.updateUtxoTotal(delta, batch, false)
		}
	}
	return nil
}

func (uv *UtxoVM) runContract(blockid []byte, tx *pb.Transaction, autogenTxList *[]*pb.Transaction, deadline int64) error {
	// 去掉高度判断，因为在CreateBlockChain的时候，没有传递矿工的地址和keys。
	if tx.Autogen && autogenTxList != nil { // 自动生成的tx, 需要校验下
		if len(*autogenTxList) == 0 {
			uv.xlog.Warn("autogenTxList has been drained")
			return ErrInvalidAutogenTx
		}
		if !bytes.Equal(tx.Desc, (*autogenTxList)[0].Desc) {
			uv.xlog.Warn("mismatch contract desc", "expected", string((*autogenTxList)[0].Desc), "got", string(tx.Desc))
			return ErrInvalidAutogenTx
		}
		uv.xlog.Debug("autogen tx contract checked ok", "desc", string(tx.Desc))
		*autogenTxList = (*autogenTxList)[1:] //pop front
	}
	if txDesc, ok := uv.isSmartContract(tx.Desc); ok { // 交易需要执行智能合约
		txDesc.Tx = tx
		txDesc.Deadline = deadline

		if uv.MustVAT(txDesc) && !txDesc.Tx.Autogen {
			return fmt.Errorf("this contract %s.%s can only be auto generated by proposal", txDesc.Module, txDesc.Method)
		}

		if scErr := uv.smartContract.Run(txDesc); scErr != nil {
			uv.xlog.Warn("failed to Run contract", "scErr", scErr, "txid", fmt.Sprintf("%x", txDesc.Tx.Txid))
			return scErr
		}
	}
	return nil
}

// RollbackContract rollback given contract
func (uv *UtxoVM) RollbackContract(blockid []byte, tx *pb.Transaction) error {
	if txDesc, ok := uv.isSmartContract(tx.Desc); ok { // 交易需要执行智能合约
		txDesc.Tx = tx
		scErr := uv.smartContract.Rollback(txDesc)
		if scErr != nil {
			uv.xlog.Warn("failed to Rollback contract", "scErr", scErr)
			return scErr
		}
	}
	return nil
}

// 同步阻塞方式执行交易
func (uv *UtxoVM) doTxSync(tx *pb.Transaction) error {
	pbTxBuf, pbErr := proto.Marshal(tx)
	if pbErr != nil {
		uv.xlog.Warn("    fail to marshal tx", "pbErr", pbErr)
		return pbErr
	}
	recvTime := time.Now().Unix()
	uv.mutex.Lock()
	defer uv.mutex.Unlock() //lock guard
	waitTime := time.Now().Unix() - recvTime
	if waitTime > TxWaitTimeout {
		uv.xlog.Warn("dotx wait too long!", "waitTime", waitTime, "txid", fmt.Sprintf("%x", tx.Txid))
	}
	_, exist := uv.unconfirmTxInMem.Load(string(tx.Txid))
	if exist {
		uv.xlog.Debug("this tx already in unconfirm table, when DoTx", "txid", fmt.Sprintf("%x", tx.Txid))
		return ErrAlreadyInUnconfirmed
	}
	batch := uv.ldb.NewBatch()
	doErr := uv.doTxInternal(tx, batch)
	if doErr != nil {
		uv.xlog.Warn("doTxInternal failed, when DoTx", "doErr", doErr)
		return doErr
	}
	batch.Put(append([]byte(pb.UnconfirmedTablePrefix), tx.Txid...), pbTxBuf)
	uv.xlog.Debug("print tx size when DoTx", "tx_size", batch.ValueSize(), "txid", fmt.Sprintf("%x", tx.Txid))
	writeErr := batch.Write()
	if writeErr != nil {
		uv.ClearCache()
		uv.xlog.Warn("fail to save to ldb", "writeErr", writeErr)
		return writeErr
	}
	uv.unconfirmTxInMem.Store(string(tx.Txid), tx)
	return nil
}

func (uv *UtxoVM) doTxAsync(tx *pb.Transaction) error {
	_, exist := uv.unconfirmTxInMem.Load(string(tx.Txid))
	if exist {
		uv.xlog.Debug("this tx already in unconfirm table, when DoTx", "txid", fmt.Sprintf("%x", tx.Txid))
		return ErrAlreadyInUnconfirmed
	}
	inboundTx := &InboundTx{tx: tx}
	uv.inboundTxChan <- inboundTx
	return nil
}

// VerifyTx check the tx signature and permission
func (uv *UtxoVM) VerifyTx(tx *pb.Transaction) (bool, error) {
	if uv.asyncMode {
		return true, nil //异步模式推迟到后面校验
	}
	isValid, err := uv.ImmediateVerifyTx(tx, false)
	if err != nil || !isValid {
		uv.xlog.Warn("ImmediateVerifyTx failed", "error", err,
			"AuthRequire ", tx.AuthRequire, "AuthRequireSigns ", tx.AuthRequireSigns,
			"Initiator", tx.Initiator, "InitiatorSigns", tx.InitiatorSigns)
	}
	return isValid, err
}

// ImmediateVerifyTx verify tx Immediately
func (uv *UtxoVM) ImmediateVerifyTx(tx *pb.Transaction, isRootTx bool) (bool, error) {
	if !isRootTx && tx.Version == RootTxVersion {
		return false, ErrVersionInvalid
	}
	if tx.Version > BetaTxVersion {
		return false, ErrVersionInvalid
	}
	if tx.Version >= TxVersion {
		// verify rwset
		ok, err := uv.verifyTxRWSets(tx)
		if err != nil && strings.HasPrefix(err.Error(), "Gas not enough") {
			err = ErrGasNotEnough
		} else {
			err = ErrRWSetInvalid
		}
		if !ok {
			return ok, err
		}

		//verify txid (由于老版本txid生成不稳定，因此只对新版本校验txid)
		txid, err := txhash.MakeTransactionID(tx)
		if err != nil {
			uv.xlog.Warn("ImmediateVerifyTx: call MakeTransactionID failed", "error", err)
			return false, err
		}
		if bytes.Compare(tx.Txid, txid) != 0 {
			uv.xlog.Warn("ImmediateVerifyTx: txid not match", "tx.Txid", tx.Txid, "txid", txid)
			return false, fmt.Errorf("Txid verify failed")
		}

		// verify contract method acl permission
		ok, err = uv.verifyContractValid(tx)
		if !ok {
			return ok, err
		}

		// verify RWSet acl permission
		ok, err = uv.verifyRWACLValid(tx)
		if !ok {
			err = ErrRWAclNotEnough
			return ok, err
		}

		// veify initiator and tx input signatures
		return uv.verifyTxUtxo(tx)
	}
	return true, nil
}

// verifyTxSign 纯密码学验证
func (uv *UtxoVM) verifyTxSign(tx *pb.Transaction) (bool, error) {
	if len(tx.GetAuthRequire()) != len(tx.GetAuthRequireSigns()) {
		return false, fmt.Errorf("tx.AuthRequire length not equal to tx.AuthRequireSigns")
	}
	digestHash, dhErr := txhash.MakeTxDigestHash(tx)
	if dhErr != nil {
		return false, dhErr
	}
	verifiedAddrs := map[string]bool{}
	for i, ak := range tx.AuthRequire {
		if verifiedAddrs[string(ak)] {
			continue
		}
		if ok, _ := pm.IdentifyAK(ak, tx.AuthRequireSigns[i], digestHash); !ok {
			return false, errors.New("utxo.verifyTxSign error")
		}
		verifiedAddrs[ak] = true
	}
	return true, nil
}

func (uv *UtxoVM) verifyAkURIPermission(name string, aksuri []string, signs []*pb.SignatureInfo, data []byte) (bool, error) {
	akType := acl.IsAccount(name)
	if akType == 1 {
		// Identify account
		acl, err := uv.queryAccountACL(name)
		if err != nil || acl == nil {
			// valid account should have ACL info, so this account might not exsit
			uv.xlog.Warn("verifyAkURIPermission error, account might not exist", "account", name, "error", err)
			return false, ErrInvalidAccount
		}
		if ok, err := pm.IdentifyAccount(string(name), aksuri, signs, data, uv.aclMgr); !ok {
			uv.xlog.Warn("verifyAkURIPermission error, failed to IdentifyAccount", "error", err)
			return false, errors.New("verifyTxUtxo veify account error")
		}
	} else if akType == 0 {
		// Identify address
		index, err := getAkIndex(name, aksuri)
		if err != nil {
			return false, errors.New("verifyTxUtxo veify address signinfo error")
		}
		if ok, err := pm.IdentifyAK(name, signs[index], data); !ok {
			uv.xlog.Warn("verifyAkURIPermission error, failed to IdentifyAK", "error", err)
			return false, errors.New("verifyTxUtxo veify address utxo error")
		}
	} else {
		uv.xlog.Warn("verifyAkURIPermission error, Invalid account/address name", "name", name)
		return false, ErrInvalidAccount
	}
	return true, nil
}

func (uv *UtxoVM) verifyInitiator(tx *pb.Transaction, digestHash []byte) (bool, error) {
	aksuri := make([]string, 0)
	name := tx.Initiator
	nameType := acl.IsAccount(name)
	if nameType < 0 {
		return false, fmt.Errorf("Invalid account/address name, name=%s", string(name))
	}
	for _, signInfo := range tx.InitiatorSigns {
		pk, err := uv.cryptoClient.GetEcdsaPublicKeyFromJSON([]byte(signInfo.PublicKey))
		if err != nil {
			return false, err
		}
		address, err := uv.cryptoClient.GetAddressFromPublicKey(pk)
		if err != nil {
			return false, err
		}
		if nameType == 1 {
			address = name + "/" + address
		}
		aksuri = append(aksuri, address)
	}
	result, err := uv.verifyAkURIPermission(name, aksuri, tx.InitiatorSigns, digestHash)
	if err != nil || !result {
		uv.xlog.Warn("verifyAkURIPermission failed", "addr", name,
			"InitiatorUri", aksuri, "InitiatorSigns", tx.InitiatorSigns, "error", err)
	}
	return result, err
}

func (uv *UtxoVM) verifyTxUtxo(tx *pb.Transaction) (bool, error) {
	if len(tx.GetAuthRequire()) != len(tx.GetAuthRequireSigns()) {
		return false, fmt.Errorf("tx.AuthRequire length not equal to tx.AuthRequireSigns")
	}
	digestHash, dhErr := txhash.MakeTxDigestHash(tx)
	if dhErr != nil {
		return false, dhErr
	}
	verifiedAddrs := map[string]bool{}
	// verify initiator signatures
	isValid, err := uv.verifyInitiator(tx, digestHash)
	if err != nil || !isValid {
		return false, err
	}
	verifiedAddrs[string(tx.Initiator)] = true
	// verify tx input signatures
	for _, txInput := range tx.TxInputs {
		if verifiedAddrs[string(txInput.FromAddr)] {
			continue
		}
		ok, err := uv.verifyAkURIPermission(string(txInput.FromAddr), tx.AuthRequire, tx.AuthRequireSigns, digestHash)
		if err != nil || !ok {
			uv.xlog.Warn("verifyAkURIPermission failed", "addr", txInput.FromAddr,
				"AuthRequied", tx.AuthRequire, "AuthSigns", tx.AuthRequireSigns)
			return false, err
		}
		verifiedAddrs[string(txInput.FromAddr)] = true
	}
	return true, nil
}

func getAkIndex(ak string, aksuri []string) (int, error) {
	for i, v := range aksuri {
		if ak == v {
			return i, nil
		}
	}
	return 0, fmt.Errorf("getAkIndex error not found")
}

func getGasLimitFromTx(tx *pb.Transaction) (int64, error) {
	for _, output := range tx.GetTxOutputs() {
		if string(output.GetToAddr()) != "$" {
			continue
		}
		gasLimit := big.NewInt(0).SetBytes(output.GetAmount()).Int64()
		// FIXME: gasLimit从大数过来的，处理溢出问题
		if gasLimit <= 0 {
			return 0, fmt.Errorf("bad gas limit %d", gasLimit)
		}
		return gasLimit, nil
	}
	// FIXME: 没有小费的tx如何得到gas limit?
	return 0, ErrGasNotEnough
}

// verifyTxRWSets verify tx read sets and write sets
func (uv *UtxoVM) verifyTxRWSets(tx *pb.Transaction) (bool, error) {
	req := tx.GetContractRequest()
	if req == nil {
		if tx.GetTxInputsExt() != nil || tx.GetTxOutputsExt() != nil {
			uv.xlog.Error("verifyTxRWSets error", "error", ErrInvalidTxExt.Error())
			return false, ErrInvalidTxExt
		}
		return true, nil
	}
	moduleName := req.GetModuleName()
	vm, err := uv.vmMgr3.GetVM(moduleName)
	if err != nil {
		return false, err
	}
	env, err := uv.model3.PrepareEnv(tx)
	if err != nil {
		return false, err
	}
	gasLimit, err := getGasLimitFromTx(tx)
	if err != nil {
		return false, err
	}
	uv.xlog.Trace("get gas limit from tx", "gasLimit", gasLimit, "txid", hex.EncodeToString(tx.Txid))
	ctx, err := vm.NewContext(req.GetContractName(), env.GetModelCache(), gasLimit)
	if err != nil {
		return false, err
	}
	_, err = ctx.Invoke(req.MethodName, req.Args)
	defer ctx.Release()
	if err != nil {
		return false, err
	}
	_, writeSet, err := env.GetModelCache().GetRWSets()
	if err != nil {
		return false, err
	}
	uv.xlog.Trace("verifyTxRWSets", "env.output", env.GetOutputs(), "writeSet", writeSet)
	ok := xmodel.Equal(env.GetOutputs(), writeSet)
	if !ok {
		return false, fmt.Errorf("Verify error")
	}
	return true, nil
}

// IsInUnConfirm check if the given txid is in unconfirm table
// return true if the txid is unconfirmed
func (uv *UtxoVM) IsInUnConfirm(txid string) bool {
	_, exit := uv.unconfirmTxInMem.Load(txid)
	return exit
}

// DoTx 执行一个交易, 影响utxo表和unconfirm-transaction表
func (uv *UtxoVM) DoTx(tx *pb.Transaction) error {
	if tx.Coinbase {
		uv.xlog.Warn("coinbase tx can not be given by PostTx", "txid", global.F(tx.Txid))
		return ErrUnexpected
	}
	if uv.asyncMode {
		return uv.doTxAsync(tx)
	}
	return uv.doTxSync(tx)
}

func (uv *UtxoVM) undoUnconfirmedTx(tx *pb.Transaction, txMap map[string]*pb.Transaction,
	txGraph TxGraph, batch kvdb.Batch, undoDone map[string]bool) error {
	if undoDone[string(tx.Txid)] == true {
		return nil // 说明已经被回滚了
	}
	uv.xlog.Info("    start to undo transaction", "txid", fmt.Sprintf("%x", tx.Txid))
	childrenTxids, exist := txGraph[string(tx.Txid)]
	if exist {
		for _, childTxid := range childrenTxids {
			childTx := txMap[childTxid]
			uv.undoUnconfirmedTx(childTx, txMap, txGraph, batch, undoDone) // 先回滚依赖“我”的交易
		}
	}
	// 下面开始回滚自身
	undoErr := uv.undoTxInternal(tx, batch)
	if undoErr != nil {
		return undoErr
	}
	if !uv.asyncMode {
		batch.Delete(append([]byte(pb.UnconfirmedTablePrefix), tx.Txid...)) // 从unconfirm表删除
	}
	undoDone[string(tx.Txid)] = true
	return nil
}

//执行一个block的时候, 处理本地未确认交易
//返回：被确认的txid集合、err
func (uv *UtxoVM) processUnconfirmTxs(block *pb.InternalBlock, batch kvdb.Batch, needRepost bool) (map[string]bool, map[string]bool, error) {
	if !bytes.Equal(block.PreHash, uv.latestBlockid) {
		uv.xlog.Warn("play failed", "block.PreHash", fmt.Sprintf("%x", block.PreHash),
			"latestBlockid", fmt.Sprintf("%x", uv.latestBlockid))
		return nil, nil, ErrPreBlockMissMatch
	}
	txidsInBlock := map[string]bool{}    // block里面所有的txid
	UTXOKeysInBlock := map[string]bool{} // block里面所有的交易需要用掉的utxo
	keysVersioinInBlock := map[string]string{}
	uv.mutex.Unlock()
	for _, tx := range block.Transactions {
		txidsInBlock[string(tx.Txid)] = true
		for _, txInput := range tx.TxInputs {
			utxoKey := genUtxoKey(txInput.FromAddr, txInput.RefTxid, txInput.RefOffset)
			if UTXOKeysInBlock[utxoKey] { //检查块内的utxo双花情况
				uv.xlog.Warn("found duplicated utxo in same block", "utxoKey", utxoKey, "txid", global.F(tx.Txid))
				uv.mutex.Lock()
				return nil, nil, ErrUTXODuplicated
			}
			UTXOKeysInBlock[utxoKey] = true
		}
		for txOutOffset, txOut := range tx.TxOutputsExt {
			valueVersion := xmodel.MakeVersion(tx.Txid, int32(txOutOffset))
			bucketAndKey := xmodel.MakeRawKey(txOut.Bucket, txOut.Key)
			keysVersioinInBlock[string(bucketAndKey)] = valueVersion
		}
	}
	uv.mutex.Lock()
	// 下面开始处理unconfirmed的交易
	unconfirmTxMap, unconfirmTxGraph, loadErr := uv.sortUnconfirmedTx()
	if loadErr != nil {
		return nil, nil, loadErr
	}
	uv.xlog.Info("unconfirm table size", "unconfirmTxMap", len(unconfirmTxMap))
	undoDone := map[string]bool{}
	unconfirmToConfirm := map[string]bool{}
	for txid, unconfirmTx := range unconfirmTxMap {
		if _, exist := txidsInBlock[string(txid)]; exist {
			// 说明这个交易已经被确认
			if !uv.asyncMode {
				batch.Delete(append([]byte(pb.UnconfirmedTablePrefix), []byte(txid)...))
			}
			uv.xlog.Trace("  delete from unconfirmed", "txid", fmt.Sprintf("%x", txid))
			// 直接从unconfirm表删除, 大部分情况是这样的
			unconfirmToConfirm[txid] = true
			continue
		}
		hasConflict := false
		for _, unconfirmTxInput := range unconfirmTx.TxInputs {
			addr := unconfirmTxInput.FromAddr
			txid := unconfirmTxInput.RefTxid
			offset := unconfirmTxInput.RefOffset
			utxoKey := genUtxoKey(addr, txid, offset)
			if _, exist := UTXOKeysInBlock[utxoKey]; exist {
				// 说明此交易和block里面的交易存在双花冲突，需要回滚, 少数情况
				uv.xlog.Warn("conflict, refuse double spent", "key", utxoKey, "txid", global.F(unconfirmTx.Txid))
				hasConflict = true
				break
			}
		}
		for _, txInputExt := range unconfirmTx.TxInputsExt {
			bucketAndKey := xmodel.MakeRawKey(txInputExt.Bucket, txInputExt.Key)
			localVersion := xmodel.MakeVersion(txInputExt.RefTxid, txInputExt.RefOffset)
			remoteVersion := keysVersioinInBlock[string(bucketAndKey)]
			if localVersion != remoteVersion && remoteVersion != "" {
				txidInVer := xmodel.GetTxidFromVersion(remoteVersion)
				if _, known := unconfirmTxMap[string(txidInVer)]; known {
					continue
				}
				uv.xlog.Warn("inputs version conflict", "key", bucketAndKey, "localVersion", localVersion, "remoteVersion", remoteVersion)
				hasConflict = true
				break
			}
		}
		for txOutOffset, txOut := range unconfirmTx.TxOutputsExt {
			bucketAndKey := xmodel.MakeRawKey(txOut.Bucket, txOut.Key)
			localVersion := xmodel.MakeVersion(unconfirmTx.Txid, int32(txOutOffset))
			remoteVersion := keysVersioinInBlock[string(bucketAndKey)]
			if localVersion != remoteVersion && remoteVersion != "" {
				txidInVer := xmodel.GetTxidFromVersion(remoteVersion)
				if _, known := unconfirmTxMap[string(txidInVer)]; known {
					continue
				}
				uv.xlog.Warn("outputs version conflict", "key", bucketAndKey, "localVersion", localVersion, "remoteVersion", remoteVersion)
				hasConflict = true
				break
			}
		}
		if hasConflict {
			undoErr := uv.undoUnconfirmedTx(unconfirmTx, unconfirmTxMap, unconfirmTxGraph, batch, undoDone)
			if undoErr != nil {
				uv.xlog.Warn("fail to undo tx", "undoErr", undoErr)
				return nil, nil, undoErr
			}
		}
	}
	if needRepost {
		go func() {
			sortTxList, unexpectedCyclic := TopSortDFS(unconfirmTxGraph)
			if len(unexpectedCyclic) > 0 {
				uv.xlog.Warn("transaction conflicted", "unexpectedCyclic", unexpectedCyclic)
				return
			}
			for _, txid := range sortTxList {
				if txidsInBlock[txid] || undoDone[txid] {
					continue
				}
				offlineTx := unconfirmTxMap[txid]
				uv.OfflineTxChan <- offlineTx
			}
		}()
	}
	return unconfirmToConfirm, undoDone, nil
}

// Play do play and repost block
func (uv *UtxoVM) Play(blockid []byte) error {
	return uv.PlayAndRepost(blockid, false, true)
}

// PlayAndRepost 执行一个新收到的block，要求block的pre_hash必须是当前vm的latest_block
// 执行后会更新latestBlockid
func (uv *UtxoVM) PlayAndRepost(blockid []byte, needRepost bool, isRootTx bool) error {
	batch := uv.ldb.NewBatch()
	block, blockErr := uv.ledger.QueryBlock(blockid)
	if blockErr != nil {
		return blockErr
	}
	uv.mutex.Lock()
	defer uv.mutex.Unlock()
	// 下面开始处理unconfirmed的交易
	unconfirmToConfirm, undoDone, err := uv.processUnconfirmTxs(block, batch, needRepost)
	if err != nil {
		return err
	}

	ctx := &contract.TxContext{UtxoBatch: batch, Block: block, LedgerObj: uv.ledger} // 将batch赋值到合约机的上下文
	uv.smartContract.SetContext(ctx)
	autoGenTxList, genErr := uv.GetVATList(block.Height, -1, block.Timestamp)
	if genErr != nil {
		uv.xlog.Warn("get autogen tx list failed", "err", genErr)
		return genErr
	}
	// 进入正题，开始执行block里面的交易，预期不会有冲突了
	uv.xlog.Debug("autogen tx list size, before play block", "len", len(autoGenTxList))
	idx, length := 0, len(block.Transactions)
	for idx < length {
		tx := block.Transactions[idx]
		txid := string(tx.Txid)
		if unconfirmToConfirm[txid] == false { // 本地没预执行过的Tx, 从block中收到的，需要Play执行
			if !tx.Autogen && !tx.Coinbase {
				if ok, err := uv.ImmediateVerifyTx(tx, isRootTx); !ok {
					uv.xlog.Warn("dotx failed to ImmediateVerifyTx", "txid", fmt.Sprintf("%x", tx.Txid), "err", err)
					return errors.New("dotx failed to ImmediateVerifyTx error")
				}
			}
			err := uv.doTxInternal(tx, batch)
			if err != nil {
				uv.xlog.Warn("dotx failed when Play", "txid", fmt.Sprintf("%x", tx.Txid), "err", err)
				return err
			}
		}
		feeErr := uv.payFee(tx, batch, block)
		if feeErr != nil {
			uv.xlog.Warn("payFee failed", "feeErr", feeErr)
			return feeErr
		}
		//如果不是矿工的话，需要执行操作
		//合约的结果校验，任何错误都可能是作恶
		var cErr error
		if idx, cErr = uv.TxOfRunningContractVerify(batch, block, tx, &autoGenTxList, idx); cErr != nil {
			uv.xlog.Warn("TxOfRunningContractVerify failed when playing", "error", cErr, "idx", idx)
			return cErr
		}
	}
	uv.xlog.Debug("autogen tx list size, after play block", "len", len(autoGenTxList))
	if err := uv.smartContract.Finalize(block.Blockid); err != nil {
		uv.xlog.Warn("smart contract.finalize failed", "blockid", fmt.Sprintf("%x", block.Blockid))
		// 合约执行失败，不影响签发块
		return err
	}
	//更新latestBlockid
	persistErr := uv.updateLatestBlockid(block.Blockid, batch, "failed to save block")
	if persistErr != nil {
		return persistErr
	}
	//写盘成功再删除unconfirm的内存镜像
	for txid := range unconfirmToConfirm {
		uv.unconfirmTxInMem.Delete(txid)
	}
	for txid := range undoDone {
		uv.unconfirmTxInMem.Delete(txid)
	}
	return nil
}

// PlayForMiner 进行合约预执行
func (uv *UtxoVM) PlayForMiner(blockid []byte, batch kvdb.Batch) error {
	block, blockErr := uv.ledger.QueryBlock(blockid)
	if blockErr != nil {
		return blockErr
	}
	if !bytes.Equal(block.PreHash, uv.latestBlockid) {
		uv.xlog.Warn("play for miner failed", "block.PreHash", fmt.Sprintf("%x", block.PreHash),
			"latestBlockid", fmt.Sprintf("%x", uv.latestBlockid))
		return ErrPreBlockMissMatch
	}
	uv.mutex.Lock()
	defer uv.mutex.Unlock() // lock guard
	var err error
	defer func() {
		if err != nil {
			uv.clearBalanceCache()
		}
	}()
	for _, tx := range block.Transactions {
		txid := string(tx.Txid)
		if tx.Coinbase {
			err = uv.doTxInternal(tx, batch)
			if err != nil {
				uv.xlog.Warn("dotx failed when PlayForMiner", "txid", fmt.Sprintf("%x", tx.Txid), "err", err)
				return err
			}
		} else {
			if !uv.asyncMode {
				batch.Delete(append([]byte(pb.UnconfirmedTablePrefix), []byte(txid)...))
			}
		}
		err = uv.payFee(tx, batch, block)
		if err != nil {
			uv.xlog.Warn("payFee failed", "feeErr", err)
			return err
		}
	}
	//继续PrePlayForMiner的合约上下文
	if err = uv.smartContract.Finalize(block.Blockid); err != nil {
		uv.xlog.Warn("smart contract.finalize failed", "blockid", fmt.Sprintf("%x", block.Blockid))
		return err
	}
	//更新latestBlockid
	err = uv.updateLatestBlockid(block.Blockid, batch, "failed to save block")
	if err != nil {
		return err
	}
	//写盘成功再清理unconfirm内存镜像
	for _, tx := range block.Transactions {
		uv.unconfirmTxInMem.Delete(string(tx.Txid))
	}
	return nil
}

// RollBackUnconfirmedTx 回滚本地未确认交易
func (uv *UtxoVM) RollBackUnconfirmedTx() (map[string]bool, error) {
	batch := uv.ldb.NewBatch()
	unconfirmTxMap, unconfirmTxGraph, loadErr := uv.sortUnconfirmedTx()
	if loadErr != nil {
		return nil, loadErr
	}
	undoDone := map[string]bool{}
	for txid, unconfirmTx := range unconfirmTxMap {
		undoErr := uv.undoUnconfirmedTx(unconfirmTx, unconfirmTxMap, unconfirmTxGraph, batch, undoDone)
		if undoErr != nil {
			uv.xlog.Warn("fail to undo tx", "undoErr", undoErr, "txid", fmt.Sprintf("%x", txid))
			return nil, undoErr
		}
	}
	writeErr := batch.Write()
	if writeErr != nil {
		uv.ClearCache()
		uv.xlog.Warn("failed to clean unconfirmed tx", "writeErr", writeErr)
		return nil, writeErr
	}
	for txid := range undoDone {
		uv.unconfirmTxInMem.Delete(txid)
	}
	return undoDone, nil
}

// Walk 从当前的latestBlockid 游走到 blockid, 会触发utxo状态的回滚。
//  执行后会更新latestBlockid
func (uv *UtxoVM) Walk(blockid []byte) error {
	uv.mutex.Lock()
	defer uv.mutex.Unlock() // lock guard
	// 首先先把所有的unconfirm回滚了。
	undoDone, err := uv.RollBackUnconfirmedTx()
	if err != nil {
		return err
	}
	uv.clearBalanceCache()
	// 然后开始寻找blockid 和 latestBlockid的最低公共祖先, 生成undoBlocks和todoBlocks
	undoBlocks, todoBlocks, findErr := uv.ledger.FindUndoAndTodoBlocks(uv.latestBlockid, blockid)
	if findErr != nil {
		uv.xlog.Warn("fail to to find common parent of two blocks", "dest_block", fmt.Sprintf("%x", blockid),
			"latestBlockid", fmt.Sprintf("%x", uv.latestBlockid), "findErr", findErr)
		return findErr
	}
	for _, undoBlk := range undoBlocks {
		batch := uv.ldb.NewBatch()
		uv.xlog.Info("start undo block", "blockid", fmt.Sprintf("%x", undoBlk.Blockid))
		ctx := &contract.TxContext{UtxoBatch: batch, Block: undoBlk, IsUndo: true, LedgerObj: uv.ledger} // 将batch赋值到合约机的上下文
		uv.smartContract.SetContext(ctx)
		for i := len(undoBlk.Transactions) - 1; i >= 0; i-- {
			tx := undoBlk.Transactions[i]
			if !undoDone[string(tx.Txid)] { //避免重复回滚
				err := uv.undoTxInternal(tx, batch)
				if err != nil {
					uv.xlog.Warn("failed to undo block", "err", err)
					return err
				}
			}
			feeErr := uv.undoPayFee(tx, batch, undoBlk)
			if feeErr != nil {
				uv.xlog.Warn("undoPayFee failed", "feeErr", feeErr)
				return feeErr
			}
			err := uv.RollbackContract(undoBlk.Blockid, tx)
			if err != nil {
				uv.xlog.Warn("failed to rollback contract, when undo block", "err", err)
			}
		}
		if err := uv.smartContract.Finalize(undoBlk.PreHash); err != nil {
			uv.xlog.Error("smart contract fianlize failed", "blockid", fmt.Sprintf("%x", undoBlk.Blockid))
			return err
		}
		updateErr := uv.updateLatestBlockid(undoBlk.PreHash, batch, "error occurs when undo blocks")
		if updateErr != nil {
			return updateErr
		}
	}
	for i := len(todoBlocks) - 1; i >= 0; i-- {
		todoBlk := todoBlocks[i]
		// 区块加解密有效性检查
		batch := uv.ldb.NewBatch()
		ctx := &contract.TxContext{UtxoBatch: batch, Block: todoBlk, LedgerObj: uv.ledger} // 将batch赋值到合约机的上下文
		uv.smartContract.SetContext(ctx)
		uv.xlog.Info("start do block", "blockid", fmt.Sprintf("%x", todoBlk.Blockid))
		autoGenTxList, genErr := uv.GetVATList(todoBlk.Height, -1, todoBlk.Timestamp)
		if genErr != nil {
			uv.xlog.Warn("get autogen tx list failed", "err", genErr)
			return genErr
		}
		idx, length := 0, len(todoBlk.Transactions)
		for idx < length {
			tx := todoBlk.Transactions[idx]
			txErr := uv.doTxInternal(tx, batch)
			if txErr != nil {
				uv.xlog.Warn("failed to do tx when Walk", "txErr", txErr, "txid", fmt.Sprintf("%x", tx.Txid))
				return txErr
			}
			feeErr := uv.payFee(tx, batch, todoBlk)
			if feeErr != nil {
				uv.xlog.Warn("payFee failed", "feeErr", feeErr)
				return feeErr
			}
			var cErr error
			if idx, cErr = uv.TxOfRunningContractVerify(batch, todoBlk, tx, &autoGenTxList, idx); cErr != nil {
				uv.xlog.Warn("TxOfRunningContractVerify failed when walking", "error", cErr, "idx", idx)
				return cErr
			}
		}
		uv.xlog.Debug("Begin to Finalize", "blockid", fmt.Sprintf("%x", todoBlk.Blockid))
		if err := uv.smartContract.Finalize(todoBlk.Blockid); err != nil {
			uv.xlog.Error("smart contract fianlize failed", "blockid", fmt.Sprintf("%x", todoBlk.Blockid))
			return err
		}
		updateErr := uv.updateLatestBlockid(todoBlk.Blockid, batch, "error occurs when do blocks") // 每do一个block,是一个原子batch写
		if updateErr != nil {
			return updateErr
		}
	}
	return nil
}

// GetLatestBlockid 返回当前vm最后一次执行到的blockid
func (uv *UtxoVM) GetLatestBlockid() []byte {
	return uv.latestBlockid
}

// HasTx 查询一笔交易是否在unconfirm表
func (uv *UtxoVM) HasTx(txid []byte) (bool, error) {
	_, exist := uv.unconfirmTxInMem.Load(string(txid))
	return exist, nil
}

// QueryTx 查询一笔交易，从unconfirm表中查询
func (uv *UtxoVM) QueryTx(txid []byte) (*pb.Transaction, error) {
	pbBuf, findErr := uv.unconfirmedTable.Get(txid)
	if findErr != nil {
		if findErr.Error() == kverr.ErrNotFound.Error() {
			return nil, ErrTxNotFound
		}
		uv.xlog.Warn("unexpected leveldb error, when do QueryTx, it may corrupted.", "findErr", findErr)
		return nil, findErr
	}
	tx := &pb.Transaction{}
	pbErr := proto.Unmarshal(pbBuf, tx)
	if pbErr != nil {
		uv.xlog.Warn("failed to unmarshal tx", "pbErr", pbErr)
		return nil, pbErr
	}
	return tx, nil
}

func (uv *UtxoVM) queryAccountACLWithConfirmed(accountName string) (*pb.Acl, bool, error) {
	if uv.aclMgr == nil {
		return nil, false, errors.New("acl manager is nil")
	}
	return uv.aclMgr.GetAccountACLWithConfirmed(accountName)
}

func (uv *UtxoVM) queryContractMethodACLWithConfirmed(contractName string, methodName string) (*pb.Acl, bool, error) {
	if uv.aclMgr == nil {
		return nil, false, errors.New("acl manager is nil")
	}
	return uv.aclMgr.GetContractMethodACLWithConfirmed(contractName, methodName)
}

func (uv *UtxoVM) queryAccountACL(accountName string) (*pb.Acl, error) {
	if uv.aclMgr == nil {
		return nil, errors.New("acl manager is nil")
	}
	return uv.aclMgr.GetAccountACL(accountName)
}

func (uv *UtxoVM) queryContractMethodACL(contractName string, methodName string) (*pb.Acl, error) {
	if uv.aclMgr == nil {
		return nil, errors.New("acl manager is nil")
	}
	return uv.aclMgr.GetContractMethodACL(contractName, methodName)
}

//获得一个账号的余额，inLock表示在调用此函数时已经对uv.mutex加过锁了
func (uv *UtxoVM) getBalance(addr string, inLock bool) (*big.Int, error) {
	cachedBalance, ok := uv.balanceCache.Get(addr)
	if ok {
		uv.xlog.Debug("hit getbalance cache", "addr", addr)
		if !inLock {
			uv.mutex.Lock()
		}
		balanceCopy := big.NewInt(0).Set(cachedBalance.(*big.Int))
		if !inLock {
			uv.mutex.Unlock()
		}
		return balanceCopy, nil
	}
	addrPrefix := fmt.Sprintf("%s%s_", pb.UTXOTablePrefix, addr)
	utxoTotal := big.NewInt(0)
	it := uv.ldb.NewIteratorWithPrefix([]byte(addrPrefix))
	defer it.Release()
	for it.Next() {
		uBinary := it.Value()
		uItem := &UtxoItem{}
		uErr := uItem.Loads(uBinary)
		if uErr != nil {
			return nil, uErr
		}
		utxoTotal.Add(utxoTotal, uItem.Amount) // utxo累加
	}
	if it.Error() != nil {
		return nil, it.Error()
	}
	if !inLock {
		uv.mutex.Lock()
		defer uv.mutex.Unlock()
	}
	if uv.balanceViewDirty[addr] {
		delete(uv.balanceViewDirty, addr)
		return utxoTotal, nil
	}
	_, exist := uv.balanceCache.Get(addr)
	if !exist {
		//填充cache
		uv.balanceCache.Add(addr, utxoTotal)
	}
	balanceCopy := big.NewInt(0).Set(utxoTotal)
	return balanceCopy, nil
}

// QueryAccountACLWithConfirmed query account's ACL with confirm status
func (uv *UtxoVM) QueryAccountACLWithConfirmed(accountName string) (*pb.Acl, bool, error) {
	return uv.queryAccountACLWithConfirmed(accountName)
}

// QueryContractMethodACLWithConfirmed query contract method's ACL with confirm status
func (uv *UtxoVM) QueryContractMethodACLWithConfirmed(contractName string, methodName string) (*pb.Acl, bool, error) {
	return uv.queryContractMethodACLWithConfirmed(contractName, methodName)
}

// QueryAccountACL query account's ACL
func (uv *UtxoVM) QueryAccountACL(accountName string) (*pb.Acl, error) {
	return uv.queryAccountACL(accountName)
}

// QueryContractMethodACL query contract method's ACL
func (uv *UtxoVM) QueryContractMethodACL(contractName string, methodName string) (*pb.Acl, error) {
	return uv.queryContractMethodACL(contractName, methodName)
}

// GetBalance 查询Address的可用余额
func (uv *UtxoVM) GetBalance(addr string) (*big.Int, error) {
	return uv.getBalance(addr, false)
}

// GetFrozenBalance 查询Address的被冻结的余额
func (uv *UtxoVM) GetFrozenBalance(addr string) (*big.Int, error) {
	addrPrefix := fmt.Sprintf("%s%s_", pb.UTXOTablePrefix, addr)
	utxoFrozen := big.NewInt(0)
	curHeight := uv.ledger.GetMeta().TrunkHeight
	it := uv.ldb.NewIteratorWithPrefix([]byte(addrPrefix))
	defer it.Release()
	for it.Next() {
		uBinary := it.Value()
		uItem := &UtxoItem{}
		uErr := uItem.Loads(uBinary)
		if uErr != nil {
			return nil, uErr
		}
		if uItem.FrozenHeight <= curHeight && uItem.FrozenHeight != -1 {
			continue
		}
		utxoFrozen.Add(utxoFrozen, uItem.Amount) // utxo累加
	}
	if it.Error() != nil {
		return nil, it.Error()
	}
	return utxoFrozen, nil
}

// Close 关闭utxo vm, 目前主要是关闭leveldb
func (uv *UtxoVM) Close() {
	uv.smartContract.Stop()
	if uv.asyncMode && uv.asyncCancel != nil {
		uv.asyncCancel()
		uv.asyncWriterWG.Wait()
		return
	}
	uv.ldb.Close()
}

// GetMeta get the utxo metadata of the blockchain
func (uv *UtxoVM) GetMeta() *pb.UtxoMeta {
	meta := &pb.UtxoMeta{}
	meta.LatestBlockid = uv.latestBlockid
	meta.UtxoTotal = uv.utxoTotal.String() // pb没有bigint，所以转换为字符串
	return meta
}

// GetTotal 返回当前vm的总资产
func (uv *UtxoVM) GetTotal() *big.Int {
	result := big.NewInt(0)
	result.SetBytes(uv.utxoTotal.Bytes())
	return result
}

// ScanWithPrefix 通过前缀获得一个连续读取的迭代器
func (uv *UtxoVM) ScanWithPrefix(prefix []byte) kvdb.Iterator {
	return uv.ldb.NewIteratorWithPrefix(prefix)
}

// GetFromTable 从一个表读取一个key的value
func (uv *UtxoVM) GetFromTable(tablePrefix []byte, key []byte) ([]byte, error) {
	realKey := append([]byte(tablePrefix), key...)
	return uv.ldb.Get(realKey)
}

// RemoveUtxoCache 清理utxoCache
func (uv *UtxoVM) RemoveUtxoCache(address string, utxoKey string) {
	uv.xlog.Trace("RemoveUtxoCache", "address", address, "utxoKey", utxoKey)
	uv.utxoCache.Remove(address, utxoKey)
}

// GetVATList return the registered VAT list
func (uv *UtxoVM) GetVATList(blockHeight int64, maxCount int, timestamp int64) ([]*pb.Transaction, error) {
	txs := []*pb.Transaction{}
	for i := 0; i < len(uv.vatHandler.HandlerList); i++ {
		name := uv.vatHandler.HandlerList[i]
		vats, err := uv.vatHandler.Handlers[name].GetVerifiableAutogenTx(blockHeight, maxCount, timestamp)
		if err != nil {
			uv.xlog.Warn("GetVATList error", "err", err)
			continue
		}
		if vats != nil {
			txs = append(txs, vats...)
		}
	}
	return txs, nil
}

// MustVAT must VAT
func (uv *UtxoVM) MustVAT(desc *contract.TxDesc) bool {
	if desc.Module == "" {
		return false //不是合约,跳过
	}
	return uv.vatHandler.MustVAT(desc.Module, desc.Method)
}

// NewBatch return batch instance
func (uv *UtxoVM) NewBatch() kvdb.Batch {
	return uv.ldb.NewBatch()
}

// GetXModel return the instance of XModel
func (uv *UtxoVM) GetXModel() *xmodel.XModel {
	return uv.model3
}

// GetACLManager return ACLManager instance
func (uv *UtxoVM) GetACLManager() *acli.Manager {
	return uv.aclMgr
}

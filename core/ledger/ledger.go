package ledger

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/golang/protobuf/proto"
	log "github.com/xuperchain/log15"
	"github.com/xuperchain/xuperchain/core/common"
	crypto_client "github.com/xuperchain/xuperchain/core/crypto/client"
	crypto_base "github.com/xuperchain/xuperchain/core/crypto/client/base"
	"github.com/xuperchain/xuperchain/core/global"
	"github.com/xuperchain/xuperchain/core/kv/kvdb"
	"github.com/xuperchain/xuperchain/core/pb"
)

var (
	// ErrBlockNotExist is returned when a block to query not exist in specific chain
	ErrBlockNotExist = errors.New("block not exist in this chain")
	// ErrTxNotFound is returned when a transaction to query not exist in confirmed table
	ErrTxNotFound = errors.New("transaction not found")
	// ErrTxDuplicated ...
	ErrTxDuplicated = errors.New("transaction duplicated in different blocks")
	// ErrRootBlockAlreadyExist is returned when two genesis block is checked in the process of confirming block
	ErrRootBlockAlreadyExist = errors.New("this ledger already has genesis block")
	// ErrMinerInterrupt is returned when IsEnablePowMinning is false
	ErrMinerInterrupt = errors.New("new block interrupts the process of the miner")
	// ErrTxNotConfirmed return tx not confirmed error
	ErrTxNotConfirmed = errors.New("transaction not confirmed")
	// NumCPU returns the number of CPU cores for the current system
	NumCPU = runtime.NumCPU()
)

var (
	// MemCacheSize baseDB memory level max size
	MemCacheSize = 128 //MB
	// FileHandlersCacheSize baseDB memory file handler cache max size
	FileHandlersCacheSize = 1024 //how many opened files-handlers cached
	// DisableTxDedup ...
	DisableTxDedup = false //whether disable dedup tx before confirm
)

const (
	// RootBlockVersion for version 1
	RootBlockVersion = 0
	// BlockVersion for version 1
	BlockVersion = 1
	// BlockCacheSize block counts in lru cache
	BlockCacheSize              = 100 //block counts in lru cache
	MaxBlockSizeKey             = "MaxBlockSize"
	ReservedContractsKey        = "ReservedContracts"
	ForbiddenContractKey        = "ForbiddenContract"
	NewAccountResourceAmountKey = "NewAccountResourceAmount"
	// Irreversible block height & slide window
	IrreversibleBlockHeightKey = "IrreversibleBlockHeight"
	IrreversibleSlideWindowKey = "IrreversibleSlideWindow"
	GasPriceKey                = "GasPrice"
	GroupChainContractKey      = "GroupChainContract"
)

// Ledger define data structure of Ledger
type Ledger struct {
	baseDB           kvdb.Database // 底层是一个leveldb实例，kvdb进行了包装
	metaTable        kvdb.Database // 记录区块链的根节点、高度、末端节点
	confirmedTable   kvdb.Database // 已确认的订单表
	blocksTable      kvdb.Database // 区块表
	mutex            *sync.RWMutex
	xlog             log.Logger       //日志库
	meta             *pb.LedgerMeta   //账本关键的元数据{genesis, tip, height}
	GenesisBlock     *GenesisBlock    //创始块
	pendingTable     kvdb.Database    //保存临时的block区块
	heightTable      kvdb.Database    //保存高度到Blockid的映射
	blockCache       *common.LRUCache // block cache, 加速QueryBlock
	blkHeaderCache   *common.LRUCache // block header cache, 加速fetchBlock
	cryptoClient     crypto_base.CryptoClient
	enablePowMinning bool
	powMutex         *sync.Mutex
	confirmBatch     kvdb.Batch //新增区块
}

// ConfirmStatus block status
type ConfirmStatus struct {
	Succ        bool  // 区块是否提交成功
	Split       bool  // 提交后是否发生了分叉
	Orphan      bool  // 是否是个孤儿节点
	TrunkSwitch bool  // 是否导致了主干分支切换
	Error       error //错误消息
}

// NewLedger create an empty ledger, if it already exists, open it directly
func NewLedger(storePath string, xlog log.Logger, otherPaths []string, kvEngineType string, cryptoType string) (*Ledger, error) {
	return newLedger(storePath, xlog, otherPaths, kvEngineType, cryptoType, true)
}

// OpenLedger open ledger which already exists
func OpenLedger(storePath string, xlog log.Logger, otherPaths []string, kvEngineType string, cryptoType string) (*Ledger, error) {
	return newLedger(storePath, xlog, otherPaths, kvEngineType, cryptoType, false)
}

func newLedger(storePath string, xlog log.Logger, otherPaths []string, kvEngineType string, cryptoType string, createIfMissing bool) (*Ledger, error) {
	ledger := &Ledger{}
	ledger.mutex = &sync.RWMutex{}
	ledger.powMutex = &sync.Mutex{}
	if xlog == nil { //如果外面没传进来log对象的话
		xlog = log.New("module", "ledger")
		xlog.SetHandler(log.StreamHandler(os.Stderr, log.LogfmtFormat()))
	}

	// new kvdb instance
	kvParam := &kvdb.KVParameter{
		DBPath:                filepath.Join(storePath, "ledger"),
		KVEngineType:          kvEngineType,
		MemCacheSize:          MemCacheSize,
		FileHandlersCacheSize: FileHandlersCacheSize,
		OtherPaths:            otherPaths,
	}
	baseDB, err := kvdb.NewKVDBInstance(kvParam)
	if err != nil {
		xlog.Warn("fail to open leveldb", "dbPath", storePath+"/ledger", "err", err)
		return nil, err
	}

	// create crypto client
	cryptoClient, cryptoErr := crypto_client.CreateCryptoClient(cryptoType)
	if cryptoErr != nil {
		xlog.Warn("fail to create crypto client", "err", cryptoErr)
		return nil, cryptoErr
	}

	ledger.baseDB = baseDB
	ledger.metaTable = kvdb.NewTable(baseDB, pb.MetaTablePrefix)
	ledger.confirmedTable = kvdb.NewTable(baseDB, pb.ConfirmedTablePrefix)
	ledger.blocksTable = kvdb.NewTable(baseDB, pb.BlocksTablePrefix)
	ledger.pendingTable = kvdb.NewTable(baseDB, pb.PendingBlocksTablePrefix)
	ledger.heightTable = kvdb.NewTable(baseDB, pb.BlockHeightPrefix)
	ledger.xlog = xlog
	ledger.meta = &pb.LedgerMeta{}
	ledger.blockCache = common.NewLRUCache(BlockCacheSize)
	ledger.blkHeaderCache = common.NewLRUCache(BlockCacheSize)
	ledger.cryptoClient = cryptoClient
	ledger.enablePowMinning = true
	ledger.confirmBatch = baseDB.NewBatch()
	metaBuf, metaErr := ledger.metaTable.Get([]byte(""))
	emptyLedger := false
	if metaErr != nil && common.NormalizedKVError(metaErr) == common.ErrKVNotFound && createIfMissing { //说明是新创建的账本
		metaBuf, pbErr := proto.Marshal(ledger.meta)
		if pbErr != nil {
			xlog.Warn("marshal meta fail", "pb_err", pbErr)
			return nil, pbErr
		}
		writeErr := ledger.metaTable.Put([]byte(""), metaBuf)
		if writeErr != nil {
			xlog.Warn("write meta_table fail", "write_err", writeErr)
			return nil, writeErr
		}
		emptyLedger = true
	} else {
		if metaErr != nil {
			xlog.Warn("unexpected kv error", "meta_err", metaErr)
			return nil, metaErr
		}
		pbErr := proto.Unmarshal(metaBuf, ledger.meta)
		if pbErr != nil {
			return nil, pbErr
		}
	}
	xlog.Info("ledger meta", "genesis_block", fmt.Sprintf("%x", ledger.meta.RootBlockid), "tip_block",
		fmt.Sprintf("%x", ledger.meta.TipBlockid), "trunk_height", ledger.meta.TrunkHeight)
	if !emptyLedger {
		gErr := ledger.loadGenesisBlock()
		if gErr != nil {
			xlog.Warn("failed to load genesis block", "g_err", gErr)
			return nil, gErr
		}
	}
	return ledger, nil
}

// Close close an instance of ledger
func (l *Ledger) Close() {
	l.baseDB.Close()
}

// GetMeta returns meta info of Ledger, such as genesis block ID, current block height, tip block ID
func (l *Ledger) GetMeta() *pb.LedgerMeta {
	return l.meta
}

// GetLDB returns the instance of underlying of kv db
func (l *Ledger) GetLDB() kvdb.Database {
	return l.baseDB
}

func (l *Ledger) loadGenesisBlock() error {
	if len(l.meta.RootBlockid) == 0 {
		return ErrBlockNotExist
	}
	rootIb, err := l.queryBlock(l.meta.RootBlockid, true)
	if err != nil {
		return err
	}
	gb, gErr := NewGenesisBlock(rootIb)
	if gErr != nil {
		return gErr
	}
	l.GenesisBlock = gb
	return nil
}

// FormatRootBlock format genesis block
func (l *Ledger) FormatRootBlock(txList []*pb.Transaction) (*pb.InternalBlock, error) {
	l.xlog.Info("begin format genesis block")
	block := &pb.InternalBlock{Version: RootBlockVersion}
	block.Transactions = txList
	block.TxCount = int32(len(txList))
	block.MerkleTree = MakeMerkleTree(txList)
	if len(block.MerkleTree) > 0 {
		block.MerkleRoot = block.MerkleTree[len(block.MerkleTree)-1]
	}
	var err error
	block.Blockid, err = MakeBlockID(block)
	if err != nil {
		return nil, err
	}
	return block, nil
}

// FormatBlock format normal block
func (l *Ledger) FormatBlock(txList []*pb.Transaction,
	proposer []byte, ecdsaPk *ecdsa.PrivateKey, /*矿工的公钥私钥*/
	timestamp int64, curTerm int64, curBlockNum int64,
	preHash []byte, utxoTotal *big.Int) (*pb.InternalBlock, error) {
	return l.formatBlock(txList, proposer, ecdsaPk, timestamp, curTerm, curBlockNum, preHash, 0, utxoTotal, true, nil, nil, 0)
}

// FormatMinerBlock format block for miner
func (l *Ledger) FormatMinerBlock(txList []*pb.Transaction,
	proposer []byte, ecdsaPk *ecdsa.PrivateKey, /*矿工的公钥私钥*/
	timestamp int64, curTerm int64, curBlockNum int64,
	preHash []byte, targetBits int32, utxoTotal *big.Int,
	qc *pb.QuorumCert, failedTxs map[string]string, blockHeight int64) (*pb.InternalBlock, error) {
	return l.formatBlock(txList, proposer, ecdsaPk, timestamp, curTerm, curBlockNum, preHash, targetBits, utxoTotal, true, qc, failedTxs, blockHeight)
}

// IsProofed check workload proof
func IsProofed(blockID []byte, targetBits int32) bool {
	given := big.NewInt(0).SetBytes(blockID)
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))
	//l.xlog.Debug("pow guess", "target", global.F(target.Bytes()), "given", global.F(given.Bytes()))
	if given.Cmp(target) == -1 {
		return true
	}
	return false
}

// FormatFakeBlock format fake block for contract pre-execution without signing
func (l *Ledger) FormatFakeBlock(txList []*pb.Transaction,
	proposer []byte, ecdsaPk *ecdsa.PrivateKey, /*矿工的公钥私钥*/
	timestamp int64, curTerm int64, curBlockNum int64,
	preHash []byte, utxoTotal *big.Int, blockHeight int64) (*pb.InternalBlock, error) {
	return l.formatBlock(txList, proposer, ecdsaPk, timestamp, curTerm, curBlockNum, preHash, 0, utxoTotal, false, nil, nil, blockHeight)
}

/*
内存中格式化一个区块
*/
func (l *Ledger) formatBlock(txList []*pb.Transaction,
	proposer []byte, ecdsaPk *ecdsa.PrivateKey, /*矿工的公钥私钥*/
	timestamp int64, curTerm int64, curBlockNum int64,
	preHash []byte, targetBits int32, utxoTotal *big.Int, needSign bool,
	qc *pb.QuorumCert, failedTxs map[string]string, blockHeight int64) (*pb.InternalBlock, error) {
	l.xlog.Info("begin format block", "preHash", fmt.Sprintf("%x", preHash))
	//编译的环境变量指定
	block := &pb.InternalBlock{Version: BlockVersion}
	block.Transactions = txList
	block.TxCount = int32(len(txList))
	block.Timestamp = timestamp
	block.Proposer = proposer
	block.CurTerm = curTerm
	block.CurBlockNum = curBlockNum
	block.TargetBits = targetBits
	block.Justify = qc
	block.Height = blockHeight
	jsPk, pkErr := l.cryptoClient.GetEcdsaPublicKeyJSONFormat(ecdsaPk)
	if pkErr != nil {
		return nil, pkErr
	}
	block.Pubkey = []byte(jsPk)
	block.PreHash = preHash
	if !needSign {
		fakeTree := make([][]byte, len(txList))
		for i, tx := range txList {
			fakeTree[i] = tx.Txid
		}
		block.MerkleTree = fakeTree
	} else {
		block.MerkleTree = MakeMerkleTree(txList)
	}
	if failedTxs != nil {
		block.FailedTxs = failedTxs
	} else {
		block.FailedTxs = map[string]string{}
	}
	if len(block.MerkleTree) > 0 {
		block.MerkleRoot = block.MerkleTree[len(block.MerkleTree)-1]
	}
	var err error
	block.Blockid, err = MakeBlockID(block)
	if err != nil {
		return nil, err
	}

	if targetBits != 0 {
		block, err = l.processFormatBlockForPOW(block, targetBits)
		if err != nil {
			return nil, err
		}
	}

	if len(preHash) > 0 && needSign {
		block.Sign, err = l.cryptoClient.SignECDSA(ecdsaPk, block.Blockid)
	}
	if err != nil {
		return nil, err
	}
	return block, nil
}

// CheckBlock check if a block is valid and support passing in a check function
func (l *Ledger) CheckBlock(blockid []byte, validator func(*pb.InternalBlock) bool) error {
	return nil
}

//保存一个区块（只包括区块头）
// 注：只是打包到一个leveldb batch write对象中
func (l *Ledger) saveBlock(block *pb.InternalBlock, batchWrite kvdb.Batch) error {
	blockBuf, pbErr := proto.Marshal(block)
	l.blkHeaderCache.Add(string(block.Blockid), block)
	if pbErr != nil {
		l.xlog.Warn("marshal block fail", "pbErr", pbErr)
		return pbErr
	}
	err := batchWrite.Put(append([]byte(pb.BlocksTablePrefix), block.Blockid...), blockBuf)
	if err != nil {
		return err
	}
	if block.InTrunk {
		sHeight := []byte(fmt.Sprintf("%020d", block.Height))
		batchWrite.Put(append([]byte(pb.BlockHeightPrefix), sHeight...), block.Blockid)
	}
	return nil
}

//根据blockid获取一个Block, 只包含区块头
func (l *Ledger) fetchBlock(blockid []byte) (*pb.InternalBlock, error) {
	blkInCache, cacheHit := l.blkHeaderCache.Get(string(blockid))
	if cacheHit {
		return blkInCache.(*pb.InternalBlock), nil
	}
	blockBuf, findErr := l.blocksTable.Get(blockid)
	if common.NormalizedKVError(findErr) == common.ErrKVNotFound {
		l.xlog.Warn("block can not be found", "findErr", findErr, "blockid", fmt.Sprintf("%x", blockid))
		return nil, findErr
	} else if findErr != nil {
		l.xlog.Warn("unkonw error", "findErr", findErr)
		return nil, findErr
	}
	block := &pb.InternalBlock{}
	pbErr := proto.Unmarshal(blockBuf, block)
	if pbErr != nil {
		l.xlog.Warn("block may corrupt", "pbErr", pbErr)
		return nil, pbErr
	}
	return block, nil
}

//当发生主干切换后，确保最长路径上的block的tx的blockid指向它
func (l *Ledger) correctTxsBlockid(blockID []byte, batchWrite kvdb.Batch) error {
	block, err := l.queryBlock(blockID, true)
	if err != nil {
		return err
	}
	for _, tx := range block.Transactions {
		if !bytes.Equal(tx.Blockid, blockID) {
			l.xlog.Warn("correct blockid of tx", "txid", fmt.Sprintf("%x", tx.Txid),
				"old_blockid", fmt.Sprintf("%x", tx.Blockid), "new_blockid", fmt.Sprintf("%x", blockID))
			tx.Blockid = blockID
			pbTxBuf, err := proto.Marshal(tx)
			if err != nil {
				l.xlog.Warn("marshal trasaction failed when confirm block", "err", err)
				return err
			}
			batchWrite.Put(append([]byte(pb.ConfirmedTablePrefix), tx.Txid...), pbTxBuf)
		}
	}
	return nil
}

//处理分叉
// P---->P---->P---->P (old tip)
//       |
//       +---->Q---->Q--->NewTip
// 处理完后，会返回分叉点的block
func (l *Ledger) handleFork(oldTip []byte, newTipPre []byte, nextHash []byte, batchWrite kvdb.Batch) (*pb.InternalBlock, error) {
	p := oldTip
	q := newTipPre
	for !bytes.Equal(p, q) {
		pBlock, pErr := l.fetchBlock(p)
		if pErr != nil {
			return nil, pErr
		}
		pBlock.InTrunk = false
		pBlock.NextHash = []byte{} //next_hash表示是主干上的下一个blockid，所以分支上的这个属性清空
		qBlock, qErr := l.fetchBlock(q)
		if qErr != nil {
			return nil, qErr
		}
		qBlock.InTrunk = true
		cerr := l.correctTxsBlockid(qBlock.Blockid, batchWrite)
		if cerr != nil {
			return nil, cerr
		}
		qBlock.NextHash = nextHash
		nextHash = q
		p = pBlock.PreHash
		q = qBlock.PreHash
		saveErr := l.saveBlock(pBlock, batchWrite)
		if saveErr != nil {
			return nil, saveErr
		}
		saveErr = l.saveBlock(qBlock, batchWrite)
		if saveErr != nil {
			return nil, saveErr
		}
	}
	splitBlock, qErr := l.fetchBlock(q)
	if qErr != nil {
		return nil, qErr
	}
	splitBlock.InTrunk = true
	splitBlock.NextHash = nextHash
	saveErr := l.saveBlock(splitBlock, batchWrite)
	if saveErr != nil {
		return nil, saveErr
	}
	return splitBlock, nil
}

// IsValidTx valid transactions of coinbase in block
func (l *Ledger) IsValidTx(idx int, tx *pb.Transaction, block *pb.InternalBlock) bool {
	if tx.Coinbase { //检查系统奖励交易的合法性
		if len(tx.TxOutputs) < 1 {
			l.xlog.Warn("invalid length of coinbase tx outputs, when ConfirmBlock", "len", len(tx.TxOutputs))
			return false
		}
		//交易奖励的金额是否符合策略?
		awardTarget := l.GenesisBlock.CalcAward(block.Height)
		amountBytes := tx.TxOutputs[0].Amount
		awardN := big.NewInt(0)
		awardN.SetBytes(amountBytes)
		if awardN.Cmp(awardTarget) != 0 {
			l.xlog.Warn("invalid block award found", "award", awardN.String(), "target", awardTarget.String())
			return false
		}
	}
	//TODO
	return true
}

// UpdateBlockChainData modify tx which txid is txid
func (l *Ledger) UpdateBlockChainData(txid string, ptxid string, publickey string, sign string, height int64) error {
	if txid == "" || ptxid == "" {
		return fmt.Errorf("invalid update blockchaindata requests")
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.xlog.Info("ledger UpdateBlockChainData", "tx", txid, "ptxid", ptxid)

	rawTxid, err := hex.DecodeString(txid)
	tx, err := l.QueryTransaction(rawTxid)
	if err != nil {
		l.xlog.Warn("ledger UpdateBlockChainData query tx error")
		return fmt.Errorf("ledger UpdateBlockChainData query tx error")
	}

	tx.ModifyBlock = &pb.ModifyBlock{
		Marked:          true,
		EffectiveTxid:   ptxid,
		EffectiveHeight: height,
		PublicKey:       publickey,
		Sign:            sign,
	}
	tx.Desc = []byte("")
	tx.TxOutputsExt = []*pb.TxOutputExt{}

	pbTxBuf, err := proto.Marshal(tx)
	if err != nil {
		l.xlog.Warn("marshal trasaction failed when UpdateBlockChainData", "err", err)
		return err
	}
	l.confirmedTable.Put(tx.Txid, pbTxBuf)

	l.xlog.Info("Update BlockChainData success", "txid", hex.EncodeToString(tx.Txid))
	return nil
}

func (l *Ledger) parallelCheckTx(txs []*pb.Transaction, block *pb.InternalBlock) (map[string]bool, map[string][]byte) {
	parallelLevel := NumCPU
	if len(txs) < parallelLevel {
		parallelLevel = len(txs)
	}
	ch := make(chan *pb.Transaction)
	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	txExist := map[string]bool{}
	txData := map[string][]byte{}
	total := len(txs)
	wg.Add(total)
	for i := 0; i <= parallelLevel; i++ {
		go func() {
			for tx := range ch {
				tx.Blockid = block.Blockid
				pbTxBuf, err := proto.Marshal(tx)
				if err != nil {
					l.xlog.Warn("marshal trasaction failed when confirm block", "err", err)
					mu.Lock()
					txData[string(tx.Txid)] = nil
					mu.Unlock()
				} else {
					mu.Lock()
					txData[string(tx.Txid)] = pbTxBuf
					mu.Unlock()
				}
				if !DisableTxDedup || !block.InTrunk {
					hasTx, _ := l.confirmedTable.Has(tx.Txid)
					mu.Lock()
					txExist[string(tx.Txid)] = hasTx
					mu.Unlock()
				}
				wg.Done()
			}
		}()
	}
	for _, tx := range txs {
		ch <- tx
	}
	wg.Wait()
	close(ch)
	return txExist, txData
}

// ConfirmBlock submit a block to ledger
func (l *Ledger) ConfirmBlock(block *pb.InternalBlock, isRoot bool) ConfirmStatus {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	blkTimer := global.NewXTimer()
	l.xlog.Info("start to confirm block", "blockid", fmt.Sprintf("%x", block.Blockid), "txCount", len(block.Transactions))
	var confirmStatus ConfirmStatus
	dummyTransactions := []*pb.Transaction{}
	realTransactions := block.Transactions // 真正的交易转存到局部变量
	block.Transactions = dummyTransactions // block表不保存transaction详情

	batchWrite := l.confirmBatch
	batchWrite.Reset()
	newMeta := proto.Clone(l.meta).(*pb.LedgerMeta)
	splitHeight := newMeta.TrunkHeight
	if isRoot { //确认创世块
		if block.PreHash != nil && len(block.PreHash) > 0 {
			confirmStatus.Succ = false
			l.xlog.Warn("genesis block shoud has no prehash")
			return confirmStatus
		}
		if len(l.meta.RootBlockid) > 0 {
			confirmStatus.Succ = false
			confirmStatus.Error = ErrRootBlockAlreadyExist
			l.xlog.Warn("already hash genesis block")
			return confirmStatus
		}
		newMeta.RootBlockid = block.Blockid
		newMeta.TrunkHeight = 0 //代表主干上块的最大高度
		newMeta.TipBlockid = block.Blockid
		block.InTrunk = true
		block.Height = 0 // 创世纪块是第0块
	} else { //非创世块,需要判断是在主干还是分支
		preHash := block.PreHash
		preBlock, findErr := l.fetchBlock(preHash)
		if findErr != nil {
			l.xlog.Warn("find pre block fail", "findErr", findErr)
			confirmStatus.Succ = false
			return confirmStatus
		}
		block.Height = preBlock.Height + 1 //不管是主干还是分支，height都是++
		if bytes.Equal(preBlock.Blockid, newMeta.TipBlockid) {
			//在主干上添加
			block.InTrunk = true
			preBlock.NextHash = block.Blockid
			newMeta.TipBlockid = block.Blockid
			newMeta.TrunkHeight++
			//因为改了pre_block的next_hash值，所以也要写回存储
			if !DisableTxDedup {
				saveErr := l.saveBlock(preBlock, batchWrite)
				l.blockCache.Del(string(preBlock.Blockid))
				if saveErr != nil {
					l.xlog.Warn("save block fail", "saveErr", saveErr)
					confirmStatus.Succ = false
					return confirmStatus
				}
			}
		} else {
			//在分支上
			if preBlock.Height+1 > newMeta.TrunkHeight {
				//分支要变成主干了
				oldTip := append([]byte{}, newMeta.TipBlockid...)
				newMeta.TrunkHeight = preBlock.Height + 1
				newMeta.TipBlockid = block.Blockid
				block.InTrunk = true
				splitBlock, splitErr := l.handleFork(oldTip, preBlock.Blockid, block.Blockid, batchWrite) //处理分叉
				if splitErr != nil {
					l.xlog.Warn("handle split failed", "splitErr", splitErr)
					confirmStatus.Succ = false
					return confirmStatus
				}
				splitHeight = splitBlock.Height
				confirmStatus.Split = true
				confirmStatus.TrunkSwitch = true
				l.xlog.Info("handle split successfully", "splitBlock", fmt.Sprintf("%x", splitBlock.Blockid))
			} else {
				// 添加在分支上, 对preblock没有影响
				block.InTrunk = false
				confirmStatus.Split = true
				confirmStatus.TrunkSwitch = false
				confirmStatus.Orphan = true
			}
		}
	}
	saveErr := l.saveBlock(block, batchWrite)
	blkTimer.Mark("saveHeader")
	if saveErr != nil {
		confirmStatus.Succ = false
		l.xlog.Warn("save current block fail", "saveErr", saveErr)
		return confirmStatus
	}
	// update branch head
	updateBranchErr := l.updateBranchInfo(block.Blockid, block.PreHash, block.Height, batchWrite)
	if updateBranchErr != nil {
		confirmStatus.Succ = false
		l.xlog.Warn("update branch info fail", "updateBranchErr", updateBranchErr)
		return confirmStatus
	}
	txExist, txData := l.parallelCheckTx(realTransactions, block)
	cbNum := 0
	oldBlockCache := map[string]*pb.InternalBlock{}
	for _, tx := range realTransactions {
		if tx.Coinbase {
			cbNum = cbNum + 1
		}
		if cbNum > 1 {
			confirmStatus.Succ = false
			l.xlog.Warn("The num of Coinbase tx should not exceed one when confirm block",
				"BlockID", global.F(tx.Blockid), "Miner", string(block.Proposer))
			return confirmStatus
		}

		pbTxBuf := txData[string(tx.Txid)]
		if pbTxBuf == nil {
			confirmStatus.Succ = false
			l.xlog.Warn("marshal trasaction failed when confirm block")
			return confirmStatus
		}
		hasTx := txExist[string(tx.Txid)]
		//l.xlog.Debug("save tx to confirmed table", "txid", fmt.Sprintf("%x", tx.Txid), "hasTx", hasTx)
		if !hasTx {
			batchWrite.Put(append([]byte(pb.ConfirmedTablePrefix), tx.Txid...), pbTxBuf)
		} else {
			//confirm表已经存在这个交易了，需要检查一下是否存在多个主干block包含同样trasnaction的情况
			oldPbTxBuf, _ := l.confirmedTable.Get(tx.Txid)
			oldTx := &pb.Transaction{}
			parserErr := proto.Unmarshal(oldPbTxBuf, oldTx)
			if parserErr != nil {
				confirmStatus.Succ = false
				confirmStatus.Error = parserErr
				return confirmStatus
			}
			oldBlock := &pb.InternalBlock{}
			if cachedBlk, cacheHit := oldBlockCache[string(oldTx.Blockid)]; cacheHit {
				oldBlock = cachedBlk
			} else {
				oldPbBlockBuf, blockErr := l.blocksTable.Get(oldTx.Blockid)
				if blockErr != nil {
					if common.NormalizedKVError(blockErr) == common.ErrKVNotFound {
						l.xlog.Warn("old block that contains the tx has been truncated", "txid", global.F(tx.Txid), "blockid", global.F(oldTx.Blockid))
						batchWrite.Put(append([]byte(pb.ConfirmedTablePrefix), tx.Txid...), pbTxBuf) //overwrite with newtx
						continue
					}
					confirmStatus.Succ = false
					confirmStatus.Error = blockErr
					return confirmStatus
				}
				parserErr = proto.Unmarshal(oldPbBlockBuf, oldBlock)
				if parserErr != nil {
					confirmStatus.Succ = false
					confirmStatus.Error = parserErr
					return confirmStatus
				}
				oldBlockCache[string(oldBlock.Blockid)] = oldBlock
			}
			if oldBlock.InTrunk && block.InTrunk && oldBlock.Height <= splitHeight {
				confirmStatus.Succ = false
				confirmStatus.Error = ErrTxDuplicated
				l.xlog.Warn("transaction duplicated in previous trunk block",
					"txid", fmt.Sprintf("%x", tx.Txid),
					"blockid", fmt.Sprintf("%x", oldBlock.Blockid))
				return confirmStatus
			} else if block.InTrunk {
				l.xlog.Info("change blockid of tx", "txid", fmt.Sprintf("%x", tx.Txid), "blockid", fmt.Sprintf("%x", block.Blockid))
				batchWrite.Put(append([]byte(pb.ConfirmedTablePrefix), tx.Txid...), pbTxBuf)
			}
		}
	}
	blkTimer.Mark("saveAllTxs")
	//删除pendingBlock中对应的数据
	batchWrite.Delete(append([]byte(pb.PendingBlocksTablePrefix), block.Blockid...))
	//改meta
	metaBuf, pbErr := proto.Marshal(newMeta)
	if pbErr != nil {
		l.xlog.Warn("marshal meta fail", "pbErr", pbErr)
		confirmStatus.Succ = false
		return confirmStatus
	}
	batchWrite.Put([]byte(pb.MetaTablePrefix), metaBuf)
	l.xlog.Debug("print block size when confirm block", "blockSize", batchWrite.ValueSize(), "blockid", fmt.Sprintf("%x", block.Blockid))
	kvErr := batchWrite.Write() // blocks, confirmed_transaction两张表原子写入
	blkTimer.Mark("saveToDisk")
	if kvErr != nil {
		confirmStatus.Succ = false
		confirmStatus.Error = kvErr
		l.xlog.Warn("batch write failed when confirm block", "kvErr", kvErr)
	} else {
		confirmStatus.Succ = true
		l.meta = newMeta
	}
	block.Transactions = realTransactions
	if isRoot { //首次confirm 创始块的时候
		lErr := l.loadGenesisBlock()
		if lErr != nil {
			confirmStatus.Succ = false
			confirmStatus.Error = lErr
		}
	}
	l.blockCache.Add(string(block.Blockid), block)
	l.xlog.Debug("confirm block cost", "blkTimer", blkTimer.Print())
	return confirmStatus
}

// ExistBlock check if a block exists in the ledger
func (l *Ledger) ExistBlock(blockid []byte) bool {
	exist, _ := l.blocksTable.Has(blockid)
	return exist
}

func (l *Ledger) queryBlock(blockid []byte, needBody bool) (*pb.InternalBlock, error) {
	pbBlockBuf, err := l.blocksTable.Get(blockid)
	if err != nil {
		if common.NormalizedKVError(err) == common.ErrKVNotFound {
			err = ErrBlockNotExist
		}
		return nil, err
	}
	block := &pb.InternalBlock{}
	parserErr := proto.Unmarshal(pbBlockBuf, block)
	if parserErr != nil {
		return nil, parserErr
	}
	if needBody {
		realTransactions := make([]*pb.Transaction, 0)
		for _, txid := range block.MerkleTree[:block.TxCount] {
			pbTxBuf, kvErr := l.confirmedTable.Get(txid)
			if kvErr != nil {
				l.xlog.Warn("tx not found", "kvErr", kvErr, "txid", fmt.Sprintf("%x", txid))
				return block, kvErr
			}
			realTx := &pb.Transaction{}
			parserErr = proto.Unmarshal(pbTxBuf, realTx)
			if parserErr != nil {
				l.xlog.Warn("tx parser err", "parserErr", parserErr)
				return block, parserErr
			}
			realTransactions = append(realTransactions, realTx)
		}
		block.Transactions = realTransactions
	}
	return block, nil
}

// QueryBlock query a block by blockID in the ledger
func (l *Ledger) QueryBlock(blockid []byte) (*pb.InternalBlock, error) {
	blkInCache, exist := l.blockCache.Get(string(blockid))
	if exist {
		l.xlog.Debug("hit queryblock cache", "blkid", global.F(blockid))
		return blkInCache.(*pb.InternalBlock), nil
	}
	blk, err := l.queryBlock(blockid, true)
	if err != nil {
		return nil, err
	}
	l.blockCache.Add(string(blockid), blk)
	return blk, nil
}

// QueryBlockHeader query a block by blockID in the ledger and return only block header
func (l *Ledger) QueryBlockHeader(blockid []byte) (*pb.InternalBlock, error) {
	return l.queryBlock(blockid, false)
}

// HasTransaction check if a transaction exists in the ledger
func (l *Ledger) HasTransaction(txid []byte) (bool, error) {
	table := l.confirmedTable
	return table.Has(txid)
}

// QueryTransaction query a transaction in the ledger and return it if exist
func (l *Ledger) QueryTransaction(txid []byte) (*pb.Transaction, error) {
	table := l.confirmedTable
	pbTxBuf, kvErr := table.Get(txid)
	if kvErr != nil {
		if common.NormalizedKVError(kvErr) == common.ErrKVNotFound {
			return nil, ErrTxNotFound
		}
		return nil, kvErr
	}
	realTx := &pb.Transaction{}
	parserErr := proto.Unmarshal(pbTxBuf, realTx)
	if parserErr != nil {
		return nil, parserErr
	}
	return realTx, nil
}

// IsTxInTrunk check if a transaction is in trunk by transaction ID
func (l *Ledger) IsTxInTrunk(txid []byte) bool {
	var blk *pb.InternalBlock
	var err error
	table := l.confirmedTable
	pbTxBuf, kvErr := table.Get(txid)
	if kvErr != nil {
		return false
	}
	realTx := &pb.Transaction{}
	pbErr := proto.Unmarshal(pbTxBuf, realTx)
	if pbErr != nil {
		l.xlog.Warn("IsTxInTrunk error", "txid", global.F(txid), "pbErr", pbErr)
		return false
	}
	blkInCache, exist := l.blockCache.Get(string(realTx.Blockid))
	if exist {
		blk = blkInCache.(*pb.InternalBlock)
	} else {
		blk, err = l.queryBlock(realTx.Blockid, false)
		if err != nil {
			l.xlog.Warn("IsTxInTrunk error", "blkid", global.F(realTx.Blockid), "kvErr", err)
			return false
		}
	}
	return blk.InTrunk
}

// FindUndoAndTodoBlocks get blocks required to undo and todo range from curBlockid to destBlockid
func (l *Ledger) FindUndoAndTodoBlocks(curBlockid []byte, destBlockid []byte) ([]*pb.InternalBlock, []*pb.InternalBlock, error) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	undoBlocks := []*pb.InternalBlock{}
	todoBlocks := []*pb.InternalBlock{}
	if bytes.Equal(destBlockid, curBlockid) { //原地踏步的情况...
		return undoBlocks, todoBlocks, nil
	}
	rootBlockid := l.meta.RootBlockid
	oldTip, oErr := l.queryBlock(curBlockid, true)
	if oErr != nil {
		l.xlog.Warn("block not found", "blockid", fmt.Sprintf("%x", curBlockid))
		return nil, nil, oErr
	}
	newTip, nErr := l.queryBlock(destBlockid, true)
	if nErr != nil {
		l.xlog.Warn("block not found", "blockid", fmt.Sprintf("%x", destBlockid))
		return nil, nil, nErr
	}
	visited := map[string]bool{}
	undoBlocks = append(undoBlocks, oldTip)
	todoBlocks = append(todoBlocks, newTip)
	visited[string(oldTip.Blockid)] = true
	visited[string(newTip.Blockid)] = true
	var splitBlockID []byte //最近的分叉点
	for {
		oldPreHash := oldTip.PreHash
		if len(oldPreHash) > 0 && oldTip.Height >= newTip.Height {
			oldTip, oErr = l.queryBlock(oldPreHash, true)
			if oErr != nil {
				return nil, nil, oErr
			}
			if _, exist := visited[string(oldTip.Blockid)]; exist {
				splitBlockID = oldTip.Blockid //从老tip开始回溯到了分叉点
				break
			} else {
				visited[string(oldTip.Blockid)] = true
				undoBlocks = append(undoBlocks, oldTip)
			}
		}
		newPreHash := newTip.PreHash
		if len(newPreHash) > 0 && newTip.Height >= oldTip.Height {
			newTip, nErr = l.queryBlock(newPreHash, true)
			if nErr != nil {
				return nil, nil, nErr
			}
			if _, exist := visited[string(newTip.Blockid)]; exist {
				splitBlockID = newTip.Blockid //从新tip开始回溯到了分叉点
				break
			} else {
				visited[string(newTip.Blockid)] = true
				todoBlocks = append(todoBlocks, newTip)
			}
		}
		if len(oldPreHash) == 0 && len(newPreHash) == 0 {
			splitBlockID = rootBlockid // 这种情况只能从roott算了
			break
		}
	}
	//收尾工作，todo_blocks, undo_blocks 如果最后一个元素是分叉点，需要去掉
	if bytes.Equal(undoBlocks[len(undoBlocks)-1].Blockid, splitBlockID) {
		undoBlocks = undoBlocks[:len(undoBlocks)-1]
	}
	if bytes.Equal(todoBlocks[len(todoBlocks)-1].Blockid, splitBlockID) {
		todoBlocks = todoBlocks[:len(todoBlocks)-1]
	}
	return undoBlocks, todoBlocks, nil
}

// Dump dump ledger structure, block height to blockid
func (l *Ledger) Dump() ([][]string, error) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	it := l.baseDB.NewIteratorWithPrefix([]byte(pb.BlocksTablePrefix))
	defer it.Release()
	blocks := make([][]string, l.meta.TrunkHeight+1)
	for it.Next() {
		block := &pb.InternalBlock{}
		parserErr := proto.Unmarshal(it.Value(), block)
		if parserErr != nil {
			return nil, parserErr
		}
		height := block.Height
		blockid := fmt.Sprintf("{ID:%x,TxCount:%d,InTrunk:%v, Tm:%d, Miner:%s}", block.Blockid, block.TxCount, block.InTrunk, block.Timestamp/1000000000, block.Proposer)
		blocks[height] = append(blocks[height], blockid)
	}
	return blocks, nil
}

// GetGenesisBlock returns genesis block if it exists
func (l *Ledger) GetGenesisBlock() *GenesisBlock {
	if l.GenesisBlock != nil {
		return l.GenesisBlock
	}
	return nil
}

// GetIrreversibleSlideWindow return irreversible slide window
func (l *Ledger) GetIrreversibleSlideWindow() int64 {
	defaultIrreversibleSlideWindow := l.GenesisBlock.GetConfig().GetIrreversibleSlideWindow()
	return defaultIrreversibleSlideWindow
}

// GetMaxBlockSize return max block size
func (l *Ledger) GetMaxBlockSize() int64 {
	defaultBlockSize := l.GenesisBlock.GetConfig().GetMaxBlockSizeInByte()
	return defaultBlockSize
}

// GetNewAccountResourceAmount return the resource amount of new an account
func (l *Ledger) GetNewAccountResourceAmount() int64 {
	defaultNewAccountResourceAmount := l.GenesisBlock.GetConfig().GetNewAccountResourceAmount()
	return defaultNewAccountResourceAmount
}

func (l *Ledger) GetReservedContracts() ([]*pb.InvokeRequest, error) {
	return l.GenesisBlock.GetConfig().GetReservedContract()
}

func (l *Ledger) GetForbiddenContract() ([]*pb.InvokeRequest, error) {
	return l.GenesisBlock.GetConfig().GetForbiddenContract()
}

func (l *Ledger) GetGroupChainContract() ([]*pb.InvokeRequest, error) {
	return l.GenesisBlock.GetConfig().GetGroupChainContract()
}

func (l *Ledger) GetGasPrice() *pb.GasPrice {
	return l.GenesisBlock.GetConfig().GetGasPrice()
}

func (l *Ledger) GetNoFee() bool {
	return l.GenesisBlock.GetConfig().NoFee
}

// SavePendingBlock put block into pending table
func (l *Ledger) SavePendingBlock(block *pb.Block) error {
	l.xlog.Debug("begin save pending block", "blockid", fmt.Sprintf("%x", block.Block.Blockid), "tx_count", len(block.Block.Transactions))
	block.Blockid = block.Block.Blockid
	blockBuf, pbErr := proto.Marshal(block)
	if pbErr != nil {
		l.xlog.Warn("save pending block fail, because marshal block fail", "pbErr", pbErr)
		return pbErr
	}
	saveErr := l.pendingTable.Put(block.Block.Blockid, blockBuf)
	if saveErr != nil {
		l.xlog.Warn("save pending block to ldb fail", "err", saveErr)
		return saveErr
	}
	return nil
}

// GetPendingBlock get block from pending table
func (l *Ledger) GetPendingBlock(blockID []byte) (*pb.Block, error) {
	l.xlog.Debug("get pending block", "bockid", fmt.Sprintf("%x", blockID))
	blockBuf, ldbErr := l.pendingTable.Get(blockID)
	if ldbErr != nil {
		if common.NormalizedKVError(ldbErr) != common.ErrKVNotFound { //其他kv错误
			l.xlog.Warn("get pending block fail", "err", ldbErr, "blockid", fmt.Sprintf("%x", blockID))
		} else { //不存在表里面
			l.xlog.Debug("the block not in pending blocks", "blocid", fmt.Sprintf("%x", blockID))
			return nil, ErrBlockNotExist
		}
		return nil, ldbErr
	}
	block := &pb.Block{}
	unMarshalErr := proto.Unmarshal(blockBuf, block)
	if unMarshalErr != nil {
		l.xlog.Warn("unmarshal block failed", "err", unMarshalErr)
		return nil, unMarshalErr
	}
	return block, nil
}

// GetEstimatedTotal estimate total assets of chain
func (l *Ledger) GetEstimatedTotal() *big.Int {
	total := big.NewInt(0)
	if l.GenesisBlock == nil {
		return total
	}
	total = total.Add(total, l.GenesisBlock.GetInternalBlock().GetCoinbaseTotal())
	var i int64
	for i = 1; i <= l.meta.TrunkHeight; i++ {
		total = total.Add(total, l.GenesisBlock.CalcAward(i))
	}
	return total
}

// QueryBlockByHeight query block by height
func (l *Ledger) QueryBlockByHeight(height int64) (*pb.InternalBlock, error) {
	sHeight := []byte(fmt.Sprintf("%020d", height))
	blockID, kvErr := l.heightTable.Get(sHeight)
	if kvErr != nil {
		if common.NormalizedKVError(kvErr) == common.ErrKVNotFound {
			return nil, ErrBlockNotExist
		}
		return nil, kvErr
	}
	return l.QueryBlock(blockID)
}

// QueryLastBlock query last block by height
func (l *Ledger) QueryLastBlock() (*pb.InternalBlock, error) {
	lastBlockHeight := l.meta.GetTrunkHeight()

	return l.QueryBlockByHeight(lastBlockHeight)
}

// GetBaseDB get internal db instance
func (l *Ledger) GetBaseDB() kvdb.Database {
	return l.baseDB
}

func (l *Ledger) removeBlocks(fromBlockid []byte, toBlockid []byte, batch kvdb.Batch) error {
	fromBlock, findErr := l.fetchBlock(fromBlockid)
	if findErr != nil {
		l.xlog.Warn("failed to find block", "findErr", findErr)
		return findErr
	}
	toBlock, findErr := l.fetchBlock(toBlockid)
	if findErr != nil {
		l.xlog.Warn("failed to find block", "findErr", findErr)
		return findErr
	}
	for fromBlock.Height > toBlock.Height {
		l.xlog.Info("remove block", "blockid", global.F(fromBlock.Blockid), "height", fromBlock.Height)
		l.blkHeaderCache.Del(string(fromBlock.Blockid))
		l.blockCache.Del(string(fromBlock.Blockid))
		batch.Delete(append([]byte(pb.BlocksTablePrefix), fromBlock.Blockid...))
		if fromBlock.InTrunk {
			sHeight := []byte(fmt.Sprintf("%020d", fromBlock.Height))
			batch.Delete(append([]byte(pb.BlockHeightPrefix), sHeight...))
		}
		//iter to prev block
		fromBlock, findErr = l.fetchBlock(fromBlock.PreHash)
		if findErr != nil {
			l.xlog.Warn("failed to find prev block", "findErr", findErr)
			return nil //ignore orphan block
		}
	}
	return nil
}

// Truncate truncate ledger and set tipblock to utxovmLastID
func (l *Ledger) Truncate(utxovmLastID []byte) error {
	batchWrite := l.baseDB.NewBatch()

	l.mutex.Lock()
	defer l.mutex.Unlock()

	newMeta := proto.Clone(l.meta).(*pb.LedgerMeta)
	newMeta.TipBlockid = utxovmLastID
	block, findErr := l.fetchBlock(utxovmLastID)
	if findErr != nil {
		l.xlog.Warn("failed to find utxovm last block", "findErr", findErr)
		return findErr
	}
	branchTips, err := l.GetBranchInfo(block.Blockid, block.Height)
	if err != nil {
		l.xlog.Warn("failed to find all branch tips", "err", err)
		return err
	}
	for _, branchTip := range branchTips {
		deletedBlockid := []byte(branchTip)
		rmErr := l.removeBlocks(deletedBlockid, block.Blockid, batchWrite)
		if rmErr != nil {
			l.xlog.Warn("failed to remove garbage blocks", "from", global.F(l.meta.TipBlockid), "to", global.F(block.Blockid))
			return rmErr
		}
		// update branch head
		updateBranchErr := l.updateBranchInfo(block.Blockid, deletedBlockid, block.Height, batchWrite)
		if updateBranchErr != nil {
			l.xlog.Warn("Truncated failed when calling updateBranchInfo", "updateBranchErr", updateBranchErr)
			return updateBranchErr
		}
	}
	newMeta.TrunkHeight = block.Height
	metaBuf, pbErr := proto.Marshal(newMeta)
	if pbErr != nil {
		l.xlog.Warn("failed to marshal pb meta")
		return pbErr
	}
	batchWrite.Put([]byte(pb.MetaTablePrefix), metaBuf)
	kvErr := batchWrite.Write()
	if kvErr != nil {
		l.xlog.Warn("batch write failed when Truncate", "kvErr", kvErr)
		return kvErr
	}

	l.meta = newMeta
	l.xlog.Info("truncate blockid succeed")
	return nil
}

// StartPowMinning set the value of enablePowMinning true to tell the miner to start minning
func (l *Ledger) StartPowMinning(height int64) {
	l.powMutex.Lock()
	defer l.powMutex.Unlock()
	l.enablePowMinning = true
	powMinningHeight = height
}

// AbortPowMinning set the value of enablePowMinning false to tell the miner to stop minning
func (l *Ledger) AbortPowMinning(height int64) {
	l.powMutex.Lock()
	defer l.powMutex.Unlock()
	if height >= powMinningHeight {
		l.enablePowMinning = false
	}
}

// IsEnablePowMinning get the value of enablePowMinning
func (l *Ledger) IsEnablePowMinning() bool {
	l.powMutex.Lock()
	defer l.powMutex.Unlock()
	return l.enablePowMinning
}

// VerifyBlock verify block
func (l *Ledger) VerifyBlock(block *pb.InternalBlock, logid string) (bool, error) {
	blkid, err := MakeBlockID(block)
	if err != nil {
		l.xlog.Warn("VerifyBlock MakeBlockID error", "logid", logid, "error", err)
		return false, nil
	}
	if !(bytes.Equal(blkid, block.Blockid)) {
		l.xlog.Warn("VerifyBlock equal blockid error", "logid", logid, "redo blockid", global.F(blkid),
			"get blockid", global.F(block.Blockid))
		return false, nil
	}

	errv := VerifyMerkle(block)
	if errv != nil {
		l.xlog.Warn("VerifyMerkle error", "logid", logid, "error", errv)
		return false, nil
	}

	k, err := l.cryptoClient.GetEcdsaPublicKeyFromJSON(block.Pubkey)
	if err != nil {
		l.xlog.Warn("VerifyBlock get ecdsa from block error", "logid", logid, "error", err)
		return false, nil
	}
	chkResult, _ := l.cryptoClient.VerifyAddressUsingPublicKey(string(block.Proposer), k)
	if chkResult == false {
		l.xlog.Warn("VerifyBlock address is not match publickey", "logid", logid)
		return false, nil
	}

	valid, err := l.cryptoClient.VerifyECDSA(k, block.Sign, block.Blockid)
	if err != nil || !valid {
		l.xlog.Warn("VerifyBlock VerifyECDSA error", "logid", logid, "error", err)
		return false, nil
	}
	return true, nil
}

// QueryBlockByTxid query block by txid after it has confirmed
func (l *Ledger) QueryBlockByTxid(txid []byte) (*pb.InternalBlock, error) {
	if exit, _ := l.HasTransaction(txid); !exit {
		return nil, ErrTxNotConfirmed
	}
	tx, err := l.QueryTransaction(txid)
	if err != nil {
		return nil, err
	}
	return l.queryBlock(tx.GetBlockid(), false)
}

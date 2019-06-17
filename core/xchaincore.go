package xchaincore

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	log "github.com/xuperchain/log15"

	"github.com/xuperchain/xuperunion/common"
	"github.com/xuperchain/xuperunion/common/config"
	"github.com/xuperchain/xuperunion/common/probe"
	"github.com/xuperchain/xuperunion/consensus"
	"github.com/xuperchain/xuperunion/consensus/tdpos"
	"github.com/xuperchain/xuperunion/contract/bridge"
	"github.com/xuperchain/xuperunion/contract/kernel"
	"github.com/xuperchain/xuperunion/contract/native"
	"github.com/xuperchain/xuperunion/contract/proposal"
	"github.com/xuperchain/xuperunion/contract/wasm"
	"github.com/xuperchain/xuperunion/crypto/account"
	crypto_client "github.com/xuperchain/xuperunion/crypto/client"
	crypto_base "github.com/xuperchain/xuperunion/crypto/client/base"
	"github.com/xuperchain/xuperunion/global"
	"github.com/xuperchain/xuperunion/ledger"
	"github.com/xuperchain/xuperunion/p2pv2"
	xuper_p2p "github.com/xuperchain/xuperunion/p2pv2/pb"
	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/utxo"
)

var (
	// ErrCannotSyncBlock used to return the error while sync block error
	ErrCannotSyncBlock = errors.New("get block error when sync block")
	// ErrConfirmBlock used to return the error while confirm block error
	ErrConfirmBlock = errors.New("Confirm block error")
	// ErrUTXOVMPlay used to return the error while utxovm play error
	ErrUTXOVMPlay = errors.New("Utxovm play error")
	// ErrWalk used to return the error while Walk error
	ErrWalk = errors.New("Walk error")
	// ErrWalkCheckMinerMatch used to return check miner match while walk error
	ErrWalkCheckMinerMatch = errors.New("Walk error check miner match")
	// ErrNotReady used to return the error while blockchain not ready error
	ErrNotReady = errors.New("BlockChain is not ready")
	// ErrBlockExist used to return the error while block already exit
	ErrBlockExist = errors.New("Block is exist")
	// ErrServiceRefused used to return the error while service refused
	ErrServiceRefused = errors.New("Service refused")
	// ErrInvalidBlock is returned when found an invalid block
	ErrInvalidBlock = errors.New("invalid block")
	// ErrProposeBlockMoreThanConfig is returned when propose block not match config
	ErrProposeBlockMoreThanConfig = errors.New("Error propose block more than config")
	// ErrBlockChainNotExist is returned when process a block for non-existent blockchain
	ErrBlockChainNotExist = errors.New("Error block chain is not exist")
	// ErrBlockChainIsExist is returned when find out blockachin has been loaded
	ErrBlockChainIsExist = errors.New("Error block chain is exist already")
)

const (
	// MaxReposting max repost times for broadcats
	MaxReposting = 300 // tx重试广播的最大并发，过多容易打爆对方的grpc连接数
	// RepostingInterval repost retry interval, ms
	RepostingInterval = 50 // 重试广播间隔ms
)

// XChainCore is the core struct of a chain
type XChainCore struct {
	con          *consensus.PluggableConsensus
	Ledger       *ledger.Ledger
	Utxovm       *utxo.UtxoVM
	P2pv2        p2pv2.P2PServer
	bcname       string
	log          log.Logger
	status       int
	privateKey   *ecdsa.PrivateKey
	publicKey    *ecdsa.PublicKey
	address      []byte
	award        string
	nodeMode     string
	CryptoClient crypto_base.CryptoClient

	mutex *sync.RWMutex

	Speed *probe.SpeedCalc
	// post_cache map[string] bool
	stopFlag      bool
	proposal      *proposal.Proposal
	NativeCodeMgr *native.GeneralSCFramework
	pipelineM     *PipelineMiner
}

// Status return the status of the chain
func (xc *XChainCore) Status() int {
	return xc.status
}

// Init init the chain
func (xc *XChainCore) Init(bcname string, xlog log.Logger, cfg *config.NodeConfig,
	p2p p2pv2.P2PServer, ker *kernel.Kernel, nodeMode string) error {

	// 设置全局随机数发生器的原始种子
	err := global.SetSeed()
	if err != nil {
		return err
	}

	xc.mutex = &sync.RWMutex{}
	xc.Speed = probe.NewSpeedCalc(bcname)
	// this.mutex.Lock()
	// defer this.mutex.Unlock()
	xc.status = global.SafeModel
	xc.bcname = bcname
	xc.log = xlog
	xc.P2pv2 = p2p
	xc.nodeMode = nodeMode
	xc.stopFlag = false
	ledger.MemCacheSize = cfg.DBCache.MemCacheSize
	ledger.FileHandlersCacheSize = cfg.DBCache.FdCacheSize
	datapath := cfg.Datapath + "/" + bcname
	datapathOthers := []string{}
	for _, dpo := range cfg.DatapathOthers {
		datapathOthers = append(datapathOthers, dpo+"/"+bcname)
	}
	utxoCacheSize := cfg.Utxo.CacheSize
	utxoTmplockSeconds := cfg.Utxo.TmpLockSeconds

	// init plugin types
	rootJs, err := ioutil.ReadFile(datapath + "/xuper.json")
	if err != nil {
		xlog.Warn("load xuper.json failed", "err", err)
		return err
	}
	kvEngineType, err := ker.GetKVEngineType(rootJs)
	if err != nil {
		xlog.Warn("parse xuper.json failed", "err", err)
		return err
	}
	cryptoType, err := ker.GetCryptoType(rootJs)
	if err != nil {
		xlog.Warn("cryptoType not found, parse xuper.json failed", "err", err)
		return err
	}

	cryptoClient, cryptoErr := crypto_client.CreateCryptoClient(cryptoType)
	if cryptoErr != nil {
		xlog.Warn("Load crypto client failed", "err", cryptoErr)
		return err
	}
	xc.CryptoClient = cryptoClient

	// 判断xuper.json和创世块参数的一致性，增加可读性
	// 暂时可以不改
	keypath := cfg.Miner.Keypath

	// this.address = utils.GetAddressFromPublicKey(1, this.publicKey)
	addr, pub, pri, err := account.GetAccInfoFromFile(keypath)
	if err != nil {
		xlog.Warn("load address and publickey and privatekey error", "path", keypath+"/address")
		return err
	}
	xc.address = addr
	xlog.Debug("Using address " + string(xc.address))
	xc.privateKey, err = cryptoClient.GetEcdsaPrivateKeyFromJSON(pri)
	if err != nil {
		return err
	}
	xc.publicKey, err = cryptoClient.GetEcdsaPublicKeyFromJSON(pub)
	if err != nil {
		return err
	}

	// write to p2p
	xchainAddrInfo := &p2pv2.XchainAddrInfo{
		Addr:   string(addr),
		Pubkey: pub,
		Prikey: pri,
	}
	xc.P2pv2.SetXchainAddr(xc.bcname, xchainAddrInfo)

	xc.Ledger, err = ledger.NewLedger(datapath, xc.log, datapathOthers, kvEngineType, cryptoType)
	if err != nil {
		xc.log.Warn("NewLedger error", "bc", xc.bcname, "datapath", datapath, "dataPathOhters", datapathOthers)
		return err
	}

	publicKeyStr, err := cryptoClient.GetEcdsaPublicKeyJSONFormat(xc.privateKey)
	if err != nil {
		return err
	}
	privateKeyStr, err := cryptoClient.GetEcdsaPrivateKeyJSONFormat(xc.privateKey)
	if err != nil {
		return err
	}

	xc.Utxovm, err = utxo.MakeUtxoVM(bcname, xc.Ledger, datapath, privateKeyStr, publicKeyStr, xc.address, xc.log,
		utxoCacheSize, utxoTmplockSeconds, cfg.Utxo.ContractExecutionTime, datapathOthers, cfg.Utxo.IsBetaTx[bcname], kvEngineType, cryptoType)

	if err != nil {
		xc.log.Warn("NewUtxoVM error", "bc", xc.bcname, "datapath", datapath, "dataPathOhters", datapathOthers)
		return err
	}
	if cfg.Utxo.AsyncMode {
		xc.Utxovm.StartAsyncWriter()
	}
	xc.Utxovm.SetMaxConfirmedDelay(cfg.Utxo.MaxConfirmedDelay)
	gBlk := xc.Ledger.GetGenesisBlock()
	if gBlk == nil {
		xc.log.Warn("GenesisBlock nil")
		return errors.New("Genesis Block is nil")
	}
	xc.award = gBlk.GetConfig().Award
	gCon, err := gBlk.GetConfig().GetGenesisConsensus()
	if err != nil {
		xc.log.Warn("Get genesis consensus error", "error", err.Error())
		return err
	}
	xc.con, err = consensus.NewPluggableConsensus(xlog, cfg, bcname, xc.Ledger, xc.Utxovm, gCon, cryptoType)
	if err != nil {
		xc.log.Warn("New PluggableConsensus Error")
		return err
	}

	xc.proposal = proposal.NewProposal(xc.log, xc.Ledger, xc.Utxovm)
	// 统一注册所有的合约虚拟机
	xc.Utxovm.RegisterVM("kernel", ker, global.VMPrivRing0)
	xc.Utxovm.RegisterVM("consensus", xc.con, global.VMPrivRing0)
	xc.Utxovm.RegisterVM("proposal", xc.proposal, global.VMPrivRing0)

	nc, err := native.New(&cfg.Native, datapath+"/native", xc.log, datapathOthers, kvEngineType)
	if err != nil {
		xc.log.Error("make native", "error", err)
		return err
	}
	xc.NativeCodeMgr = nc

	xc.Utxovm.RegisterVM("native", nc, global.VMPrivRing0)

	xbridge := bridge.New()
	wasmvm, err := wasm.New(&cfg.Wasm, filepath.Join(datapath, "wasm"), xbridge, xc.Utxovm.GetXModel())
	if err != nil {
		xc.log.Error("initialize WASM error", "error", err)
		return err
	}

	xbridge.RegisterExecutor("native", nc)
	xbridge.RegisterExecutor("wasm", wasmvm)
	xbridge.RegisterToXCore(xc.Utxovm.RegisterVM3)

	// 统一注册xuper3合约虚拟机
	x3kernel, xerr := kernel.NewKernel(wasmvm)
	if xerr != nil {
		return xerr
	}
	xc.Utxovm.RegisterVM3(x3kernel.GetName(), x3kernel)

	// 统一注册VAT
	xc.Utxovm.RegisterVAT("Propose", xc.proposal, nil)
	xc.Utxovm.RegisterVAT("consensus", xc.con, xc.con.GetVATWhiteList())
	xc.Utxovm.RegisterVAT("kernel", ker, ker.GetVATWhiteList())

	xc.pipelineM = NewPipelineMiner(xc)
	go xc.pipelineM.Start()
	go xc.Speed.ShowLoop(xc.log)
	go xc.repostOfflineTx()
	return nil
}

//周期repost本地未上链的交易
func (xc *XChainCore) repostOfflineTx() {
	batchChan := common.NewBatchChan(MaxReposting, RepostingInterval, xc.Utxovm.OfflineTxChan)
	for txList := range batchChan.GetQueue() {
		header := &pb.Header{Logid: global.Glogid()}
		batchTxMsg := &pb.BatchTxs{Header: header}
		//将txList包装为rpc message
		for _, tx := range txList {
			if inUnconfirm, _ := xc.Utxovm.HasTx(tx.Txid); !inUnconfirm {
				continue //跳过已经confirm的
			}
			txStatus := &pb.TxStatus{
				Header: header,
				Bcname: xc.bcname,
				Txid:   tx.Txid,
				Tx:     tx,
			}
			batchTxMsg.Txs = append(batchTxMsg.Txs, txStatus)
		}
		xc.log.Debug("repost batch tx list", "size", len(batchTxMsg.Txs))
		msgInfo, _ := proto.Marshal(batchTxMsg)
		msg, _ := xuper_p2p.NewXuperMessage(xuper_p2p.XuperMsgVersion1, xc.bcname, header.GetLogid(), xuper_p2p.XuperMessage_BATCHPOSTTX, msgInfo, xuper_p2p.XuperMessage_SUCCESS)
		go xc.P2pv2.SendMessage(context.Background(), msg, p2pv2.DefaultStrategy) //p2p广播出去
	}
}

// SendBlock send block
func (xc *XChainCore) SendBlock(in *pb.Block, hd *global.XContext) error {
	if xc.Status() != global.Normal {
		xc.log.Debug("refused a connection at function call GenerateTx", "logid", in.Header.Logid, "cost", hd.Timer.Print())
		return ErrServiceRefused
	}
	xc.mutex.Lock()
	defer xc.mutex.Unlock()

	nonVerify := (xc.nodeMode == config.NodeModeFastSync)

	// 如果已经存在，则立即返回
	if xc.Ledger.ExistBlock(in.Blockid) {
		xc.log.Debug("Block is exist", "logid", in.Header.Logid, "cost", hd.Timer.Print())
		return ErrBlockExist
	}

	//在锁外校验block中的tx合法性
	xc.mutex.Unlock()
	for idx, tx := range in.Block.Transactions {
		if !xc.Ledger.IsValidTx(idx, tx, in.Block) {
			xc.log.Warn("invalid tx got from the block", "txid", global.F(tx.Txid), "blkid", global.F(in.Block.Blockid))
			xc.mutex.Lock()
			return ErrInvalidBlock
		}
	}
	xc.mutex.Lock()
	if xc.Ledger.ExistBlock(in.Blockid) {
		//放锁期间，可能这个块已经被另外一个线程存进去了，所以需要再次判断
		xc.log.Debug("Block is exist", "logid", in.Header.Logid, "cost", hd.Timer.Print())
		return ErrBlockExist
	}
	blocksIds := []string{}
	//如果是接受到老的block（版本是1）, TODO
	blocksIds = append(blocksIds, string(in.Block.Blockid))
	err := xc.Ledger.SavePendingBlock(in)
	if err != nil {
		xc.log.Warn("Save Pending Block error! ", "logid", in.Header.Logid, "blockid", in.Block.Blockid)
		return ErrCannotSyncBlock
	}
	// var preblk *pb.InternalBlock = in.Block
	proposeBlockMoreThanConfig := false //块是否为非法块
	preblkhash := in.Block.PreHash
	for {
		xc.log.Debug("Start to Find ExistBlock", "logid", in.Header.Logid, "cost", hd.Timer.Print(), "prehash", global.F(preblkhash))
		if xc.Ledger.ExistBlock(preblkhash) {
			xc.log.Debug("Find Same Block", "logid", in.Header.Logid, "prehash", global.F(preblkhash))
			break
		}
		// call for prehash
		ib, _ := xc.Ledger.GetPendingBlock(preblkhash)
		if ib == nil {
			xc.log.Debug("Start to BroadCastGetBlock", "logid", in.Header.Logid, "cost", hd.Timer.Print())
			ib = xc.BroadCastGetBlock(&pb.BlockID{Header: in.Header, Bcname: in.Bcname, Blockid: preblkhash, NeedContent: true})
			if ib == nil {
				xc.log.Warn("Can't Get a Block", "logid", in.Header.Logid, "blockid", global.F(preblkhash))
				return ErrCannotSyncBlock
			} else if ib.Block == nil {
				xc.log.Warn("Get a Block Content error", "logid", in.Header.Logid, "blokid", global.F(preblkhash), "error", in.Header.Error)
				return ErrCannotSyncBlock
			} else {
				err := xc.Ledger.SavePendingBlock(ib)
				if err != nil {
					xc.log.Warn("Save Pending Block error, after got it from network! ", "logid", in.Header.Logid, "blockid", in.Block.Blockid)
					return ErrCannotSyncBlock
				}
			}
		}
		preblkhash = ib.Block.PreHash
		blocksIds = append(blocksIds, string(ib.Block.Blockid))
	}

	xc.log.Debug("End to Find the same", "logid", in.Header.Logid, "blocks size", len(blocksIds), "cost", hd.Timer.Print(),
		"genesis", global.F(xc.Ledger.GetMeta().RootBlockid),
		"prehash", global.F(preblkhash), "utxo", global.F(xc.Utxovm.GetLatestBlockid()))
	rbErr := xc.pipelineM.RollbackPrePlay()
	if rbErr != nil {
		xc.log.Warn("fail to rollback preplay contract", "err", rbErr)
		return rbErr
	}
	// preblk 是跟区块同步的交点，判断preblk是不是当前utxo的位置
	if bytes.Equal(xc.Utxovm.GetLatestBlockid(), preblkhash) {
		xc.log.Debug("Equal The Same", "logid", in.Header.Logid, "cost", hd.Timer.Print())
		for i := len(blocksIds) - 1; i >= 0; i-- {
			block, err := xc.Ledger.GetPendingBlock([]byte(blocksIds[i]))
			if block == nil {
				xc.log.Warn("GetPendingBlock from ledger error", "logid", in.Header.Logid, "cost", hd.Timer.Print(), "err", err)
				return ErrConfirmBlock
			}
			// 区块加解密有效性检查
			if !nonVerify {
				if res, _ := xc.con.CheckMinerMatch(in.Header, block.Block); !res {
					xc.log.Warn("refused a connection becausefo check miner error", "logid", in.Header.Logid, "cost", hd.Timer.Print())
					return ErrServiceRefused
				}
			}
			cs := xc.Ledger.ConfirmBlock(block.Block, false)
			xc.log.Debug("ConfirmBlock Time", "logid", in.Header.Logid, "cost", hd.Timer.Print())
			if !cs.Succ {
				xc.log.Warn("confirm error", "logid", in.Header.Logid)
				return ErrConfirmBlock
			}
			isTipBlock := (i == 0)
			err = xc.Utxovm.PlayAndRepost(block.Blockid, isTipBlock, false)
			xc.log.Debug("Play Time", "logid", in.Header.Logid, "cost", hd.Timer.Print())
			if err != nil {
				xc.log.Warn("utxo vm play err", "logid", in.Header.Logid, "err", err)
				return ErrUTXOVMPlay
			}
		}
	} else {
		//交点不等于utxo latest block
		xc.log.Debug("XXXXXXXXX The NO Same", "logid", in.Header.Logid, "cost", hd.Timer.Print())
		block0 := &pb.Block{}
		trunkSwitch := false //是否发生主干切换
		for i := len(blocksIds) - 1; i >= 0; i-- {
			block, err := xc.Ledger.GetPendingBlock([]byte(blocksIds[i]))
			if err != nil {
				xc.log.Warn("GetPendingBlock from leadger error", "logid", in.Header.Logid, "cost", hd.Timer.Print())
				return ErrConfirmBlock
			}
			if i == 0 {
				block0 = block
			}

			if res, err := xc.con.CheckMinerMatch(in.Header, block.Block); !res {
				if err != nil && err == tdpos.ErrProposeBlockMoreThanConfig {
					proposeBlockMoreThanConfig = true
					xc.log.Warn("CheckMinerMatch ErrProposeBlockMoreThanConfig", "logid", in.Header.Logid, "cost", hd.Timer.Print())
					break
				}
				xc.log.Warn("refused a connection becausefo check miner error", "logid", in.Header.Logid, "cost", hd.Timer.Print())
				return ErrServiceRefused
			}

			cs := xc.Ledger.ConfirmBlock(block.Block, false)
			xc.log.Debug("ConfirmBlock Time", "logid", in.Header.Logid, "cost", hd.Timer.Print(), "blockid", global.F(block.Blockid))
			if !cs.Succ {
				xc.log.Warn("confirm error", "logid", in.Header.Logid)
				return ErrConfirmBlock
			}
			trunkSwitch = (cs.TrunkSwitch || block.Block.InTrunk)
		}
		if !trunkSwitch {
			xc.log.Warn("no need to do walk", "trunkSwitch", trunkSwitch, "blockid", global.F(block0.Blockid))
			if proposeBlockMoreThanConfig {
				return ErrProposeBlockMoreThanConfig
			}
			return nil
		}
		err := xc.Utxovm.Walk(block0.Blockid)
		xc.log.Debug("Walk Time", "logid", in.Header.Logid, "cost", hd.Timer.Print())
		if err != nil {
			xc.log.Warn("Walk error", "logid", in.Header.Logid, "err", err)
			return ErrWalk
		}
	}
	// 待块确认后, 共识执行相应的操作
	xc.con.ProcessConfirmBlock(in.Block)
	if proposeBlockMoreThanConfig {
		return ErrProposeBlockMoreThanConfig
	}
	return nil
}

func (xc *XChainCore) doMiner() {
	minerTimer := global.NewXTimer()
	xc.mutex.Lock()
	lockHold := true
	minerTimer.Mark("GetLock")
	defer func() {
		if lockHold {
			xc.mutex.Unlock()
		}
	}()
	ledgerLastID := xc.Ledger.GetMeta().TipBlockid
	utxovmLastID := xc.Utxovm.GetLatestBlockid()

	// 如果Walk一直失败，建议不要挖矿了，而是报警处理
	for !bytes.Equal(ledgerLastID, utxovmLastID) {
		xc.log.Warn("ledger last blockid is not equal utxovm last id")
		rbErr := xc.pipelineM.RollbackPrePlay()
		if rbErr != nil {
			xc.log.Warn("fail to rollback preplay contract", "err", rbErr)
			return
		}
		err := xc.Utxovm.Walk(ledgerLastID)
		if err != nil {
			xc.log.Error("Walk error ", "ledger blockid", global.F(ledgerLastID),
				"utxo blockid", global.F(utxovmLastID))
			return
		}
		ledgerLastID = xc.Ledger.GetMeta().TipBlockid
		utxovmLastID = xc.Utxovm.GetLatestBlockid()
	}

	header := &pb.Header{Logid: global.Glogid()}

	// 直接使用maxBlockSize约束tx的体积，为单个tx体积过大的情况兜底
	// 问题：1.实际上区块大小仍可能超过maxBlockSize(因为还有区块头)；
	//      2.不能控制merkleTree的体积
	// todo: byj, 优化序列化方法与存储结构?
	// 打包块起始时间
	t := time.Now()
	txs := []*pb.Transaction{}
	// 挖矿前共识的预处理
	var curTerm, curBlockNum int64
	var targetBits int32
	data, ok := xc.con.ProcessBeforeMiner(xc.Ledger.GetMeta().TrunkHeight+1, t.UnixNano())
	minerTimer.Mark("ProcessBeforeMiner")
	if ok {
		if data != nil {
			if v, ok := data["type"]; ok {
				switch v {
				case consensus.ConsensusTypeTdpos:
					xc.log.Trace("Minning tdpos ProcessBeforeMiner!")
					curTerm = data["curTerm"].(int64)
					curBlockNum = data["curBlockNum"].(int64)
				case consensus.ConsensusTypePow:
					xc.log.Trace("Minning tdpos ProcessBeforeMiner!")
					targetBits = data["targetBits"].(int32)
				}
			}
		}
	} else {
		xc.log.Trace("Minning ProcessBeforeMiner not ok!")
		return
	}
	meta := xc.Ledger.GetMeta()
	//1. 查询自动生成的交易
	vatList, err := xc.Utxovm.GetVATList(xc.Ledger.GetMeta().TrunkHeight+1, -1, t.UnixNano())
	minerTimer.Mark("GetAutogenTxs")
	if err != nil {
		xc.log.Warn("[Minning] fail to get triggered tx list", "logid", header.Logid)
		return
	}
	xc.log.Trace("[Minning] get vatList success", "vatList", vatList)
	txs = append(txs, vatList...)
	// make fake block
	fakeBlock, err := xc.Ledger.FormatFakeBlock(txs, xc.address, xc.privateKey,
		t.UnixNano(), curTerm, curBlockNum, meta.TipBlockid, xc.Utxovm.GetTotal())
	if err != nil {
		xc.log.Warn("[Minning] format block error", "logid")
		return
	}
	allFailedTxs := map[string]string{}
	if txs, _, err = xc.Utxovm.TxOfRunningContractGenerate(txs, fakeBlock, xc.pipelineM.batch, xc.pipelineM.NeedInitCtx()); err != nil {
		if err.Error() != common.ErrContractExecutionTimeout.Error() {
			xc.log.Warn("PrePlay failed", "error", err)
			return
		}
	}
	for txid, txErr := range fakeBlock.FailedTxs {
		allFailedTxs[txid] = txErr
	}
	//2. 打包已经预执行过的未确认交易
	batch, txsUnconf, failedTxs := xc.pipelineM.FetchTxs()
	txs = append(txs, txsUnconf...)
	for txid, txErr := range failedTxs {
		allFailedTxs[txid] = txErr
	}
	minerTimer.Mark("PrePlay")
	//3. 统一在最后插入矿工奖励
	blockAward := xc.Ledger.GenesisBlock.CalcAward(xc.Ledger.GetMeta().TrunkHeight + 1)
	awardtx, err := xc.Utxovm.GenerateAwardTx(xc.address, blockAward.String(), []byte{'1'})
	minerTimer.Mark("GenAwardTx")
	txs = append(txs, awardtx)
	b, err := xc.Ledger.FormatPOWBlock(txs, xc.address, xc.privateKey,
		t.UnixNano(), curTerm, curBlockNum, meta.TipBlockid, targetBits, xc.Utxovm.GetTotal(), allFailedTxs)
	if err != nil {
		xc.log.Warn("[Minning] format block error", "logid", header.Logid, "err", err)
		return
	}
	minerTimer.Mark("Formatblock2")

	xc.log.Debug("[Minning] Start to ConfirmBlock", "logid", header.Logid)
	confirmStatus := xc.Ledger.ConfirmBlock(b, false)
	minerTimer.Mark("ConfirmBlock")
	if confirmStatus.Succ {
		if confirmStatus.Orphan {
			xc.log.Warn("[Minning] the mined blocked was attached to branch, no need to play")
			return
		}
		xc.log.Info("[Minning] ConfirmBlock Success", "logid", header.Logid, "Height", meta.TrunkHeight+1)
	} else {
		xc.log.Warn("[Minning] ConfirmBlock Fail", "logid", header.Logid, "confirm_status", confirmStatus)
		return
	}
	xc.mutex.Unlock() //后面放开锁
	lockHold = false
	xc.Utxovm.SetBlockGenEvent()
	defer xc.Utxovm.NotifyFinishBlockGen()
	err = xc.Utxovm.PlayForMiner(b.Blockid, batch)
	if err != nil {
		xc.log.Warn("[Minning] utxo play error ", "logid", header.Logid, "error", err, "blockid", fmt.Sprintf("%x", b.Blockid))
		return
	}
	minerTimer.Mark("PlayForMiner")
	xc.con.ProcessConfirmBlock(b)
	minerTimer.Mark("ProcessConfirmBlock")
	xc.log.Debug("[Minning] Start to BroadCast", "logid", header.Logid)

	// broadcast block
	block := &pb.Block{
		Bcname:  xc.bcname,
		Blockid: b.Blockid,
		Block:   b,
	}
	msgInfo, _ := proto.Marshal(block)
	msg, _ := xuper_p2p.NewXuperMessage(xuper_p2p.XuperMsgVersion1, xc.bcname, "", xuper_p2p.XuperMessage_SENDBLOCK, msgInfo, xuper_p2p.XuperMessage_NONE)
	go xc.P2pv2.SendMessage(context.Background(), msg, p2pv2.DefaultStrategy)
	minerTimer.Mark("BroadcastBlock")
	if xc.Utxovm.IsAsync() {
		xc.log.Warn("doMiner cost", "cost", minerTimer.Print(), "txCount", b.TxCount)
	} else {
		xc.log.Debug("doMiner cost", "cost", minerTimer.Print(), "txCount", b.TxCount)
	}
}

// Miner start to miner
func (xc *XChainCore) Miner() int {
	// 1 强制walk到最新状态
	ledgerLastID := xc.Ledger.GetMeta().TipBlockid
	utxovmLastID := xc.Utxovm.GetLatestBlockid()
	if !bytes.Equal(ledgerLastID, utxovmLastID) {
		xc.log.Warn("ledger last blockid is not equal utxovm last id")
		xc.Utxovm.Walk(ledgerLastID)
	}
	// 2 FAST_SYNC模式下需要回滚掉本地所有的未确认交易
	if xc.nodeMode == config.NodeModeFastSync {
		if _, err := xc.Utxovm.RollBackUnconfirmedTx(); err != nil {
			xc.log.Warn("FAST_SYNC mode RollBackUnconfirmedTx error", "error", err)
		}
	}
	// 3 开始同步
	xc.status = global.Normal
	xc.SyncBlocks()
	for {
		// 重要: 首次出块前一定要同步到最新的状态
		xc.log.Trace("Miner type of consensus", "type", xc.con.Type(xc.Ledger.GetMeta().TrunkHeight+1))
		b, s := xc.con.CompeteMaster(xc.Ledger.GetMeta().TrunkHeight + 1)
		xc.log.Debug("competemaster", "blockchain", xc.bcname, "master", b, "needSync", s)
		if b {
			// todo 首次切换为矿工时SyncBlcok, Bug: 可能会导致第一次出块失败
			if s {
				xc.SyncBlocks()
			}
			xc.pipelineM.Resume()
			xc.doMiner()
		} else {
			xc.pipelineM.Pause()
		}
		meta := xc.Ledger.GetMeta()
		xc.log.Info("Minner", "genesis", fmt.Sprintf("%x", meta.RootBlockid), "last", fmt.Sprintf("%x", meta.TipBlockid), "height", meta.TrunkHeight, "utxovm", fmt.Sprintf("%x", xc.Utxovm.GetLatestBlockid()))
		if xc.stopFlag {
			break
		}
	}
	return 0
}

// Stop stop one xchain instance
func (xc *XChainCore) Stop() {
	xc.Utxovm.Close()
	xc.stopFlag = true
}

// PostTx post transaction to utxo and broad cast the transaction
func (xc *XChainCore) PostTx(in *pb.TxStatus, hd *global.XContext) (*pb.CommonReply, bool) {
	out := &pb.CommonReply{Header: in.Header}
	out.Header.Error = pb.XChainErrorEnum_SUCCESS
	if xc.Status() != global.Normal {
		out.Header.Error = pb.XChainErrorEnum_CONNECT_REFUSE // 拒绝
		xc.log.Debug("refused a connection at function call GenerateTx", "logid", in.Header.Logid)
		return out, false
	}

	// 对Tx进行的签名, 1 如果utxo属于用户，则走原来的验证逻辑 2 如果utxo属于账户，则走账户acl验证逻辑
	txValid, validErr := xc.Utxovm.VerifyTx(in.Tx)
	if !txValid {
		switch validErr {
		case utxo.ErrGasNotEnough:
			out.Header.Error = pb.XChainErrorEnum_GAS_NOT_ENOUGH_ERROR
		case utxo.ErrRWSetInvalid, utxo.ErrInvalidTxExt:
			out.Header.Error = pb.XChainErrorEnum_RWSET_INVALID_ERROR
		case utxo.ErrRWAclNotEnough:
			out.Header.Error = pb.XChainErrorEnum_RWACL_INVALID_ERROR
		case utxo.ErrVersionInvalid:
			out.Header.Error = pb.XChainErrorEnum_TX_VERSION_INVALID_ERROR
		default:
			out.Header.Error = pb.XChainErrorEnum_TX_SIGN_ERROR
		}
		xc.log.Warn("post tx verify tx error", "txid", global.F(in.Tx.Txid),
			"valid_err", validErr, "logid", in.Header.Logid)
		return out, false
	}

	err := xc.Utxovm.DoTx(in.Tx)
	xc.log.Debug("Utxovm DoTx", "logid", in.Header.Logid, "cost", hd.Timer.Print())
	if err != nil {
		out.Header.Error = HandlerUtxoError(err)
		if err != utxo.ErrAlreadyInUnconfirmed {
			xc.log.Warn("utxo vm do tx error", "logid", in.Header.Logid, "error", err)
		}
		return out, false
	}
	xc.Speed.Add("PostTx")
	if xc.Utxovm.IsAsync() {
		return out, xc.Utxovm.IsInUnConfirm(string(in.Txid))
	}
	return out, true
}

// QueryTx query transaction from ledger
func (xc *XChainCore) QueryTx(in *pb.TxStatus) *pb.TxStatus {
	out := &pb.TxStatus{Header: in.Header}
	out.Header.Error = pb.XChainErrorEnum_SUCCESS
	out.Status = pb.TransactionStatus_UNDEFINE
	out.Bcname = in.Bcname
	out.Txid = in.Txid
	if xc.Status() != global.Normal {
		out.Header.Error = pb.XChainErrorEnum_CONNECT_REFUSE
		xc.log.Debug("refused a connection a function call QueryTx", "logid", in.Header.Logid)
		return out
	}

	t, err := xc.Ledger.QueryTransaction(out.Txid)
	if err != nil {
		xc.log.Debug("Query Transaction Error", "logid", in.Header.Logid, "Txid", global.F(out.Txid), "error", err)
		out.Status = pb.TransactionStatus_NOEXIST
		/*if err == ledger.parser_err {
			this.log.Warn("Parser error")
		}*/
		if err == ledger.ErrTxNotFound {
			// 查询unconfirm表，看看
			t, err = xc.Utxovm.QueryTx(out.Txid)
			if err != nil {
				xc.log.Debug("Query Transaction Unconfirm table Error", "logid", in.Header.Logid, "Txid", global.F(out.Txid), "error", err)
				return out
			}
			xc.log.Debug("Query Transaction Unconfirm table Success", "logid", in.Header.Logid, "Txid", global.F(out.Txid))
			out.Status = pb.TransactionStatus_UNCONFIRM
			out.Tx = t
			return out
		}
	} else {
		xc.log.Debug("Query Transaction Successa", "logid", in.Header.Logid, "Txid", global.F(out.Txid))
		out.Status = pb.TransactionStatus_CONFIRM
		// 根据blockid查block状态，看是否被分叉
		ib, err := xc.Ledger.QueryBlockHeader(t.Blockid)
		if err != nil {
			xc.log.Debug("Query Block Error", "logid", in.Header.Logid, "Txid", global.F(out.Txid), "blockid", global.F(t.Blockid), "error", err)
			out.Header.Error = pb.XChainErrorEnum_UNKNOW_ERROR
		} else {
			xc.log.Debug("Query Block Success", "logid", in.Header.Logid, "Txid", global.F(out.Txid), "blockid", global.F(t.Blockid))
			meta := xc.Ledger.GetMeta()
			out.Tx = t
			if ib.InTrunk {
				// out.Distance =  height - ib.height
				out.Distance = meta.TrunkHeight - ib.Height
				out.Status = pb.TransactionStatus_CONFIRM
			} else {
				out.Status = pb.TransactionStatus_FURCATION
			}
		}
	}

	return out
}

// GetBlock get block from ledger
func (xc *XChainCore) GetBlock(in *pb.BlockID) *pb.Block {
	out := &pb.Block{Header: global.GHeader()}
	out.Header.Error = pb.XChainErrorEnum_SUCCESS
	out.Bcname = in.Bcname

	if xc.Status() != global.Normal {
		out.Header.Error = pb.XChainErrorEnum_CONNECT_REFUSE
		xc.log.Debug("refused a connection a function call GetBlock", "logid", in.Header.Logid)
		return out
	}

	ib, err := xc.Ledger.QueryBlock(in.Blockid)
	if err != nil {
		switch err {
		case ledger.ErrBlockNotExist:
			out.Header.Error = pb.XChainErrorEnum_SUCCESS
			out.Status = pb.Block_NOEXIST
			return out
		default:
			xc.log.Warn("getblock", "logid", in.Header.Logid, "error", err)
			out.Header.Error = pb.XChainErrorEnum_UNKNOW_ERROR
			return out
		}
	} else {
		xc.log.Debug("debug needcontent", "logid", in.Header.Logid, "needcontent", in.NeedContent)
		if in.NeedContent {
			out.Block = ib
		}
		if ib.InTrunk {
			out.Status = pb.Block_TRUNK
		} else {
			out.Status = pb.Block_BRANCH
		}
	}
	return out
}

// GetBlockChainStatus get block status from ledger
func (xc *XChainCore) GetBlockChainStatus(in *pb.BCStatus) *pb.BCStatus {
	if in.GetHeader() == nil {
		in.Header = global.GHeader()
	}
	out := &pb.BCStatus{Header: in.Header}
	out.Bcname = in.Bcname
	out.Header.Error = pb.XChainErrorEnum_SUCCESS
	if xc.Status() != global.Normal {
		out.Header.Error = pb.XChainErrorEnum_CONNECT_REFUSE
		xc.log.Debug("refused a connection a function call GetBlock", "logid", in.Header.Logid)
		return out
	}

	meta := xc.Ledger.GetMeta()
	out.Meta = meta
	utxoMeta := xc.Utxovm.GetMeta()
	out.UtxoMeta = utxoMeta

	ib, err := xc.Ledger.QueryBlock(meta.TipBlockid)
	if err != nil {
		out.Header.Error = HandlerLedgerError(err)
		return out
	}
	out.Block = ib

	return out
}

// ConfirmTipBlockChainStatus check tip block status
func (xc *XChainCore) ConfirmTipBlockChainStatus(in *pb.BCStatus) *pb.BCTipStatus {
	out := &pb.BCTipStatus{Header: global.GHeader()}
	out.Header.Error = pb.XChainErrorEnum_SUCCESS
	meta := xc.Ledger.GetMeta()
	if string(in.Block.GetBlockid()) == string(meta.TipBlockid) {
		out.IsTrunkTip = true
	} else {
		out.IsTrunkTip = false
	}
	return out
}

// QueryContractMethodACL get ACL for a contract method
func (xc *XChainCore) QueryContractMethodACL(contractName string, methodName string) (*pb.Acl, bool, error) {
	if xc == nil {
		return nil, false, errors.New("xchain core is nil")
	}
	if xc.Status() != global.Normal {
		return nil, false, ErrNotReady
	}
	acl, confirmed, err := xc.Utxovm.QueryContractMethodACLWithConfirmed(contractName, methodName)
	if err != nil {
		return nil, false, err
	}
	return acl, confirmed, nil
}

// QueryAccountACL get ACL for an account
func (xc *XChainCore) QueryAccountACL(accountName string) (*pb.Acl, bool, error) {
	if xc == nil {
		return nil, false, errors.New("xchain core is nil")
	}
	if xc.Status() != global.Normal {
		return nil, false, ErrNotReady
	}
	acl, confirmed, err := xc.Utxovm.QueryAccountACLWithConfirmed(accountName)
	if err != nil {
		return nil, false, err
	}
	return acl, confirmed, nil
}

// GetBalance get balance from utxo
func (xc *XChainCore) GetBalance(addr string) (string, error) {
	if xc.Status() != global.Normal {
		return "", ErrNotReady
	}
	bint, err := xc.Utxovm.GetBalance(addr)
	if err != nil {
		return "", err
	}
	return bint.String(), nil
}

// GetFrozenBalance get balance that still be frozen from utxo
func (xc *XChainCore) GetFrozenBalance(addr string) (string, error) {
	if xc.Status() != global.Normal {
		return "", ErrNotReady
	}
	bint, err := xc.Utxovm.GetFrozenBalance(addr)
	if err != nil {
		return "", err
	}
	return bint.String(), nil
}

// GetConsType get consensus type for specific block chain
func (xc *XChainCore) GetConsType() string {
	return xc.con.Type(xc.Ledger.GetMeta().TrunkHeight + 1)
}

// GetDposCandidates get all candidates
func (xc *XChainCore) GetDposCandidates() ([]string, error) {
	candidates := []string{}
	it := xc.Utxovm.ScanWithPrefix([]byte(tdpos.GenCandidateBallotsPrefix()))
	defer it.Release()
	for it.Next() {
		key := string(it.Key())
		addr, err := tdpos.ParseCandidateBallotsKey(key)
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, addr)
	}
	return candidates, nil
}

// GetDposNominateRecords get nominate(positively) record infos for specific address
func (xc *XChainCore) GetDposNominateRecords(addr string) ([]*pb.DposNominateInfo, error) {
	nominateRecords := []*pb.DposNominateInfo{}
	it := xc.Utxovm.ScanWithPrefix([]byte(tdpos.GenNominateRecordsPrefix(addr)))
	defer it.Release()
	for it.Next() {
		key := string(it.Key())
		addrCandidate, txid, err := tdpos.ParseNominateRecordsKey(key)
		if err != nil {
			return nil, err
		}
		nominateRecord := &pb.DposNominateInfo{
			Candidate: addrCandidate,
			Txid:      txid,
		}
		nominateRecords = append(nominateRecords, nominateRecord)
	}
	return nominateRecords, nil
}

// GetDposNominatedRecords get nominated(passively) record infos for specific address
func (xc *XChainCore) GetDposNominatedRecords(addr string) (string, error) {
	key := tdpos.GenCandidateNominateKey(addr)
	val, err := xc.Utxovm.GetFromTable(nil, []byte(key))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(val), err
}

// GetDposVoteRecords get vote(positively) record infos for specific address
func (xc *XChainCore) GetDposVoteRecords(addr string) ([]*pb.VoteRecord, error) {
	voteRecords := []*pb.VoteRecord{}
	it := xc.Utxovm.ScanWithPrefix([]byte(tdpos.GenVoteCandidatePrefix(addr)))
	defer it.Release()
	for it.Next() {
		key := string(it.Key())
		candidate, txid, err := tdpos.ParseVoteCandidateKey(key)
		voteRecord := &pb.VoteRecord{
			Candidate: candidate,
			Txid:      txid,
		}
		if err != nil {
			return nil, err
		}
		voteRecords = append(voteRecords, voteRecord)
	}
	return voteRecords, nil
}

// GetDposVotedRecords get voted(passively) record infos for specific address
func (xc *XChainCore) GetDposVotedRecords(addr string) ([]*pb.VotedRecord, error) {
	votedRecords := []*pb.VotedRecord{}
	it := xc.Utxovm.ScanWithPrefix([]byte(tdpos.GenCandidateVotePrefix(addr)))
	defer it.Release()
	for it.Next() {
		key := string(it.Key())
		voter, txid, err := tdpos.ParseCandidateVoteKey(key)
		votedRecord := &pb.VotedRecord{
			Voter: voter,
			Txid:  txid,
		}
		if err != nil {
			return nil, err
		}
		votedRecords = append(votedRecords, votedRecord)
	}
	return votedRecords, nil
}

// GetCheckResults get all proposers for specific term
func (xc *XChainCore) GetCheckResults(term int64) ([]string, error) {
	proposers := []string{}
	version := xc.con.Version(xc.Ledger.GetMeta().TrunkHeight + 1)
	key := tdpos.GenTermCheckKey(version, term)
	val, err := xc.Utxovm.GetFromTable(nil, []byte(key))
	if err != nil || val == nil {
		return nil, err
	}
	err = json.Unmarshal(val, &proposers)
	if err != nil {
		return nil, err
	}
	return proposers, nil
}

// GetNodeMode get node running mode, such as Normal mode, FastSync mode
func (xc *XChainCore) GetNodeMode() string {
	return xc.nodeMode
}

// PreExec get read/write set for smart contract could be run in parallel
func (xc *XChainCore) PreExec(req *pb.InvokeRPCRequest, hd *global.XContext) (*pb.InvokeResponse, error) {
	return xc.Utxovm.PreExec(req, hd)
}

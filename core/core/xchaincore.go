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
	"syscall"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/patrickmn/go-cache"
	log "github.com/xuperchain/log15"

	"github.com/xuperchain/xuperchain/core/common"
	"github.com/xuperchain/xuperchain/core/common/config"
	"github.com/xuperchain/xuperchain/core/common/events"
	"github.com/xuperchain/xuperchain/core/common/probe"
	"github.com/xuperchain/xuperchain/core/consensus"
	cons_base "github.com/xuperchain/xuperchain/core/consensus/base"
	"github.com/xuperchain/xuperchain/core/consensus/tdpos"
	"github.com/xuperchain/xuperchain/core/contract/bridge"
	"github.com/xuperchain/xuperchain/core/contract/kernel"
	"github.com/xuperchain/xuperchain/core/contract/proposal"
	"github.com/xuperchain/xuperchain/core/crypto/account"
	crypto_client "github.com/xuperchain/xuperchain/core/crypto/client"
	crypto_base "github.com/xuperchain/xuperchain/core/crypto/client/base"
	"github.com/xuperchain/xuperchain/core/global"
	"github.com/xuperchain/xuperchain/core/kv/kvdb"
	"github.com/xuperchain/xuperchain/core/ledger"
	p2p_base "github.com/xuperchain/xuperchain/core/p2p/base"
	xuper_p2p "github.com/xuperchain/xuperchain/core/p2p/pb"
	"github.com/xuperchain/xuperchain/core/pb"
	"github.com/xuperchain/xuperchain/core/utxo"
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
	// ErrBlockTooLarge is returned when its size greater than the max block size defined
	ErrBlockTooLarge = errors.New("block is too large")
)

const (
	TxidCacheGcTime = 180 * time.Second

	// DefaultMessageCacheSize for p2p message
	DefaultMessageCacheSize = 1000
)

// XChainCore is the core struct of a chain
type XChainCore struct {
	con          *consensus.PluggableConsensus
	Ledger       *ledger.Ledger
	Utxovm       *utxo.UtxoVM
	P2pSvr       p2p_base.P2PServer
	LedgerKeeper *LedgerKeeper
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
	stopFlag bool
	proposal *proposal.Proposal

	// isCoreMiner if current node is one of the core miners
	isCoreMiner bool
	// enable core peer connection or not
	coreConnection bool
	// if failSkip is false, you will execute loop of walk, or just only once walk
	failSkip bool
	// add a lru cache of tx gotten for xchaincore
	txidCache            *cache.Cache
	txidCacheExpiredTime time.Duration
	enableCompress       bool
	pruneOption          config.PruneOption

	// cache for duplicate block message
	msgCache           *common.LRUCache
	blockBroadcaseMode uint8
	// group chain involved
	groupChain GroupChainRegister
}

// Status return the status of the chain
func (xc *XChainCore) Status() int {
	return xc.status
}

// Init init the chain
func (xc *XChainCore) Init(bcname string, xlog log.Logger, cfg *config.NodeConfig,
	p2p p2p_base.P2PServer, ker *kernel.Kernel, nodeMode string, groupChain GroupChainRegister) error {

	// 设置全局随机数发生器的原始种子
	err := global.SetSeed()
	if err != nil {
		return err
	}

	xc.mutex = &sync.RWMutex{}
	xc.Speed = probe.NewSpeedCalc(bcname)
	// this.mutex.Lock()
	// defer this.mutex.Unlock()
	xc.groupChain = groupChain
	xc.status = global.SafeModel
	xc.bcname = bcname
	xc.log = xlog
	xc.P2pSvr = p2p
	xc.nodeMode = nodeMode
	xc.stopFlag = false
	xc.coreConnection = cfg.CoreConnection
	xc.failSkip = cfg.FailSkip
	xc.txidCacheExpiredTime = cfg.TxidCacheExpiredTime
	xc.txidCache = cache.New(xc.txidCacheExpiredTime, TxidCacheGcTime)
	xc.enableCompress = cfg.EnableCompress
	xc.pruneOption = cfg.Prune
	xc.msgCache = common.NewLRUCache(DefaultMessageCacheSize)
	xc.blockBroadcaseMode = cfg.BlockBroadcaseMode
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
	xchainAddrInfo := &p2p_base.XchainAddrInfo{
		Addr:   string(addr),
		Pubkey: pub,
		Prikey: pri,
	}
	xc.log.Debug("+++++++setbefore", "P2PSvr", xc.P2pSvr, "xchainAddrInfo", xchainAddrInfo)
	xc.P2pSvr.SetXchainAddr(xc.bcname, xchainAddrInfo)

	xc.Ledger, err = ledger.OpenLedger(datapath, xc.log, datapathOthers, kvEngineType, cryptoType)
	if err != nil {
		xc.log.Warn("OpenLedger error", "bc", xc.bcname, "datapath", datapath, "dataPathOhters", datapathOthers)
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

	// init events with handler
	xc.initEvents()

	xc.Utxovm, err = utxo.MakeUtxoVM(bcname, xc.Ledger, datapath, privateKeyStr, publicKeyStr, xc.address, xc.log,
		utxoCacheSize, utxoTmplockSeconds, cfg.Utxo.ContractExecutionTime, datapathOthers, cfg.Utxo.IsBetaTx[bcname], kvEngineType, cryptoType)

	if err != nil {
		xc.log.Warn("NewUtxoVM error", "bc", xc.bcname, "datapath", datapath, "dataPathOhters", datapathOthers)
		return err
	}
	if cfg.Utxo.AsyncMode {
		xc.Utxovm.StartAsyncWriter()
	} else if cfg.Utxo.AsyncBlockMode {
		//
		xc.Utxovm.StartAsyncBlockMode()
	}
	xc.Utxovm.SetMaxConfirmedDelay(cfg.Utxo.MaxConfirmedDelay)
	xc.Utxovm.SetModifyBlockAddr(cfg.Kernel.ModifyBlockAddr)
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
	xc.con, err = consensus.NewPluggableConsensus(xlog, cfg, bcname, xc.Ledger, xc.Utxovm, gCon, cryptoType, p2p)
	if err != nil {
		xc.log.Warn("New PluggableConsensus Error")
		return err
	}

	xc.proposal = proposal.NewProposal(xc.log, xc.Ledger, xc.Utxovm)
	// 统一注册所有的合约虚拟机
	xc.Utxovm.RegisterVM("kernel", ker, global.VMPrivRing0)
	xc.Utxovm.RegisterVM("consensus", xc.con, global.VMPrivRing0)
	xc.Utxovm.RegisterVM("proposal", xc.proposal, global.VMPrivRing0)

	basedir, err := filepath.Abs(datapath)
	if err != nil {
		return err
	}
	xbridge, err := bridge.New(&bridge.XBridgeConfig{
		Basedir: basedir,
		VMConfigs: map[bridge.ContractType]bridge.VMConfig{
			bridge.TypeWasm:   &cfg.Wasm,
			bridge.TypeNative: &cfg.Native,
		},
		XModel: xc.Utxovm.GetXModel(),
		Config: cfg.Contract,
	})
	if err != nil {
		return err
	}
	xbridge.RegisterToXCore(xc.Utxovm.RegisterVM3)

	// 统一注册xuper3合约虚拟机
	x3kernel, xerr := kernel.NewKernel(xbridge)
	if xerr != nil {
		return xerr
	}
	xc.Utxovm.RegisterVM3(x3kernel.GetName(), x3kernel)

	// 统一注册VAT
	xc.Utxovm.RegisterVAT("Propose", xc.proposal, nil)
	xc.Utxovm.RegisterVAT("consensus", xc.con, xc.con.GetVATWhiteList())
	xc.Utxovm.RegisterVAT("kernel", ker, ker.GetVATWhiteList())

	// 启动Sync节点，负责区块同步
	sn := NewLedgerKeeper(xc.bcname, xc.log, xc.P2pSvr, xc.Ledger.GetMaxBlockSize(), xc.Ledger, xc.nodeMode, xc.Utxovm, xc.con)
	if err := sn.Init(); err != nil {
		xc.log.Warn("Xchaincore::Init::NewLedgerKeeper init error", "err", err)
		return err
	}
	sn.Start()
	xc.LedgerKeeper = sn

	go xc.Speed.ShowLoop(xc.log)
	go xc.repostOfflineTx()
	return nil
}

//周期repost本地未上链的交易
func (xc *XChainCore) repostOfflineTx() {
	for txList := range xc.Utxovm.OfflineTxChan {
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
		msg, _ := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion1, xc.bcname, header.GetLogid(), xuper_p2p.XuperMessage_BATCHPOSTTX, msgInfo, xuper_p2p.XuperMessage_SUCCESS)

		filters := []p2p_base.FilterStrategy{p2p_base.DefaultStrategy}
		if xc.NeedCoreConnection() {
			filters = append(filters, p2p_base.CorePeersStrategy)
		}
		whiteList := xc.groupChain.GetAllowedPeersWithBcname(xc.bcname)
		opts := []p2p_base.MessageOption{
			p2p_base.WithFilters(filters),
			p2p_base.WithBcName(xc.bcname),
			p2p_base.WithCompress(xc.enableCompress),
			p2p_base.WithWhiteList(whiteList),
		}
		go xc.P2pSvr.SendMessage(context.Background(), msg, opts...) //p2p广播出去
	}
}

func (xc *XChainCore) ProcessSendBlock(in *pb.Block, hd *global.XContext) error {
	if xc.pruneOption.Switch && xc.pruneOption.Bcname == xc.bcname {
		return errors.New("the chain is dong ledger pruning")
	}
	return xc.SendBlock(in, hd)
}

/*
	当收到一个区块广播之后，若该区块刚好和本地账本tipID对接上，则此时直接触发LedgerKeeper追加
	若该区块并不是本地tipID的下一个，则主动触发同步GET_HEADERS
	若该区块为tipID高度以下的区块，直接忽略
*/
func (xc *XChainCore) SendBlock(in *pb.Block, hd *global.XContext) error {
	if xc.Ledger.ExistBlock(in.GetBlock().GetBlockid()) {
		xc.log.Debug("Block is exist", "logid", in.Header.Logid, "cost", hd.Timer.Print())
		return nil
	}
	tipId := global.F(xc.Ledger.GetMeta().GetTipBlockid())
	if bytes.Equal(xc.Utxovm.GetLatestBlockid(), in.GetBlock().GetPreHash()) {
		xc.log.Trace("Appending block in SendBlock", "time", time.Now().UnixNano(), "blockname", xc.bcname, "tipID", tipId)
		appendTask := NewLedgerTask(tipId, xc.Ledger.GetMeta().GetTrunkHeight(), Appending, NewLedgerTaskContext(
			&map[string]*SimpleBlock{
				global.F(in.GetBlock().GetBlockid()): &SimpleBlock{
					internalBlock: in.GetBlock(),
					header:        in.GetHeader(),
				}}, nil, hd))
		xc.LedgerKeeper.PutTask(appendTask)
		return nil
	}
	xc.log.Trace("sync blocks in SendBlock", "time", time.Now().UnixNano(), "blockname", xc.bcname, "tipID", tipId)
	syncTask := NewLedgerTask(tipId, xc.Ledger.GetMeta().GetTrunkHeight(), Syncing, NewLedgerTaskContext(nil,
		&[]string{in.GetHeader().GetFromNode()}, hd))
	xc.LedgerKeeper.PutTask(syncTask)
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

	if !bytes.Equal(ledgerLastID, utxovmLastID) {
		xc.log.Warn("ledger last blockid is not equal utxovm last id")
		err := xc.Utxovm.Walk(ledgerLastID, false)
		// if xc.failSkip = false, then keep logic, if not equal, retry
		if err != nil {
			if !xc.failSkip {
				xc.log.Error("Walk error at", "ledger blockid", global.F(ledgerLastID),
					"utxo blockid", global.F(utxovmLastID))
				return
			} else {
				err := xc.LedgerKeeper.DoTruncateTask(utxovmLastID)
				if err != nil {
					return
				}
			}
		}

		ledgerLastID = xc.Ledger.GetMeta().TipBlockid
		utxovmLastID = xc.Utxovm.GetLatestBlockid()
	}

	header := &pb.Header{Logid: global.Glogid()}

	// 打包块起始时间
	t := time.Now()
	// 挖矿前共识的预处理
	var curTerm, curBlockNum int64
	var targetBits int32
	qc := (*pb.QuorumCert)(nil)
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
					if qci, ok := data["quorum_cert"].(*pb.QuorumCert); ok {
						qc = qci
					}
				case consensus.ConsensusTypePow:
					xc.log.Trace("Minning pow ProcessBeforeMiner!")
					targetBits = data["targetBits"].(int32)
				case consensus.ConsensusTypeXpoa:
					xc.log.Trace("Minning xpoa ProcessBeforeMiner!", "quorum_cert", data["quorum_cert"])
					if qci, ok := data["quorum_cert"].(*pb.QuorumCert); ok {
						qc = qci
					}
				}
			}
		}
	} else {
		xc.log.Trace("Minning ProcessBeforeMiner not ok!")
		return
	}
	meta := xc.Ledger.GetMeta()
	accumulatedTxSize := 0
	txSizeTotalLimit, txSizeTotalLimitErr := xc.Utxovm.MaxTxSizePerBlock()
	if txSizeTotalLimitErr != nil {
		return
	}
	var freshBlock *pb.InternalBlock
	var freshBatch kvdb.Batch
	txs := []*pb.Transaction{}
	//1. 查询自动生成的交易
	vatList, err := xc.Utxovm.GetVATList(xc.Ledger.GetMeta().TrunkHeight+1, -1, t.UnixNano())
	minerTimer.Mark("GetAutogenTxs")
	if err != nil {
		xc.log.Warn("[Minning] fail to get triggered tx list", "logid", header.Logid)
		return
	}
	xc.log.Trace("[Minning] get vatList success", "vatList", vatList)
	txs = append(txs, vatList...)
	for _, vatTx := range txs {
		accumulatedTxSize += proto.Size(vatTx)
	}
	txsUnconf, err := xc.Utxovm.GetUnconfirmedTx(false)
	if err != nil {
		xc.log.Warn("[Minning] fail to get unconfirmedtx")
		return
	}
	for _, ucTx := range txsUnconf {
		accumulatedTxSize += proto.Size(ucTx)
		if accumulatedTxSize > txSizeTotalLimit {
			xc.log.Warn("already got enough tx to produce block", "acct", accumulatedTxSize, "limit", txSizeTotalLimit)
			break
		}
		txs = append(txs, ucTx)
	}
	fakeBlock, err := xc.Ledger.FormatFakeBlock(txs, xc.address, xc.privateKey,
		t.UnixNano(), curTerm, curBlockNum, xc.Utxovm.GetLatestBlockid(), xc.Utxovm.GetTotal(), xc.Ledger.GetMeta().TrunkHeight+1)
	if err != nil {
		xc.log.Warn("[Minning] format fake block error", "logid")
		return
	}
	//2. pre-execute the contract
	freshBatch = xc.Utxovm.NewBatch()
	if txs, _, err = xc.Utxovm.TxOfRunningContractGenerate(txs, fakeBlock, freshBatch, true); err != nil {
		if err.Error() != common.ErrContractExecutionTimeout.Error() {
			xc.log.Warn("PrePlay fake block failed", "error", err) //unexpected error
			return
		}
	}
	minerTimer.Mark("PrePlay")
	//3. 统一在最后插入矿工奖励
	blockAward := xc.Ledger.GenesisBlock.CalcAward(xc.Ledger.GetMeta().TrunkHeight + 1)
	awardtx, err := xc.Utxovm.GenerateAwardTx(xc.address, blockAward.String(), []byte{'1'})
	minerTimer.Mark("GenAwardTx")
	txs = append(txs, awardtx)
	freshBlock, err = xc.Ledger.FormatMinerBlock(txs, xc.address, xc.privateKey,
		t.UnixNano(), curTerm, curBlockNum, xc.Utxovm.GetLatestBlockid(), targetBits,
		xc.Utxovm.GetTotal(), qc, fakeBlock.FailedTxs, xc.Ledger.GetMeta().TrunkHeight+1)
	if err != nil {
		xc.log.Warn("[Minning] format block error", "logid", header.Logid, "err", err)
		return
	}
	minerTimer.Mark("Formatblock2")
	xc.log.Debug("[Minning] Start to ConfirmBlock", "logid", header.Logid)
	confirmStatus := xc.Ledger.ConfirmBlock(freshBlock, false)
	minerTimer.Mark("ConfirmBlock")
	if confirmStatus.Succ {
		if confirmStatus.Orphan {
			xc.log.Warn("[Minning] the mined blocked was attached to branch, no need to play")
			return
		}
		xc.log.Info("[Minning] ConfirmBlock Success", "logid", header.Logid, "Height", meta.TrunkHeight+1)
	} else {
		xc.log.Warn("[Minning] ConfirmBlock Fail", "logid", header.Logid, "confirm_status", confirmStatus)
		err := xc.Utxovm.Walk(xc.Utxovm.GetLatestBlockid(), false)
		if err != nil {
			xc.log.Warn("[Mining] failed to walk when confirming block has error", "err", err)
		}
		return
	}
	xc.mutex.Unlock() //后面放开锁
	lockHold = false
	xc.Utxovm.SetBlockGenEvent()
	defer xc.Utxovm.NotifyFinishBlockGen()
	err = xc.Utxovm.PlayForMiner(freshBlock.Blockid, freshBatch)
	if err != nil {
		xc.log.Warn("[Minning] utxo play error ", "logid", header.Logid, "error", err, "blockid", fmt.Sprintf("%x", freshBlock.Blockid))
		return
	}
	minerTimer.Mark("PlayForMiner")
	xc.con.ProcessConfirmBlock(freshBlock)
	minerTimer.Mark("ProcessConfirmBlock")
	xc.log.Debug("[Minning] Start to BroadCast", "logid", header.Logid)

	go func() {
		// 这里提出两种块传播模式：
		//  1. 一种是完全块广播模式(Full_BroadCast_Mode)，即直接广播原始块给所有相邻节点，
		//     适用于出块矿工在知道周围节点都不具备该块的情况下；
		//  2. 一种是问询式块广播模式(Interactive_BroadCast_Mode)，即先广播新块的头部给相邻节点，
		//     相邻节点在没有相同块的情况下通过GetBlock主动获取块数据。
		//  3. Mixed_BroadCast_Mode是指出块节点将新块用Full_BroadCast_Mode模式广播，
		//     其他节点使用Interactive_BroadCast_Mode
		// broadcast block in Full_BroadCast_Mode since it's the original miner
		block := &pb.Block{
			Header:  &pb.Header{FromNode: xc.P2pSvr.GetLocalUrl()},
			Bcname:  xc.bcname,
			Blockid: freshBlock.Blockid,
		}
		if xc.blockBroadcaseMode == 1 {
			// send block id in Interactive_BroadCast_Mode
			msgInfo, _ := proto.Marshal(block)
			msg, _ := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion1, xc.bcname, "", xuper_p2p.XuperMessage_NEW_BLOCKID, msgInfo, xuper_p2p.XuperMessage_NONE)
			filters := []p2p_base.FilterStrategy{p2p_base.DefaultStrategy}
			if xc.NeedCoreConnection() {
				filters = append(filters, p2p_base.CorePeersStrategy)
			}
			whiteList := xc.groupChain.GetAllowedPeersWithBcname(xc.bcname)
			opts := []p2p_base.MessageOption{
				p2p_base.WithFilters(filters),
				p2p_base.WithBcName(xc.bcname),
				p2p_base.WithWhiteList(whiteList),
			}
			xc.P2pSvr.SendMessage(context.Background(), msg, opts...)
		} else {
			// send full block in Full_BroadCast_Mode or Mixed_Broadcast_Mode
			block.Block = freshBlock
			msgInfo, _ := proto.Marshal(block)
			msg, _ := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion1, xc.bcname, "", xuper_p2p.XuperMessage_SENDBLOCK, msgInfo, xuper_p2p.XuperMessage_NONE)
			filters := []p2p_base.FilterStrategy{p2p_base.DefaultStrategy}
			if xc.NeedCoreConnection() {
				filters = append(filters, p2p_base.CorePeersStrategy)
			}
			whiteList := xc.groupChain.GetAllowedPeersWithBcname(xc.bcname)
			opts := []p2p_base.MessageOption{
				p2p_base.WithFilters(filters),
				p2p_base.WithBcName(xc.bcname),
				p2p_base.WithCompress(xc.enableCompress),
				p2p_base.WithWhiteList(whiteList),
			}
			xc.P2pSvr.SendMessage(context.Background(), msg, opts...)
		}
	}()

	minerTimer.Mark("BroadcastBlock")
	if xc.Utxovm.IsAsync() || xc.Utxovm.IsAsyncBlock() {
		xc.log.Warn("doMiner cost", "cost", minerTimer.Print(), "txCount", freshBlock.TxCount)
	} else {
		xc.log.Debug("doMiner cost", "cost", minerTimer.Print(), "txCount", freshBlock.TxCount)
	}
}

// Miner start to miner
func (xc *XChainCore) Miner() int {
	// 1 强制walk到最新状态
	ledgerLastID := xc.Ledger.GetMeta().TipBlockid
	utxovmLastID := xc.Utxovm.GetLatestBlockid()
	if !bytes.Equal(ledgerLastID, utxovmLastID) {
		xc.log.Warn("ledger last blockid is not equal utxovm last id")
		xc.Utxovm.Walk(ledgerLastID, false)
	}
	// 2 FAST_SYNC模式下需要回滚掉本地所有的未确认交易
	if xc.nodeMode == config.NodeModeFastSync {
		if _, _, err := xc.Utxovm.RollBackUnconfirmedTx(); err != nil {
			xc.log.Warn("FAST_SYNC mode RollBackUnconfirmedTx error", "error", err)
		}
	}
	// 3 开始同步
	xc.status = global.Normal
	xc.SyncBlocks()
	xc.dataInitReady()
	for {
		// 重要: 首次出块前一定要同步到最新的状态
		xc.log.Trace("Miner type of consensus", "type", xc.con.Type(xc.Ledger.GetMeta().TrunkHeight+1))
		// 账本裁剪入口
		if xc.pruneOption.Switch && xc.pruneOption.Bcname == xc.bcname {
			rawBlockid, err := hex.DecodeString(xc.pruneOption.TargetBlockid)
			if err != nil {
				return -1
			}
			err = xc.pruneLedger(rawBlockid)
			if err != nil {
				xc.log.Warn("pruning ledger failed", "err", err)
				return -1
			}
			xc.log.Trace("pruning ledger success")
			xc.pruneOption.Switch = false
			xc.SyncBlocks()
			// 裁剪账本可能需要时间，做完之后直接返回
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			return -1
		}
		b, s := xc.con.CompeteMaster(xc.Ledger.GetMeta().TrunkHeight + 1)
		xc.log.Debug("competemaster", "blockchain", xc.bcname, "master", b, "needSync", s, "compete height", xc.Ledger.GetMeta().TrunkHeight+1)
		xc.updateIsCoreMiner()
		if b {
			// todo 首次切换为矿工时SyncBlcok, Bug: 可能会导致第一次出块失败
			if s {
				xc.SyncBlocks()
			}
			xc.doMiner()
		}
		meta := xc.Ledger.GetMeta()
		xc.log.Info("Minner", "genesis", fmt.Sprintf("%x", meta.RootBlockid), "last", fmt.Sprintf("%x", meta.TipBlockid), "height", meta.TrunkHeight, "utxovm", fmt.Sprintf("%x", xc.Utxovm.GetLatestBlockid()))
		if xc.stopFlag {
			break
		}
	}
	return 0
}

// dataInitReady do some preparation work after blockchain data init ready
func (xc *XChainCore) dataInitReady() {
	eb := events.GetEventBus()
	miners := xc.con.GetCoreMiners()
	msg := &cons_base.MinersChangedEvent{
		BcName:        xc.bcname,
		CurrentMiners: miners,
		NextMiners:    miners,
	}
	em := &events.EventMessage{
		BcName:   xc.bcname,
		Type:     events.ProposerReady,
		Priority: 0,
		Sender:   xc,
		Message:  msg,
	}
	_, err := eb.FireEventAsync(em)
	if err != nil {
		xc.log.Warn("dataInitReady fire event failed", "error", err)
	}
}

func (xc *XChainCore) updateIsCoreMiner() {
	miners := xc.con.GetCoreMiners()
	for _, miner := range miners {
		if miner.Address == string(xc.address) {
			xc.isCoreMiner = true
			xc.log.Debug("updateIsCoreMiner", "bcname", xc.bcname, "isCoreMiner", xc.isCoreMiner)
			return
		}
	}
	xc.isCoreMiner = false
}

// Stop stop one xchain instance
func (xc *XChainCore) Stop() {
	xc.Utxovm.Close()
	xc.Ledger.Close()
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
	txid := in.GetTxid()
	if txid == nil {
		out.Header.Error = pb.XChainErrorEnum_CONNECT_REFUSE // 拒绝
		xc.log.Debug("refused a tx with the txid being nil", "logid", in.GetHeader().GetLogid())
		return out, false
	}
	txidStr := string(txid)
	if _, exist := xc.txidCache.Get(txidStr); exist {
		out.Header.Error = pb.XChainErrorEnum_TX_DUPLICATE_ERROR // tx重复
		xc.log.Debug("refused to accept a repeated transaction recently")
		return out, false
	}
	xc.txidCache.Set(txidStr, true, xc.txidCacheExpiredTime)
	// 对Tx进行的签名, 1 如果utxo属于用户，则走原来的验证逻辑 2 如果utxo属于账户，则走账户acl验证逻辑
	txValid, validErr := xc.Utxovm.VerifyTx(in.Tx)
	if !txValid {
		switch validErr {
		case utxo.ErrGasNotEnough:
			out.Header.Error = pb.XChainErrorEnum_GAS_NOT_ENOUGH_ERROR
		case utxo.ErrRWSetInvalid, utxo.ErrInvalidTxExt:
			out.Header.Error = pb.XChainErrorEnum_RWSET_INVALID_ERROR
		case utxo.ErrACLNotEnough:
			out.Header.Error = pb.XChainErrorEnum_RWACL_INVALID_ERROR
		case utxo.ErrVersionInvalid:
			out.Header.Error = pb.XChainErrorEnum_TX_VERSION_INVALID_ERROR
		case utxo.ErrInvalidSignature:
			out.Header.Error = pb.XChainErrorEnum_TX_SIGN_ERROR
		default:
			out.Header.Error = pb.XChainErrorEnum_TX_VERIFICATION_ERROR
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
			xc.txidCache.Delete(txidStr)
			xc.log.Info("utxo vm do tx error", "logid", in.Header.Logid, "error", err)
		}
		return out, false
	}
	xc.Speed.Add("PostTx")
	if xc.Utxovm.IsAsync() || xc.Utxovm.IsAsyncBlock() {
		return out, false //no need to repost tx immediately
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
func (xc *XChainCore) GetBlockChainStatus(in *pb.BCStatus, viewOption pb.ViewOption) *pb.BCStatus {
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
	if viewOption == pb.ViewOption_NONE || viewOption == pb.ViewOption_LEDGER || viewOption == pb.ViewOption_BRANCHINFO {
		out.Meta = meta
	}
	if viewOption == pb.ViewOption_NONE || viewOption == pb.ViewOption_UTXOINFO {
		utxoMeta := xc.Utxovm.GetMeta()
		out.UtxoMeta = utxoMeta
	}
	ib, err := xc.Ledger.QueryBlock(meta.TipBlockid)
	if err != nil {
		out.Header.Error = HandlerLedgerError(err)
		return out
	}
	if viewOption == pb.ViewOption_NONE {
		out.Block = ib
	}
	if viewOption == pb.ViewOption_NONE || viewOption == pb.ViewOption_BRANCHINFO {
		// fetch all branches info
		branchManager, branchErr := xc.Ledger.GetBranchInfo([]byte("0"), int64(0))
		if branchErr != nil {
			out.Header.Error = HandlerLedgerError(branchErr)
			return out
		}
		for _, branchID := range branchManager {
			out.BranchBlockid = append(out.BranchBlockid, fmt.Sprintf("%x", branchID))
		}
	}
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

func (xc *XChainCore) QueryContractStatData() (*pb.ContractStatDataResponse, error) {
	contractStatDataResponse := &pb.ContractStatDataResponse{
		Header: global.GHeader(),
		Bcname: xc.bcname,
	}

	if xc.Status() != global.Normal {
		return contractStatDataResponse, ErrNotReady
	}

	contractStatData, contractStatDataErr := xc.Utxovm.QueryContractStatData()
	if contractStatDataErr != nil {
		return contractStatDataResponse, contractStatDataErr
	}

	contractStatDataResponse.Data = contractStatData
	return contractStatDataResponse, nil
}

// QueryUtxoRecord get utxo record for an account
func (xc *XChainCore) QueryUtxoRecord(accountName string, displayCount int64) (*pb.UtxoRecordDetail, error) {
	defaultUtxoRecord := &pb.UtxoRecordDetail{Header: &pb.Header{}}
	if xc == nil {
		return defaultUtxoRecord, errors.New("xchaincore is nil")
	}
	if xc.Status() != global.Normal {
		return defaultUtxoRecord, ErrNotReady
	}
	utxoRecord, err := xc.Utxovm.QueryUtxoRecord(accountName, displayCount)
	if err != nil {
		return defaultUtxoRecord, err
	}

	return utxoRecord, nil
}

// QueryAccountContainAK get all accounts contain a specific address
func (xc *XChainCore) QueryAccountContainAK(address string) ([]string, error) {
	if xc == nil {
		return nil, errors.New("xchain core is nil")
	}
	if xc.Status() != global.Normal {
		return nil, ErrNotReady
	}
	accounts, err := xc.Utxovm.QueryAccountContainAK(address)
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

// QueryTxFromForbidden query if the tx has been forbidden
func (xc *XChainCore) QueryTxFromForbidden(txid []byte) bool {
	if xc.Status() != global.Normal {
		return false
	}
	exist, confirmed, _ := xc.Utxovm.QueryTxFromForbiddenWithConfirmed(txid)
	// only forbid exist && confirmed transaction
	if exist && confirmed {
		return true
	}
	return false
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

// GetBalanceDetail get balance that still be frozen from utxo
func (xc *XChainCore) GetBalanceDetail(addr string) (*pb.TokenFrozenDetails, error) {
	if xc.Status() != global.Normal {
		return nil, ErrNotReady
	}
	tokenDetails, err := xc.Utxovm.GetBalanceDetail(addr)
	if err != nil {
		return nil, err
	}

	tokenFrozenDetails := &pb.TokenFrozenDetails{
		Bcname: xc.bcname,
		Tfd:    tokenDetails,
	}

	return tokenFrozenDetails, nil
}

// GetConsType get consensus type for specific block chain
func (xc *XChainCore) GetConsType() string {
	return xc.con.Type(xc.Ledger.GetMeta().TrunkHeight + 1)
}

// GetDposCandidates get all candidates
func (xc *XChainCore) GetDposCandidates() ([]string, error) {
	candidates := []string{}
	it := xc.Utxovm.ScanWithPrefix([]byte(tdpos.GenCandidateNominatePrefix()))
	defer it.Release()
	for it.Next() {
		key := string(it.Key())
		addr, err := tdpos.ParseCandidateNominateKey(key)
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
	res := []string{}
	proposers := []*cons_base.CandidateInfo{}
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
	for _, proposer := range proposers {
		res = append(res, proposer.Address)
	}
	return res, nil
}

// GetConsStatus get current consensus status
func (xc *XChainCore) GetConsStatus() *cons_base.ConsensusStatus {
	return xc.con.GetStatus()
}

// GetNodeMode get node running mode, such as Normal mode, FastSync mode
func (xc *XChainCore) GetNodeMode() string {
	return xc.nodeMode
}

// PreExec get read/write set for smart contract could be run in parallel
func (xc *XChainCore) PreExec(req *pb.InvokeRPCRequest, hd *global.XContext) (*pb.InvokeResponse, error) {
	return xc.Utxovm.PreExec(req, hd)
}

// IsCoreMiner return true if current node is one of the current core miners
// Note that is could be a little delay since it updated at each CompeteMaster.
func (xc *XChainCore) IsCoreMiner() bool {
	return xc.isCoreMiner
}

// NeedCoreConnection return true if current node is one of the core miners
// and coreConnection configure to true. True means block and batch tx messages
// need to send to core peers using p2p core peer connections
func (xc *XChainCore) NeedCoreConnection() bool {
	return xc.isCoreMiner && xc.coreConnection
}

// GetBlockByHeight get block from ledger on trunk, by Block Height
func (xc *XChainCore) GetBlockByHeight(in *pb.BlockHeight) *pb.Block {
	out := &pb.Block{Header: in.Header}
	out.Header.Error = pb.XChainErrorEnum_SUCCESS
	out.Bcname = in.Bcname
	if xc.Status() != global.Normal {
		out.Header.Error = pb.XChainErrorEnum_CONNECT_REFUSE
		xc.log.Debug("refused a connection a function call GetBlock", "logid", in.Header.Logid)
		return out
	}
	ib, err := xc.Ledger.QueryBlockByHeight(in.Height)
	if err != nil {
		switch err {
		case ledger.ErrBlockNotExist:
			out.Header.Error = pb.XChainErrorEnum_SUCCESS
			out.Status = pb.Block_NOEXIST
			return out
		default:
			xc.log.Warn("getblock by height", "logid", in.Header.Logid, "error", err)
			out.Header.Error = pb.XChainErrorEnum_UNKNOW_ERROR
			return out
		}
	} else {
		out.Block = ib
		if ib.InTrunk {
			out.Status = pb.Block_TRUNK
		} else {
			out.Status = pb.Block_BRANCH
		}
	}
	return out
}

// GetAccountContractsStatus query account contracts
func (xc *XChainCore) GetAccountContractsStatus(account string, needContent bool) ([]*pb.ContractStatus, error) {
	res := []*pb.ContractStatus{}
	contracts, err := xc.Utxovm.GetAccountContracts(account)
	if err != nil {
		xc.log.Warn("GetAccountContractsStatus error", "error", err.Error())
		return nil, err
	}
	for _, v := range contracts {
		if !needContent {
			cs := &pb.ContractStatus{
				ContractName: v,
			}
			res = append(res, cs)
		} else {
			contractStatus, err := xc.Utxovm.GetContractStatus(v)
			if err != nil {
				return nil, err
			}
			res = append(res, contractStatus)
		}
	}

	return res, nil
}

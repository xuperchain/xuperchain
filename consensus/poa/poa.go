//Copyright 2019 Baidu, Inc.

package poa

import (
	"bytes"
	"errors"
	"fmt"
	log "github.com/xuperchain/log15"
	"github.com/xuperchain/xuperunion/consensus/poa/bft"
	"io/ioutil"
	"os"
	"strconv"
	"sync"
	"time"

	"encoding/hex"

	"github.com/xuperchain/xuperunion/common/config"
	cons_base "github.com/xuperchain/xuperunion/consensus/base"
	"github.com/xuperchain/xuperunion/consensus/common/chainedbft"
	bft_config "github.com/xuperchain/xuperunion/consensus/common/chainedbft/config"
	"github.com/xuperchain/xuperunion/contract"
	crypto_base "github.com/xuperchain/xuperunion/crypto/client/base"
	"github.com/xuperchain/xuperunion/global"
	"github.com/xuperchain/xuperunion/ledger"
	"github.com/xuperchain/xuperunion/p2pv2"
	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/utxo"
)

// Init init poa
func (poa *Poa) Init() {
	poa.config = PoaConfig{
		initProposer: make([]*cons_base.CandidateInfo, 0),
	}
	poa.isProduce = make(map[int64]bool)
	poa.revokeCache = new(sync.Map)
	poa.mutex = new(sync.RWMutex)
}

// Type return the type of Poa consensus
func (poa *Poa) Type() string {
	return TYPE
}

// Version return the version of Poa consensus
func (poa *Poa) Version() int64 {
	return poa.version
}

// Configure is the specific implementation of ConsensusInterface
func (poa *Poa) Configure(xlog log.Logger, cfg *config.NodeConfig, consCfg map[string]interface{},
	extParams map[string]interface{}) error {
	// xLog used for logging
	if xlog == nil {
		xlog = log.New("module", "consensus")
		xlog.SetHandler(log.StreamHandler(os.Stderr, log.LogfmtFormat()))
	}
	poa.log = xlog

	// load address from "/address"
	address, err := ioutil.ReadFile(cfg.Miner.Keypath + "/address")
	if err != nil {
		xlog.Warn("load address error", "path", cfg.Miner.Keypath+"/address")
		return err
	}
	poa.address = address

	// parse extParams including "crypto_client", "ledger", "utxovm", "bcname", "timestamp", "p2psvr", "height"
	switch extParams["crypto_client"].(type) {
	case crypto_base.CryptoClient:
		poa.cryptoClient = extParams["crypto_client"].(crypto_base.CryptoClient)
	default:
		errMsg := "invalid type of crypto_client"
		xlog.Warn(errMsg)
		return errors.New(errMsg)
	}

	switch extParams["ledger"].(type) {
	case *ledger.Ledger:
		poa.ledger = extParams["ledger"].(*ledger.Ledger)
	default:
		errMsg := "invalid type of ledger"
		xlog.Warn(errMsg)
		return errors.New(errMsg)
	}

	switch extParams["utxovm"].(type) {
	case *utxo.UtxoVM:
		poa.utxoVM = extParams["utxovm"].(*utxo.UtxoVM)
	default:
		errMsg := "invalid type of utxovm"
		xlog.Warn(errMsg)
		return errors.New(errMsg)
	}

	switch extParams["bcname"].(type) {
	case string:
		poa.bcname = extParams["bcname"].(string)
	default:
		errMsg := "invalid type of bcname"
		xlog.Warn(errMsg)
		return errors.New(errMsg)
	}

	switch extParams["timestamp"].(type) {
	case int64:
		poa.initTimestamp = extParams["timestamp"].(int64)
	default:
		errMsg := "invalid type of timestamp"
		xlog.Warn(errMsg)
		return errors.New(errMsg)
	}

	if p2psvr, ok := extParams["p2psvr"].(p2pv2.P2PServer); ok {
		poa.p2psvr = p2psvr
	}

	if height, ok := extParams["height"].(int64); ok {
		poa.height = height
	}

	// buildConfigs parses the configs in poa
	if err = poa.buildConfigs(xlog, nil, consCfg); err != nil {
		return err
	}

	if err = poa.initBFT(cfg); err != nil {
		xlog.Warn("init chained-bft failed!", "error", err)
		return err
	}

	poa.log.Trace("Configure", "Poa", poa)
	return nil
}

func (poa *Poa) buildConfigs(xlog log.Logger, cfg *config.NodeConfig, consCfg map[string]interface{}) error {
	// assemble consensus config including "period", "alternate_interval", "block_num", "account_name", "init_proposer", "version"
	if consCfg["period"] == nil {
		return errors.New("Parse Poa period error, can not be null")
	}

	if consCfg["alternate_interval"] == nil {
		return errors.New("Parse Poa alternate_interval error, can not be null")
	}

	if consCfg["block_num"] == nil {
		return errors.New("Parse Poa block_num error, can not be null")
	}

	if consCfg["account_name"] == nil {
		return errors.New("Parse Poa account_name error, can not be null")
	}

	if consCfg["init_proposer"] == nil {
		return errors.New("Parse Poa init_proposer error, can not be null")
	}

	if consCfg["version"] == nil {
		consCfg["version"] = "0"
	}

	// parse parameters including "period", "alternate_interval", "block_num", "account_name", "init_proposer", "version"
	period, err := strconv.ParseInt(consCfg["period"].(string), 10, 64)
	if err != nil {
		xlog.Warn("Parse Poa config period error", "error", err.Error())
		return err
	}
	poa.config.period = period * 1e6

	alternateInterval, err := strconv.ParseInt(consCfg["alternate_interval"].(string), 10, 64)
	if err != nil {
		xlog.Warn("Parse Poa config alternateInterval error", "error", err.Error())
		return err
	}
	if alternateInterval%period != 0 {
		xlog.Warn("Parse Poa config alternateInterval error", "error", "alternateInterval should be eliminated by period")
		return errors.New("alternateInterval should be eliminated by period")
	}
	poa.config.alternateInterval = alternateInterval * 1e6

	blockNum, err := strconv.ParseInt(consCfg["block_num"].(string), 10, 64)
	if err != nil {
		xlog.Warn("Parse Poa block_num period error", "error", err.Error())
		return err
	}
	poa.config.blockNum = blockNum

	poa.config.accountName = consCfg["account_name"].(string)


	// read config of need_neturl
	needNetURL := false
	if needNetURLVal, ok := consCfg["need_neturl"]; ok {
		needNetURL = needNetURLVal.(bool)
	}
	poa.config.needNetURL = needNetURL

	initProposer := consCfg["init_proposer"].([]interface{})
	xlog.Trace("initProposer", "initProposer", initProposer)

	for _, v := range initProposer {
		canInfo := &cons_base.CandidateInfo{}
		canInfo.Address = v.(string)
		poa.config.initProposer = append(poa.config.initProposer, canInfo)
		poa.proposerInfos = append(poa.proposerInfos, canInfo)
	}

	// if have init_proposer_neturl, this info can be used for core peers connection
	if _, ok := consCfg["init_proposer_neturl"]; ok {
		proposerNeturls := consCfg["init_proposer_neturl"].([]interface{})
		for idx, v := range proposerNeturls {
			poa.config.initProposer[idx].PeerAddr = v.(string)
			poa.proposerInfos[idx].PeerAddr = v.(string)
			poa.log.Debug("Poa proposer info", "index", idx, "proposer", poa.config.initProposer[idx])
		}
	} else {
		poa.log.Warn("Poa have no neturl info for proposers",
			"need_neturl", needNetURL)
		if needNetURL {
			return errors.New("config error, init_proposer_neturl could not be empty")
		}
	}
	poa.proposerNum = int64(len(poa.proposerInfos))

	version, err := strconv.ParseInt(consCfg["version"].(string), 10, 64)
	if err != nil {
		xlog.Warn("Parse Poa config version error", "error", err.Error())
		return err
	}
	poa.version = version

	// parse bft related config
	poa.config.enableBFT = false
	if bftConfData, ok := consCfg["bft_config"].(map[string]interface{}); ok {
		bftconf := bft_config.MakeConfig(bftConfData)
		// if bft_config is not empty, enable bft
		poa.config.enableBFT = true
		poa.config.bftConfig = bftconf
	}

	poa.log.Trace("Poa after config", "TTDpos.config", poa.config)
	return nil
}

// CompeteMaster is the specific implementation of ConsensusInterface
func (poa *Poa) CompeteMaster(height int64) (bool, bool) {
	time.Sleep(time.Duration(poa.config.alternateInterval))
	poa.mutex.RLock()
	defer poa.mutex.RUnlock()
	if string(poa.address) == poa.proposerInfos[poa.curPos].Address {
		poa.log.Trace("CompeteMaster now xterm infos", "term", poa.curTerm, "pos", poa.curPos, "blockPos", poa.curBlockNum,
			"master", true)
		return true, poa.needSync()
	}
	poa.log.Trace("CompeteMaster now xterm infos", "term", poa.curTerm, "pos", poa.curPos, "blockPos", poa.curBlockNum,
		"master", false)
	return false, false
}

func (poa *Poa) needSync() bool {
	meta := poa.ledger.GetMeta()
	if meta.TrunkHeight == 0 {
		return true
	}
	blockTip, err := poa.ledger.QueryBlock(meta.TipBlockid)
	if err != nil {
		return true
	}
	if string(blockTip.Proposer) == string(poa.address) {
		return false
	}
	return true
}

// CheckMinerMatch is the specific implementation of ConsensusInterface
func (poa *Poa) CheckMinerMatch(header *pb.Header, in *pb.InternalBlock) (bool, error) {
	// 1 验证块信息是否合法
	blkid, err := ledger.MakeBlockID(in)
	if err != nil {
		poa.log.Warn("CheckMinerMatch MakeBlockID error", "logid", header.Logid, "error", err)
		return false, nil
	}
	if !(bytes.Equal(blkid, in.Blockid)) {
		poa.log.Warn("CheckMinerMatch equal blockid error", "logid", header.Logid, "redo blockid", global.F(blkid),
			"get blockid", global.F(in.Blockid))
		return false, nil
	}

	k, err := poa.cryptoClient.GetEcdsaPublicKeyFromJSON(in.Pubkey)
	if err != nil {
		poa.log.Warn("CheckMinerMatch get ecdsa from block error", "logid", header.Logid, "error", err)
		return false, nil
	}
	chkResult, _ := poa.cryptoClient.VerifyAddressUsingPublicKey(string(in.Proposer), k)
	if chkResult == false {
		poa.log.Warn("CheckMinerMatch address is not match publickey", "logid", header.Logid)
		return false, nil
	}

	valid, err := poa.cryptoClient.VerifyECDSA(k, in.Sign, in.Blockid)
	if err != nil || !valid {
		poa.log.Warn("CheckMinerMatch VerifyECDSA error", "logid", header.Logid, "error", err)
		return false, nil
	}

	if poa.config.enableBFT && !poa.isFirstBlock(in.GetHeight()) {
		// if BFT enabled and it's not the first proposal
		// check whether previous block's QuorumCert is valid
		ok, err := poa.bftPaceMaker.GetChainedBFT().IsQuorumCertValidate(in.GetJustify())
		if err != nil || !ok {
			poa.log.Warn("CheckMinerMatch bft IsQuorumCertValidate failed", "logid", header.Logid, "error", err)
			return false, nil
		}
	}

	// 2 验证轮数信息
	preBlock, err := poa.ledger.QueryBlock(in.PreHash)
	if err != nil {
		poa.log.Warn("CheckMinerMatch failed, get preblock error")
		return false, nil
	}
	poa.log.Trace("CheckMinerMatch", "preBlock.CurTerm", preBlock.CurTerm, "in.CurTerm", in.CurTerm, " in.Proposer",
		string(in.Proposer), "blockid", fmt.Sprintf("%x", in.Blockid))
	term, pos, _ := poa.minerScheduling(in.Timestamp)
	if poa.isProposer(term, pos, in.Proposer) {
		// curTermProposerProduceNumCache is not thread safe, lock before use it.
		poa.mutex.Lock()
		defer poa.mutex.Unlock()
		// 当不是第一轮时需要和前面的
		if in.CurTerm != 1 {
			// 减少矿工50%概率恶意地输入时间
			if preBlock.CurTerm > term {
				poa.log.Warn("CheckMinerMatch failed, preBlock.CurTerm is bigger than this!")
				return false, nil
			}
			// 当系统切轮时初始化 curTermProposerProduceNum
			if preBlock.CurTerm < term || (poa.curTerm == term && poa.curTermProposerProduceNumCache == nil) {
				poa.curTermProposerProduceNumCache = make(map[int64]map[string]map[string]bool)
				poa.curTermProposerProduceNumCache[in.CurTerm] = make(map[string]map[string]bool)
			}
		}
		// 判断某个矿工是否恶意出块
		if poa.curTermProposerProduceNumCache != nil && poa.curTermProposerProduceNumCache[in.CurTerm] != nil {
			if _, ok := poa.curTermProposerProduceNumCache[in.CurTerm][string(in.Proposer)]; !ok {
				poa.curTermProposerProduceNumCache[in.CurTerm][string(in.Proposer)] = make(map[string]bool)
				poa.curTermProposerProduceNumCache[in.CurTerm][string(in.Proposer)][hex.EncodeToString(in.Blockid)] = true
			} else {
				if !poa.curTermProposerProduceNumCache[in.CurTerm][string(in.Proposer)][hex.EncodeToString(in.Blockid)] {
					poa.curTermProposerProduceNumCache[in.CurTerm][string(in.Proposer)][hex.EncodeToString(in.Blockid)] = true
				}
			}
			if int64(len(poa.curTermProposerProduceNumCache[in.CurTerm][string(in.Proposer)])) > poa.config.blockNum+1 {
				poa.log.Warn("CheckMinerMatch failed, proposer produce more than config blockNum!", "blockNum", len(poa.curTermProposerProduceNumCache[in.CurTerm][string(in.Proposer)]))
				return false, ErrProposeBlockMoreThanConfig
			}
		}
	} else {
		poa.log.Warn("CheckMinerMatch failed, received block shouldn't proposed!")
		return false, nil
	}
	return true, nil
}

// ProcessBeforeMiner is the specific implementation of ConsensusInterface
func (poa *Poa) ProcessBeforeMiner(timestamp int64) (map[string]interface{}, bool) {
	res := make(map[string]interface{})
	// check bft status
	if poa.config.enableBFT {
		// TODO: what if IsLastViewConfirmed failed in competeMaster, but succeed in ProcessBeforeMiner?
		if !poa.isFirstBlock(poa.ledger.GetMeta().GetTrunkHeight() + 1) {
			if ok, _ := poa.bftPaceMaker.IsLastViewConfirmed(); !ok {
				poa.log.Warn("ProcessBeforeMiner last block not confirmed, walk to previous block")
				lastBlockId := poa.ledger.GetMeta().GetTipBlockid()
				lastBlock, err := poa.ledger.QueryBlock(lastBlockId)
				if err != nil {
					poa.log.Warn("ProcessBeforeMiner tip block query failed", "error", err)
					return nil, false
				}
				err = poa.utxoVM.Walk(lastBlock.GetPreHash(), false)
				if err != nil {
					poa.log.Warn("ProcessBeforeMiner utxo walk failed", "error", err)
					return nil, false
				}
				err = poa.ledger.Truncate(poa.utxoVM.GetLatestBlockid())
				if err != nil {
					poa.log.Warn("ProcessBeforeMiner ledger truncate failed", "error", err)
					return nil, false
				}
			}
		}

		qc, err := poa.bftPaceMaker.CurrentQCHigh([]byte(""))
		if err != nil {
			return nil, false
		}
		res["quorum_cert"] = qc
	}

	res["type"] = TYPE
	res["curTerm"] = poa.curTerm
	res["curBlockNum"] = poa.curBlockNum
	poa.log.Trace("ProcessBeforeMiner", "res", res)
	return res, true
}

// ProcessConfirmBlock is the specific implementation of ConsensusInterface
func (poa *Poa) ProcessConfirmBlock(block *pb.InternalBlock) error {
	// send bft NewProposal if bft enable and it's the miner
	if poa.config.enableBFT && bytes.Compare(block.GetProposer(), poa.address) == 0 {
		blockData := &pb.Block{
			Bcname:  poa.bcname,
			Blockid: block.Blockid,
			Block:   block,
		}

		err := poa.bftPaceMaker.NextNewProposal(block.Blockid, blockData)
		if err != nil {
			poa.log.Warn("ProcessConfirmBlock: bft next proposal failed", "error", err)
			return err
		}
	}
	return nil
}

// InitCurrent is the specific implementation of ConsensusInterface
func (poa *Poa) InitCurrent(block *pb.InternalBlock) error {
	return nil
}

// Stop is the specific implementation of interface contract
func (poa *Poa) Stop() {
	if poa.config.enableBFT && poa.bftPaceMaker != nil {
		poa.bftPaceMaker.Stop()
	}
}

// ReadOutput is the specific implementation of interface contract
func (poa *Poa) ReadOutput(desc *contract.TxDesc) (contract.ContractOutputInterface, error) {
	return nil, nil
}

// GetCoreMiners get the information of core miners
func (poa *Poa) GetCoreMiners() []*cons_base.MinerInfo {
	var res []*cons_base.MinerInfo
	for _, proposer := range poa.proposerInfos {
		minerInfo := &cons_base.MinerInfo{
			Address:  proposer.Address,
			PeerInfo: proposer.PeerAddr,
		}
		res = append(res, minerInfo)
	}
	return res
}

// GetStatus get the current status of consensus
func (poa *Poa) GetStatus() *cons_base.ConsensusStatus {
	status := &cons_base.ConsensusStatus{
		Term:     poa.curTerm,
		BlockNum: poa.curBlockNum,
	}
	if int(poa.curPos) < 0 || poa.curPos >= poa.proposerNum {
		poa.log.Warn("current pos illegal", "pos", poa.curPos)
	} else {
		status.Proposer = poa.proposerInfos[int(poa.curPos)].Address
	}
	return status
}

func (poa *Poa) initBFT(cfg *config.NodeConfig) error {
	// BFT not enabled
	if !poa.config.enableBFT {
		return nil
	}

	// read keys
	pkpath := cfg.Miner.Keypath + "/public.key"
	pkJSON, err := ioutil.ReadFile(pkpath)
	if err != nil {
		poa.log.Warn("load private key error", "path", pkpath)
		return err
	}
	skpath := cfg.Miner.Keypath + "/private.key"
	skJSON, err := ioutil.ReadFile(skpath)
	if err != nil {
		poa.log.Warn("load private key error", "path", skpath)
		return err
	}
	sk, err := poa.cryptoClient.GetEcdsaPrivateKeyFromJSON(skJSON)
	if err != nil {
		poa.log.Warn("parse private key failed", "privateKey", skJSON)
		return err
	}

	// initialize bft
	bridge := bft.NewCbftBridge(poa.bcname, poa.ledger, poa.log, poa)
	qcNeeded := 3
	qc := make([]*pb.QuorumCert, qcNeeded)
	meta := poa.ledger.GetMeta()
	if meta.TrunkHeight != 0 {
		blockid := meta.TipBlockid
		for qcNeeded > 0 {
			qcNeeded--
			block, err := poa.ledger.QueryBlock(blockid)
			if err != nil {
				poa.log.Warn("initBFT: get block failed", "error", err, "blockid", string(blockid))
				return err
			}
			qc[qcNeeded] = block.GetJustify()
			blockid = block.GetPreHash()
			if blockid == nil {
				break
			}
		}
	}

	cbft, err := chainedbft.NewChainedBft(
		poa.log,
		poa.config.bftConfig,
		poa.bcname,
		string(poa.address),
		string(pkJSON),
		sk,
		poa.config.initProposer,
		bridge,
		poa.cryptoClient,
		poa.p2psvr,
		qc[2], qc[1], qc[0])

	if err != nil {
		poa.log.Warn("initBFT: create ChainedBft failed", "error", err)
		return err
	}

	paceMaker, err := bft.NewDPoSPaceMaker(poa.bcname, poa.height, meta.TrunkHeight,
		string(poa.address), cbft, poa.log, poa, poa.ledger)
	if err != nil {
		if err != nil {
			poa.log.Warn("initBFT: create DPoSPaceMaker failed", "error", err)
			return err
		}
	}
	poa.bftPaceMaker = paceMaker
	bridge.SetPaceMaker(paceMaker)
	return poa.bftPaceMaker.Start()
}

func (poa *Poa) isFirstBlock(BlockHeight int64) bool {
	consStartHeight := poa.height
	if consStartHeight == 0 {
		consStartHeight++
	}
	poa.log.Debug("isFirstBlock check", "consStartHeight", consStartHeight,
		"targetHeight", BlockHeight)
	return consStartHeight == BlockHeight
}

func (poa *Poa) minerScheduling(timestamp int64) (term int64, pos int64, blockPos int64) {
	return poa.curTerm, poa.curPos, poa.curBlockNum
}

func (poa *Poa) Run(desc *contract.TxDesc) error {
	return nil
}

func (poa *Poa) Rollback(desc *contract.TxDesc) error {
	return nil
}
// Finalize is the specific implementation of interface contract
func (poa *Poa) Finalize(blockid []byte) error {
	return nil
}

// SetContext is the specific implementation of interface contract
func (poa *Poa) SetContext(context *contract.TxContext) error {
	return nil
}

//Copyright 2019 Baidu, Inc.

package xpoa

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"strconv"
	"sync"
	"time"

	log "github.com/xuperchain/log15"
	"github.com/xuperchain/xuperchain/core/common"
	"github.com/xuperchain/xuperchain/core/common/config"
	"github.com/xuperchain/xuperchain/core/consensus/base"
	cons_base "github.com/xuperchain/xuperchain/core/consensus/base"
	bft "github.com/xuperchain/xuperchain/core/consensus/common/chainedbft"
	bft_config "github.com/xuperchain/xuperchain/core/consensus/common/chainedbft/config"
	crypto_base "github.com/xuperchain/xuperchain/core/crypto/client/base"
	"github.com/xuperchain/xuperchain/core/ledger"
	p2p_base "github.com/xuperchain/xuperchain/core/p2p/base"
	"github.com/xuperchain/xuperchain/core/pb"
	"github.com/xuperchain/xuperchain/core/utxo"
)

var (
	// ErrXlogIsEmpty xlog instance empty error
	ErrXlogIsEmpty = errors.New("Term publish proposer num less than config")
	// ErrUpdateValidates update validates error
	ErrUpdateValidates = errors.New("Update validates error")
	// ErrContractNotFound contract of validates not found
	ErrContractNotFound = errors.New("Xpoa validate contract not found")
)

// Type return the type of Poa consensus
func (xpoa *XPoa) Type() string {
	return TYPE
}

// Version return the version of Poa consensus
func (xpoa *XPoa) Version() int64 {
	return xpoa.xpoaConf.version
}

// Configure is the specific implementation of ConsensusInterface
func (xpoa *XPoa) Configure(xlog log.Logger, cfg *config.NodeConfig, consCfg map[string]interface{},
	extParams map[string]interface{}) error {
	if xlog == nil {
		return ErrXlogIsEmpty
	}
	xpoa.lg = xlog
	xpoa.mutex = new(sync.RWMutex)
	xpoa.isProduce = make(map[int64]bool)
	address, err := ioutil.ReadFile(cfg.Miner.Keypath + "/address")
	if err != nil {
		xpoa.lg.Warn("load address error", "path", cfg.Miner.Keypath+"/address")
		return err
	}
	xpoa.address = string(address)

	err = xpoa.buildExtParams(extParams)
	if err != nil {
		xpoa.lg.Warn("xpoa buildExtParams error", "error", err.Error())
		return err
	}

	err = xpoa.buildXPoaConfig(consCfg)
	if err != nil {
		xpoa.lg.Warn("xpoa buildXPoaConfig error", "error", err.Error())
		return err
	}

	cryptoClient, ok := extParams["crypto_client"].(crypto_base.CryptoClient)
	if !ok {
		return errors.New("invalid type of crypto_client")
	}

	err = xpoa.initBFT(cfg, cryptoClient)
	if err != nil {
		xpoa.lg.Warn("xpoa initBFT error", "error", err.Error())
		return err
	}

	xpoa.lg.Trace("Configure", "Poa", xpoa)
	return nil
}

func (xpoa *XPoa) buildExtParams(extParams map[string]interface{}) error {

	if ledger, ok := extParams["ledger"].(*ledger.Ledger); ok {
		xpoa.ledger = ledger
	} else {
		return errors.New("invalid type of ledger")
	}

	if utxovm, ok := extParams["utxovm"].(*utxo.UtxoVM); ok {
		xpoa.utxoVM = utxovm
	} else {
		return errors.New("invalid type of utxovm")
	}
	if bcname, ok := extParams["bcname"].(string); ok {
		xpoa.bcname = bcname
	} else {
		return errors.New("invalid type of bcname")
	}
	if p2psvr, ok := extParams["p2psvr"].(p2p_base.P2PServer); ok {
		xpoa.p2psvr = p2psvr
	} else {
		return errors.New("invalid type of p2psvr")
	}
	if height, ok := extParams["height"].(int64); ok {
		xpoa.startHeight = height
	} else {
		return errors.New("invalid type of heights")
	}
	if timestamp, ok := extParams["timestamp"].(int64); ok {
		xpoa.xpoaConf.initTimestamp = timestamp
	} else {
		return errors.New("invalid type of timestamp")
	}
	return nil
}

func (xpoa *XPoa) buildXPoaConfig(consCfg map[string]interface{}) error {
	if consCfg["version"] == nil {
		xpoa.xpoaConf.version = 0
	} else {
		version, err := strconv.ParseInt(consCfg["version"].(string), 10, 64)
		if err != nil {
			xpoa.lg.Warn("Parse XPoa config version error", "error", err.Error())
			return err
		}
		xpoa.xpoaConf.version = version
	}

	if consCfg["period"] == nil {
		return errors.New("Parse XPoa period error, can not be null")
	}
	period, err := strconv.ParseInt(consCfg["period"].(string), 10, 64)
	if err != nil {
		xpoa.lg.Warn("Parse XPoa config period error", "error", err.Error())
		return err
	}
	xpoa.xpoaConf.period = period * 1e6

	if consCfg["block_num"] == nil {
		return errors.New("Parse XPoa block_num error, can not be null")
	}
	blockNum, err := strconv.ParseInt(consCfg["block_num"].(string), 10, 64)
	if err != nil {
		xpoa.lg.Warn("Parse XPoa block_num error", "error", err.Error())
		return err
	}
	xpoa.xpoaConf.blockNum = blockNum

	if consCfg["contract_name"] == nil {
		return errors.New("Parse XPoa contract_name error, can not be null")
	}
	xpoa.xpoaConf.contractName = consCfg["contract_name"].(string)

	if consCfg["method_name"] == nil {
		return errors.New("Parse XPoa method_name error, can not be null")
	}
	xpoa.xpoaConf.methodName = consCfg["method_name"].(string)

	// init proposers
	xpoa.lg.Trace("Config init_proposer", "init_proposer", consCfg["init_proposer"])
	if consCfg["init_proposer"] == nil {
		return errors.New("Parse XPoa init_proposer error, can not be null")
	}
	if initProposers, ok := consCfg["init_proposer"].([]interface{}); ok {
		for idx := 0; idx < len(initProposers); idx++ {
			p, _ := initProposers[idx].(map[string]interface{})
			proposer := &cons_base.CandidateInfo{}
			if p["address"] == nil || p["neturl"] == nil {
				return errors.New("Parse XPoa init_proposer error, neturl and address can not be null")
			}
			proposer.Address = p["address"].(string)
			proposer.PeerAddr = p["neturl"].(string)
			xpoa.xpoaConf.initProposers = append(xpoa.xpoaConf.initProposers, proposer)
		}
	} else {
		return errors.New("The type of XPoa config init proposer error")
	}

	// bft enable
	xpoa.enableBFT = false
	if _, ok := consCfg["bft_config"].(map[string]interface{}); ok {
		// if bft_config is not empty, enable bft
		xpoa.enableBFT = true
		xpoa.xpoaConf.bftConfig = &bft_config.Config{}
	}
	return nil
}

func (xpoa *XPoa) initBFT(cfg *config.NodeConfig, cryptoClient crypto_base.CryptoClient) error {
	if !xpoa.enableBFT {
		xpoa.lg.Info("no need to init bft for haven't configed")
		return nil
	}

	// read keys
	pkpath := cfg.Miner.Keypath + "/public.key"
	pkJSON, err := ioutil.ReadFile(pkpath)
	if err != nil {
		xpoa.lg.Warn("load private key error", "path", pkpath)
		return err
	}
	skpath := cfg.Miner.Keypath + "/private.key"
	skJSON, err := ioutil.ReadFile(skpath)
	if err != nil {
		xpoa.lg.Warn("load private key error", "path", skpath)
		return err
	}
	sk, err := cryptoClient.GetEcdsaPrivateKeyFromJSON(skJSON)
	if err != nil {
		xpoa.lg.Warn("parse private key failed", "privateKey", skJSON)
		return err
	}

	// initialize bft
	bridge := bft.NewDefaultCbftBridge(xpoa.bcname, xpoa.ledger, xpoa.lg, xpoa)
	qcNeeded := 3
	qc := make([]*pb.QuorumCert, qcNeeded)
	meta := xpoa.ledger.GetMeta()
	if meta.TrunkHeight != 0 {
		blockid := meta.TipBlockid
		for qcNeeded > 0 {
			qcNeeded--
			block, err := xpoa.ledger.QueryBlock(blockid)
			if err != nil {
				xpoa.lg.Warn("initBFT: get block failed", "error", err, "blockid", string(blockid))
				return err
			}
			qc[qcNeeded] = block.GetJustify()
			blockid = block.GetPreHash()
			if blockid == nil {
				break
			}
		}
	}

	cbft, err := bft.NewChainedBft(
		xpoa.lg,
		xpoa.xpoaConf.bftConfig,
		xpoa.bcname,
		string(xpoa.address),
		string(pkJSON),
		sk,
		xpoa.proposerInfos,
		bridge,
		cryptoClient,
		xpoa.p2psvr,
		qc[2], qc[1], qc[0])

	if err != nil {
		xpoa.lg.Warn("initBFT: create ChainedBft failed", "error", err)
		return err
	}

	paceMaker, err := bft.NewDefaultPaceMaker(xpoa.bcname, xpoa.startHeight, meta.TrunkHeight,
		string(xpoa.address), cbft, xpoa.lg, xpoa, xpoa.ledger)
	if err != nil {
		if err != nil {
			xpoa.lg.Warn("initBFT: create DefaultPaceMaker failed", "error", err)
			return err
		}
	}

	xpoa.bftPaceMaker = paceMaker
	bridge.SetPaceMaker(paceMaker)
	return xpoa.bftPaceMaker.Start()
}

// CompeteMaster is the specific implementation of ConsensusInterface
func (xpoa *XPoa) CompeteMaster(height int64) (bool, bool) {
	if !xpoa.IsActive() {
		xpoa.lg.Info("XPoa CompeteMaster consensus instance not active", "state", xpoa.state)
		return false, false
	}
Again:
	t := time.Now()
	un := t.UnixNano()
	key := un / xpoa.xpoaConf.period
	sleep := xpoa.xpoaConf.period - un%xpoa.xpoaConf.period
	maxsleeptime := time.Millisecond * 10
	if sleep > int64(maxsleeptime) {
		sleep = int64(maxsleeptime)
	}
	v, ok := xpoa.isProduce[key]
	if !ok || v == false {
		xpoa.isProduce[key] = true
	} else {
		time.Sleep(time.Duration(sleep))
		goto Again
	}

	// update validates
	if _, err := xpoa.updateValidates(height); err != nil {
		xpoa.lg.Error("Xpoa update validates error", "error", err.Error())
		return false, false
	}

	// get current proposers info
	curTerm, curPos, curBlockPos := xpoa.minerScheduling(time.Now().UnixNano())

	// update view
	if err := xpoa.updateViews(height); err != nil {
		xpoa.lg.Error("Xpoa update views error", "error", err.Error())
		return false, false
	}

	// master check
	if xpoa.isProposer(curPos, xpoa.address) {
		xpoa.lg.Trace("Xpoa CompeteMaster now xterm infos", "term", curTerm, "pos", curPos, "blockPos", curBlockPos, "un2", time.Now().UnixNano(),
			"master", true, "height", height)
		s := xpoa.needSync()
		return true, s
	}

	xpoa.lg.Trace("Xpoa CompeteMaster now xterm infos", "term", curTerm, "pos", curPos, "blockPos", curBlockPos, "un2", time.Now().UnixNano(),
		"master", false, "height", height)
	return false, false
}

// getCurrentValidates return current validates from xmodel
// 注意：当查不到的时候或者一个候选人都没查到则默认取初始化的值
// TODO: zq needs to be optimized in future because
func (xpoa *XPoa) getCurrentValidates() ([]*cons_base.CandidateInfo, int64, int64, error) {
	contractRes, confirmedTime, confirmedHeight, err := xpoa.utxoVM.SystemCall(xpoa.xpoaConf.contractName, xpoa.xpoaConf.methodName, nil, true)
	if common.NormalizedKVError(err) == common.ErrKVNotFound {
		xpoa.lg.Warn("Xpoa getCurrentValidates not found")
		return xpoa.xpoaConf.initProposers, xpoa.xpoaConf.initTimestamp, 0, ErrContractNotFound
	}
	if err == utxo.ErrorNotConfirm {
		xpoa.lg.Warn("Xpoa getCurrentValidates not confirmed")
		return xpoa.proposerInfos, xpoa.termTimestamp, xpoa.termHeight, nil
	}
	if err != nil {
		xpoa.lg.Error("Xpoa getCurrentValidates error", "err", err.Error())
		return nil, 0, 0, err
	}

	candidateInfos := &cons_base.CandidateInfos{}
	if err = json.Unmarshal(contractRes, candidateInfos); err != nil {
		xpoa.lg.Warn("Xpoa getCurrentValidates Unmarshal error", "err", err.Error(), "contractRes", contractRes)
		return xpoa.xpoaConf.initProposers, xpoa.xpoaConf.initTimestamp, 0, nil
	}

	if len(candidateInfos.Proposers) == 0 {
		xpoa.lg.Warn("Xpoa getCurrentValidates len(proposers) is 0")
		return xpoa.xpoaConf.initProposers, xpoa.xpoaConf.initTimestamp, 0, nil
	}
	for i := range candidateInfos.Proposers {
		xpoa.lg.Trace("Xpoa getCurrentValidates res", "Proposer", candidateInfos.Proposers[i])
	}
	xpoa.lg.Trace("Xpoa getCurrentValidates res", "confirmedTime", confirmedTime, "confirmedHeight", confirmedHeight)
	return candidateInfos.Proposers, confirmedTime, confirmedHeight, nil
}

// updateValidates update validates
// param: initTime time of the latest changed
// param: curValidates validates of the latest changed
// return: bool refers whether validates has changed
func (xpoa *XPoa) updateValidates(curHeight int64) (bool, error) {
	curValidates, initTime, initHeight, err := xpoa.getCurrentValidates()
	if err != nil && err != ErrContractNotFound {
		xpoa.lg.Error("Xpoa updateValidates getCurrentValidates error", "error", err.Error())
		return false, err
	}

	if err != ErrContractNotFound && curHeight < initHeight+3 {
		xpoa.lg.Debug("Xpoa updateValidates no need to update", "initHeight", initHeight, "curHeight", curHeight)
		return true, nil
	}

	if len(xpoa.proposerInfos) == len(curValidates) {
		if base.CandidateInfoEqual(xpoa.proposerInfos, curValidates) {
			return true, nil
		}
	}
	err = xpoa.bftPaceMaker.UpdateValidatorSet(curValidates)
	if err != nil {
		return false, ErrUpdateValidates
	}
	xpoa.termHeight = initHeight
	xpoa.termTimestamp = initTime
	xpoa.proposerInfos = curValidates
	return true, nil
}

// minerScheduling return current term, pos, blockPos from last changed
func (xpoa *XPoa) minerScheduling(timestamp int64) (term int64, pos int64, blockPos int64) {
	if timestamp < xpoa.termTimestamp {
		return
	}
	// 每一轮的时间
	termTime := xpoa.xpoaConf.period * int64(len(xpoa.proposerInfos)) * xpoa.xpoaConf.blockNum
	// 每个矿工轮值时间
	posTime := xpoa.xpoaConf.period * xpoa.xpoaConf.blockNum
	term = (timestamp-xpoa.termTimestamp)/termTime + 1
	resTime := (timestamp - xpoa.termTimestamp) - (term-1)*termTime
	pos = resTime / posTime
	resTime = resTime - (resTime/posTime)*posTime
	blockPos = resTime/xpoa.xpoaConf.period + 1
	xpoa.lg.Trace("getTermPos", "timestamp", timestamp, "term", term, "pos", pos, "blockPos", blockPos)
	return
}

// updateViews update view
func (xpoa *XPoa) updateViews(viewNum int64) error {
	proposer, err := xpoa.getCurProposer()
	if err != nil {
		xpoa.lg.Error("updateViews getCurProposer error", "err", err.Error())
		return err
	}

	nextProposer, err := xpoa.getNextProposer()
	if err != nil {
		xpoa.lg.Error("updateViews getNextProposer error", "err", err.Error())
		return err
	}

	// old height might out-of-date, use current trunkHeight when NewView
	return xpoa.bftPaceMaker.NextNewView(viewNum, nextProposer, proposer)
}

// getCurProposer get current proposer
func (xpoa *XPoa) getCurProposer() (string, error) {
	_, pos, _ := xpoa.minerScheduling(time.Now().UnixNano())
	if int(pos) > len(xpoa.proposerInfos)-1 {
		return "", errors.New("xpoa proposer infos idx error")
	}
	return xpoa.proposerInfos[pos].Address, nil
}

// getNextProposer get next proposer
func (xpoa *XPoa) getNextProposer() (string, error) {
	_, pos, _ := xpoa.minerScheduling(time.Now().UnixNano())
	if int(pos) > len(xpoa.proposerInfos)-1 {
		return "", errors.New("xpoa proposer infos idx error")
	}
	if int(pos) == len(xpoa.proposerInfos)-1 {
		return xpoa.proposerInfos[0].Address, nil
	}
	return xpoa.proposerInfos[pos+1].Address, nil
}

// getProposerWithTime get proposer with timestamp
// 注意：这里的time需要是一个同步的时间戳
func (xpoa *XPoa) getProposerWithTime(timestamp, height int64) (string, error) {
	// update validates
	if _, err := xpoa.updateValidates(height); err != nil {
		xpoa.lg.Error("Xpoa getProposerWithTime update validates error", "error", err.Error())
		return "", err
	}
	_, pos, _ := xpoa.minerScheduling(timestamp)
	if int(pos) > len(xpoa.proposerInfos)-1 {
		xpoa.lg.Error("Xpoa getProposerWithTime minerScheduling error")
		return "", errors.New("Xpoa getProposerWithTime minerScheduling error")
	}
	return xpoa.proposerInfos[pos].Address, nil
}

// isProposer return whether the node is the proposer
func (xpoa *XPoa) isProposer(pos int64, address string) bool {
	if int(pos) > len(xpoa.proposerInfos)-1 {
		xpoa.lg.Warn("xpoa isProposer error for out of index")
		return false
	}
	return xpoa.proposerInfos[pos].Address == address
}

// needSync return whether
func (xpoa *XPoa) needSync() bool {
	meta := xpoa.ledger.GetMeta()
	if meta.TrunkHeight == 0 {
		return true
	}
	blockTip, err := xpoa.ledger.QueryBlock(meta.TipBlockid)
	if err != nil {
		return true
	}
	if string(blockTip.Proposer) == string(xpoa.address) {
		return false
	}
	return true
}

// CheckMinerMatch is the specific implementation of ConsensusInterface
func (xpoa *XPoa) CheckMinerMatch(header *pb.Header, in *pb.InternalBlock) (bool, error) {
	if !xpoa.IsActive() {
		xpoa.lg.Info("XPoa CheckMinerMatch consensus instance not active", "state", xpoa.state)
		return false, nil
	}
	// verify block
	if ok, err := xpoa.ledger.VerifyBlock(in, header.GetLogid()); !ok || err != nil {
		xpoa.lg.Info("XPoa CheckMinerMatch VerifyBlock not ok")
		return ok, err
	}

	// verify bft
	if xpoa.enableBFT && !xpoa.isFirstBlock(in.GetHeight()) {
		// if BFT enabled and it's not the first proposal
		// check whether previous block's QuorumCert is valid
		ok, err := xpoa.bftPaceMaker.GetChainedBFT().IsQuorumCertValidate(in.GetJustify())
		if err != nil || !ok {
			xpoa.lg.Warn("CheckMinerMatch bft IsQuorumCertValidate failed", "logid", header.Logid, "error", err)
			return false, nil
		}
	}

	// 验证矿工身份
	// get current validates from model
	proposer, err := xpoa.getProposerWithTime(in.GetTimestamp(), in.GetHeight())
	if err != nil {
		xpoa.lg.Warn("CheckMinerMatch getProposerWithTime error", "error", err.Error())
		return false, nil
	}
	return bytes.Equal(in.GetProposer(), []byte(proposer)), nil
}

// ProcessBeforeMiner is the specific implementation of ConsensusInterface
func (xpoa *XPoa) ProcessBeforeMiner(timestamp int64) (map[string]interface{}, bool) {
	if !xpoa.IsActive() {
		xpoa.lg.Info("XPoa ProcessBeforeMiner consensus intance not active", "state", xpoa.state)
		return nil, false
	}

	res := make(map[string]interface{})
	_, pos, blockPos := xpoa.minerScheduling(timestamp)
	if blockPos > xpoa.xpoaConf.blockNum || int(pos) >= len(xpoa.proposerInfos) {
		return res, false
	}
	if !xpoa.isProposer(pos, xpoa.address) {
		xpoa.lg.Warn("ProcessBeforeMiner prepare too long, omit!")
		return nil, false
	}
	res["type"] = TYPE
	if xpoa.enableBFT {
		if !xpoa.isFirstBlock(xpoa.ledger.GetMeta().GetTrunkHeight() + 1) {
			if ok, _ := xpoa.bftPaceMaker.IsLastViewConfirmed(); !ok {
				if len(xpoa.proposerInfos) == 1 {
					res["quorum_cert"] = nil
					return res, true
				}
				xpoa.lg.Warn("ProcessBeforeMiner last block not confirmed, walk to previous block")
				lastBlockid := xpoa.ledger.GetMeta().GetTipBlockid()
				lastBlock, err := xpoa.ledger.QueryBlock(lastBlockid)
				if err != nil {
					xpoa.lg.Warn("ProcessBeforeMiner tip block query failed", "error", err)
					return nil, false
				}
				err = xpoa.utxoVM.Walk(lastBlock.GetPreHash(), false)
				if err != nil {
					xpoa.lg.Warn("ProcessBeforeMiner utxo walk failed", "error", err)
					return nil, false
				}
				err = xpoa.ledger.Truncate(xpoa.utxoVM.GetLatestBlockid())
				if err != nil {
					xpoa.lg.Warn("ProcessBeforeMiner ledger truncate failed", "error", err)
					return nil, false
				}
			}
			qc, err := xpoa.bftPaceMaker.CurrentQCHigh([]byte(""))
			if err != nil || qc == nil {
				return nil, false
			}
			res["quorum_cert"] = qc
		}
	}
	xpoa.lg.Trace("ProcessBeforeMiner", "res", res)
	return res, true
}

// ProcessConfirmBlock is the specific implementation of ConsensusInterface
func (xpoa *XPoa) ProcessConfirmBlock(block *pb.InternalBlock) error {
	if !xpoa.IsActive() {
		xpoa.lg.Info("Xpoa ProcessConfirmBlock consensus instance not active", "state", xpoa.state)
		return errors.New("Xpoa ProcessConfirmBlock consensus not active")
	}
	// send bft NewProposal if bft enable and it's the miner
	if xpoa.enableBFT && bytes.Compare(block.GetProposer(), []byte(xpoa.address)) == 0 {
		blockData := &pb.Block{
			Bcname:  xpoa.bcname,
			Blockid: block.Blockid,
			Block:   block,
		}

		err := xpoa.bftPaceMaker.NextNewProposal(block.Blockid, blockData)
		if err != nil {
			xpoa.lg.Warn("ProcessConfirmBlock: bft next proposal failed", "error", err)
			return err
		}
	}
	// update bft smr status
	if xpoa.enableBFT && !xpoa.isInValidateSets() {
		xpoa.bftPaceMaker.UpdateSmrState(block.GetJustify())
	}
	return nil
}

// isInValidateSets return whether
func (xpoa *XPoa) isInValidateSets() bool {
	for idx := range xpoa.proposerInfos {
		if xpoa.address == xpoa.proposerInfos[idx].Address {
			return true
		}
	}
	return false
}

// InitCurrent is the specific implementation of ConsensusInterface
func (xpoa *XPoa) InitCurrent(block *pb.InternalBlock) error {
	if !xpoa.IsActive() {
		xpoa.lg.Info("Xpoa InitCurrent consensus instance not active", "state", xpoa.state)
		return errors.New("Xpoa InitCurrent consensus not active")
	}
	return nil
}

// Suspend will suspend the consensus instance while consensus update
func (xpoa *XPoa) Suspend() error {
	xpoa.mutex.Lock()
	xpoa.state = cons_base.SUSPEND
	if xpoa.enableBFT {
		xpoa.bftPaceMaker.GetChainedBFT().UnRegisterToNetwork()
	}
	xpoa.mutex.Unlock()
	return nil
}

// Activate will activate the consensus instance while consensus rollback
func (xpoa *XPoa) Activate() error {
	xpoa.mutex.Lock()
	xpoa.state = cons_base.RUNNING
	if xpoa.enableBFT {
		xpoa.bftPaceMaker.GetChainedBFT().RegisterToNetwork()
	}
	xpoa.mutex.Unlock()
	return nil
}

// IsActive return whether the cosensus is active
func (xpoa *XPoa) IsActive() bool {
	return xpoa.state == cons_base.RUNNING
}

// Stop is the specific implementation of interface contract
func (xpoa *XPoa) Stop() {
	if xpoa.bftPaceMaker != nil {
		err := xpoa.bftPaceMaker.Stop()
		if err != nil {
			xpoa.lg.Error("the xpoa stops unsuccessfully", "error", err)
		}
	}
}

// GetCoreMiners get the information of core miners
func (xpoa *XPoa) GetCoreMiners() []*cons_base.MinerInfo {
	var res []*cons_base.MinerInfo
	for _, proposer := range xpoa.proposerInfos {
		minerInfo := &cons_base.MinerInfo{
			Address:  proposer.Address,
			PeerInfo: proposer.PeerAddr,
		}
		res = append(res, minerInfo)
	}
	return res
}

// GetStatus get the current status of consensus
func (xpoa *XPoa) GetStatus() *cons_base.ConsensusStatus {
	if !xpoa.IsActive() {
		xpoa.lg.Info("XPoa GetStatus consensus instance not active", "state", xpoa.state)
		return nil
	}
	timestamp := time.Now().UnixNano()
	term, pos, blockPos := xpoa.minerScheduling(timestamp)
	proposers := xpoa.proposerInfos
	status := &cons_base.ConsensusStatus{
		Term:     term,
		BlockNum: blockPos,
	}
	if int(pos) < 0 || int(pos) >= len(proposers) {
		xpoa.lg.Warn("current pos illegal", "pos", pos)
	} else {
		status.Proposer = proposers[int(pos)].Address
	}
	return status
}

// isFirstBlock return whether is the first after validates has changed
func (xpoa *XPoa) isFirstBlock(BlockHeight int64) bool {
	xpoa.lg.Debug("isFirstBlock check", "consStartHeight", xpoa.startHeight+1,
		"targetHeight", BlockHeight)
	return xpoa.startHeight+1 == BlockHeight
}

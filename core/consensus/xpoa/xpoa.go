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

const XPOA_CONTRACT_NAME = "xpoa_validates"

var (
	// ErrXlogIsEmpty xlog instance empty error
	ErrXlogIsEmpty = errors.New("Term publish proposer num less than config")
	// ErrUpdateValidates update validates error
	ErrUpdateValidates = errors.New("Update validates error")
	// ErrContractNotFound contract of validates not found
	ErrContractNotFound = errors.New("Xpoa validate contract not found")
	// ErrNotConfirmed validates not confirmed
	ErrNotConfirmed = errors.New("Xpoa validate contract not confirmed")
)

var curHeight = int64(0)

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
	xpoa.effectiveDelay = 1
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
		version, ok := consCfg["version"].(string)
		if !ok {
			return errors.New("invalid type of version")
		}
		versionInt, err := strconv.ParseInt(version, 10, 64)
		if err != nil {
			xpoa.lg.Warn("Parse XPoa config version error", "error", err.Error())
			return err
		}
		xpoa.xpoaConf.version = versionInt
	}

	if consCfg["period"] == nil {
		return errors.New("Parse XPoa period error, can not be null")
	}

	period, ok := consCfg["period"].(string)
	if !ok {
		return errors.New("invalid type of period")
	}
	periodInt, err := strconv.ParseInt(period, 10, 64)
	if err != nil {
		xpoa.lg.Warn("Parse XPoa config period error", "error", err.Error())
		return err
	}
	xpoa.xpoaConf.period = periodInt * 1e6

	if consCfg["block_num"] == nil {
		return errors.New("Parse XPoa block_num error, can not be null")
	}
	blockNum, ok := consCfg["block_num"].(string)
	if !ok {
		return errors.New("invalid type of period")
	}
	blockNumInt, err := strconv.ParseInt(blockNum, 10, 64)
	if err != nil {
		xpoa.lg.Warn("Parse XPoa block_num error", "error", err.Error())
		return err
	}
	xpoa.xpoaConf.blockNum = blockNumInt

	if consCfg["contract_name"] == nil {
		return errors.New("Parse XPoa contract_name error, can not be null")
	}
	contractName, ok := consCfg["contract_name"].(string)
	if !ok {
		return errors.New("invalid type of contract_name")
	}
	xpoa.xpoaConf.contractName = contractName

	if consCfg["method_name"] == nil {
		return errors.New("Parse XPoa method_name error, can not be null")
	}
	methodName, ok := consCfg["method_name"].(string)
	if !ok {
		return errors.New("invalid type of method_name")
	}
	xpoa.xpoaConf.methodName = methodName

	// init proposers
	xpoa.lg.Debug("Config init_proposer", "init_proposer", consCfg["init_proposer"])
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
		qc[2], qc[1], qc[0],
		xpoa.effectiveDelay)

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

	xpoa.lg.Info("Compete Master", "height", height)
	xpoa.mutex.Lock()
	defer xpoa.mutex.Unlock()
	curHeight = height

	// update validates
	preBlockId, err := xpoa.ledger.QueryBlockByHeight(height - 1)
	if err != nil {
		xpoa.lg.Error("xpoa.getCurrentValidates", "getBlock", err)
		return false, false
	}
	xpoa.lg.Info("Miner Propcess update validates")
	if _, err := xpoa.updateValidates(preBlockId.GetBlockid()); err != nil {
		xpoa.lg.Error("Xpoa update validates error", "error", err.Error())
		return false, false
	}

	candidate, err := xpoa.getProposer(time.Now().UnixNano(), xpoa.proposerInfos)
	preproposer := string(preBlockId.GetProposer())
	if err = xpoa.bftPaceMaker.NextNewView(height, candidate.Address, preproposer); err != nil {
		return false, false
	}

	if candidate.Address == xpoa.address {
		xpoa.lg.Trace("Xpoa CompeteMaster now xterm infos", "master", true, "height", height)
		s := xpoa.needSync()
		return true, s
	}

	xpoa.lg.Trace("Xpoa CompeteMaster now xterm infos", "master", false, "height", height)
	return false, false
}

// getValidatesByBlockId 根据当前输入blockid，用快照的方式在xmodel中寻找<=当前blockid的最新的候选人值，若无则使用xuper.json中指定的初始值
func (xpoa *XPoa) getValidatesByBlockId(blockId []byte) ([]*cons_base.CandidateInfo, bool, error) {
	reader, err := xpoa.utxoVM.GetSnapShotWithBlock(blockId)
	if err != nil {
		xpoa.lg.Error("Xpoa updateValidates getCurrentValidates error", "CreateSnapshot err:", err)
		return nil, false, err
	}
	contractRes, err := xpoa.utxoVM.SystemCall(reader, xpoa.xpoaConf.contractName, xpoa.xpoaConf.methodName, nil)
	if err == ErrNotConfirmed {
		xpoa.lg.Error("Xpoa updateValidates getCurrentValidates not confirmed no need to update")
		return nil, true, nil
	}
	if err != nil && common.NormalizedKVError(err) != common.ErrKVNotFound {
		xpoa.lg.Error("Xpoa getCurrentValidates error", "err", err)
		return nil, false, err
	}
	if common.NormalizedKVError(err) == common.ErrKVNotFound {
		xpoa.lg.Warn("Xpoa getCurrentValidates not found")
		return xpoa.xpoaConf.initProposers, true, nil
	}
	candidateInfos := &cons_base.CandidateInfos{}
	if err = json.Unmarshal(contractRes, candidateInfos); err != nil {
		xpoa.lg.Warn("Xpoa getCurrentValidates Unmarshal error", "err", err.Error(), "contractRes", string(contractRes))
		return xpoa.xpoaConf.initProposers, true, nil
	}
	if len(candidateInfos.Proposers) == 0 {
		xpoa.lg.Warn("Xpoa getCurrentValidates len(proposers) is 0")
		return xpoa.xpoaConf.initProposers, true, nil
	}
	return candidateInfos.Proposers, true, nil
}

// updateValidates update validates
// param: initTime time of the latest changed
// param: curValidates validates of the latest changed
// return: bool refers whether validates has changed
func (xpoa *XPoa) updateValidates(blockId []byte) (bool, error) {
	validates, ok, err := xpoa.getValidatesByBlockId(blockId)
	if err != nil {
		return ok, err
	}
	if validates != nil && !base.CandidateInfoEqual(xpoa.proposerInfos, validates) {
		err := xpoa.bftPaceMaker.UpdateValidatorSet(validates)
		if err != nil {
			return false, ErrUpdateValidates
		}
		xpoa.proposerInfos = validates
		xpoa.lg.Debug("Xpoa updateValidates Successfully", "base on", blockId)
		for _, v := range xpoa.proposerInfos {
			xpoa.lg.Debug("updateValidates", "curValidates", v.Address)
		}
	}
	return true, nil
}

// minerScheduling return current term, pos, blockPos from last changed
func (xpoa *XPoa) minerScheduling(timestamp int64, length int64) (term int64, pos int64, blockPos int64) {
	// 每一轮的时间
	termTime := xpoa.xpoaConf.period * length * xpoa.xpoaConf.blockNum
	// 每个矿工轮值时间
	posTime := xpoa.xpoaConf.period * xpoa.xpoaConf.blockNum
	term = (timestamp)/termTime + 1
	//10640483 180000
	resTime := timestamp - (term-1)*termTime
	pos = resTime / posTime
	resTime = resTime - (resTime/posTime)*posTime
	blockPos = resTime/xpoa.xpoaConf.period + 1
	xpoa.lg.Trace("getTermPos", "timestamp", timestamp, "term", term, "pos", pos, "blockPos", blockPos)
	return
}

// getProposer 根据时间与当前候选人集合计算proposer
func (xpoa *XPoa) getProposer(nextTime int64, proposerInfos []*cons_base.CandidateInfo) (*cons_base.CandidateInfo, error) {
	term, pos, blockPos := xpoa.minerScheduling(nextTime, int64(len(proposerInfos)))
	if int(pos) > len(proposerInfos)-1 {
		return nil, errors.New("xpoa proposer infos idx error")
	}
	xpoa.lg.Trace("Xpoa getProposer", "term", term, "pos", pos, "blockPos", blockPos, "time", nextTime)
	return proposerInfos[pos], nil
}

// getProposerWithTime get proposer with timestamp
// 注意：这里的time需要是一个同步的时间戳
func (xpoa *XPoa) getProposerWithTime(timestamp int64, blockId []byte) (string, error) {
	xpoa.lg.Info("ConfirmBlock Propcess update validates")
	if _, err := xpoa.updateValidates(blockId); err != nil {
		xpoa.lg.Error("Xpoa getProposerWithTime update validates error", "error", err.Error())
		return "", err
	}
	candidate, err := xpoa.getProposer(timestamp, xpoa.proposerInfos)
	if err != nil {
		xpoa.lg.Error("Xpoa getProposerWithTime minerScheduling error")
		return "", err
	}
	return candidate.Address, nil
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
	if in.GetHeight() <= xpoa.ledger.GetMeta().GetTrunkHeight()-3 {
		xpoa.lg.Warn("refuse short chain of blocks", "remote", in.GetHeight(), "local", xpoa.ledger.GetMeta().GetTrunkHeight())
		return false, nil
	}

	if !xpoa.IsActive() {
		xpoa.lg.Info("XPoa CheckMinerMatch consensus instance not active", "state", xpoa.state)
		return false, nil
	}
	// verify block
	if ok, err := xpoa.ledger.VerifyBlock(in, header.GetLogid()); !ok || err != nil {
		xpoa.lg.Info("XPoa CheckMinerMatch VerifyBlock not ok")
		return ok, err
	}

	// 验证矿工身份
	// get current validates from model
	proposer, err := xpoa.getProposerWithTime(in.GetTimestamp(), in.GetPreHash())
	if err != nil {
		xpoa.lg.Warn("CheckMinerMatch getProposerWithTime error", "error", err.Error())
		return false, nil
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
	return bytes.Equal(in.GetProposer(), []byte(proposer)), nil
}

// ProcessBeforeMiner is the specific implementation of ConsensusInterface
func (xpoa *XPoa) ProcessBeforeMiner(timestamp int64) (map[string]interface{}, bool) {
	if !xpoa.IsActive() {
		xpoa.lg.Info("XPoa ProcessBeforeMiner consensus intance not active", "state", xpoa.state)
		return nil, false
	}

	if curHeight != xpoa.ledger.GetMeta().GetTrunkHeight()+1 {
		xpoa.lg.Warn("ProcessBeforeMiner error", "curHeight", curHeight, "xpoa.ledger.GetMeta().GetTrunkHeight()", xpoa.ledger.GetMeta().GetTrunkHeight())
		return nil, false
	}

	res := make(map[string]interface{})
	_, pos, blockPos := xpoa.minerScheduling(timestamp, int64(len(xpoa.proposerInfos)))
	if blockPos > xpoa.xpoaConf.blockNum || int(pos) >= len(xpoa.proposerInfos) {
		return res, false
	}
	if !xpoa.isProposer(pos, xpoa.address) {
		xpoa.lg.Warn("ProcessBeforeMiner prepare too long, omit!")
		return nil, false
	}
	res["type"] = TYPE
	if xpoa.enableBFT {
		height := xpoa.ledger.GetMeta().GetTrunkHeight() + 1
		if !xpoa.isFirstBlock(height) {
			if ok, _ := xpoa.bftPaceMaker.IsLastViewConfirmed(); !ok {
				// 若view number未更新则先暂停
				if xpoa.bftPaceMaker.CheckViewNumer(height) {
					xpoa.lg.Warn("Haven't received preLeader's NextViewMsg, hold first.")
					return nil, false
				}
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

		// 如果是当前矿工，检测到下一轮需变更validates，且下一轮proposer并不在节点列表中，此时需在广播列表中新加入节点
		validates := xpoa.proposerInfos
		nextValidates, _, err := xpoa.getValidatesByBlockId(block.GetBlockid())
		if err == nil {
			nextTime := time.Now().UnixNano() + xpoa.xpoaConf.period
			nextProposer, err := xpoa.getProposer(nextTime, nextValidates)
			xpoa.lg.Warn("Cal nextProposer:", "proposer", nextProposer)
			if err == nil && !xpoa.isInValidateSets(nextProposer.Address) {
				// 更新发送节点
				xpoa.lg.Info("Send Proposal to new Validates")
				nextValidates := append(xpoa.proposerInfos, nextProposer)
				validates = nextValidates
			}
		}

		err = xpoa.bftPaceMaker.NextNewProposal(block.Blockid, blockData, validates)
		if err != nil {
			xpoa.lg.Warn("ProcessConfirmBlock: bft next proposal failed", "error", err)
			return err
		}
		xpoa.lg.Info("Now Confirm finish", "ledger height", xpoa.ledger.GetMeta().TrunkHeight, "viewNum", xpoa.bftPaceMaker.CurrentView())
		return nil
	}
	// update bft smr status
	if xpoa.enableBFT {
		xpoa.bftPaceMaker.UpdateSmrState(block.GetJustify())
	}
	xpoa.lg.Debug("Now Confirm finish", "ledger height", xpoa.ledger.GetMeta().TrunkHeight, "viewNum", xpoa.bftPaceMaker.CurrentView())
	return nil
}

// isInValidateSets return whether
func (xpoa *XPoa) isInValidateSets(address string) bool {
	for _, proposerInfo := range xpoa.proposerInfos {
		if address == proposerInfo.Address {
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
	term, pos, blockPos := xpoa.minerScheduling(timestamp, int64(len(xpoa.proposerInfos)))
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

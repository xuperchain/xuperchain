//Copyright 2019 Baidu, Inc.

package poa

import (
	"bytes"
	"encoding/json"
	"errors"
	log "github.com/xuperchain/log15"
	"github.com/xuperchain/xuperunion/common/config"
	cons_base "github.com/xuperchain/xuperunion/consensus/base"
	"github.com/xuperchain/xuperunion/consensus/common/chainedbft"
	bft_config "github.com/xuperchain/xuperunion/consensus/common/chainedbft/config"
	"github.com/xuperchain/xuperunion/consensus/poa/bft"
	"github.com/xuperchain/xuperunion/contract"
	crypto_base "github.com/xuperchain/xuperunion/crypto/client/base"
	"github.com/xuperchain/xuperunion/global"
	"github.com/xuperchain/xuperunion/ledger"
	"github.com/xuperchain/xuperunion/p2pv2"
	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/utxo"
	"io/ioutil"
	"os"
	"strconv"
	"sync"
	"time"
)

// Init init poa
func (poa *Poa) Init() {
	poa.config = Config{
		initProposer: make([]*cons_base.CandidateInfo, 0),
	}
	poa.curTerm = 1
	poa.curPos = 0
	poa.curBlockNum = 0
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
	paramsNeed := []string{"period", "alternate_interval", "block_num", "account_name", "init_proposer", "init_proposer_neturl"}
	for _, param := range paramsNeed {
		if consCfg[param] == nil {
			return errors.New("parse Poa " + param + " error, can not be null")
		}
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

	poa.config.blockNum, err = strconv.ParseInt(consCfg["block_num"].(string), 10, 64)
	if err != nil {
		xlog.Warn("Parse Poa block_num period error", "error", err.Error())
		return err
	}
	poa.config.accountName = consCfg["account_name"].(string)
	poa.accountName = poa.config.accountName

	if proposers, err := poa.getProposersFromACL(); err == nil {
		// if the proposers are already in the acl of account, get it from chain.
		poa.proposerInfos = proposers
	} else {
		// otherwise got it from the initial configuration
		initProposers := consCfg["init_proposer"].([]interface{})

		xlog.Trace("initProposers", "initProposers", initProposers)
		initProposerUrls := consCfg["init_proposer_neturl"].([]interface{})
		xlog.Trace("initProposerUrls", "initProposerUrls", initProposerUrls)
		if len(initProposers) != len(initProposerUrls) {
			return errors.New("the lengths of initProposers and initProposerUrls should be equal")
		}
		poa.proposerNum = int64(len(initProposers))

		for idx := int64(0); idx < poa.proposerNum; idx++ {
			canInfo := &cons_base.CandidateInfo{}
			canInfo.Address = initProposers[idx].(string)
			canInfo.PeerAddr = initProposerUrls[idx].(string)
			poa.config.initProposer = append(poa.config.initProposer, canInfo)
			poa.log.Debug("Poa proposer info", "index", idx, "proposer", poa.config.initProposer[idx])
		}
		poa.proposerInfos = poa.config.initProposer
	}

	version, err := strconv.ParseInt(consCfg["version"].(string), 10, 64)
	if err != nil {
		xlog.Warn("Parse Poa config version error", "error", err.Error())
		return err
	}
	poa.version = version

	// enable bft
	poa.config.bftConfig = bft_config.MakeConfig(make(map[string]interface{}))

	poa.log.Trace("Poa after config", "Poa.config", poa.config)
	return nil
}

// CompeteMaster is the specific implementation of ConsensusInterface
func (poa *Poa) CompeteMaster(height int64) (bool, bool) {
	poa.mutex.RLock()
	defer poa.mutex.RUnlock()
	if string(poa.address) == poa.proposerInfos[poa.curPos].Address {
		poa.log.Trace("CompeteMaster now xterm infos", "term", poa.curTerm, "pos", poa.curPos, "blockPos", poa.curBlockNum,
			"master", true)
		return true, poa.needSync()
	} else {
		time.Sleep(time.Duration(poa.config.alternateInterval))
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

	if !poa.isFirstBlock(in.GetHeight()) {
		// if BFT enabled and it's not the first proposal
		// check whether previous block's QuorumCert is valid
		ok, err := poa.bftPaceMaker.GetChainedBFT().IsQuorumCertValidate(in.GetJustify())
		if err != nil || !ok {
			poa.log.Warn("CheckMinerMatch bft IsQuorumCertValidate failed", "logid", header.Logid, "error", err)
			return false, nil
		}
	}

	// 2 验证轮数信息
	if !poa.isProposer(poa.curTerm, poa.curPos, in.Proposer) {
		poa.log.Warn("CheckMinerMatch failed, received block shouldn't proposed!")
		return false, nil
	}
	return true, nil
}

// ProcessBeforeMiner is the specific implementation of ConsensusInterface
func (poa *Poa) ProcessBeforeMiner(timestamp int64) (map[string]interface{}, bool) {
	res := make(map[string]interface{})
	// check bft status
	// TODO: what if IsLastViewConfirmed failed in competeMaster, but succeed in ProcessBeforeMiner?
	if !poa.isFirstBlock(poa.ledger.GetMeta().GetTrunkHeight() + 1) {
		if ok, _ := poa.bftPaceMaker.IsLastViewConfirmed(); !ok {
			poa.log.Warn("ProcessBeforeMiner last block not confirmed, walk to previous block")
			lastBlockID := poa.ledger.GetMeta().GetTipBlockid()
			lastBlock, err := poa.ledger.QueryBlock(lastBlockID)
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
	res["type"] = TYPE
	res["curTerm"] = poa.curTerm
	res["curBlockNum"] = poa.curBlockNum
	poa.log.Trace("ProcessBeforeMiner", "res", res)
	return res, true
}

// ProcessConfirmBlock is the specific implementation of ConsensusInterface
func (poa *Poa) ProcessConfirmBlock(block *pb.InternalBlock) error {
	// send bft NewProposal if bft enable and it's the miner
	if bytes.Compare(block.GetProposer(), poa.address) == 0 {
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
	// increase the blockNum and change the candidateInfo when new term comes.
	poa.curBlockNum++
	if poa.curBlockNum >= poa.config.blockNum {
		poa.curBlockNum = 0
		poa.curPos++
		if poa.curPos >= poa.proposerNum {
			poa.curTerm++
			poa.curPos = 0

			poa.log.Debug("the accountName of poa", "account name", poa.accountName)
		}
	}
	time.Sleep(time.Duration(poa.config.alternateInterval))
	poa.log.Debug("current pos", "term", poa.curTerm, "pos", poa.curPos, "blockNum", poa.curBlockNum)
	return nil
}

// InitCurrent is the specific implementation of ConsensusInterface
func (poa *Poa) InitCurrent(block *pb.InternalBlock) error {
	return nil
}

// Stop is the specific implementation of interface contract
func (poa *Poa) Stop() {
	if poa.bftPaceMaker != nil {
		err := poa.bftPaceMaker.Stop()
		if err != nil {
			poa.log.Error("the poa stops unsuccessfully", "error", err)
		}
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

	paceMaker, err := bft.NewPoaPaceMaker(poa.bcname, poa.height, meta.TrunkHeight,
		string(poa.address), cbft, poa.log, poa, poa.ledger)
	if err != nil {
		poa.log.Warn("initBFT: create PoaPaceMaker failed", "error", err)
		return err
	}
	poa.bftPaceMaker = paceMaker
	bridge.SetPaceMaker(paceMaker)
	return poa.bftPaceMaker.Start()
}

func (poa *Poa) isFirstBlock(BlockHeight int64) bool {
	consStartHeight := poa.height
	consStartHeight++
	poa.log.Debug("isFirstBlock check", "consStartHeight", consStartHeight,
		"targetHeight", BlockHeight)
	return poa.height+1 == BlockHeight
}

func (poa *Poa) getProposersFromACL() ([]*cons_base.CandidateInfo, error) {
	acl, confirmed, err := poa.utxoVM.QueryAccountACLWithConfirmed(poa.accountName)
	if err != nil {
		//poa.log.Error(err.Error())
		return nil, err
	}
	if acl == nil || !confirmed {
		poa.log.Warn("no acl in current account", "acl", acl, "confirmed", confirmed)
		return nil, errors.New("no acl in current account")
	}
	poa.mutex.Lock()
	defer poa.mutex.Unlock()
	l := 0
	r := len(acl.AksWeight)
	tmpSet := make([]*cons_base.CandidateInfo, r)
	for address, weight := range acl.AksWeight {
		if weight > 0 {
			tmpSet[l] = &cons_base.CandidateInfo{
				Address:  address,
				PeerAddr: "",
			}
			l++
		} else {
			tmpSet[r] = &cons_base.CandidateInfo{
				Address:  address,
				PeerAddr: "",
			}
			r--
		}
	}
	if l > 1 {
		poa.log.Warn("more than one weights are greater than 0, means there are more than one CAs")
	}
	proposers, _ := json.Marshal(poa.proposerInfos)
	poa.log.Info("the proposers is now updated", "proposers", proposers)
	return tmpSet, nil
}

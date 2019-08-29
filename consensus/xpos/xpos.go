package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	log "github.com/xuperchain/log15"
	"github.com/xuperchain/xuperunion/common/config"
	cons_base "github.com/xuperchain/xuperunion/consensus/base"
	crypto_base "github.com/xuperchain/xuperunion/crypto/client/base"
	"github.com/xuperchain/xuperunion/global"
	"github.com/xuperchain/xuperunion/ledger"
	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/permission/acl/utils"
	"github.com/xuperchain/xuperunion/utxo"
)

// TYPE consensus type
const TYPE = "xpos"

// XPoSConsensus structure for xpos consensus
type XPoSConsensus struct {
	// 日志
	log log.Logger
	// 共识作用的链名
	bcname string
	// 节点的address
	address []byte
	// 账本实例
	ledger *ledger.Ledger
	// utxo实例
	utxoVM *utxo.UtxoVM
	// 加密模块
	cryptoClient crypto_base.CryptoClient
	// xpos共识配置
	config xposConfig
}

// xpos共识机制的配置
type xposConfig struct {
	initProposer string
}

// GetInstance get an instance of xpos consensus
func GetInstance() interface{} {
	return &XPoSConsensus{}
}

// Type get the type of consensus
func (xp *XPoSConsensus) Type() string {
	return TYPE
}

// Version get the version of consensus
func (xp *XPoSConsensus) Version() int64 {
	return 1
}

// InitCurrent ...
func (xp *XPoSConsensus) InitCurrent(block *pb.InternalBlock) error {
	return nil
}

// Configure get configuration param from config file
func (xp *XPoSConsensus) Configure(xlog log.Logger, cfg *config.NodeConfig, consCfg map[string]interface{},
	extParams map[string]interface{}) error {
	if xlog == nil {
		xlog = log.New("module", "consensus")
		xlog.SetHandler(log.StreamHandler(os.Stderr, log.LogfmtFormat()))
	}
	address, err := ioutil.ReadFile(cfg.Miner.Keypath + "/address")
	if err != nil {
		xlog.Warn("load address error", "path", cfg.Miner.Keypath+"/address")
		return err
	}
	xp.log = xlog
	xp.address = address

	ok := false
	// get initProposer
	xp.config.initProposer, ok = extParams["initProposer"].(string)
	if !ok {
		errMsg := "invalid type of initProposer"
		xlog.Warn(errMsg)
		return errors.New(errMsg)
	}
	// get cryptoClient
	xp.cryptoClient, ok = extParams["crypto_client"].(crypto_base.CryptoClient)
	if !ok {
		errMsg := "invalid type of crypto_client"
		xlog.Warn(errMsg)
		return errors.New(errMsg)
	}
	// get ledger
	xp.ledger, ok = extParams["ledger"].(*ledger.Ledger)
	if !ok {
		errMsg := "invalid type of ledger"
		xlog.Warn(errMsg)
		return errors.New(errMsg)
	}
	// get utxovm
	xp.utxoVM, ok = extParams["utxovm"].(*utxo.UtxoVM)
	if !ok {
		errMsg := "invalid type of utxovm"
		xlog.Warn(errMsg)
		return errors.New(errMsg)
	}
	// get bcname
	xp.bcname, ok = extParams["bcname"].(string)
	if !ok {
		errMsg := "invalid type of bcname"
		xlog.Warn(errMsg)
		return errors.New(errMsg)
	}

	return nil
}

/*
 * check if it's your turn to mint a new block
 * step1: get current xchain node's address
 * step2: check if the xchain node's address has been occupied by a slot
 * step3: calculate the wait time before minting a new block
 */
func (xp *XPoSConsensus) CompeteMaster(height int64) (bool, bool) {
	t := time.Now()
	un := t.UnixNano()
	versionData, err := xp.utxoVM.GetXModel().Get(utils.GetAddress2SlotBucket(), xp.address)
	if err != nil || versionData == nil {
		ret, _ := xp.utxoVM.QuerySlot2Address()
		if len(ret) <= 0 {
			if string(xp.address) == xp.config.initProposer {
				time.Sleep(time.Duration(3) * time.Second)
				return true, false
			}
			return false, false
		}
	}
	pureData := versionData.GetPureData()
	confirmed := versionData.GetConfirmed()
	if !confirmed || pureData == nil {
		ret, _ := xp.utxoVM.QuerySlot2Address()
		if len(ret) <= 0 {
			if string(xp.address) == xp.config.initProposer {
				time.Sleep(time.Duration(3) * time.Second)
				return true, false
			}
			return false, false
		}
		return false, false
	}
	slotIdStr := string(pureData.GetValue())
	slotId, slotIdErr := strconv.ParseInt(slotIdStr, 10, 64)
	if slotIdErr != nil {
		ret, _ := xp.utxoVM.QuerySlot2Address()
		if len(ret) <= 0 {
			if string(xp.address) == xp.config.initProposer {
				time.Sleep(time.Duration(3) * time.Second)
				return true, false
			}
			return false, false
		}
		return false, false
	}
	waitTime := (slotId*3 + 60 - (un/1000000000)%60) % 60
	// 如果不加3,矿工节点会利用出块的3s时间出无限个块
	time.Sleep(time.Duration(waitTime+3) * time.Second)
	return true, xp.needSync(time.Now().UnixNano())
}

// CheckMinerMatch check if miner's block is ok
func (xp *XPoSConsensus) CheckMinerMatch(header *pb.Header, in *pb.InternalBlock) (bool, error) {
	// 1 验证块信息是否合法
	blkid, err := ledger.MakeBlockID(in)
	if err != nil {
		xp.log.Warn("CheckMinerMatch MakeBlockID error", "logid", header.Logid, "error", err)
		return false, err
	}
	if !(bytes.Equal(blkid, in.Blockid)) {
		xp.log.Warn("CheckMinerMatch equal blockid error", "logid", header.Logid, "redo blockid", global.F(blkid), "get blockid", global.F(in.Blockid))
		return false, errors.New("block doesn't match")
	}
	k, err := xp.cryptoClient.GetEcdsaPublicKeyFromJSON(in.Pubkey)
	if err != nil {
		xp.log.Warn("CheckMinerMatch get ecdsa from block error", "logid", header.Logid, "error", err)
		return false, err
	}
	chkResult, _ := xp.cryptoClient.VerifyAddressUsingPublicKey(string(in.Proposer), k)
	if chkResult == false {
		xp.log.Warn("CheckMinerMatch address is not match publickey", "logid", header.Logid)
		return false, errors.New("CheckMinerMatch address is not match publickey")
	}
	valid, err := xp.cryptoClient.VerifyECDSA(k, in.Sign, in.Blockid)
	if err != nil || !valid {
		xp.log.Warn("CheckMinerMatch VerifyECDSA error", "logid", header.Logid, "error", err)
		return false, err
	}
	// step1: check preblock minting time
	preBlock, preBlockErr := xp.ledger.QueryBlock(in.GetPreHash())
	if preBlockErr != nil {
		return false, preBlockErr
	}
	preBlockMintingTime := preBlock.GetTimestamp()
	currBlockMintingTime := in.GetTimestamp()
	if currBlockMintingTime <= preBlockMintingTime {
		xp.log.Warn("CheckMinerMatch failed, such block comes from the past")
		return false, errors.New("get a block coming from the past")
	}
	// 当只有一个initProposer的情况
	// 因为initProposer没有占用slot, 不需要验证slot信息
	// 获取当前槽位占用情况(不包括unconfirmed状态)
	ret, _ := xp.utxoVM.QuerySlot2Address()
	if len(ret) <= 0 {
		if string(in.GetProposer()) == xp.config.initProposer {
			return true, nil
		}
		return false, errors.New("all slots are empty but the block proposer is not init proposer")
	}
	// step2: check if minting time is ok
	t := time.Now()
	localTimestamp := t.UnixNano()
	startLocalTimestamp := localTimestamp/1000000000 - 15
	endLocalTimestamp := localTimestamp/1000000000 + 15
	startSlotId := startLocalTimestamp % 60 / 3
	endSlotId := endLocalTimestamp % 60 / 3
	versionData, err := xp.utxoVM.GetXModel().Get(utils.GetAddress2SlotBucket(), in.GetProposer())
	if (err != nil || versionData == nil) && string(in.GetProposer()) != xp.config.initProposer {
		return false, errors.New("such address doesn't own a slot")
	}
	pureData := versionData.GetPureData()
	confirmed := versionData.GetConfirmed()
	if pureData == nil || !confirmed {
		return false, errors.New("such address doesn't own a slot")
	}
	slotIdStr := string(pureData.GetValue())
	currBlockAddrSlotId, slotIdErr := strconv.ParseInt(slotIdStr, 10, 64)
	if slotIdErr != nil && string(in.GetProposer()) != xp.config.initProposer {
		return false, errors.New("such address doesn't own a slot")
	}
	// the proposer match with the slotId
	if currBlockAddrSlotId < startSlotId && currBlockAddrSlotId > endSlotId {
		return false, errors.New("such block came from distant future or past")
	}
	// TODO, @ToWorld check xpower
	// 需要先在block中新增字段xpower,针对不同的共识算法xpower每次新增的值不一样
	// xpower是xpos分叉管理的核心

	return true, nil
}

// ProcessBeforeMiner something before miner
func (xp *XPoSConsensus) ProcessBeforeMiner(timestamp int64) (map[string]interface{}, bool) {
	return map[string]interface{}{}, true
}

// ProcessConfirmBlock ...
func (xp *XPoSConsensus) ProcessConfirmBlock(block *pb.InternalBlock) error {
	return nil
}

// GetCoreMiners ...
func (xp *XPoSConsensus) GetCoreMiners() []*cons_base.MinerInfo {
	res := []*cons_base.MinerInfo{}
	return res
}

// GetStatus ...
func (xp *XPoSConsensus) GetStatus() *cons_base.ConsensusStatus {
	return &cons_base.ConsensusStatus{}
}

func (xp *XPoSConsensus) needSync(currTimestamp int64) bool {
	tipBlockid := xp.ledger.GetMeta().TipBlockid
	tipBlock, err := xp.ledger.QueryBlock(tipBlockid)
	if err != nil {
		return true
	}
	preBlockTimestamp := tipBlock.GetTimestamp()
	if currTimestamp-preBlockTimestamp > 0 && currTimestamp-preBlockTimestamp < 4 {
		return false
	}
	return true
}

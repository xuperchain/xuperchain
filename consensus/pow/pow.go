/**
 * @filename pow.go
 * @desc pow共识, 固定难度系数
**/
package main

import (
	"bytes"
	"errors"
	log "github.com/xuperchain/log15"
	"github.com/xuperchain/xuperunion/common/config"
	cons_base "github.com/xuperchain/xuperunion/consensus/base"
	crypto_base "github.com/xuperchain/xuperunion/crypto/client/base"
	"github.com/xuperchain/xuperunion/global"
	"github.com/xuperchain/xuperunion/ledger"
	"github.com/xuperchain/xuperunion/pb"
	"io/ioutil"
	"math/big"
	"strconv"
)

// TYPE is the type of the pow consensus
const TYPE = "pow"

// PowConsensus is struct of pow consensus
type PowConsensus struct {
	log          log.Logger
	address      []byte
	config       powConfig
	cryptoClient crypto_base.CryptoClient
	ledger       *ledger.Ledger
}

// pow 共识机制的配置
type powConfig struct {
	defaultTarget   int32
	adjustHeightGap int32
	expectedPeriod  int32
	maxTarget       int32
}

// GetInstance implement plugin framework
func GetInstance() interface{} {
	return &PowConsensus{}
}

// Type return the type of pow consensus
func (pc *PowConsensus) Type() string {
	return TYPE
}

// Version return the version of pow consensus
func (pc *PowConsensus) Version() int64 {
	return 0
}

// Configure is the specific implementation of ConsensusInterface
func (pc *PowConsensus) Configure(xlog log.Logger, cfg *config.NodeConfig, consCfg map[string]interface{},
	extParams map[string]interface{}) error {
	pc.log = xlog
	address, err := ioutil.ReadFile(cfg.Miner.Keypath + "/address")
	if err != nil {
		xlog.Warn("load address error", "path", cfg.Miner.Keypath+"/address")
		return err
	}

	if extParams["crypto_client"] == nil {
		errMsg := "crypto_client not found in extParams"
		xlog.Warn(errMsg)
		return errors.New(errMsg)
	}
	if extParams["ledger"] == nil {
		errMsg := "ledger not found in extParams"
		xlog.Warn(errMsg)
		return errors.New(errMsg)
	}

	switch extParams["crypto_client"].(type) {
	case crypto_base.CryptoClient:
		pc.cryptoClient = extParams["crypto_client"].(crypto_base.CryptoClient)
	default:
		errMsg := "invalid type of crypto_client"
		xlog.Warn(errMsg)
		return errors.New(errMsg)
	}
	switch extParams["ledger"].(type) {
	case *ledger.Ledger:
		pc.ledger = extParams["ledger"].(*ledger.Ledger)
	default:
		errMsg := "invalid type of ledger"
		xlog.Warn(errMsg)
		return errors.New(errMsg)
	}

	pc.address = address
	err = pc.buildConsConfig(xlog, consCfg)
	if err != nil {
		return err
	}
	return nil
}

func (pc *PowConsensus) buildConsConfig(xlog log.Logger, consCfg map[string]interface{}) error {
	for _, paraName := range []string{"expectedPeriod", "defaultTarget", "adjustHeightGap", "maxTarget"} {
		switch consCfg[paraName].(type) {
		case string:
			xlog.Trace("load pow parameter", paraName, consCfg[paraName])
		default:
			xlog.Warn("miss parameter or type is not string formated int", "paraName", paraName)
			return errors.New("miss:" + paraName)
		}
	}
	expectedPeriod, intErr := strconv.ParseUint(consCfg["expectedPeriod"].(string), 10, 64)
	if intErr != nil {
		return intErr
	}
	defaultTarget, intErr := strconv.ParseUint(consCfg["defaultTarget"].(string), 10, 64)
	if intErr != nil {
		return intErr
	}
	adjustHeightGap, intErr := strconv.ParseUint(consCfg["adjustHeightGap"].(string), 10, 64)
	if intErr != nil {
		return intErr
	}
	maxTarget, intErr := strconv.ParseUint(consCfg["maxTarget"].(string), 10, 64)
	if intErr != nil {
		return intErr
	}
	pc.config.expectedPeriod = int32(expectedPeriod)
	pc.config.defaultTarget = int32(defaultTarget)
	pc.config.adjustHeightGap = int32(adjustHeightGap)
	pc.config.maxTarget = int32(maxTarget)
	return nil
}

// CompeteMaster is the specific implementation of ConsensusInterface
func (pc *PowConsensus) CompeteMaster(height int64) (bool, bool) {
	return true, true
}

// CheckMinerMatch is the specific implementation of ConsensusInterface
func (pc *PowConsensus) CheckMinerMatch(header *pb.Header, in *pb.InternalBlock) (bool, error) {
	blkid, err := ledger.MakeBlockID(in)
	if err != nil {
		pc.log.Warn("MakeBlockID error", "logid", header.Logid, "error", err)
		return false, nil
	}
	if !(bytes.Equal(blkid, in.Blockid)) {
		pc.log.Warn("equal blockid error", "logid", header.Logid, "redo blockid", global.F(blkid), "get blockid", global.F(in.Blockid))
		return false, nil
	}

	targetBits := pc.calDifficulty(in)
	if targetBits != in.TargetBits {
		pc.log.Warn("unexpected target bits", "expect", targetBits, "got", in.TargetBits, "proposer", string(in.Proposer))
		return false, nil
	}
	preBlock, err := pc.ledger.QueryBlock(in.PreHash)
	if err != nil {
		pc.log.Warn("CheckMinerMatch failed, get preblock error")
		return false, nil
	}
	if in.Timestamp < preBlock.Timestamp {
		pc.log.Warn("unexpected block timestamp", "pre", preBlock.Timestamp, "next", in.Timestamp)
		return false, nil
	}
	// 验证前导0
	if !ledger.IsProofed(in.Blockid, targetBits) {
		pc.log.Warn(" blockid IsProofed error")
		return false, nil
	}

	//验证签名
	//1 验证一下签名和公钥是不是能对上
	k, err := pc.cryptoClient.GetEcdsaPublicKeyFromJSON(in.Pubkey)
	if err != nil {
		pc.log.Warn("get ecdsa from block error", "logid", header.Logid, "error", err)
		return false, nil
	}
	//Todo 跟address比较
	chkResult, _ := pc.cryptoClient.VerifyAddressUsingPublicKey(string(in.Proposer), k)
	if chkResult == false {
		pc.log.Warn("address is not match publickey", "logid", header.Logid)
		return false, nil
	}

	//2 验证一下签名是否正确
	valid, err := pc.cryptoClient.VerifyECDSA(k, in.Sign, in.Blockid)
	if err != nil {
		pc.log.Warn("VerifyECDSA error", "logid", header.Logid, "error", err)
	}
	return valid, nil

}

// InitCurrent is the specific implementation of ConsensusInterface
func (pc *PowConsensus) InitCurrent(block *pb.InternalBlock) error {
	return nil
}

// ProcessBeforeMiner is the specific implementation of ConsensusInterface
func (pc *PowConsensus) ProcessBeforeMiner(timestamp int64) (map[string]interface{}, bool) {
	res := make(map[string]interface{})
	res["type"] = TYPE
	res["targetBits"] = pc.calDifficulty(nil)
	return res, true
}

// ProcessConfirmBlock is the specific implementation of ConsensusInterface
func (pc *PowConsensus) ProcessConfirmBlock(block *pb.InternalBlock) error {
	return nil
}

func (pc *PowConsensus) getTargetBitsFromBlock(block *pb.InternalBlock) int32 {
	return block.TargetBits
}

func (pc *PowConsensus) getPrevBlock(curBlock *pb.InternalBlock, gap int32) (prevBlock *pb.InternalBlock, err error) {
	for i := int32(0); i < gap; i++ {
		prevBlock, err = pc.ledger.QueryBlockHeader(curBlock.PreHash)
		if err != nil {
			return
		}
		curBlock = prevBlock
	}
	return
}

// reference of bitcoin's pow: https://github.com/bitcoin/bitcoin/blob/master/src/pow.cpp#L49
func (pc *PowConsensus) calDifficulty(curBlock *pb.InternalBlock) int32 {
	if curBlock == nil {
		curBlock = &pb.InternalBlock{
			PreHash: pc.ledger.GetMeta().TipBlockid,
			Height:  pc.ledger.GetMeta().TrunkHeight + 1,
		}
	}
	if curBlock.Height <= int64(pc.config.adjustHeightGap) {
		return pc.config.defaultTarget
	}
	height := curBlock.Height
	preBlock, err := pc.getPrevBlock(curBlock, 1)
	if err != nil {
		pc.log.Warn("query prev block failed", "err", err, "height", height-1)
		return pc.config.defaultTarget
	}
	prevTargetBits := pc.getTargetBitsFromBlock(preBlock)
	if height%int64(pc.config.adjustHeightGap) == 0 {
		farBlock, err := pc.getPrevBlock(curBlock, pc.config.adjustHeightGap)
		if err != nil {
			pc.log.Warn("query far block failed", "err", err, "height", height-int64(pc.config.adjustHeightGap))
			return pc.config.defaultTarget
		}
		expectedTimeSpan := pc.config.expectedPeriod * (pc.config.adjustHeightGap - 1)
		actualTimeSpan := int32((preBlock.Timestamp - farBlock.Timestamp) / 1e9)
		pc.log.Info("timespan diff", "expectedTimeSpan", expectedTimeSpan, "actualTimeSpan", actualTimeSpan)
		//at most adjust two bits, left or right direction
		if actualTimeSpan < expectedTimeSpan/4 {
			actualTimeSpan = expectedTimeSpan / 4
		}
		if actualTimeSpan > expectedTimeSpan*4 {
			actualTimeSpan = expectedTimeSpan * 4
		}
		difficulty := big.NewInt(1)
		difficulty.Lsh(difficulty, uint(prevTargetBits))
		difficulty.Mul(difficulty, big.NewInt(int64(expectedTimeSpan)))
		difficulty.Div(difficulty, big.NewInt(int64(actualTimeSpan)))
		newTargetBits := int32(difficulty.BitLen() - 1)
		if newTargetBits > pc.config.maxTarget {
			pc.log.Info("retarget", "newTargetBits", newTargetBits)
			newTargetBits = pc.config.maxTarget
		}
		pc.log.Info("adjust targetBits", "height", height, "targetBits", newTargetBits, "prevTargetBits", prevTargetBits)
		return newTargetBits
	} else {
		pc.log.Info("prev targetBits", "prevTargetBits", prevTargetBits)
		return prevTargetBits
	}
}

// GetCoreMiners get the information of core miners
func (pc *PowConsensus) GetCoreMiners() []*cons_base.MinerInfo {
	// PoW don't have definite miner
	res := []*cons_base.MinerInfo{}
	return res
}

// GetStatus get current status of consensus
func (pc *PowConsensus) GetStatus() *cons_base.ConsensusStatus {
	return &cons_base.ConsensusStatus{}
}

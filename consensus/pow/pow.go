/**
 * @filename pow.go
 * @desc pow共识, 固定难度系数
**/
package main

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
	"io/ioutil"

	log "github.com/xuperchain/log15"
	"github.com/xuperchain/xuperunion/common/config"
	cons_base "github.com/xuperchain/xuperunion/consensus/base"
	crypto_base "github.com/xuperchain/xuperunion/crypto/client/base"
	"github.com/xuperchain/xuperunion/global"
	"github.com/xuperchain/xuperunion/ledger"
	"github.com/xuperchain/xuperunion/pb"
)

// TYPE is the type of the pow consensus
const TYPE = "pow"

//TargetBits is the targetBits of pow difficulty
// todo: 后续修改为弹性调整
const TargetBits = 16

// PowConsensus is struct of pow consensus
type PowConsensus struct {
	log          log.Logger
	address      []byte
	config       powConfig
	cryptoClient crypto_base.CryptoClient
}

// pow 共识机制的配置
type powConfig struct {
	targetBits int32
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

	switch extParams["crypto_client"].(type) {
	case crypto_base.CryptoClient:
		pc.cryptoClient = extParams["crypto_client"].(crypto_base.CryptoClient)
	default:
		errMsg := "invalid type of crypto_client"
		xlog.Warn(errMsg)
		return errors.New(errMsg)
	}

	pc.address = address
	pc.config.targetBits = TargetBits
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

	// 验证前导0
	if !ledger.IsProofed(in.Blockid, pc.config.targetBits) {
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
	ks := []*ecdsa.PublicKey{}
	ks = append(ks, k)
	valid, err := pc.cryptoClient.VerifyXuperSignature(ks, in.Sign, in.Blockid)
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
	res["targetBits"] = pc.calDifficulty()
	return res, true
}

// ProcessConfirmBlock is the specific implementation of ConsensusInterface
func (pc *PowConsensus) ProcessConfirmBlock(block *pb.InternalBlock) error {
	return nil
}

// todo: 后续增加难度系数动态调整
func (pc *PowConsensus) calDifficulty() int32 {
	return pc.config.targetBits
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

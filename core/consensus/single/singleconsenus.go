// Package main is the plugin implementation of single consensus
package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"strconv"
	"time"

	log "github.com/xuperchain/log15"
	"github.com/xuperchain/xuperchain/core/common/config"
	cons_base "github.com/xuperchain/xuperchain/core/consensus/base"
	crypto_base "github.com/xuperchain/xuperchain/core/crypto/client/base"
	"github.com/xuperchain/xuperchain/core/global"
	"github.com/xuperchain/xuperchain/core/ledger"
	"github.com/xuperchain/xuperchain/core/pb"
)

// TYPE is the type of the pow consensus
const TYPE = "single"

// SingleConsensus is struct of single consensus
type SingleConsensus struct {
	log                log.Logger
	master             bool
	address            []byte
	masterAddr         []byte
	blockProducePeriod int64
	isProduce          map[int64]bool
	cryptoClient       crypto_base.CryptoClient
	state              cons_base.ConsensusState
}

// GetInstance : implement plugin framework
func GetInstance() interface{} {
	return &SingleConsensus{}
}

// Type return the type of single consensus
func (sc *SingleConsensus) Type() string {
	return TYPE
}

// Version return the version of single consensus
func (sc *SingleConsensus) Version() int64 {
	return 0
}

// Configure is the specific implementation of ConsensusInterface
func (sc *SingleConsensus) Configure(xlog log.Logger, cfg *config.NodeConfig, consCfg map[string]interface{},
	extParams map[string]interface{}) error {
	sc.log = xlog
	if consCfg["miner"] == nil {
		return errors.New("Parse SingleConsensus miner error, can not be null")
	}

	if consCfg["period"] == nil {
		return errors.New("Parse SingleConsensus period error, can not be null")
	}

	switch consCfg["miner"].(type) {
	case string:
		sc.masterAddr = []byte(consCfg["miner"].(string))
	default:
		return errors.New("the type of miner should be string")
	}

	switch consCfg["period"].(type) {
	case string:
		period, err := strconv.ParseInt(consCfg["period"].(string), 10, 64)
		if err != nil {
			xlog.Warn("Parse SingleConsensus config period error", "error", err.Error())
			return err
		}
		sc.blockProducePeriod = period * 1e6
	default:
		return errors.New("the type of period should be string")
	}

	address, err := ioutil.ReadFile(cfg.Miner.Keypath + "/address")
	if err != nil {
		xlog.Warn("load address error", "path", cfg.Miner.Keypath+"/address")
		return err
	}
	sc.address = address
	sc.isProduce = make(map[int64]bool)

	switch extParams["crypto_client"].(type) {
	case crypto_base.CryptoClient:
		sc.cryptoClient = extParams["crypto_client"].(crypto_base.CryptoClient)
	default:
		errMsg := "invalid type of crypto_client"
		xlog.Warn(errMsg)
		return errors.New(errMsg)
	}

	log.Trace("block produce period " + strconv.FormatInt(sc.blockProducePeriod, 10) + "ms")
	return nil
}

// CompeteMaster is the specific implementation of ConsensusInterface
func (sc *SingleConsensus) CompeteMaster(height int64) (bool, bool) {
Again:
	t := time.Now()
	un := t.UnixNano()
	key := un / sc.blockProducePeriod
	sleep := un % sc.blockProducePeriod
	if sleep > int64(time.Second) {
		sleep = int64(time.Second)
	}
	v, ok := sc.isProduce[key]
	if !ok || v == false {
		sc.isProduce[key] = true
	} else {
		time.Sleep(time.Duration(sleep))
		goto Again
	}
	if string(sc.address) == string(sc.masterAddr) {
		sc.log.Trace("CompeteMaster", "UnixNano", un, "key", key, "sleep", sleep, "master", sc.master)
		return true, false
	}
	sc.log.Trace("CompeteMaster is not master", "master", sc.master)
	return false, false
}

// CheckMinerMatch is the specific implementation of ConsensusInterface
func (sc *SingleConsensus) CheckMinerMatch(header *pb.Header, in *pb.InternalBlock) (bool, error) {
	blkid, err := ledger.MakeBlockID(in)
	if err != nil {
		sc.log.Warn("MakeBlockID error", "logid", header.Logid, "error", err)
		return false, nil
	}
	if !(bytes.Equal(blkid, in.Blockid) && bytes.Equal(in.Proposer, sc.masterAddr)) {
		sc.log.Warn("equal blockid error", "logid", header.Logid, "redo blockid", global.F(blkid), "get blockid", global.F(in.Blockid), "in.proposer", global.F(in.Proposer), "proposer", global.F(sc.masterAddr))
		return false, nil
	}
	//验证签名
	//1 验证一下签名和公钥是不是能对上
	k, err := sc.cryptoClient.GetEcdsaPublicKeyFromJSON(in.Pubkey)
	if err != nil {
		sc.log.Warn("get ecdsa from block error", "logid", header.Logid, "error", err)
		return false, nil
	}
	//Todo 跟address比较
	chkResult, _ := sc.cryptoClient.VerifyAddressUsingPublicKey(string(in.Proposer), k)
	if chkResult == false {
		sc.log.Warn("address is not match publickey", "logid", header.Logid)
		return false, nil
	}

	//2 验证一下签名是否正确
	valid, err := sc.cryptoClient.VerifyECDSA(k, in.Sign, in.Blockid)
	if err != nil {
		sc.log.Warn("VerifyECDSA error", "logid", header.Logid, "error", err)
	}
	return valid, nil
}

// InitCurrent is the specific implementation of ConsensusInterface
func (sc *SingleConsensus) InitCurrent(block *pb.InternalBlock) error {
	return nil
}

// ProcessBeforeMiner is the specific implementation of ConsensusInterface
func (sc *SingleConsensus) ProcessBeforeMiner(timestamp int64) (map[string]interface{}, bool) {
	return nil, true
}

// ProcessConfirmBlock is the specific implementation of ConsensusInterface
func (sc *SingleConsensus) ProcessConfirmBlock(block *pb.InternalBlock) error {
	return nil
}

// GetCoreMiners get the information of core miners
func (sc *SingleConsensus) GetCoreMiners() []*cons_base.MinerInfo {
	// single consensus only have one definite miner
	res := []*cons_base.MinerInfo{}
	master := &cons_base.MinerInfo{
		Address:  string(sc.masterAddr),
		PeerInfo: "",
	}
	res = append(res, master)
	return res
}

// GetStatus get current status of consensus
func (sc *SingleConsensus) GetStatus() *cons_base.ConsensusStatus {
	return &cons_base.ConsensusStatus{
		Proposer: string(sc.masterAddr),
	}
}

// Suspend is the specific implementation of ConsensusInterface
func (sc *SingleConsensus) Suspend() error {
	sc.state = cons_base.SUSPEND
	return nil
}

// Activate is the specific implementation of ConsensusInterface
func (sc *SingleConsensus) Activate() error {
	sc.state = cons_base.RUNNING
	return nil
}

// IsActive return whether the state of consensus is active
func (sc *SingleConsensus) IsActive() bool {
	return sc.state == cons_base.RUNNING
}

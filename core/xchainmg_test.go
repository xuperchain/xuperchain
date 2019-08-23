package xchaincore

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"

	log "github.com/xuperchain/log15"
	"github.com/xuperchain/xuperunion/common/config"
	"github.com/xuperchain/xuperunion/global"
	"github.com/xuperchain/xuperunion/p2pv2"
	xuper_p2p "github.com/xuperchain/xuperunion/p2pv2/pb"
	"github.com/xuperchain/xuperunion/pb"

	"github.com/xuperchain/xuperunion/contract/kernel"
	crypto_client "github.com/xuperchain/xuperunion/crypto/client"
	"github.com/xuperchain/xuperunion/ledger"
	"github.com/xuperchain/xuperunion/utxo"
)

const BobAddress = "dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN"
const BobPubkey = `{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571}`
const BobPrivateKey = `{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571,"D":29079635126530934056640915735344231956621504557963207107451663058887647996601}`
const DefaultKvEngine = "default"

var baseDir = os.Getenv("XCHAIN_ROOT")

func Init(t *testing.T) *XChainMG {
	logger := log.New("module", "xchain")
	logger.SetHandler(log.StreamHandler(os.Stderr, log.LogfmtFormat()))
	cfg := config.NewNodeConfig()

	l, _ := net.Listen("tcp", ":0")
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	t.Log("port: ", port)
	cfg.P2pV2.Port = int32(port)

	p2pV2Service, p2pV2Err := p2pv2.NewP2PServerV2(cfg.P2pV2, nil)
	t.Log("cfg: ", cfg.P2pV2)
	if p2pV2Err != nil {
		t.Error("new p2pv2 server error ", p2pV2Err.Error())
	}
	xcmg := &XChainMG{}
	cfg.Datapath = "../core/data/blockchain"
	if err := xcmg.Init(logger, cfg, p2pV2Service); err != nil {
		t.Error("XChainMG init error ", err.Error())
	}
	return xcmg
}

func TestXChainMgBasic(t *testing.T) {
	/*
		c := exec.Command("sh", "-c", "../xchain-cli netUrl gen")
		_, cmdErr := c.Output()
		if cmdErr != nil {
			t.Error("netUrl gen error ", cmdErr.Error())
		}
	*/
	InitCreateBlockChain(t)
	xcmg := Init(t)
	defer func() {
		if xcmg != nil {
			xcmg.Stop()
			defer os.RemoveAll(fmt.Sprintf("%s/core/data", baseDir))
		}
	}()
	if xcmg == nil {
		t.Error("create XChainMG error")
	}
	// test for Get
	rootXCore := xcmg.Get("xuper")
	if rootXCore == nil {
		t.Error("expect not nil, but got ", rootXCore)
	}
	xcore := xcmg.Get("Dog")
	if xcore != nil {
		t.Error("expect nil, but got ", xcore)
	}
	// test for GetAll
	bcs := xcmg.GetAll()
	if len(bcs) != 1 {
		t.Error("expect size but got ", len(bcs), "expect xuper xchain but got ", bcs[0])
	}
	// test for Start
	xcmg.Start()
	// test for CreateBlockChain
	// create exist chain
	xcore2, xcoreErr := xcmg.CreateBlockChain("xuper", []byte("todo"))
	if xcoreErr != ErrBlockChainIsExist {
		t.Error("expect ErrBlockChainIsExist, but got ", xcoreErr)
	} else {
		t.Log("xcore2: ", xcore2)
	}
	// create non exist chain
	rootJs, _ := ioutil.ReadFile(fmt.Sprintf("%s/core/data/config", baseDir) + "/xuper.json")
	xcore2, xcoreErr = xcmg.CreateBlockChain("dog", rootJs)
	if xcoreErr != nil {
		t.Error("create non exist chain error ", xcoreErr.Error())
	} else {
		defer os.RemoveAll(fmt.Sprintf("%s/core/data/blockchain/dog", baseDir))
		t.Log("dog chain ", xcore2)
	}
	status := rootXCore.Status()
	t.Log("xuper chain status ", status)

	txReq := &pb.TxData{}
	txReq.Bcname = "xuper"
	txReq.FromAddr = BobAddress
	txReq.FromPubkey = BobPubkey
	txReq.FromScrkey = BobPrivateKey
	txReq.Nonce = "nonce"
	txReq.Timestamp = time.Now().UnixNano()
	// tx_req.Desc = []byte("")
	txReq.Account = []*pb.TxDataAccount{
		{Address: BobAddress, Amount: "1"},
	}
	txReq.Header = &pb.Header{}
	hd := &global.XContext{Timer: global.NewXTimer()}
	txStatus := rootXCore.GenerateTx(txReq, hd)
	t.Log("tx status ", txStatus)

	ret, balErr := rootXCore.GetBalance(BobAddress)
	if balErr != nil {
		t.Error("get balance error ", balErr.Error())
	} else {
		t.Log("address ", BobAddress, " balance ", ret)
	}
	// test for GetFrozenBalance
	ret, balErr = rootXCore.GetFrozenBalance(BobAddress)
	if balErr != nil {
		t.Error("get frozen balance error ", balErr.Error())
	} else {
		t.Log("address ", BobAddress, " frozen balance ", ret)
	}

	// test for GetConsType
	consType := rootXCore.GetConsType()
	if consType != "single" {
		t.Error("expect consType is single, but got ", consType)
	}
	// test for GetDposCandidates
	strArr, _ := rootXCore.GetDposCandidates()
	if len(strArr) != 0 {
		t.Error("expect 0 candidates, but got ", len(strArr))
	}
	// test for GetDposNominateRecords
	strArr2, _ := rootXCore.GetDposNominateRecords(BobAddress)
	if len(strArr2) != 0 {
		t.Error("expect 0 moninate records, but got ", len(strArr2))
	}
	strArr3, _ := rootXCore.GetDposNominatedRecords(BobAddress)
	if len(strArr3) != 0 {
		t.Error("expect 0 moninated records, but got ", len(strArr3))
	}
	strArr4, _ := rootXCore.GetDposVoteRecords(BobAddress)
	if len(strArr4) != 0 {
		t.Error("expect 0 vote records, but got ", len(strArr4))
	}
	strArr5, _ := rootXCore.GetDposVotedRecords(BobAddress)
	if len(strArr5) != 0 {
		t.Error("expect 0 voted records, but got ", len(strArr4))
	}
	// test for GetCheckResults
	proposers, _ := rootXCore.GetCheckResults(0)
	if len(proposers) != 0 {
		t.Error("expect 0 proposers, but got ", len(proposers))
	}
	// test for PostTx -> txStatus
	reply, state := rootXCore.PostTx(txStatus, hd)
	t.Log("postTx reply ", reply)
	t.Log("postTx status ", state)
	// test for QueryTx -> txStatus
	result := rootXCore.QueryTx(txStatus)
	t.Log("query non exist tx result ", result)
	// query ok tx
	ecdsaPk, pkErr := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if pkErr != nil {
		t.Error("generate key error ", pkErr.Error())
	}
	t.Log("tip blockid ", rootXCore.Ledger.GetMeta().TipBlockid)
	t.Log("tx ", txStatus.Tx)
	block, formatBlockErr := rootXCore.Ledger.FormatBlock([]*pb.Transaction{txStatus.Tx},
		[]byte(BobAddress),
		ecdsaPk,
		223456789,
		0,
		0,
		rootXCore.Ledger.GetMeta().TipBlockid,
		big.NewInt(1))
	if formatBlockErr != nil {
		t.Error("format block error", formatBlockErr.Error())
	}
	confirmStatus := rootXCore.Ledger.ConfirmBlock(block, false)
	if !confirmStatus.Succ {
		t.Error("confirm block error ", confirmStatus)
	}
	playErr := rootXCore.Utxovm.Play(block.Blockid)
	if playErr != nil {
		t.Error("utxo play error ", playErr.Error())
	}
	rootXCore.nodeMode = config.NodeModeFastSync
	globalBlock := &pb.Block{
		Header:  &pb.Header{},
		Bcname:  "xuper",
		Blockid: block.Blockid,
		Status:  pb.Block_TRUNK,
		Block:   block,
	}
	result = rootXCore.QueryTx(txStatus)
	t.Log("query exist tx result ", result)
	// test for GetNodeMode
	nodeMode := rootXCore.GetNodeMode()
	if nodeMode != config.NodeModeFastSync {
		t.Error("expect FAST_SYNC, but got ", nodeMode)
	}
	// test for SendBlock
	// sendblock exist block
	sendBlockErr := rootXCore.SendBlock(globalBlock, hd)
	if sendBlockErr != nil && sendBlockErr != ErrBlockExist {
		t.Error("send block error ", sendBlockErr.Error())
	}
	// sendblock non exist block but trunk block
	txReq.Nonce = "nonce1"
	txReq.Timestamp = time.Now().UnixNano()
	txStatus = rootXCore.GenerateTx(txReq, &global.XContext{Timer: global.NewXTimer()})
	block, formatBlockErr = rootXCore.Ledger.FormatBlock([]*pb.Transaction{txStatus.Tx},
		[]byte(BobAddress),
		ecdsaPk,
		223456789,
		0,
		0,
		rootXCore.Ledger.GetMeta().TipBlockid,
		big.NewInt(1))
	if formatBlockErr != nil {
		t.Error("format block error", formatBlockErr.Error())
	}
	block.Height = rootXCore.Ledger.GetMeta().TrunkHeight + 1
	globalBlock = &pb.Block{
		Header:  &pb.Header{},
		Bcname:  "xuper",
		Blockid: block.Blockid,
		Status:  pb.Block_TRUNK,
		Block:   block,
	}
	sendBlockErr = rootXCore.SendBlock(globalBlock, &global.XContext{Timer: global.NewXTimer()})
	if sendBlockErr != nil {
		t.Error("send block error ", sendBlockErr.Error())
		return
	}
	status = rootXCore.Status()
	t.Log("xuper chain status ", status)
	// sendblock non exist block and no trunk block
	txReq.Nonce = "nonce2"
	txReq.Timestamp = time.Now().UnixNano()
	txStatus = rootXCore.GenerateTx(txReq, &global.XContext{Timer: global.NewXTimer()})
	block, formatBlockErr = rootXCore.Ledger.FormatBlock([]*pb.Transaction{txStatus.Tx},
		[]byte(BobAddress),
		ecdsaPk,
		223456789,
		0,
		0,
		rootXCore.Ledger.GetMeta().TipBlockid,
		big.NewInt(1))
	if formatBlockErr != nil {
		t.Error("format block error", formatBlockErr.Error())
	}
	block.Height = rootXCore.Ledger.GetMeta().TrunkHeight + 1
	globalBlock = &pb.Block{
		Header:  &pb.Header{},
		Bcname:  "xuper",
		Blockid: block.Blockid,
		Status:  pb.Block_TRUNK,
		Block:   block,
	}
	// save in pending table
	saveErr := rootXCore.Ledger.SavePendingBlock(globalBlock)
	if saveErr != nil {
		t.Error("save error ", saveErr.Error())
	}
	txReq.Nonce = "nonce3"
	txReq.Timestamp = time.Now().UnixNano()
	txStatus = rootXCore.GenerateTx(txReq, &global.XContext{Timer: global.NewXTimer()})
	block2, formatBlockErr2 := rootXCore.Ledger.FormatBlock([]*pb.Transaction{txStatus.Tx},
		[]byte(BobAddress),
		ecdsaPk,
		223456789,
		0,
		0,
		globalBlock.Block.Blockid,
		big.NewInt(1))

	if formatBlockErr2 != nil {
		t.Error("format block error", formatBlockErr.Error())
	}
	globalBlock = &pb.Block{
		Header:  &pb.Header{},
		Bcname:  "xuper",
		Blockid: block2.Blockid,
		Status:  pb.Block_TRUNK,
		Block:   block2,
	}
	block2.Height = rootXCore.Ledger.GetMeta().TrunkHeight + 1
	sendBlockErr = rootXCore.SendBlock(globalBlock, &global.XContext{Timer: global.NewXTimer()})
	if sendBlockErr != nil {
		t.Error("send block error ", sendBlockErr.Error())
	}
	// test for GetBlock
	// get exist block
	blockID := &pb.BlockID{
		Header:      &pb.Header{},
		Bcname:      "xuper",
		Blockid:     globalBlock.Blockid,
		NeedContent: true,
	}
	existBlock := rootXCore.GetBlock(blockID)
	if existBlock == nil {
		t.Error("expect no nil, but got ", existBlock)
	}
	// get non exist block
	blockID = &pb.BlockID{
		Header:      &pb.Header{},
		Bcname:      "dog",
		Blockid:     []byte("123456"),
		NeedContent: true,
	}
	existBlock = rootXCore.GetBlock(blockID)
	if existBlock.Status != pb.Block_NOEXIST {
		t.Error("expect NOEXIST but got ", existBlock.Status)
	}
	// test for GetBlockChainStatus
	blkStatus := &pb.BCStatus{
		Header: &pb.Header{},
		Bcname: "xuper",
	}
	out := rootXCore.GetBlockChainStatus(blkStatus)
	t.Log("out:  ", out)
	out2 := rootXCore.ConfirmTipBlockChainStatus(blkStatus)
	t.Log("out: ", out2)
	// test for countGetBlockChainStatus
	res, _ := xuper_p2p.NewXuperMessage(xuper_p2p.XuperMsgVersion2, "xuper", "123456789",
		xuper_p2p.XuperMessage_GET_BLOCK_RES, nil, xuper_p2p.XuperMessage_CHECK_SUM_ERROR)
	countGetBlockChainStatus([]*xuper_p2p.XuperMessage{res})
	// test for countConfirmBlockRes
	t.Log("state value ", countConfirmBlockRes([]*xuper_p2p.XuperMessage{res}))
	t.Log("is accepted: ", rootXCore.syncConfirm(blkStatus))
	res2, _ := rootXCore.syncForOnce()
	t.Log("sync for once", res2)
	rootXCore.doMiner()
	// test fot BroadCastGetBlock
	rootXCore.BroadCastGetBlock(blockID)
	// test for xchainmg_net.go
	// assemble POSTTX,SENDBLOCK,BATCHPOSTTX,
	msgInfo, _ := proto.Marshal(txStatus)
	t.Log("msg info ", msgInfo)
	msgTmp, _ := xuper_p2p.NewXuperMessage(xuper_p2p.XuperMsgVersion1, "xuper", "123456", xuper_p2p.XuperMessage_POSTTX,
		msgInfo, xuper_p2p.XuperMessage_NONE)
	xcmg.msgChan <- msgTmp
	//batch BatchPostTx
	batchTxs := &pb.BatchTxs{
		Header: &pb.Header{},
		Txs:    []*pb.TxStatus{txStatus},
	}
	t.Log("batchTxs: ", batchTxs)
	msgInfos, _ := proto.Marshal(batchTxs)
	msgsTmp, _ := xuper_p2p.NewXuperMessage(xuper_p2p.XuperMsgVersion1, "xuper", "123457", xuper_p2p.XuperMessage_BATCHPOSTTX,
		msgInfos, xuper_p2p.XuperMessage_NONE)
	xcmg.msgChan <- msgsTmp
	sendBlockMsgInfo, _ := proto.Marshal(globalBlock)
	sendBlockMsgTmp, _ := xuper_p2p.NewXuperMessage(xuper_p2p.XuperMsgVersion1, "xuper", "123458", xuper_p2p.XuperMessage_SENDBLOCK,
		sendBlockMsgInfo, xuper_p2p.XuperMessage_NONE)
	xcmg.msgChan <- sendBlockMsgTmp
}

func InitCreateBlockChain(t *testing.T) {
	//defer os.RemoveAll(workSpace)
	workSpace := fmt.Sprintf("%s/core/data/blockchain/xuper", baseDir)
	os.RemoveAll(fmt.Sprintf("%s/core/data", baseDir))
	cmd := fmt.Sprintf("cp -rf %s/data %s/core/", baseDir, baseDir)
	fmt.Println(cmd)
	c := exec.Command("sh", "-c", cmd)

	_, cmdErr := c.Output()
	if cmdErr != nil {
		//	t.Error("cp error ", cmdErr.Error())
	}
	os.RemoveAll(fmt.Sprintf("%s/core/data/blockchain/xuper/", baseDir))
	ledger, err := ledger.NewLedger(workSpace, nil, nil, DefaultKvEngine, crypto_client.CryptoTypeDefault)
	if err != nil {
		t.Fatal(err)
	}
	defer ledger.Close()
	kl := &kernel.Kernel{}
	kLogger := log.New("module", "kernel")
	kLogger.SetHandler(log.StreamHandler(os.Stderr, log.LogfmtFormat()))
	kl.Init(workSpace, kLogger, nil, "xuper")
	kl.SetNewChainWhiteList(map[string]bool{BobAddress: true})
	utxoVM, _ := utxo.MakeUtxoVM("xuper", ledger, workSpace, "", "", []byte(""), nil, 5000, 60, 500, nil, false, DefaultKvEngine, crypto_client.CryptoTypeDefault)
	utxoVM.RegisterVM("kernel", kl, global.VMPrivRing0)
	//创建链的时候分配财富
	tx, err2 := utxoVM.GenerateRootTx([]byte(`
        {
            "version" : "1"
            , "consensus" : {
                "type"  : "single",
                "miner" : "dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN"
            }
            , "predistribution":[
                {
                    "address" : "dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN"
                    , "quota" : "100000000000000000000"
                }
            ]
            , "maxblocksize" : "128"
            , "period" : "3000"
            , "award" : "428100000000"
            , "decimals" : "8"
            , "award_decay": {
                "height_gap": 31536000,
                "ratio": 1
            }
        }
    `))
	if err2 != nil {
		t.Fatal(err2)
	}
	defer utxoVM.Close()
	block, _ := ledger.FormatRootBlock([]*pb.Transaction{tx})
	confirmStatus := ledger.ConfirmBlock(block, true)
	if !confirmStatus.Succ {
		t.Fatal("confirm block fail")
	}
	playErr := utxoVM.Play(block.Blockid)
	if playErr != nil {
		t.Fatal(playErr)
	}
	c = exec.Command("sh", "-c", fmt.Sprintf("cp %s/core/data/config/xuper.json %s/core/data/blockchain/xuper/", baseDir, baseDir))
	_, cmdErr = c.Output()
	if cmdErr != nil {
		t.Error("cp error ", cmdErr.Error())
	}
}

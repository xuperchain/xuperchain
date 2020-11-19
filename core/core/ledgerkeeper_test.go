package xchaincore

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	log "github.com/xuperchain/log15"
	"github.com/xuperchain/xuperchain/core/common/config"
	"github.com/xuperchain/xuperchain/core/consensus"
	crypto_client "github.com/xuperchain/xuperchain/core/crypto/client"
	"github.com/xuperchain/xuperchain/core/global"
	"github.com/xuperchain/xuperchain/core/ledger"
	p2p_base "github.com/xuperchain/xuperchain/core/p2p/base"
	"github.com/xuperchain/xuperchain/core/p2p/p2pv2"
	xuper_p2p "github.com/xuperchain/xuperchain/core/p2p/pb"
	"github.com/xuperchain/xuperchain/core/pb"
	"github.com/xuperchain/xuperchain/core/utxo"
	"github.com/xuperchain/xuperchain/core/utxo/txhash"
)

const (
	engine        = "default"
	bobAddress    = "dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN"
	bobPubkey     = `{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571}`
	bobPrivateKey = `{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571,"D":29079635126530934056640915735344231956621504557963207107451663058887647996601}`
)

type fakeBlockChainHolder struct {
	Ledger     *ledger.Ledger
	UtxoVM     *utxo.UtxoVM
	B0         *pb.InternalBlock
	B1         *pb.InternalBlock
	B2         *pb.InternalBlock
	PrivateKey *ecdsa.PrivateKey
}

func generateTx(coinbase bool, utxo *utxo.UtxoVM) *pb.Transaction {
	txReq := &pb.TxData{}
	txReq.Bcname = "xuper"
	txReq.FromAddr = bobAddress
	txReq.FromPubkey = bobPubkey
	txReq.FromScrkey = bobPrivateKey
	txReq.Nonce = "nonce"
	txReq.Timestamp = time.Now().UnixNano()
	txReq.Account = []*pb.TxDataAccount{
		{Address: bobAddress, Amount: "1000000", FrozenHeight: int64(1000000)},
	}
	tx, _ := utxo.GenerateTx(txReq)
	return tx
}

func prepareBlockchain() *fakeBlockChainHolder {
	workSpace, _ := ioutil.TempDir("/tmp", "")
	os.RemoveAll(workSpace)
	defer os.RemoveAll(workSpace)
	// 准备账本
	l, _ := ledger.NewLedger(workSpace, nil, nil, engine, crypto_client.CryptoTypeDefault)
	cryptoClient, _ := crypto_client.CreateCryptoClient(crypto_client.CryptoTypeDefault)
	privateKey, _ := cryptoClient.GetEcdsaPrivateKeyFromJSON([]byte(bobPrivateKey))

	t1 := &pb.Transaction{}
	t1.TxOutputs = append(t1.TxOutputs, &pb.TxOutput{Amount: []byte("1000000000000"), ToAddr: []byte(bobAddress)})
	t1.Coinbase = true
	t1.Desc = []byte(`{"maxblocksize" : "128"}`)
	t1.Txid, _ = txhash.MakeTransactionID(t1)
	block0, _ := l.FormatRootBlock([]*pb.Transaction{t1})
	status := l.ConfirmBlock(block0, true)
	// 准备utxovm
	address, _ := hex.DecodeString(bobAddress)
	utxovm, _ := utxo.NewUtxoVM("xuper", l, workSpace, bobPrivateKey, bobPubkey, address, nil, false, "default", crypto_client.CryptoTypeDefault)
	utxovm.Play(block0.GetBlockid())
	fmt.Print(status)

	t2 := generateTx(false, utxovm)
	block1, _ := l.FormatFakeBlock([]*pb.Transaction{t2}, []byte(bobAddress), privateKey, time.Now().UnixNano(), 1, 1, block0.GetBlockid(), big.NewInt(0), 1)
	status = l.ConfirmBlock(block1, false)
	utxovm.Play(block1.GetBlockid())

	t3 := generateTx(false, utxovm)
	block2, _ := l.FormatFakeBlock([]*pb.Transaction{t3}, []byte(bobAddress), privateKey, time.Now().UnixNano(), 2, 2, block1.GetBlockid(), big.NewInt(0), 2)
	status = l.ConfirmBlock(block2, false)
	utxovm.Play(block2.GetBlockid())

	return &fakeBlockChainHolder{
		Ledger:     l,
		UtxoVM:     utxovm,
		B0:         block0,
		B1:         block1,
		B2:         block2,
		PrivateKey: privateKey,
	}
}

func prepareSingleCon(path string, ledger *ledger.Ledger, utxoVM *utxo.UtxoVM) *consensus.PluggableConsensus {
	cfg := config.NewNodeConfig()
	cfg.Miner.Keypath = path
	rootConfig := map[string]interface{}{
		"name": consensus.ConsensusTypeSingle,
		"config": map[string]interface{}{
			"period": "3000",
			"miner":  bobAddress,
		},
	}
	plugCons, _ := consensus.NewPluggableConsensus(nil, cfg, "xuper", ledger, utxoVM, rootConfig, crypto_client.CryptoTypeDefault, nil)
	return plugCons
}

func prepareLedgerKeeper(port int32, path string) (*LedgerKeeper, *fakeBlockChainHolder) {
	bcHolder := prepareBlockchain()
	// 准备共识
	consensus_path := "../data/keys"
	consensus := prepareSingleCon(consensus_path, bcHolder.Ledger, bcHolder.UtxoVM)
	// 准备p2p节点
	testCases := map[string]struct {
		in config.P2PConfig
	}{
		"testNewServer": {
			in: config.P2PConfig{
				Port:            port,
				KeyPath:         path,
				IsNat:           true,
				IsHidden:        false,
				BootNodes:       []string{},
				MaxStreamLimits: 32,
			},
		},
	}

	p2pSrv := p2pv2.NewP2PServerV2()
	p2pSrv.Init(testCases["testNewServer"].in, nil, nil)
	l, _ := NewLedgerKeeper("xuper", nil, p2pSrv, bcHolder.Ledger, "Normal", bcHolder.UtxoVM, consensus)
	return l, bcHolder
}

func TestDoTruncateTask(t *testing.T) {
	lk, holder := prepareLedgerKeeper(47101, "../data/netkeys/")
	lk.DoTruncateTask(holder.B1.GetBlockid())
	if holder.Ledger.GetMeta().GetTrunkHeight() == 2 {
		t.Error("TestDoTruncateTask truncate error")
	}
}

func TestGetBlockIdsWithGetHeadersMsg(t *testing.T) {
	lk, holder := prepareLedgerKeeper(47101, "../data/netkeys/")
	body := &pb.GetBlockIdsRequest{
		Count:   100,
		BlockId: holder.B0.GetBlockid(),
	}
	bodyBuf, _ := proto.Marshal(body)
	msg, _ := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion2, "xuper", "", xuper_p2p.XuperMessage_GET_BLOCKIDS, bodyBuf, xuper_p2p.XuperMessage_NONE)
	xmsg, err := lk.handleGetBlockIds(nil, msg)

	headerMsgBody := &pb.GetBlockIdsResponse{}
	err = proto.Unmarshal(xmsg.GetData().GetMsgInfo(), headerMsgBody)
	if err != nil {
		t.Error("TestGetBlockIdsWithGetHeadersMsg ErrUnmarshal")
	}
	blockIds := headerMsgBody.GetBlockIds()
	tipId := headerMsgBody.GetTipBlockId()
	if len(blockIds) == 0 || int64(len(blockIds)) > 100 {
		t.Error("TestGetBlockIdsWithGetHeadersMsg Internal Error", "tipId", global.F(tipId), "len", len(blockIds), "B0", global.F(holder.B0.GetBlockid()))
		return
	}
	if int64(len(blockIds)) == 100 {
		t.Log("TestGetBlockIdsWithGetHeadersMsg return Error")
		return
	}
	holder.Ledger.Close()
	holder.UtxoVM.Close()
}

func TestGetBlocksWithGetDataMsg(t *testing.T) {
	lk, holder := prepareLedgerKeeper(47101, "../data/netkeys/")
	body := &pb.GetBlocksRequest{
		BlockIds: [][]byte{holder.B1.GetBlockid(), holder.B2.GetBlockid()},
	}
	bodyBuf, _ := proto.Marshal(body)
	msg, err := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion2, "xuper", "", xuper_p2p.XuperMessage_GET_BLOCKS, bodyBuf, xuper_p2p.XuperMessage_NONE)
	lk.syncMsgChan <- msg
	returnMsg := <-lk.syncMsgChan
	msgType := returnMsg.GetHeader().GetType()
	if msgType != xuper_p2p.XuperMessage_GET_BLOCKS {
		t.Error("TestGetBlocksWithGetDataMsg test Internal error")
	}
	xmsg, err := lk.handleGetBlocks(nil, returnMsg)
	peerSyncMap := &map[string]*pb.InternalBlock{
		global.F(holder.B1.GetBlockid()): nil,
		global.F(holder.B2.GetBlockid()): nil,
	}
	t.Log("peerSyncMap LEN=", len(*peerSyncMap), "INFO=", *peerSyncMap)
	blocksMsgBody := &pb.GetBlocksResponse{}
	err = proto.Unmarshal(xmsg.GetData().GetMsgInfo(), blocksMsgBody)
	if err != nil {
		t.Error("TestGetBlocksWithGetDataMsg ErrUnmarshal")
	}
	if len(blocksMsgBody.GetBlocksInfo()) == 0 {
		t.Error("TestGetBlocksWithGetDataMsg ErrTargetDataNotFound")
	}
	blocks := blocksMsgBody.GetBlocksInfo()
	for _, block := range blocks {
		blockId := global.F(block.GetBlockid())
		mapValue, ok := (*peerSyncMap)[blockId]
		if !ok || mapValue != nil {
			t.Error("TestGetBlocksWithGetDataMsg KeyNotFound", "mapValue", mapValue, "ok", ok)
		}
		(*peerSyncMap)[blockId] = block
	}
	t.Log("peerSyncMap", peerSyncMap)
	holder.Ledger.Close()
	holder.UtxoVM.Close()
}

func TestHandleGetHeadersMsg(t *testing.T) {
	lk, holder := prepareLedgerKeeper(47101, "../data/netkeys/")
	t.Log("gBlk:", global.F(holder.B0.GetBlockid()), " nextBlk:", global.F(holder.B1.GetBlockid()), " nextNextBlk:", global.F(holder.B2.GetBlockid()))
	body := &pb.GetBlockIdsRequest{
		Count:   2,
		BlockId: holder.B0.GetBlockid(),
	}
	bodyBuf, _ := proto.Marshal(body)
	msg, _ := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion2, "xuper", "", xuper_p2p.XuperMessage_GET_BLOCKIDS, bodyBuf, xuper_p2p.XuperMessage_NONE)
	returnMsg, err := lk.handleGetBlockIds(nil, msg)
	if err != nil {
		t.Error("TestHandleGetHeadersMsg test Internal error: ", err.Error())
	}
	returnBody := &pb.GetBlockIdsResponse{}
	if err := proto.Unmarshal(returnMsg.GetData().GetMsgInfo(), returnBody); err != nil {
		t.Error("TestHandleGetHeadersMsg test Internal error: ", err.Error())
	}
	returnHeaders := returnBody.GetBlockIds()
	holder.Ledger.Close()
	holder.UtxoVM.Close()
	if len(returnHeaders) != 2 {
		t.Error("TestHandleGetHeadersMsg returnHeaders != 2, headers=", returnHeaders)
		return
	}
	if bytes.Compare(returnHeaders[0], holder.B1.GetBlockid()) != 0 || bytes.Compare(returnHeaders[1], holder.B2.GetBlockid()) != 0 {
		t.Error("TestHandleGetHeadersMsg return different block. ", " block1:", global.F(returnHeaders[0]), " block2:", global.F(returnHeaders[1]))
	}
	return
}

func TestHandleGetDataMsg(t *testing.T) {
	lk, holder := prepareLedgerKeeper(47101, "../data/netkeys/")
	body := &pb.GetBlocksRequest{
		BlockIds: [][]byte{holder.B2.GetBlockid(), holder.B1.GetBlockid()},
	}
	bodyBuf, _ := proto.Marshal(body)
	msg, _ := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion2, "xuper", "", xuper_p2p.XuperMessage_GET_BLOCKS, bodyBuf, xuper_p2p.XuperMessage_NONE)
	returnMsg, err := lk.handleGetBlocks(nil, msg)
	if err != nil {
		t.Error("TestHandleGetDataMsg test Internal error: ", err.Error())
	}
	returnBody := &pb.GetBlocksResponse{}
	if err := proto.Unmarshal(returnMsg.GetData().GetMsgInfo(), returnBody); err != nil {
		t.Error("TestHandleGetDataMsg test Internal error: ", err.Error())
	}
	blocks := returnBody.GetBlocksInfo()
	holder.Ledger.Close()
	holder.UtxoVM.Close()
	if len(blocks) != 2 {
		t.Error("TestHandleGetDataMsg blocks != 2, LEN=", len(blocks))
		//return
	}
	if bytes.Compare(blocks[0].GetBlockid(), holder.B2.GetBlockid()) != 0 {
		t.Error("TestHandleGetDataMsg BlocksInfo, 0=", global.F(blocks[0].GetBlockid()))
		return
	}
	if blocks[0].GetTxCount() != 1 {
		t.Error("TestHandleGetDataMsg GetTxCount=", blocks[0].GetTxCount())
	}
	return
}

func TestAssignTaskRandomly(t *testing.T) {
	targetPeers := []string{"NodeA", "NodeB", "NodeC", "NodeD", "NodeE", "NodeF"}
	headersList := [][]byte{
		[]byte{1}, []byte{2}, []byte{3}, []byte{4},
	}
	result, err := assignTaskRandomly(targetPeers, headersList)
	if len(result) == 0 {
		t.Error("assignTaskRandomly test error: NONE")
	}
	if err != nil {
		t.Error("assignTaskRandomly test error:", err.Error())
	}
}

func TestRandomPickPeers(t *testing.T) {
	randomNumber := int64(3)
	targetSyncBlocksPeers := new(sync.Map)
	targetSyncBlocksPeers.Store("NodeA", true)
	targetSyncBlocksPeers.Store("NodeB", true)
	targetSyncBlocksPeers.Store("NodeC", false)
	targetSyncBlocksPeers.Store("NodeD", true)
	targetSyncBlocksPeers.Store("NodeE", true)
	targetSyncBlocksPeers.Store("NodeF", false)
	result, err := randomPickPeers(randomNumber, targetSyncBlocksPeers)
	t.Log("LEN=", len(result))
	if err != nil || int64(len(result)) != randomNumber {
		t.Error("randomPickPeers test error:", err.Error())
	}

	randomNumber = int64(0)
	result, err = randomPickPeers(randomNumber, targetSyncBlocksPeers)
	t.Log("LEN=", len(result))
	if err != nil {
		t.Error("randomPickPeers test ZERO error")
	}
}

func TestGetValidPeersNumber(t *testing.T) {
	targetSyncBlocksPeers := new(sync.Map)
	targetSyncBlocksPeers.Store("NodeA", true)
	targetSyncBlocksPeers.Store("NodeB", true)
	targetSyncBlocksPeers.Store("NodeC", false)
	number := getValidPeersNumber(targetSyncBlocksPeers)
	if number == 0 {
		t.Error("getValidPeersNumber test ZERO error")
	}
}

func TestPickIndexes(t *testing.T) {
	list := []int{3, 1, 1, 5, 6, 7, 8}
	result := pickIndexes(int64(10), list)
	if len(result) != 4 {
		t.Error("TestPickIndexes test ZERO error")
	}
	for _, v := range result {
		if v != 0 && v != 1 && v != 2 && v != 3 {
			t.Error("TestPickIndexes test logic error")
		}
	}
}

func TestCheckAndComfirm(t *testing.T) {
	lk, holder := prepareLedgerKeeper(47101, "../data/netkeys/")
	tx := generateTx(false, holder.UtxoVM)
	b3, err := holder.Ledger.FormatFakeBlock([]*pb.Transaction{tx}, []byte(bobAddress), holder.PrivateKey, time.Now().UnixNano(), 3, 3, holder.B2.GetBlockid(), big.NewInt(0), 3)
	if err != nil {
		t.Error("TestCheckAndComfirm make fake newblock error")
		return
	}
	qc := &pb.QuorumCert{
		ProposalId:  b3.GetBlockid(),
		ProposalMsg: nil,
		ViewNumber:  3,
		Type:        pb.QCState_PREPARE,
		SignInfos:   &pb.QCSignInfos{},
	}
	signedBlock, err := holder.Ledger.FormatMinerBlock([]*pb.Transaction{tx}, []byte(bobAddress), holder.PrivateKey, time.Now().UnixNano(), 3, 3, holder.B2.GetBlockid(), 0, big.NewInt(0), qc, map[string]string{}, 3)
	if err != nil {
		t.Error("TestCheckAndComfirm make singed newblock error", "error", err)
		return
	}

	simpleBlock := &SimpleBlock{
		logid:         "TestCheckAndComfirm",
		internalBlock: signedBlock,
	}
	err, trunkSwitch := lk.checkAndConfirm(true, simpleBlock)
	holder.Ledger.Close()
	holder.UtxoVM.Close()
	if err != nil {
		t.Error("TestCheckAndComfirm", "error", err)
		return
	}
	if !signedBlock.InTrunk && trunkSwitch {
		t.Error("TestCheckAndComfirm::invalid switch")
		return
	}
}

func TestConfirmBlocks(t *testing.T) {
	lk, holder := prepareLedgerKeeper(47101, "../data/netkeys/")

	tx := generateTx(false, holder.UtxoVM)
	b3, err := holder.Ledger.FormatFakeBlock([]*pb.Transaction{tx}, []byte(bobAddress), holder.PrivateKey, time.Now().UnixNano(), 3, 3, holder.B2.GetBlockid(), big.NewInt(0), 3)
	if err != nil {
		t.Error("TestConfirmForkingBlock make fake newblock error")
		return
	}
	qc := &pb.QuorumCert{
		ProposalId:  b3.GetBlockid(),
		ProposalMsg: nil,
		ViewNumber:  3,
		Type:        pb.QCState_PREPARE,
		SignInfos:   &pb.QCSignInfos{},
	}
	signedBlock, err := holder.Ledger.FormatMinerBlock([]*pb.Transaction{tx}, []byte(bobAddress), holder.PrivateKey,
		time.Now().UnixNano(), 3, 3, holder.B2.GetBlockid(), 0, big.NewInt(0), qc, map[string]string{}, 3)
	if err != nil {
		t.Error("TestConfirmAppendingBlock make singed newblock error", "error", err)
		return
	}
	tmpSlice := []*SimpleBlock{&SimpleBlock{
		logid:         "TestConfirmBlocks",
		internalBlock: signedBlock,
	}}
	newBeginId, err := lk.confirmBlocks(&global.XContext{Timer: global.NewXTimer()}, tmpSlice, true)
	if bytes.Compare(newBeginId, signedBlock.GetBlockid()) != 0 {
		t.Error("TestConfirmBlocks", "error", err)
	}

	// 模拟分叉情况
	tx = generateTx(false, holder.UtxoVM)
	b3, _ = holder.Ledger.FormatFakeBlock([]*pb.Transaction{tx}, []byte(bobAddress), holder.PrivateKey, time.Now().UnixNano(), 3, 3, holder.B2.GetBlockid(), big.NewInt(0), 3)
	qc = &pb.QuorumCert{
		ProposalId:  b3.GetBlockid(),
		ProposalMsg: nil,
		ViewNumber:  3,
		Type:        pb.QCState_PREPARE,
		SignInfos:   &pb.QCSignInfos{},
	}
	signedBlock, _ = holder.Ledger.FormatMinerBlock([]*pb.Transaction{tx}, []byte(bobAddress), holder.PrivateKey,
		time.Now().UnixNano(), 3, 3, holder.B2.GetBlockid(), 0, big.NewInt(0), qc, map[string]string{}, 3)
	tmpSlice = []*SimpleBlock{
		&SimpleBlock{
			logid:         "TestConfirmBlocks",
			internalBlock: signedBlock,
		}}
	newBeginId, err = lk.confirmBlocks(&global.XContext{Timer: global.NewXTimer()}, tmpSlice, true)
	if bytes.Compare(newBeginId, signedBlock.GetBlockid()) != 0 {
		t.Error("TestConfirmBlocks", "error", err)
	}
}

func TestPushBack(t *testing.T) {
	l := NewTasksList()
	task := &LedgerTask{}
	l.PushBack(task)
	if !l.PushBack(task) {
		t.Error("TestPushBack::repeat action")
	}
}

func TestFix(t *testing.T) {
	l := NewTasksList()
	task1 := &LedgerTask{
		targetHeight: 1,
	}
	task2 := &LedgerTask{
		targetHeight: 2,
	}
	task3 := &LedgerTask{
		targetHeight: 3,
	}
	l.PushBack(task1)
	l.PushBack(task2)
	l.PushBack(task3)
	l.fix(2)
	if l.Len() == 3 {
		t.Error("TestFix::fix failed.")
	}
}

func TestPutAndGet(t *testing.T) {
	slog := log.New("module", "syncnode")
	slog.SetHandler(log.StreamHandler(os.Stderr, log.LogfmtFormat()))
	stm := newSyncTaskManager(slog)
	task1 := &LedgerTask{
		targetHeight: 1,
		action:       Syncing,
	}
	if !stm.Put(task1) {
		t.Error("TestPutAndGet::put failed.")
	}
	task2 := &LedgerTask{
		targetHeight: 2,
		action:       Appending,
	}
	if !stm.Put(task2) {
		t.Error("TestPutAndGet::put failed.")
	}
	if stm.Pop().GetAction() != Appending {
		t.Error("TestPutAndGet::get failed.")
	}
}

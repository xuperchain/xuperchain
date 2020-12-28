package ledger

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	crypto_client "github.com/xuperchain/xuperchain/core/crypto/client"
	"github.com/xuperchain/xuperchain/core/pb"
	"github.com/xuperchain/xuperchain/core/utxo/txhash"
)

const AliceAddress = "dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN"
const AlicePubkey = `{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571}`
const AlicePrivateKey = `{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571,"D":29079635126530934056640915735344231956621504557963207107451663058887647996601}`
const BobAddress = "WNWk3ekXeM5M2232dY2uCJmEqWhfQiDYT"

const DefaultKvEngine = "default"

func TestOpenClose(t *testing.T) {
	workSpace, dirErr := ioutil.TempDir("/tmp", "")
	if dirErr != nil {
		t.Fatal(dirErr)
	}
	os.RemoveAll(workSpace)
	defer os.RemoveAll(workSpace)
	ledger, err := NewLedger(workSpace, nil, nil, DefaultKvEngine, crypto_client.CryptoTypeDefault)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ledger)
	ledger.Close()
}

func TestBasicFunc(t *testing.T) {
	workSpace, dirErr := ioutil.TempDir("/tmp", "")
	if dirErr != nil {
		t.Fatal(dirErr)
	}
	os.RemoveAll(workSpace)
	defer os.RemoveAll(workSpace)
	ledger, err := NewLedger(workSpace, nil, nil, DefaultKvEngine, crypto_client.CryptoTypeDefault)
	if err != nil {
		t.Fatal(err)
	}
	/*
		cryptoClient, err := crypto_client.CreateCryptoClient(crypto_client.CryptoTypeDefault)
		if err != nil {
			t.Fatal(err)
		}*/
	t1 := &pb.Transaction{}
	t2 := &pb.Transaction{}
	t1.TxOutputs = append(t1.TxOutputs, &pb.TxOutput{Amount: []byte("888"), ToAddr: []byte(BobAddress)})
	t1.Coinbase = true
	t1.Desc = []byte(`{"maxblocksize" : "128"}`)
	t1.Txid, _ = txhash.MakeTransactionID(t1)
	t2.TxInputs = append(t2.TxInputs, &pb.TxInput{RefTxid: t1.Txid, RefOffset: 0, FromAddr: []byte(AliceAddress)})
	t2.Txid, _ = txhash.MakeTransactionID(t2)
	ecdsaPk, pkErr := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	t.Logf("pkSk: %v", ecdsaPk)
	if pkErr != nil {
		t.Fatal("fail to generate publice/private key")
	}
	block, err := ledger.FormatRootBlock([]*pb.Transaction{t1})
	if err != nil {
		t.Fatalf("format block fail, %v", err)
	}
	t.Logf("block id %x", block.Blockid)
	confirmStatus := ledger.ConfirmBlock(block, true)
	if !confirmStatus.Succ {
		t.Fatal("confirm block fail")
	}
	hasTx, _ := ledger.HasTransaction(t1.Txid)
	if !hasTx {
		t.Fatal("genesis tx not exist")
	}
	confirmTwice := ledger.ConfirmBlock(block, true)
	if confirmTwice.Succ {
		t.Fatal("confirm should fail, unexpected")
	}
	t1 = &pb.Transaction{}
	t2 = &pb.Transaction{}
	t1.TxOutputs = append(t1.TxOutputs, &pb.TxOutput{Amount: []byte("666"), ToAddr: []byte(BobAddress)})
	t1.Txid, _ = txhash.MakeTransactionID(t1)
	t2.TxInputs = append(t2.TxInputs, &pb.TxInput{RefTxid: t1.Txid, RefOffset: 0, FromAddr: []byte(AliceAddress)})
	t2.Txid, _ = txhash.MakeTransactionID(t2)
	block2, err := ledger.FormatBlock([]*pb.Transaction{t1, t2},
		[]byte("xchain-Miner-222222"),
		ecdsaPk,
		223456789,
		0,
		0,
		block.Blockid, big.NewInt(0),
	)
	t.Logf("bolock2 id %x", block2.Blockid)
	confirmStatus = ledger.ConfirmBlock(block2, false)
	if !confirmStatus.Succ {
		t.Fatal("confirm block fail 2")
	}

	blockCopy, readErr := ledger.QueryBlock(block.Blockid)
	if readErr != nil {
		t.Fatalf("read block fail, %v", readErr)
	} else {
		t.Logf("block detail: %v", proto.MarshalTextString(blockCopy))
	}
	blockByHeight, _ := ledger.QueryBlockByHeight(block.Height)
	if string(blockByHeight.Blockid) != string(blockCopy.Blockid) {
		t.Fatalf("query block by height failed")
	}
	lastBlock, _ := ledger.QueryLastBlock()
	t.Logf("query last block %x", lastBlock.Blockid)
	t.Logf("block1 next hash %x", blockCopy.NextHash)
	blockCopy2, readErr2 := ledger.QueryBlock(blockCopy.NextHash)
	if readErr2 != nil {
		t.Fatalf("read block fail, %v", readErr2)
	} else {
		t.Logf("block2 detail: %v", proto.MarshalTextString(blockCopy2))
	}
	txCopy, txErr := ledger.QueryTransaction(t1.Txid)
	if txErr != nil {
		t.Fatalf("query tx fail, %v", txErr)
	}
	t.Logf("tx detail: %v", txCopy)
	maxBlockSize := ledger.GetMaxBlockSize()
	if maxBlockSize != (128 << 20) {
		t.Fatalf("maxBlockSize unexpected: %v", maxBlockSize)
	}
	/*
		upErr := ledger.UpdateMaxBlockSize(0, ledger.baseDB.NewBatch())

		if upErr == nil {
			t.Fatal("unexpected")
		}
		ledger.UpdateMaxBlockSize(123, ledger.baseDB.NewBatch())
		if ledger.GetMaxBlockSize() != 123 {
			t.Fatalf("unexpected block size, %d", ledger.GetMeta().MaxBlockSize)
		}*/

	// coinbase txs > 1
	t1 = &pb.Transaction{}
	t2 = &pb.Transaction{}
	t1.TxOutputs = append(t1.TxOutputs, &pb.TxOutput{Amount: []byte("666"), ToAddr: []byte(BobAddress)})
	t1.Coinbase = true
	t1.Desc = []byte("{}")
	t1.Txid, _ = txhash.MakeTransactionID(t1)
	t2.TxInputs = append(t2.TxInputs, &pb.TxInput{RefTxid: t1.Txid, RefOffset: 0, FromAddr: []byte(AliceAddress)})
	t2.Coinbase = true
	t2.Txid, _ = txhash.MakeTransactionID(t2)
	block3, err := ledger.FormatBlock([]*pb.Transaction{t1, t2},
		[]byte("xchain-Miner-222222"),
		ecdsaPk,
		223456789,
		0,
		0,
		block.Blockid, big.NewInt(0),
	)
	t.Logf("bolock3 id %x", block3.Blockid)
	confirmStatus = ledger.ConfirmBlock(block3, false)
	if confirmStatus.Succ {
		t.Fatal("The num of coinbase txs error")
	}

	ledger.Close()
}

func TestSplitFunc(t *testing.T) {
	workSpace, dirErr := ioutil.TempDir("/tmp", "")
	if dirErr != nil {
		t.Fatal(dirErr)
	}
	os.RemoveAll(workSpace)
	defer os.RemoveAll(workSpace)
	ledger, err := NewLedger(workSpace, nil, nil, DefaultKvEngine, crypto_client.CryptoTypeDefault)
	if err != nil {
		t.Fatal(err)
	}
	/*
		cryptoClient, err := crypto_client.CreateCryptoClient(crypto_client.CryptoTypeDefault)
		if err != nil {
			t.Fatal(err)
		}*/
	t1 := &pb.Transaction{}
	t2 := &pb.Transaction{}
	t1.TxOutputs = append(t1.TxOutputs, &pb.TxOutput{Amount: []byte("666"), ToAddr: []byte(BobAddress)})
	t1.Coinbase = true
	t1.Desc = []byte("{}")
	t1.Txid, _ = txhash.MakeTransactionID(t1)
	t2.TxInputs = append(t2.TxInputs, &pb.TxInput{RefTxid: t1.Txid, RefOffset: 0, FromAddr: []byte(AliceAddress)})
	t2.Txid, _ = txhash.MakeTransactionID(t2)
	ecdsaPk, pkErr := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	t.Logf("pkSk: %v", ecdsaPk)
	if pkErr != nil {
		t.Fatal("fail to generate publice/private key")
	}
	block, err := ledger.FormatBlock([]*pb.Transaction{t1, t2},
		[]byte("xchain-Miner-1"),
		ecdsaPk,
		123456789,
		0,
		0,
		[]byte{}, big.NewInt(0),
	)
	if err != nil {
		t.Fatalf("format block fail, %v", err)
	}
	t.Logf("block id %x", block.Blockid)
	confirmStatus := ledger.ConfirmBlock(block, true)
	if !confirmStatus.Succ {
		t.Fatal("confirm block fail", confirmStatus.Error)
	}
	t1 = &pb.Transaction{}
	t2 = &pb.Transaction{}
	t1.TxOutputs = append(t1.TxOutputs, &pb.TxOutput{Amount: []byte("999"), ToAddr: []byte(BobAddress)})
	t1.Txid, _ = txhash.MakeTransactionID(t1)
	t2.TxInputs = append(t2.TxInputs, &pb.TxInput{RefTxid: t1.Txid, RefOffset: 0, FromAddr: []byte(AliceAddress)})
	t2.Txid, _ = txhash.MakeTransactionID(t2)
	block2, err := ledger.FormatBlock([]*pb.Transaction{t1, t2},
		[]byte("xchain-Miner-222222"),
		ecdsaPk,
		223456789,
		0,
		0,
		block.Blockid, big.NewInt(0),
	)
	t.Logf("bolock2 id %x", block2.Blockid)
	confirmStatus = ledger.ConfirmBlock(block2, false)
	if !confirmStatus.Succ {
		t.Fatal("confirm block fail 2", confirmStatus.Error)
	}
	//伪造一个新的txid
	t1.Txid = append(t1.Txid, []byte("a")...)
	t2.Txid = append(t2.Txid, []byte("b")...)

	block3, err := ledger.FormatBlock([]*pb.Transaction{t1, t2},
		[]byte("xchain-Miner-222223"),
		ecdsaPk,
		2234567899,
		0,
		0,
		block.Blockid, big.NewInt(0),
	)
	t.Logf("bolock3 id %x", block3.Blockid)
	confirmStatus = ledger.ConfirmBlock(block3, false)
	if !confirmStatus.Succ {
		t.Fatal("confirm block fail 3")
	}

	block4, err := ledger.FormatBlock([]*pb.Transaction{t1, t2},
		[]byte("xchain-Miner-222224"),
		ecdsaPk,
		2234567999,
		0,
		0,
		block3.Blockid, big.NewInt(0),
	)
	t.Logf("bolock4 id %x", block4.Blockid)
	confirmStatus = ledger.ConfirmBlock(block4, false)
	if !confirmStatus.Succ {
		t.Fatal("confirm block fail 4")
	}
	dumpLayer, dumpErr := ledger.Dump()
	if dumpErr != nil {
		t.Fatal("dump ledger fail")
	} else {
		for height, blocks := range dumpLayer {
			t.Log("Height", height, "blocks", blocks)
		}
	}

	totalProperty := ledger.GetEstimatedTotal()
	t.Log("ledger total property", totalProperty)
	gensisBlock := ledger.GetGenesisBlock()
	if gensisBlock != nil {
		t.Log("gensisBlock ", gensisBlock)
	} else {
		t.Fatal("gensis_block is expected to be not nil")
	}
	curBlockid := block4.Blockid
	destBlockid := block2.Blockid

	undoBlocks, todoBlocks, findErr := ledger.FindUndoAndTodoBlocks(curBlockid, destBlockid)
	if findErr != nil {
		t.Fatal("fail to to find common parent of two blocks", "destBlockid",
			fmt.Sprintf("%x", destBlockid), "latestBlockid", fmt.Sprintf("%x", curBlockid))
	} else {
		t.Log("print undo block")
		for _, undoBlk := range undoBlocks {
			t.Log(undoBlk.Blockid)
		}
		t.Log("print todo block")
		for _, todoBlk := range todoBlocks {
			t.Log(todoBlk.Blockid)
		}
	}
	// test for IsTxInTrunk
	// t1 is in block3 and block3 is in branch
	if isOnChain := ledger.IsTxInTrunk(t1.Txid); !isOnChain {
		t.Fatal("expect true, got ", isOnChain)
	}
	// test for QueryBlockHeader
	blkHeader, err := ledger.QueryBlockHeader(block4.Blockid)
	if err != nil {
		t.Fatal("Query Block error")
	} else {
		t.Log("blkHeader ", blkHeader)
	}
	// test for ExistBlock
	if exist := ledger.ExistBlock(block3.Blockid); !exist {
		t.Fatal("expect block3 exist, got ", exist)
	}

	ledger.Close()
}

func TestTruncate(t *testing.T) {
	workSpace, dirErr := ioutil.TempDir("/tmp", "")
	if dirErr != nil {
		t.Fatal(dirErr)
	}
	os.RemoveAll(workSpace)
	defer os.RemoveAll(workSpace)
	ledger, err := NewLedger(workSpace, nil, nil, DefaultKvEngine, crypto_client.CryptoTypeDefault)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ledger.meta)

	t1 := &pb.Transaction{}
	t2 := &pb.Transaction{}
	t1.TxOutputs = append(t1.TxOutputs, &pb.TxOutput{Amount: []byte("888"), ToAddr: []byte(BobAddress)})
	t1.Coinbase = true
	t1.Desc = []byte(`{"maxblocksize" : "128"}`)
	t1.Txid, _ = txhash.MakeTransactionID(t1)
	t2.TxInputs = append(t2.TxInputs, &pb.TxInput{RefTxid: t1.Txid, RefOffset: 0, FromAddr: []byte(AliceAddress)})
	t2.Txid, _ = txhash.MakeTransactionID(t2)
	ecdsaPk, pkErr := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	t.Logf("pkSk: %v", ecdsaPk)
	if pkErr != nil {
		t.Fatal("fail to generate publice/private key")
	}
	block1, err := ledger.FormatRootBlock([]*pb.Transaction{t1})
	if err != nil {
		t.Fatalf("format block fail, %v", err)
	}
	t.Logf("block1 id %x", block1.Blockid)
	confirmStatus := ledger.ConfirmBlock(block1, true)
	if !confirmStatus.Succ {
		t.Fatal("confirm block fail")
	}

	t1 = &pb.Transaction{}
	t2 = &pb.Transaction{}
	t1.TxOutputs = append(t1.TxOutputs, &pb.TxOutput{Amount: []byte("666"), ToAddr: []byte(BobAddress)})
	t1.Txid, _ = txhash.MakeTransactionID(t1)
	t2.TxInputs = append(t2.TxInputs, &pb.TxInput{RefTxid: t1.Txid, RefOffset: 0, FromAddr: []byte(AliceAddress)})
	t2.Txid, _ = txhash.MakeTransactionID(t2)
	//block2
	block2, err := ledger.FormatBlock([]*pb.Transaction{t1, t2},
		[]byte("xchain-Miner-222222"),
		ecdsaPk,
		223456789,
		0,
		0,
		block1.Blockid, big.NewInt(0),
	)
	t.Logf("bolock2 id %x", block2.Blockid)
	confirmStatus = ledger.ConfirmBlock(block2, false)
	if !confirmStatus.Succ {
		t.Fatal("confirm block fail 2")
	}

	//block2 <- block3
	block3, err := ledger.FormatBlock([]*pb.Transaction{&pb.Transaction{Txid: []byte("dummy1")}},
		[]byte("xchain-Miner-333333"),
		ecdsaPk,
		223456790,
		0,
		0,
		block2.Blockid, big.NewInt(0),
	)
	confirmStatus = ledger.ConfirmBlock(block3, false)
	if !confirmStatus.Succ {
		t.Fatal("confirm block fail 2")
	}

	//block2 <- block4
	block4, err := ledger.FormatBlock([]*pb.Transaction{&pb.Transaction{Txid: []byte("dummy2")}},
		[]byte("xchain-Miner-444444"),
		ecdsaPk,
		223456791,
		0,
		0,
		block2.Blockid, big.NewInt(0),
	)
	confirmStatus = ledger.ConfirmBlock(block4, false)
	if !confirmStatus.Succ {
		t.Fatal("confirm block fail 2")
	}

	layers, _ := ledger.Dump()
	t.Log("Before truncate", layers)
	if len(layers) != 3 {
		t.Fatal("layers unexpected", len(layers))
	}
	err = ledger.Truncate(block1.Blockid)
	if err != nil {
		t.Fatalf("Trucate error")
	}
	layers, _ = ledger.Dump()
	if len(layers) != 1 {
		t.Fatal("layers unexpected", len(layers))
	}
	t.Log("After truncate", layers)

	metaBuf, _ := ledger.metaTable.Get([]byte(""))
	_ = proto.Unmarshal(metaBuf, ledger.meta)
	t.Log(ledger.meta)

	ledger.Close()
}

func TestBlockHeader(t *testing.T) {
	path := os.Getenv("DB")
	startstr := os.Getenv("START")
	limitstr := os.Getenv("LIMIT")
	limit, _ := strconv.Atoi(limitstr)
	ledger, err := NewLedger(path, nil, nil, DefaultKvEngine, crypto_client.CryptoTypeDefault)
	if err != nil {
		return
	}
	fmt.Printf("start:%s\n", startstr)
	fmt.Printf("limit:%s\n", limitstr)
	startBlockid, _ := hex.DecodeString(startstr)
	blockid := startBlockid
	tstart := time.Now()
	sizem := make(map[int]int)
	for i := 0; i < limit; i++ {
		// blk, err := ledger.QueryBlockHeader(blockid)
		blk, err := ledger.QueryBlock(blockid)
		if err != nil {
			blockid = startBlockid
			continue
		}
		blockid = blk.GetPreHash()
		size := proto.Size(blk)
		for _, s := range []int{100, 50, 25, 10, 5, 2, 1} {
			if size < (1 << 20 * s) {
				sizem[s]++
			}
		}
	}
	fmt.Printf("%v\n", sizem)
	fmt.Printf("used:%s\n", time.Now().Sub(tstart))
}

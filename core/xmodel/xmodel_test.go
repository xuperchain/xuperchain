package xmodel

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	log "github.com/xuperchain/log15"
	crypto_client "github.com/xuperchain/xuperchain/core/crypto/client"
	"github.com/xuperchain/xuperchain/core/ledger"
	"github.com/xuperchain/xuperchain/core/pb"
)

const DefaultKvEngine = "default"

var logger log.Logger

func TestInitLogger(t *testing.T) {
	logger = log.New("module", "xmodel")
	logger.SetHandler(log.StreamHandler(os.Stderr, log.LogfmtFormat()))
}

func TestBaiscFunc(t *testing.T) {
	//path := "/tmp/xmodel_test"
	path, pathErr := ioutil.TempDir("/tmp", "")
	if pathErr != nil {
		t.Fatal(pathErr)
	}
	os.RemoveAll(path)
	defer os.RemoveAll(path)
	ledger, err := ledger.NewLedger(path, nil, nil, DefaultKvEngine, crypto_client.CryptoTypeDefault)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ledger)
	stateDB, err := openDB(path+"/utxo_vm", logger)
	if err != nil {
		t.Fatal(err)
	}
	xModel, err := NewXuperModel(ledger, stateDB, logger)
	if err != nil {
		t.Fatal(err)
	}
	verData, err := xModel.Get("bucket1", []byte("hello"))
	logger.Info("test xmodel get", "verData", verData, "err", err)
	if !IsEmptyVersionedData(verData) {
		t.Fatal("unexpected")
	}
	tx1 := &pb.Transaction{
		Txid: []byte("Tx1"),
		TxInputsExt: []*pb.TxInputExt{
			&pb.TxInputExt{
				Bucket: "bucket1",
				Key:    []byte("hello"),
			},
		},
		TxOutputsExt: []*pb.TxOutputExt{
			&pb.TxOutputExt{
				Bucket: "bucket1",
				Key:    []byte("hello"),
				Value:  []byte("you are the best!"),
			},
		},
	}
	batch := stateDB.NewBatch()
	err = xModel.DoTx(tx1, batch)
	if err != nil {
		t.Fatal(err)
	}
	saveUnconfirmTx(tx1, batch)
	err = batch.Write()
	if err != nil {
		t.Fatal(err)
	}
	verData, err = xModel.Get("bucket1", []byte("hello"))
	logger.Info("afetr dotx, test xmodel get", "verData", verData, "err", err)
	if GetVersion(verData) != fmt.Sprintf("%x_%d", "Tx1", 0) {
		t.Fatal("unexpected", GetVersion(verData))
	}
	tx2 := &pb.Transaction{
		Txid: []byte("Tx2"),
		TxInputsExt: []*pb.TxInputExt{
			&pb.TxInputExt{
				Bucket:    "bucket1",
				Key:       []byte("hello"),
				RefTxid:   []byte("Tx1"),
				RefOffset: 0,
			},
			&pb.TxInputExt{
				Bucket: "bucket1",
				Key:    []byte("world"),
			},
		},
		TxOutputsExt: []*pb.TxOutputExt{
			&pb.TxOutputExt{
				Bucket: "bucket1",
				Key:    []byte("hello"),
				Value:  []byte("\x00"),
			},
			&pb.TxOutputExt{
				Bucket: "bucket1",
				Key:    []byte("world"),
				Value:  []byte("world is full of love!"),
			},
		},
	}
	batch2 := stateDB.NewBatch()
	err = xModel.DoTx(tx2, batch2)
	if err != nil {
		t.Fatal(err)
	}
	saveUnconfirmTx(tx2, batch2)
	err = batch2.Write()
	if err != nil {
		t.Fatal(err)
	}
	verData, err = xModel.Get("bucket1", []byte("hello"))
	logger.Info("afetr dotx again, test xmodel get", "verData", verData, "err", err)
	if GetVersion(verData) != fmt.Sprintf("%x_%d", "Tx2", 0) {
		t.Fatal("unexpected", GetVersion(verData))
	}
	iter, err := xModel.Select("bucket1", []byte(""), []byte("\xff"))
	defer iter.Release()
	validKvCount := 0
	for iter.Next() {
		t.Logf("iter:  data %v, key: %s\n", iter.Data(), iter.Key())
		validKvCount++
	}
	if validKvCount != 1 {
		t.Fatal("unexpected", validKvCount)
	}
	ledger.Close()
}

package xmodel

import (
	"os"
	"testing"

	log "github.com/xuperchain/log15"
	crypto_client "github.com/xuperchain/xuperchain/core/crypto/client"
	"github.com/xuperchain/xuperchain/core/kv/kvdb"
	"github.com/xuperchain/xuperchain/core/ledger"
	"github.com/xuperchain/xuperchain/core/pb"
)

const pathDB = "/tmp/xmodel_cache_test"

func prepareData(xModel *XModel, stateDB kvdb.Database, t *testing.T) {
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
	err := xModel.DoTx(tx1, batch)
	if err != nil {
		t.Fatal(err)

	}
	saveUnconfirmTx(tx1, batch)
	err = batch.Write()
	if err != nil {
		t.Fatal(err)

	}

	tx3 := &pb.Transaction{
		Txid: []byte("Tx3"),
		TxInputsExt: []*pb.TxInputExt{
			&pb.TxInputExt{
				Bucket: "bucket1",
				Key:    []byte("hello7"),
			},
		},
		TxOutputsExt: []*pb.TxOutputExt{
			&pb.TxOutputExt{
				Bucket: "bucket1",
				Key:    []byte("hello7"),
				Value:  []byte("you are the best!"),
			},
		},
	}
	batch = stateDB.NewBatch()
	err = xModel.DoTx(tx3, batch)
	if err != nil {
		t.Fatal(err)

	}
	saveUnconfirmTx(tx3, batch)
	err = batch.Write()
	if err != nil {
		t.Fatal(err)

	}

	tx4 := &pb.Transaction{
		Txid: []byte("Tx4"),
		TxInputsExt: []*pb.TxInputExt{
			&pb.TxInputExt{
				Bucket: "bucket1",
				Key:    []byte("hello8"),
			},
		},
		TxOutputsExt: []*pb.TxOutputExt{
			&pb.TxOutputExt{
				Bucket: "bucket1",
				Key:    []byte("hello8"),
				Value:  []byte("you are the best!"),
			},
		},
	}
	batch = stateDB.NewBatch()
	err = xModel.DoTx(tx4, batch)
	if err != nil {
		t.Fatal(err)

	}
	saveUnconfirmTx(tx4, batch)
	err = batch.Write()
	if err != nil {
		t.Fatal(err)

	}
}

func prepareXmodel(t *testing.T) *XModel {
	logger := log.New("module", "xmodel")
	logger.SetHandler(log.StreamHandler(os.Stderr, log.LogfmtFormat()))
	remove(pathDB)
	leg, err := ledger.NewLedger(pathDB, nil, nil, DefaultKvEngine, crypto_client.CryptoTypeDefault)
	if err != nil {
		t.Fatal(err)
	}
	stateDB, err := openDB(pathDB+"/utxo_vm", logger)
	if err != nil {
		t.Fatal(err)
	}
	xModel, err := NewXuperModel(leg, stateDB, logger)
	if err != nil {
		t.Fatal(err)
	}
	prepareData(xModel, stateDB, t)
	return xModel
}

func remove(path string) {
	os.RemoveAll(path)
}

func prepareXModelCache(t *testing.T) *XMCache {
	xmodel := prepareXmodel(t)
	mc, err := NewXModelCache(xmodel, nil)
	if err != nil {
		t.Fatal(err)
	}
	return mc
}

func TestNewXModelCache(t *testing.T) {
	defer remove(pathDB)
	xmodel := prepareXmodel(t)
	_, err := NewXModelCache(xmodel, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = NewXModelCache(xmodel, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestXModelCacheGet(t *testing.T) {
	defer remove(pathDB)
	mc := prepareXModelCache(t)
	// 测试基本操作
	vd, _ := mc.Get("bucket1", []byte("hello"))
	if vd == nil {
		t.Error("vd can not be null!")
	}
	t.Log(vd)

	err := mc.Put("bucket1", []byte("hello"), []byte("change"))
	if err != nil {
		t.Error(err)
	}
	vd, err = mc.Get("bucket1", []byte("hello"))
	t.Log(vd)
	vd2, err := mc.model.Get("bucket1", []byte("hello"))
	if string(vd.GetPureData().GetValue()) == string(vd2.GetPureData().GetValue()) {
		t.Error("value can not be equal")
	}

	err = mc.Del("bucket1", []byte("hello"))
	if err != nil {
		t.Error(err)
	}
	vd, err = mc.Get("bucket1", []byte("hello"))
	t.Log(vd, err)
	if err != ErrHasDel {
		t.Error("The key has been del")
	}
	vd2, err = mc.model.Get("bucket1", []byte("hello"))
	t.Log(vd2, err)

	err = mc.Put("bucket1", []byte("hello"), []byte("change"))
	if err != nil {
		t.Error(err)
	}

	err = mc.Del("bucket1", []byte("hello"))
	if err != nil {
		t.Error(err)
	}

	// 测试读写集
	rs, ws, err := mc.GetRWSets()
	if len(rs) != 1 || len(ws) != 1 {
		t.Error("GetRWSets error!")
	}

	// 测试迭代器
	mc.isPenetrate = false
	mc.Get("bucket1", []byte("hello2"))
	mc.Put("bucket1", []byte("hello3"), []byte("Xuper chain is the best blockchain!"))
	mc.Put("bucket1", []byte("hello5"), []byte("Xuper chain is the best blockchain!"))
	mc.Put("bucket1", []byte("hello4"), []byte("Xuper chain is the best blockchain!"))
	iter, err := mc.Select("bucket1", []byte("hello"), []byte("i"))
	if err != nil {
		t.Error("Get iterator error")
	}

	keys := []string{}
	for iter.Next() {
		t.Log(string(iter.Key()))
		keys = append(keys, string(iter.Key()))
	}

	if len(keys) != 3 {
		t.Error("Iterator error", keys)
	}
	iter.Release()

	mc.isPenetrate = true
	mc.Put("bucket1", []byte("hello2"), []byte("change"))
	iter2, err := mc.Select("bucket1", []byte("hello"), []byte("i"))
	keys2 := []string{}

	for iter2.Next() {
		t.Log(string(iter2.Key()))
		keys2 = append(keys2, string(iter2.Key()))
	}
	if len(keys2) != 6 {
		t.Error("Iterator error", keys2)
	}
}

func TestIsContractUtxoEffective(t *testing.T) {
	testCases := map[string]struct {
		conTxInputs  []*pb.TxInput
		conTxOutputs []*pb.TxOutput
		tx           *pb.Transaction
		res          bool
	}{
		"test check ok": {
			conTxInputs: []*pb.TxInput{
				&pb.TxInput{
					RefTxid:   []byte("txInput1_RefTxid"),
					RefOffset: 1,
					FromAddr:  []byte("txInput1_FromAddr"),
				},
				&pb.TxInput{
					RefTxid:   []byte("txInput2_RefTxid"),
					RefOffset: 2,
					FromAddr:  []byte("txInput2_FromAddr"),
				},
				&pb.TxInput{
					RefTxid:   []byte("txInput3_RefTxid"),
					RefOffset: 3,
					FromAddr:  []byte("txInput3_FromAddr"),
				},
			},
			conTxOutputs: []*pb.TxOutput{
				&pb.TxOutput{
					Amount: []byte("txOutput1_Amount"),
					ToAddr: []byte("txOutput1_ToAddr"),
				},
				&pb.TxOutput{
					Amount: []byte("txOutput2_Amount"),
					ToAddr: []byte("txOutput2_ToAddr"),
				},
				&pb.TxOutput{
					Amount: []byte("txOutput2_Amount"),
					ToAddr: []byte("txOutput2_ToAddr"),
				},
			},
			tx: &pb.Transaction{
				TxInputs: []*pb.TxInput{
					&pb.TxInput{
						RefTxid:   []byte("txInput1_RefTxid"),
						RefOffset: 1,
						FromAddr:  []byte("txInput1_FromAddr"),
					},
					&pb.TxInput{
						RefTxid:   []byte("txInput2_RefTxid"),
						RefOffset: 2,
						FromAddr:  []byte("txInput2_FromAddr"),
					},
					&pb.TxInput{
						RefTxid:   []byte("txInput3_RefTxid"),
						RefOffset: 3,
						FromAddr:  []byte("txInput3_FromAddr"),
					},
					&pb.TxInput{
						RefTxid:   []byte("txInput4_RefTxid"),
						RefOffset: 4,
						FromAddr:  []byte("txInput4_FromAddr"),
					},
				},
				TxOutputs: []*pb.TxOutput{
					&pb.TxOutput{
						Amount: []byte("txOutput1_Amount"),
						ToAddr: []byte("txOutput1_ToAddr"),
					},
					&pb.TxOutput{
						Amount: []byte("txOutput2_Amount"),
						ToAddr: []byte("txOutput2_ToAddr"),
					},
					&pb.TxOutput{
						Amount: []byte("txOutput2_Amount"),
						ToAddr: []byte("txOutput2_ToAddr"),
					},
					&pb.TxOutput{
						Amount: []byte("txOutput3_Amount"),
						ToAddr: []byte("txOutput3_ToAddr"),
					},
				},
			},
			res: true,
		},
		"test check failed1": {
			conTxInputs: []*pb.TxInput{
				&pb.TxInput{
					RefTxid:   []byte("txInput1_RefTxid"),
					RefOffset: 1,
					FromAddr:  []byte("txInput1_FromAddr"),
				},
				&pb.TxInput{
					RefTxid:   []byte("txInput2_RefTxid"),
					RefOffset: 2,
					FromAddr:  []byte("txInput2_FromAddr"),
				},
				&pb.TxInput{
					RefTxid:   []byte("txInput3_RefTxid"),
					RefOffset: 3,
					FromAddr:  []byte("txInput3_FromAddr"),
				},
			},
			conTxOutputs: []*pb.TxOutput{
				&pb.TxOutput{
					Amount: []byte("txOutput1_Amount"),
					ToAddr: []byte("txOutput1_ToAddr"),
				},
				&pb.TxOutput{
					Amount: []byte("txOutput2_Amount"),
					ToAddr: []byte("txOutput2_ToAddr"),
				},
				&pb.TxOutput{
					Amount: []byte("txOutput2_Amount"),
					ToAddr: []byte("txOutput2_ToAddr"),
				},
			},
			tx: &pb.Transaction{
				TxInputs: []*pb.TxInput{
					&pb.TxInput{
						RefTxid:   []byte("txInput1_RefTxid"),
						RefOffset: 1,
						FromAddr:  []byte("txInput1_FromAddr"),
					},
					&pb.TxInput{
						RefTxid:   []byte("txInput2_RefTxid"),
						RefOffset: 2,
						FromAddr:  []byte("txInput2_FromAddr"),
					},
					&pb.TxInput{
						RefTxid:   []byte("txInput3_RefTxid"),
						RefOffset: 3,
						FromAddr:  []byte("txInput3_FromAddr"),
					},
					&pb.TxInput{
						RefTxid:   []byte("txInput4_RefTxid"),
						RefOffset: 4,
						FromAddr:  []byte("txInput4_FromAddr"),
					},
				},
				TxOutputs: []*pb.TxOutput{
					&pb.TxOutput{
						Amount: []byte("txOutput1_Amount"),
						ToAddr: []byte("txOutput1_ToAddr"),
					},
					&pb.TxOutput{
						Amount: []byte("txOutput2_Amount"),
						ToAddr: []byte("txOutput2_ToAddr"),
					},
					&pb.TxOutput{
						Amount: []byte("txOutput3_Amount"),
						ToAddr: []byte("txOutput3_ToAddr"),
					},
				},
			},
			res: false,
		},
		"test check failed2": {
			conTxInputs: []*pb.TxInput{
				&pb.TxInput{
					RefTxid:   []byte("txInput1_RefTxid"),
					RefOffset: 1,
					FromAddr:  []byte("txInput1_FromAddr"),
				},
				&pb.TxInput{
					RefTxid:   []byte("txInput2_RefTxid"),
					RefOffset: 2,
					FromAddr:  []byte("txInput2_FromAddr"),
				},
				&pb.TxInput{
					RefTxid:   []byte("txInput3_RefTxid"),
					RefOffset: 3,
					FromAddr:  []byte("txInput3_FromAddr"),
				},
			},
			conTxOutputs: []*pb.TxOutput{
				&pb.TxOutput{
					Amount: []byte("txOutput1_Amount"),
					ToAddr: []byte("txOutput1_ToAddr"),
				},
				&pb.TxOutput{
					Amount: []byte("txOutput2_Amount"),
					ToAddr: []byte("txOutput2_ToAddr"),
				},
				&pb.TxOutput{
					Amount: []byte("txOutput4_Amount"),
					ToAddr: []byte("txOutput4_ToAddr"),
				},
			},
			tx: &pb.Transaction{
				TxInputs: []*pb.TxInput{
					&pb.TxInput{
						RefTxid:   []byte("txInput1_RefTxid"),
						RefOffset: 1,
						FromAddr:  []byte("txInput1_FromAddr"),
					},
					&pb.TxInput{
						RefTxid:   []byte("txInput2_RefTxid"),
						RefOffset: 2,
						FromAddr:  []byte("txInput2_FromAddr"),
					},
					&pb.TxInput{
						RefTxid:   []byte("txInput3_RefTxid"),
						RefOffset: 3,
						FromAddr:  []byte("txInput3_FromAddr"),
					},
					&pb.TxInput{
						RefTxid:   []byte("txInput4_RefTxid"),
						RefOffset: 4,
						FromAddr:  []byte("txInput4_FromAddr"),
					},
				},
				TxOutputs: []*pb.TxOutput{
					&pb.TxOutput{
						Amount: []byte("txOutput1_Amount"),
						ToAddr: []byte("txOutput1_ToAddr"),
					},
					&pb.TxOutput{
						Amount: []byte("txOutput2_Amount"),
						ToAddr: []byte("txOutput2_ToAddr"),
					},
					&pb.TxOutput{
						Amount: []byte("txOutput3_Amount"),
						ToAddr: []byte("txOutput3_ToAddr"),
					},
				},
			},
			res: false,
		},
	}
	for k, v := range testCases {
		res := IsContractUtxoEffective(v.conTxInputs, v.conTxOutputs, v.tx)
		if res != v.res {
			t.Error("TestIsConUtxoEffective failed case=", k, "expect=", v.res, "actual=", res)
		}
	}
}

// package util 包含了一些测试用的便捷工具

package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/protobuf/proto"
	log "github.com/xuperchain/log15"
	crypto_client "github.com/xuperchain/xuperunion/crypto/client"
	"github.com/xuperchain/xuperunion/kv/kvdb"
	"github.com/xuperchain/xuperunion/ledger"
	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/pluginmgr"
	"github.com/xuperchain/xuperunion/xmodel"
)

// XModelContext context for XModel
type XModelContext struct {
	Basedir string
	StateDB kvdb.Database
	Cache   *xmodel.XMCache
	Model   *xmodel.XModel
}

func saveUnconfirmTx(tx *pb.Transaction, db kvdb.Database) error {
	buf, err := proto.Marshal(tx)
	if err != nil {
		return err
	}
	rawKey := append([]byte(pb.UnconfirmedTablePrefix), []byte(tx.Txid)...)
	return db.Put(rawKey, buf)
}

// CommitCache commit model cache
func (x *XModelContext) CommitCache() error {
	tx := &pb.Transaction{
		Txid: []byte("fake_tx"),
	}
	rset, wset, _ := x.Cache.GetRWSets()
	for _, r := range rset {
		tx.TxInputsExt = append(tx.TxInputsExt, &pb.TxInputExt{
			Bucket:    r.GetPureData().GetBucket(),
			Key:       r.GetPureData().GetKey(),
			RefTxid:   r.GetRefTxid(),
			RefOffset: r.GetRefOffset(),
		})
	}
	for _, w := range wset {
		tx.TxOutputsExt = append(tx.TxOutputsExt, &pb.TxOutputExt{
			Bucket: w.GetBucket(),
			Key:    w.GetKey(),
			Value:  w.GetValue(),
		})
	}
	err := saveUnconfirmTx(tx, x.StateDB)
	if err != nil {
		return err
	}
	batch := x.StateDB.NewBatch()
	err = x.Model.DoTx(tx, batch)
	if err != nil {
		return err
	}
	return batch.Write()
}

// NewCache create model cache instance
func (x *XModelContext) NewCache() error {
	cache, err := xmodel.NewXModelCache(x.Model, true)
	if err != nil {
		return err
	}
	x.Cache = cache
	return nil
}

func openDB(dbPath string, logger log.Logger) (kvdb.Database, error) {
	plgMgr, plgErr := pluginmgr.GetPluginMgr()
	if plgErr != nil {
		logger.Warn("fail to get plugin manager")
		return nil, plgErr
	}
	var baseDB kvdb.Database
	soInst, err := plgMgr.PluginMgr.CreatePluginInstance("kv", "default")
	if err != nil {
		logger.Warn("fail to create plugin instance", "kvtype", "default")
		return nil, err
	}
	baseDB = soInst.(kvdb.Database)
	err = baseDB.Open(dbPath, map[string]interface{}{
		"cache":     128,
		"fds":       512,
		"dataPaths": []string{},
	})
	if err != nil {
		logger.Warn("xmodel::openDB failed to open db", "dbPath", dbPath)
		return nil, err
	}
	return baseDB, nil
}

// WithXModelContext set xmodel context
func WithXModelContext(t testing.TB, callback func(x *XModelContext)) {
	logger := log.New("module", "xmodel")
	logger.SetHandler(log.StreamHandler(os.Stderr, log.LogfmtFormat()))
	basedir, err := ioutil.TempDir("", "xmodel-data")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(basedir)

	ledgerdir := filepath.Join(basedir, "ledger")
	ledger, err := ledger.NewLedger(ledgerdir, logger, nil, "default", crypto_client.CryptoTypeDefault)
	if err != nil {
		t.Fatal(err)
	}

	utxodir := filepath.Join(basedir, "utxo")
	stateDB, err := openDB(utxodir, logger)
	if err != nil {
		t.Fatal(err)
	}
	defer stateDB.Close()
	model, err := xmodel.NewXuperModel("xuper", ledger, stateDB, logger)
	if err != nil {
		t.Fatal(err)
	}
	cache, err := xmodel.NewXModelCache(model, true)
	if err != nil {
		t.Fatal(err)
	}

	callback(&XModelContext{
		Basedir: basedir,
		StateDB: stateDB,
		Model:   model,
		Cache:   cache,
	})
}

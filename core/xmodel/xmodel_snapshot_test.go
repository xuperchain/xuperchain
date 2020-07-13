package xmodel

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	log "github.com/xuperchain/log15"
	"github.com/xuperchain/xuperchain/core/kv/kvdb"
	"github.com/xuperchain/xuperchain/core/ledger"
	"github.com/xuperchain/xuperchain/core/pb"
	"github.com/xuperchain/xuperchain/core/utxo/txhash"
)

var (
	ledg    *ledger.Ledger
	stateDB kvdb.Database
	slog    log.Logger
)

// 如果指定一个存在的账本目录，就会从这个目录加载，默认创建一个新的空账本
const (
	// 账本存储目录
	ledgPath = ""
)

const (
	BobAddress = "WNWk3ekXeM5M2232dY2uCJmEqWhfQiDYT"
)

func initTestEnv() (isTmp bool, rootPath string, err error) {
	slog = log.New("module", "xmodel")
	slog.SetHandler(log.StreamHandler(os.Stderr, log.LogfmtFormat()))

	isTmp, rootPath, err = createTmpLedger()
	if err != nil {
		return isTmp, rootPath, err
	}

	ledg, err = ledger.NewLedger(rootPath, nil, nil, "default", "default")
	if err != nil {
		return isTmp, rootPath, err
	}

	stateDB, err = openDB(rootPath+"/utxoVM", slog)
	if err != nil {
		return isTmp, rootPath, err
	}

	return isTmp, rootPath, err
}

func createTmpLedger() (bool, string, error) {
	if ledgPath != "" {
		return false, ledgPath, nil
	}

	rootPath, err := ioutil.TempDir("/tmp", "")
	if err != nil {
		return false, "", err
	}
	os.RemoveAll(rootPath)

	ledg, err = ledger.NewLedger(rootPath, nil, nil, "default", "default")
	if err != nil {
		return true, rootPath, err
	}

	t1 := &pb.Transaction{}
	t1.TxOutputs = append(t1.TxOutputs, &pb.TxOutput{Amount: []byte("888"), ToAddr: []byte(BobAddress)})
	t1.Coinbase = true
	t1.Desc = []byte(`{"maxblocksize" : "128"}`)
	t1.Txid, _ = txhash.MakeTransactionID(t1)
	block, err := ledg.FormatRootBlock([]*pb.Transaction{t1})
	if err != nil {
		return true, rootPath, err
	}

	confirmStatus := ledg.ConfirmBlock(block, true)
	if !confirmStatus.Succ {
		return true, rootPath, fmt.Errorf("confirm block fail")
	}

	ledg.Close()
	return true, rootPath, nil
}

func getBlkIdByHeight(height int64) ([]byte, error) {
	blkInfo, err := ledg.QueryBlockByHeight(height)
	if err != nil {
		return nil, err
	}

	return blkInfo.Blockid, nil
}

func TestGet(t *testing.T) {
	isTmp, rootPath, err := initTestEnv()
	if err != nil {
		t.Fatal(err)
	}
	if isTmp {
		defer os.RemoveAll(rootPath)
	}

	blkId, err := getBlkIdByHeight(0)
	if err != nil {
		t.Fatal(err)
	}

	xmsp, err := NewSnapshot(blkId, ledg, stateDB, slog)
	if err != nil {
		t.Fatal(err)
	}

	vData, err := xmsp.Get("proftestc", []byte("key_1"))
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(vData)
	fmt.Println(hex.EncodeToString(vData.RefTxid))
}

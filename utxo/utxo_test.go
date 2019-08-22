package utxo

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/xuperchain/xuperunion/contract"
	crypto_client "github.com/xuperchain/xuperunion/crypto/client"
	"github.com/xuperchain/xuperunion/crypto/hash"
	"github.com/xuperchain/xuperunion/global"
	ledger_pkg "github.com/xuperchain/xuperunion/ledger"
	"github.com/xuperchain/xuperunion/pb"
)

// common test data
const (
	BobAddress      = "dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN"
	BobPubkey       = `{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571}`
	BobPrivateKey   = `{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571,"D":29079635126530934056640915735344231956621504557963207107451663058887647996601}`
	AliceAddress    = "WNWk3ekXeM5M2232dY2uCJmEqWhfQiDYT"
	AlicePubkey     = `{"Curvname":"P-256","X":38583161743450819602965472047899931736724287060636876073116809140664442044200,"Y":73385020193072990307254305974695788922719491565637982722155178511113463088980}`
	AlicePrivateKey = `{"Curvname":"P-256","X":38583161743450819602965472047899931736724287060636876073116809140664442044200,"Y":73385020193072990307254305974695788922719491565637982722155178511113463088980,"D":98698032903818677365237388430412623738975596999573887926929830968230132692775}`

	minerPrivateKey = `{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571,"D":29079635126530934056640915735344231956621504557963207107451663058887647996601}`
	minerPublicKey  = `{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571}`
	minerAddress    = `dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN`

	DefaultKVEngine = "default"
)

// Users predefined user
var Users = map[string]struct {
	Address    string
	Pubkey     string
	PrivateKey string
}{
	"bob": {
		Address:    BobAddress,
		Pubkey:     BobPubkey,
		PrivateKey: BobPrivateKey,
	},
	"alice": {
		Address:    AliceAddress,
		Pubkey:     AlicePubkey,
		PrivateKey: AlicePrivateKey,
	},
}

func TestUtxoNew(t *testing.T) {
	workspace, dirErr := ioutil.TempDir("/tmp", "")
	if dirErr != nil {
		t.Fatal(dirErr)
	}
	os.RemoveAll(workspace)
	defer os.RemoveAll(workspace)
	ledger, err := ledger_pkg.NewLedger(workspace, nil, nil, DefaultKVEngine, crypto_client.CryptoTypeDefault)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ledger)
	utxoVM, _ := NewUtxoVM("xuper", ledger, workspace, minerPrivateKey, minerPublicKey, []byte(minerAddress),
		nil, false, DefaultKVEngine, crypto_client.CryptoTypeDefault)
	// test for QueryTx
	_, err1 := utxoVM.QueryTx([]byte("123"))
	if err1 != nil {
		t.Log("query tx[123] error ", err1.Error())
	}
}

func transfer(from string, to string, t *testing.T, utxoVM *UtxoVM, ledger *ledger_pkg.Ledger, amount string, preHash []byte, desc string, frozenHeight int64) ([]byte, error) {
	t.Logf("preHash of this block: %x", preHash)
	txReq := &pb.TxData{}
	txReq.Bcname = "xuper-chain"
	txReq.FromAddr = Users[from].Address
	txReq.FromPubkey = Users[from].Pubkey
	txReq.FromScrkey = Users[from].PrivateKey
	txReq.Nonce = "nonce"
	txReq.Timestamp = time.Now().UnixNano()
	txReq.Desc = []byte(desc)
	txReq.Account = []*pb.TxDataAccount{
		{Address: Users[to].Address, Amount: amount, FrozenHeight: frozenHeight},
	}
	timer := global.NewXTimer()
	tx, err := utxoVM.GenerateTx(txReq)
	if err != nil {
		return nil, err
	}
	t.Log("version: ", tx.Version)
	// test for amount as negative value
	txReq.Account = []*pb.TxDataAccount{
		{Address: Users[to].Address, Amount: "-1", FrozenHeight: int64(0)},
	}
	var negativeErr error
	_, negativeErr = utxoVM.GenerateTx(txReq)
	if negativeErr != nil {
		t.Log("Generate negative value error ", negativeErr.Error())
	}
	// test for very big amount
	txReq.Account = []*pb.TxDataAccount{
		{Address: Users[to].Address, Amount: "100000", FrozenHeight: int64(0)},
	}
	_, bigErr := utxoVM.GenerateTx(txReq)
	if bigErr != nil {
		t.Log("Generate very big value error ", bigErr.Error())
	}

	cryptoClient, err := crypto_client.CreateCryptoClient(crypto_client.CryptoTypeDefault)
	if err != nil {
		return nil, err
	}
	utxoVM.cryptoClient = cryptoClient
	timer.Mark("GenerateTx")
	verifyOK, vErr := utxoVM.ImmediateVerifyTx(tx, false)
	t.Log("VerifyTX", timer.Print())
	if !verifyOK || vErr != nil {
		t.Log("verify tx fail, ignore in unit test here", vErr)
	}
	// do query tx before do tx
	_, err = utxoVM.QueryTx(tx.Txid)
	if err != nil {
		t.Log("query tx ", tx.Txid, "error ", err.Error())
	}

	// test for asyncMode
	utxoVM.asyncMode = true
	errDo := utxoVM.DoTx(tx)
	timer.Mark("DoTx")
	if errDo != nil {
		t.Fatal(errDo)
		return nil, errDo
	}
	utxoVM.asyncMode = false
	utxoVM.DoTx(tx)

	// do query tx after do tx
	_, err = utxoVM.QueryTx(tx.Txid)
	if err != nil {
		t.Log("query tx ", tx.Txid, "error ", err.Error())
	}

	txlist, packErr := utxoVM.GetUnconfirmedTx(true)
	timer.Mark("GetUnconfirmedTx")
	if packErr != nil {
		return nil, packErr
	}
	//奖励矿工
	awardTx, minerErr := utxoVM.GenerateAwardTx([]byte("miner-1"), "1000", []byte("award,onyeah!"))
	timer.Mark("GenerateAwardTx")
	if minerErr != nil {
		return nil, minerErr
	}

	// case: award_amount is negative
	_, negativeErr = utxoVM.GenerateAwardTx([]byte("miner-1"), "-2", []byte("negative award!"))
	if negativeErr != nil {
		t.Log("GenerateAwardTx error ", negativeErr.Error())
	}
	txlist = append(txlist, awardTx)
	ecdsaPk, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	timer.Mark("GenerateKey")
	block, _ := ledger.FormatBlock(txlist, []byte("miner-1"), ecdsaPk, 123456789, 0, 0, preHash, utxoVM.GetTotal())
	timer.Mark("FormatBlock")
	confirmStatus := ledger.ConfirmBlock(block, false)
	timer.Mark("ConfirmBlock")
	if !confirmStatus.Succ {
		t.Log("confirmStatus", confirmStatus)
		return nil, errors.New("fail to confirm block")
	}
	t.Log("performance metric", timer.Print())
	return block.Blockid, nil
}

func TestUtxoWorkWithLedgerBasic(t *testing.T) {
	workspace, dirErr := ioutil.TempDir("/tmp", "")
	if dirErr != nil {
		t.Fatal(dirErr)
	}
	os.RemoveAll(workspace)
	defer os.RemoveAll(workspace)
	ledger, err := ledger_pkg.NewLedger(workspace, nil, nil, DefaultKVEngine, crypto_client.CryptoTypeDefault)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ledger)
	utxoVM, _ := NewUtxoVM("xuper", ledger, workspace, minerPrivateKey, minerPublicKey, []byte(minerAddress),
		nil, false, DefaultKVEngine, crypto_client.CryptoTypeDefault)
	_, err = utxoVM.QueryTx([]byte("123"))
	if err != ErrTxNotFound {
		t.Fatal("unexpected err", err)
	}
	//创建链的时候分配财富
	tx, err := utxoVM.GenerateRootTx([]byte(`
       {
        "version" : "1"
        , "consensus" : {
                "miner" : "0x00000000000"
        }
        , "predistribution":[
                {
                        "address" : "` + BobAddress + `",
                        "quota" : "100"
                },
				{
                        "address" : "` + AliceAddress + `",
                        "quota" : "200"
                }

        ]
        , "maxblocksize" : "128"
        , "period" : "5000"
        , "award" : "1000"
		} 
    `))
	if err != nil {
		t.Fatal(err)
	}

	// test for HasTx
	exist, _ := utxoVM.HasTx(tx.Txid)
	t.Log("Has tx ", tx.Txid, exist)
	err = utxoVM.DoTx(tx)
	if err != nil {
		t.Log("coinbase do tx error ", err.Error())
	}

	block, _ := ledger.FormatRootBlock([]*pb.Transaction{tx})
	t.Logf("blockid %x", block.Blockid)
	confirmStatus := ledger.ConfirmBlock(block, true)
	if !confirmStatus.Succ {
		t.Fatal("confirm block fail")
	}

	// test for isConfirmed
	status := utxoVM.isConfirmed(tx)
	t.Log("award tx ", tx.Txid, "is confirmed ? ", status)

	playErr := utxoVM.Play(block.Blockid)
	if playErr != nil {
		t.Fatal(playErr)
	}
	// test for GetLatestBlockid
	tipBlock := utxoVM.GetLatestBlockid()
	t.Log("current tip block ", tipBlock)
	t.Log("last tip block ", block.Blockid)

	bobBalance, _ := utxoVM.GetBalance(BobAddress)
	aliceBalance, _ := utxoVM.GetBalance(AliceAddress)
	if bobBalance.String() != "100" || aliceBalance.String() != "200" {
		t.Fatal("unexpected balance", bobBalance, aliceBalance)
	}
	t.Logf("bob balance: %s, alice balance: %s", bobBalance.String(), aliceBalance.String())
	rootBlockid := block.Blockid
	t.Logf("rootBlockid: %x", rootBlockid)
	//bob再给alice转5
	nextBlockid, blockErr := transfer("bob", "alice", t, utxoVM, ledger, "5", rootBlockid, "", 0)
	if blockErr != nil {
		t.Fatal(blockErr)
	} else {
		t.Logf("next block id: %x", nextBlockid)
	}
	utxoVM.Play(nextBlockid)
	bobBalance, _ = utxoVM.GetBalance(BobAddress)
	aliceBalance, _ = utxoVM.GetBalance(AliceAddress)
	t.Logf("bob balance: %s, alice balance: %s", bobBalance.String(), aliceBalance.String())
	//bob再给alice转6
	nextBlockid, blockErr = transfer("bob", "alice", t, utxoVM, ledger, "6", nextBlockid, "", 0)
	if blockErr != nil {
		t.Fatal(blockErr)
	} else {
		t.Logf("next block id: %x", nextBlockid)
	}
	utxoVM.Play(nextBlockid)
	bobBalance, _ = utxoVM.GetBalance(BobAddress)
	aliceBalance, _ = utxoVM.GetBalance(AliceAddress)
	t.Logf("bob balance: %s, alice balance: %s", bobBalance.String(), aliceBalance.String())

	//再创建一个新账本，从前面一个账本复制数据
	workspace2, dirErr := ioutil.TempDir("/tmp", "")
	if dirErr != nil {
		t.Fatal(dirErr)
	}
	os.RemoveAll(workspace2)
	defer os.RemoveAll(workspace2)
	ledger2, lErr := ledger_pkg.NewLedger(workspace2, nil, nil, DefaultKVEngine,
		crypto_client.CryptoTypeDefault)
	if lErr != nil {
		t.Fatal(lErr)
	}
	utxoVM2, _ := NewUtxoVM("xuper", ledger2, workspace2, minerPrivateKey, minerPublicKey, []byte(minerAddress),
		nil, false, DefaultKVEngine, crypto_client.CryptoTypeDefault)
	pBlockid := ledger.GetMeta().RootBlockid
	for len(pBlockid) > 0 { //这个for完成把第一个账本的数据同步给第二个
		t.Logf("replicating... %x", pBlockid)
		pBlock, pErr := ledger.QueryBlock(pBlockid)
		if pErr != nil {
			t.Fatal(pErr)
		}
		isRoot := bytes.Equal(pBlockid, ledger.GetMeta().RootBlockid)
		cStatus := ledger2.ConfirmBlock(pBlock, isRoot)
		if !cStatus.Succ {
			t.Fatal(cStatus)
		}
		pBlockid = pBlock.NextHash
	}
	utxoVM2.Play(ledger2.GetMeta().RootBlockid) //先做一下根节点
	dummyBlockid, dummyErr := transfer("bob", "alice", t, utxoVM2, ledger2, "7", ledger2.GetMeta().RootBlockid, "", 0)
	if dummyErr != nil {
		t.Fatal(dummyErr)
	}
	utxoVM2.Play(dummyBlockid)
	utxoVM2.Walk(ledger2.GetMeta().TipBlockid) //再游走到末端 ,预期会导致dummmy block回滚
	bobBalance, _ = utxoVM2.GetBalance(BobAddress)
	aliceBalance, _ = utxoVM2.GetBalance(AliceAddress)
	minerBalance, _ := utxoVM2.GetBalance("miner-1")
	t.Logf("bob balance: %s, alice balance: %s, miner-1: %s", bobBalance.String(), aliceBalance.String(), minerBalance.String())
	if bobBalance.String() != "89" || aliceBalance.String() != "211" {
		t.Fatal("unexpected balance", bobBalance, aliceBalance)
	}
	transfer("bob", "alice", t, utxoVM2, ledger2, "7", ledger2.GetMeta().TipBlockid, "", 0)
	transfer("bob", "alice", t, utxoVM2, ledger2, "7", ledger2.GetMeta().TipBlockid, "", 0)
	utxoVM2.Walk(ledger2.GetMeta().TipBlockid)
	bobBalance, _ = utxoVM2.GetBalance(BobAddress)
	aliceBalance, _ = utxoVM2.GetBalance(AliceAddress)
	minerBalance, _ = utxoVM2.GetBalance("miner-1")
	t.Logf("bob balance: %s, alice balance: %s, miner-1: %s", bobBalance.String(), aliceBalance.String(), minerBalance.String())
	if bobBalance.String() != "75" || aliceBalance.String() != "225" {
		t.Fatal("unexpected balance", bobBalance, aliceBalance)
	}
	t.Log(ledger.Dump())

	aliceBalance2, _ := utxoVM.GetBalance(AliceAddress)
	t.Log("get alice balance ", aliceBalance2)

	// test for RemoveUtxoCache
	utxoVM.RemoveUtxoCache("bob", "123")
	// test for GetTotal
	total := utxoVM.GetTotal()
	t.Log("total ", total)
	iter := utxoVM.ScanWithPrefix([]byte("UWNWk3ekXeM5M2232dY2uCJmEqWhfQiDYT_"))
	for iter.Next() {
		t.Log("ScanWithPrefix  ", iter.Key())
	}

	ledger.Close()
}

func TestTSort(t *testing.T) {
	g := TxGraph{}
	g["tx3"] = []string{"tx1", "tx2"}
	g["tx2"] = []string{"tx1", "tx0"}
	g["tx1"] = []string{"tx0"}
	output, cylic, _ := TopSortDFS(g)
	t.Log(output)
	if !reflect.DeepEqual(output, []string{"tx3", "tx2", "tx1", "tx0"}) {
		t.Fatal("sort fail")
	}
	if cylic {
		t.Fatal("sort fail2")
	}
}

func TestCheckCylic(t *testing.T) {
	g := TxGraph{}
	g["tx3"] = []string{"tx1", "tx2"}
	g["tx2"] = []string{"tx1", "tx0"}
	g["tx1"] = []string{"tx0", "tx2"}
	output, cylic, _ := TopSortDFS(g)
	if output != nil {
		t.Fatal("sort fail1")
	}
	t.Log(cylic)
	//if len(cylic) != 2 {
	if cylic == false {
		t.Fatal("sort fail2")
	}
}

func TestFrozenHeight(t *testing.T) {
	workspace, dirErr := ioutil.TempDir("/tmp", "")
	if dirErr != nil {
		t.Fatal(dirErr)
	}
	os.RemoveAll(workspace)
	defer os.RemoveAll(workspace)
	ledger, err := ledger_pkg.NewLedger(workspace, nil, nil, DefaultKVEngine, crypto_client.CryptoTypeDefault)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ledger)
	utxoVM, _ := NewUtxoVM("xuper", ledger, workspace, minerPrivateKey, minerPublicKey, []byte(minerAddress),
		nil, false, DefaultKVEngine, crypto_client.CryptoTypeDefault)
	//创建链的时候分配, bob:100, alice:200
	tx, err := utxoVM.GenerateRootTx([]byte(`
       {
        "version" : "1"
        , "consensus" : {
                "miner" : "0x00000000000"
        }
        , "predistribution":[
                {
                        "address" : "` + BobAddress + `",
                        "quota" : "100"
                },
				{
                        "address" : "` + AliceAddress + `",
                        "quota" : "200"
                }

        ]
        , "maxblocksize" : "128"
        , "period" : "5000"
        , "award" : "1000"
		} 
    `))
	if err != nil {
		t.Fatal(err)
	}
	block, _ := ledger.FormatRootBlock([]*pb.Transaction{tx})
	t.Logf("blockid %x", block.Blockid)
	confirmStatus := ledger.ConfirmBlock(block, true)
	if !confirmStatus.Succ {
		t.Fatal("confirm block fail")
	}
	playErr := utxoVM.Play(block.Blockid)
	if playErr != nil {
		t.Fatal(playErr)
	}
	bobBalance, _ := utxoVM.GetBalance(BobAddress)
	aliceBalance, _ := utxoVM.GetBalance(AliceAddress)
	if bobBalance.String() != "100" || aliceBalance.String() != "200" {
		t.Fatal("unexpected balance", bobBalance, aliceBalance)
	}
	//bob 给alice转100，账本高度=2的时候才能解冻
	nextBlockid, blockErr := transfer("bob", "alice", t, utxoVM, ledger, "100", ledger.GetMeta().TipBlockid, "", 2)
	if blockErr != nil {
		t.Fatal(blockErr)
	} else {
		t.Logf("next block id: %x", nextBlockid)
	}

	// test for GetFrozenBalance
	frozenBalance, frozenBalanceErr := utxoVM.GetFrozenBalance(AliceAddress)
	if frozenBalanceErr != nil {
		t.Log("get frozen balance error ", frozenBalanceErr.Error())
	} else {
		t.Log("alice frozen balance ", frozenBalance)
	}

	//alice给bob转300, 预期失败，因为无法使用被冻住的utxo
	nextBlockid, blockErr = transfer("alice", "bob", t, utxoVM, ledger, "300", ledger.GetMeta().TipBlockid, "", 0)
	if blockErr != ErrNoEnoughUTXO {
		t.Fatal("unexpected ", blockErr)
	}
	//alice先给自己转1块钱，让块高度增加
	nextBlockid, blockErr = transfer("alice", "alice", t, utxoVM, ledger, "1", ledger.GetMeta().TipBlockid, "", 0)
	if blockErr != nil {
		t.Fatal(blockErr)
	}
	//然后alice再次尝试给bob转300,预期utxo解冻可用了
	nextBlockid, blockErr = transfer("alice", "bob", t, utxoVM, ledger, "300", ledger.GetMeta().TipBlockid, "", 0)
	if blockErr != nil {
		t.Fatal(blockErr)
	}
}

type testInterface struct{}

func (ti *testInterface) Run(desc *contract.TxDesc) error {
	return nil
}

func (ti *testInterface) Rollback(desc *contract.TxDesc) error {
	return nil
}

func (ti *testInterface) ReadOutput(desc *contract.TxDesc) (contract.ContractOutputInterface, error) {
	return nil, nil
}

func (ti *testInterface) Finalize(blockid []byte) error {
	return nil
}

func (ti *testInterface) SetContext(context *contract.TxContext) error {
	return nil
}

func (ti *testInterface) Stop() {

}

func (ti *testInterface) CreateChainBlock() {

}

type testVat struct{}

func (tv *testVat) GetVerifiableAutogenTx(blockHeight int64, maxCount int, timestamp int64) ([]*pb.Transaction, error) {
	return nil, nil
}

func (tv *testVat) GetVATWhiteList() map[string]bool {
	whiteList := map[string]bool{
		"update_consensus": true,
	}
	return whiteList
}

type testOutput struct {
	Outputs string `json:"outputs"`
	GasUsed uint64 `json:"gasused"`
	Error   error  `json:"error"`
}

func (to *testOutput) Decode(data []byte) error {
	return json.Unmarshal(data, to)
}

func (to *testOutput) Encode() ([]byte, error) {
	return json.Marshal(to)
}

/*
func (to *testOutput) Decode(data []byte) error {
    // return json.Unmarshal(data, to)
    return nil
}*/

func (to *testOutput) GetGasUsed() uint64 {
	return to.GasUsed
}

func (to *testOutput) Digest() ([]byte, error) {
	buf := new(bytes.Buffer)
	encoder := json.NewEncoder(buf)
	err := encoder.Encode(*to)
	if err == nil {
		return hash.DoubleSha256(buf.Bytes()), err
	}
	return nil, err
}

func (to *testOutput) GetTxGeneratedByContract() ([]*pb.Transaction, error) {
	return nil, nil
}

func (to *testOutput) VerifyTxGeneratedByContract(block *pb.InternalBlock, tx *pb.Transaction, idx int) (int, error) {
	return idx, nil
}

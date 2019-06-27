package utxo

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"os"
	"testing"
	"time"

	crypto_client "github.com/xuperchain/xuperunion/crypto/client"
	"github.com/xuperchain/xuperunion/kv/kvdb"
	ledger_pkg "github.com/xuperchain/xuperunion/ledger"
	"github.com/xuperchain/xuperunion/pb"
)

// FakeUtxoVM define a fake UTXO for test purpose
type FakeUtxoVM struct {
	U     *UtxoVM
	L     *ledger_pkg.Ledger
	Users map[string]struct {
		Address    string
		Pubkey     string
		PrivateKey string
	}
	BobPrivateKey, BobPubkey, BobAddress       string
	AliceAddress, AlicePrivateKey, AlicePubkey string
}

// NewFakeUtxoVM create instance of FakeUtxoVM
func NewFakeUtxoVM(t *testing.T, workspace string, recreate bool) *FakeUtxoVM {
	if recreate {
		os.RemoveAll(workspace)
	}
	ledger, err := ledger_pkg.NewLedger(workspace, nil, nil, "default", crypto_client.CryptoTypeDefault)
	if err != nil {
		t.Fatal(err)
	}
	// load public key and address
	privateKey := `{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571,"D":29079635126530934056640915735344231956621504557963207107451663058887647996601}`
	publicKey := `{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571}`
	address := []byte(`dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN`)

	utxoVM, err := NewUtxoVM("xuper", ledger, workspace, privateKey, publicKey, address, nil, false, "default",
		crypto_client.CryptoTypeDefault)
	if err != nil {
		t.Fatal(err)
	}
	f := &FakeUtxoVM{
		U:               utxoVM,
		L:               ledger,
		BobAddress:      "dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN",
		BobPubkey:       `{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571}`,
		BobPrivateKey:   `{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571,"D":29079635126530934056640915735344231956621504557963207107451663058887647996601}`,
		AliceAddress:    "WNWk3ekXeM5M2232dY2uCJmEqWhfQiDYT",
		AlicePubkey:     `{"Curvname":"P-256","X":38583161743450819602965472047899931736724287060636876073116809140664442044200,"Y":73385020193072990307254305974695788922719491565637982722155178511113463088980}`,
		AlicePrivateKey: `{"Curvname":"P-256","X":38583161743450819602965472047899931736724287060636876073116809140664442044200,"Y":73385020193072990307254305974695788922719491565637982722155178511113463088980,"D":98698032903818677365237388430412623738975596999573887926929830968230132692775}`,
	}

	f.Users = map[string]struct {
		Address    string
		Pubkey     string
		PrivateKey string
	}{
		"bob": {
			Address:    f.BobAddress,
			Pubkey:     f.BobPubkey,
			PrivateKey: f.BobPrivateKey,
		},
		"alice": {
			Address:    f.AliceAddress,
			Pubkey:     f.AlicePubkey,
			PrivateKey: f.AlicePrivateKey,
		},
	}

	return f
}

// PlayForMiner generate miner award tx and create block
func (f *FakeUtxoVM) PlayForMiner(t *testing.T, txs []*pb.Transaction, preHash []byte, miner string) error {
	cryptoClient, err := crypto_client.CreateCryptoClientFromJSONPrivateKey([]byte(f.Users[miner].PrivateKey))
	if err != nil {
		return err
	}
	pk, _ := cryptoClient.GetEcdsaPrivateKeyFromJSON([]byte(f.Users[miner].PrivateKey))
	pendingBlock, err := f.L.FormatBlock(txs, []byte(f.Users[miner].Address), pk,
		22121212, 0, 0, preHash, f.U.GetTotal())
	if err != nil {
		return err
	}
	var batch kvdb.Batch
	blockAward := f.L.GenesisBlock.CalcAward(f.L.GetMeta().TrunkHeight + 1)
	awardtx, err := f.U.GenerateAwardTx([]byte(f.Users[miner].Address), blockAward.String(), []byte{'1'})
	if txs, batch, err = f.U.TxOfRunningContractGenerate(txs, pendingBlock, nil, true); err != nil {
		return err
	}
	txs = append(txs, awardtx)

	b, err := f.L.FormatBlock(txs, []byte(f.Users[miner].Address), pk,
		22121212, 0, 0, preHash, f.U.GetTotal())
	if err != nil {
		return err
	}

	confirmStatus := f.L.ConfirmBlock(b, false)
	if !confirmStatus.Succ {
		return confirmStatus.Error
	}

	return f.U.PlayForMiner(b.Blockid, batch)
}

// GenerateTx generate a transaction using given params
func (f *FakeUtxoVM) GenerateTx(from string, to string, t *testing.T, amount string, preHash []byte, desc []byte, frozenHeight int64) ([]*pb.Transaction, *pb.Transaction, *ecdsa.PrivateKey, error) {
	utxoVM := f.U
	t.Logf("preHash of this block: %x", preHash)
	txReq := &pb.TxData{}
	txReq.Bcname = "xuper-chain"
	txReq.FromAddr = f.Users[from].Address
	txReq.FromPubkey = f.Users[from].Pubkey
	txReq.FromScrkey = f.Users[from].PrivateKey
	txReq.Nonce = "nonce"
	txReq.Timestamp = time.Now().UnixNano()
	txReq.Desc = desc
	txReq.Account = []*pb.TxDataAccount{
		{Address: f.Users[to].Address, Amount: amount, FrozenHeight: frozenHeight},
		{Address: FeePlaceholder, Amount: "300000000"},
	}
	tx, err := utxoVM.GenerateTx(txReq)
	if err != nil {
		return nil, nil, nil, err
	}

	verifyOK, vErr := utxoVM.ImmediateVerifyTx(tx, false)
	if !verifyOK || vErr != nil {
		t.Log("verify tx fail, ignore in unit test here", vErr)
		t.Error("verify tx failed")
	}
	errDo := utxoVM.DoTx(tx)
	if errDo != nil {
		return nil, nil, nil, errDo
	}
	txlist, packErr := utxoVM.GetUnconfirmedTx(true)
	if packErr != nil {
		return nil, nil, nil, packErr
	}

	awardTx, minerErr := utxoVM.GenerateAwardTx([]byte("miner-1"), "1", []byte("award,onyeah!"))
	if minerErr != nil {
		return nil, nil, nil, minerErr
	}
	txlist = append(txlist, awardTx)
	ecdsdPk, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	return txlist, tx, ecdsdPk, nil
}

// Transfer used to transfer betweet address/account
func (f *FakeUtxoVM) Transfer(from string, to string, t *testing.T, amount string, preHash []byte, desc []byte, frozenHeight int64) (*pb.InternalBlock, []*pb.Transaction, *pb.Transaction, error) {
	ledger := f.L
	txlist, tx, ecdsdPk, err := f.GenerateTx(from, to, t, amount, preHash, desc, frozenHeight)
	if err != nil {
		return nil, nil, nil, err
	}
	t.Logf("txs after GenerateTx %d", len(txlist))
	block, _ := ledger.FormatBlock(txlist, []byte("miner-1"), ecdsdPk, 123456789, 0, 0, preHash, f.U.GetTotal())
	confirmStatus := ledger.ConfirmBlock(block, false)
	if !confirmStatus.Succ {
		t.Log("confirmStatus", confirmStatus)
		return nil, nil, nil, errors.New("fail to confirm block")
	}
	return block, txlist, tx, nil
}

// UtxoWorkWithLedgerBasic basic operations of UTXO
func (f *FakeUtxoVM) UtxoWorkWithLedgerBasic(t *testing.T) {
	utxoVM, ledger := f.U, f.L
	//创建链的时候分配财富
	tx, err := utxoVM.GenerateRootTx([]byte(`
       {
        "version" : "1"
        , "consensus" : {
                "miner" : "0x00000000000"
        }
        , "predistribution":[
                {
                        "address" : "` + f.BobAddress + `",
                        "quota" : "10000000000000000"
                },
				{
                        "address" : "` + f.AliceAddress + `",
                        "quota" : "20000000000000000"
                }

        ]
        , "maxblocksize" : "128"
        , "period" : "5000"
        , "award" : "1"
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
	bobBalance, _ := utxoVM.GetBalance(f.BobAddress)
	aliceBalance, _ := utxoVM.GetBalance(f.AliceAddress)
	t.Logf("bob balance: %s, alice balance: %s", bobBalance.String(), aliceBalance.String())
	rootBlockid := block.Blockid
	t.Logf("rootBlockid: %x", rootBlockid)
	//bob再给alice转5
	nextBlock, _, _, blockErr := f.Transfer("bob", "alice", t, "5", rootBlockid, []byte(""), 0)
	nextBlockid := nextBlock.Blockid
	if blockErr != nil {
		t.Fatal(blockErr)
	} else {
		t.Logf("next block id: %x", nextBlockid)
	}
	utxoVM.Play(nextBlockid)
	bobBalance, _ = utxoVM.GetBalance(f.BobAddress)
	aliceBalance, _ = utxoVM.GetBalance(f.AliceAddress)
	t.Logf("bob balance: %s, alice balance: %s", bobBalance.String(), aliceBalance.String())
	//bob再给alice转6
	nextBlock, _, _, blockErr = f.Transfer("bob", "alice", t, "6", nextBlockid, []byte(""), 0)
	if blockErr != nil {
		t.Fatal(blockErr)
	} else {
		t.Logf("next block id: %x", nextBlockid)
	}
	nextBlockid = nextBlock.Blockid
	utxoVM.Play(nextBlockid)
	bobBalance, _ = utxoVM.GetBalance(f.BobAddress)
	aliceBalance, _ = utxoVM.GetBalance(f.AliceAddress)
	t.Logf("bob balance: %s, alice balance: %s", bobBalance.String(), aliceBalance.String())
}

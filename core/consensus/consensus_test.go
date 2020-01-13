package consensus

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"

	"github.com/xuperchain/xuperunion/common/config"
	"github.com/xuperchain/xuperunion/consensus/tdpos"
	"github.com/xuperchain/xuperunion/contract"
	crypto_client "github.com/xuperchain/xuperunion/crypto/client"
	"github.com/xuperchain/xuperunion/kv/kvdb"
	"github.com/xuperchain/xuperunion/ledger"
	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/utxo"

	//"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/xuperchain/xuperunion/pluginmgr"
)

const (
	minerPrivateKey = `{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571,"D":29079635126530934056640915735344231956621504557963207107451663058887647996601}`
	minerPublicKey  = `{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571}`
	minerAddress    = `dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN`
)

const BobAddress = "dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN"
const BobPubkey = `{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571}`
const BobPrivateKey = `{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571,"D":29079635126530934056640915735344231956621504557963207107451663058887647996601}`
const AliceAddress = "WNWk3ekXeM5M2232dY2uCJmEqWhfQiDYT"
const AlicePubkey = `{"Curvname":"P-256","X":38583161743450819602965472047899931736724287060636876073116809140664442044200,"Y":73385020193072990307254305974695788922719491565637982722155178511113463088980}`
const AlicePrivateKey = `{"Curvname":"P-256","X":38583161743450819602965472047899931736724287060636876073116809140664442044200,"Y":73385020193072990307254305974695788922719491565637982722155178511113463088980,"D":98698032903818677365237388430412623738975596999573887926929830968230132692775}`

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

var (
	//workspace   = "/tmp/test_workspace"
	//testspace   = "/tmp/testspace"
	kvengine    = "default"
	tCryptoType = crypto_client.CryptoTypeDefault
)

var workspace, workSpaceErr = ioutil.TempDir("/tmp", "")
var testspace, testSpaceErr = ioutil.TempDir("/tmp", "")

func plugPrepareWithGensisBlock(t *testing.T) *PluggableConsensus {
	ledger, err := ledger.NewLedger(workspace, nil, nil, kvengine, tCryptoType)
	if err != nil {
		t.Fatal(err)
	}
	tx, gensisErr := utxo.GenerateRootTx([]byte(`
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
	if gensisErr != nil {
		t.Fatal(err)
	}

	block, _ := ledger.FormatRootBlock([]*pb.Transaction{tx})
	confirmStatus := ledger.ConfirmBlock(block, true)
	if !confirmStatus.Succ {
		t.Fatal("confirm block fail")
	} else {
		t.Log("trunk height ", ledger.GetMeta().TrunkHeight)
	}
	utxoVM, _ := utxo.NewUtxoVM("xuper", ledger, workspace, minerPrivateKey, minerPublicKey, []byte(minerAddress), nil, false, kvengine, tCryptoType)
	playErr := utxoVM.Play(block.Blockid)
	if playErr != nil {
		t.Fatal(playErr)
	}

	// 一般的交易
	from := "alice"
	to := "bob"
	txReq := &pb.TxData{}
	txReq.Bcname = "xuper-chain"
	txReq.FromAddr = Users[from].Address
	txReq.FromPubkey = Users[from].Pubkey
	txReq.FromScrkey = Users[from].PrivateKey
	txReq.Nonce = "nonce"
	txReq.Timestamp = time.Now().UnixNano()
	txReq.Desc = []byte("hello world")
	txReq.Account = []*pb.TxDataAccount{
		{Address: Users[to].Address, Amount: "2", FrozenHeight: int64(0)},
		{Address: "$", Amount: "3"},
	}
	bobBalance, _ := utxoVM.GetBalance(BobAddress)
	aliceBalance, _ := utxoVM.GetBalance(AliceAddress)
	t.Log("get bob balance ", bobBalance)
	t.Log("get alice balance ", aliceBalance)
	tx2, err2 := utxoVM.GenerateTx(txReq)
	if err2 != nil {
		t.Fatal(err2)
	}

	errDo := utxoVM.DoTx(tx2)
	if errDo != nil {
		t.Fatal(errDo)
	}
	txList, _ := utxoVM.GetUnconfirmedTx(true)
	awardTx, errMiner := utxoVM.GenerateAwardTx([]byte("miner-1"), "1", []byte("award,onyeah!"))
	if errMiner != nil {
		t.Fatal(errMiner)
	}
	txList = append(txList, awardTx)
	ecdsaPk, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	preHash := utxoVM.GetLatestBlockid()
	block, _ = ledger.FormatBlock(txList, []byte("miner-1"), ecdsaPk, 123456789, 0, 0, preHash, utxoVM.GetTotal())
	confirmStatus = ledger.ConfirmBlock(block, false)
	if !confirmStatus.Succ {
		t.Fatal("confirm block fail")
	} else {
		t.Log("trunk height ", ledger.GetMeta().TrunkHeight)
	}

	bobBalance, _ = utxoVM.GetBalance(BobAddress)
	aliceBalance, _ = utxoVM.GetBalance(AliceAddress)
	t.Log("get bob balance ", bobBalance)
	t.Log("get alice balance ", aliceBalance)

	cfg := config.NewNodeConfig()
	cfg.Miner.Keypath = "../data/keys/"
	bcname := "xuper"
	rootConfig := map[string]interface{}{
		"name": ConsensusTypeSingle,
		"config": map[string]interface{}{
			"period": "3000",
			"miner":  "dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN",
		},
	}
	plugCons, plugErr := NewPluggableConsensus(nil, cfg, bcname, ledger, utxoVM, rootConfig, tCryptoType, nil)
	if plugErr != nil {
		t.Fatal(plugErr)
	}
	return plugCons
}

func plugPrepare(t *testing.T) *PluggableConsensus {
	ldg, err := ledger.NewLedger(workspace, nil, nil, kvengine, tCryptoType)
	if err != nil {
		t.Fatal(err)
	}
	// Generate Root Tx
	tx, err := utxo.GenerateRootTx([]byte(`
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
	// FormatBlock
	if err != nil {
		t.Fatal(err)
	}
	block, _ := ldg.FormatRootBlock([]*pb.Transaction{tx})
	t.Logf("blockid %x", block.Blockid)
	// ConfirmBlock
	confirmStatus := ldg.ConfirmBlock(block, true)
	if !confirmStatus.Succ {
		t.Fatal("confirm block fail")
	}
	utxoVM, _ := utxo.NewUtxoVM("xuper", ldg, workspace, minerPrivateKey, minerPublicKey, []byte(minerAddress), nil, false, kvengine, tCryptoType)
	cfg := config.NewNodeConfig()
	cfg.Miner.Keypath = "../data/keys/"
	bcname := "xuper"
	rootConfig := map[string]interface{}{
		"name": ConsensusTypeSingle,
		"config": map[string]interface{}{
			// 出块周期
			"period": "3000",
			// 矿工id
			"miner": "dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN",
		},
	}
	plugCons, err := NewPluggableConsensus(nil, cfg, bcname, ldg, utxoVM, rootConfig, tCryptoType, nil)
	if err != nil {
		t.Fatal(err)
	}
	return plugCons
}

func plugClear() {
	for _, k := range []string{workspace, testspace} {
		os.RemoveAll(k)
	}
}

func TestGenPlugConsKeyWithPrefix(t *testing.T) {
	plugClear()
	defer plugClear()
	testCases := map[string]struct {
		height    int64
		timestamp int64
		expect    string
	}{
		"test1": {
			height:    int64(0),
			timestamp: 1234567,
			expect:    "P00000000000000000000_1234567",
		},
		"test2": {
			height:    int64(9223372036854775800),
			timestamp: 1234567,
			expect:    "P09223372036854775800_1234567",
		},
	}

	for k, v := range testCases {
		actual := genPlugConsKeyWithPrefix(v.height, v.timestamp)
		if actual != v.expect {
			t.Errorf("%s genPlugConsKeyWithPrefix failed, expect %s, actual %s", k, v.expect, actual)
		}
		h, time, _ := parsePlugConsKeyWithPrefix(actual)

		if h != v.height || time != v.timestamp {
			t.Errorf("%s parsePlugConsKeyWithPrefix failed, expect h=%d, t=%d, actual h=%d t=%d", k, v.height, v.timestamp, h, time)
		}
	}
}

func TestNewPluggableConsensus(t *testing.T) {
	plugClear()
	defer plugClear()
	pc := plugPrepare(t)
	t.Log(pc)
}

func TestNewPluggableConsensusWithTrunkHeight(t *testing.T) {
	plugClear()
	defer plugClear()
	pc := plugPrepareWithGensisBlock(t)
	t.Log(pc)
}

func TestPlugConsType(t *testing.T) {
	plugClear()
	defer plugClear()
	pc := plugPrepare(t)
	second := &StepConsensus{}
	second.StartHeight = 100
	second.Conn = &tdpos.TDpos{}
	pc.cons = append(pc.cons, second)

	tp := pc.Type(150)
	t.Log("current consensus type ", tp)
	if tp != tdpos.TYPE {
		t.Errorf("Type failed, expect %v, actual %v", tdpos.TYPE, tp)
	}
}

func TestPlugConsVersion(t *testing.T) {
	plugClear()
	defer plugClear()
	pc := plugPrepare(t)
	second := &StepConsensus{}
	second.StartHeight = 100
	second.Conn = &tdpos.TDpos{}
	pc.cons = append(pc.cons, second)

	t.Log("current consensus version ", pc.Version(150))
	t.Log("current consensus version ", pc.Version(-1))
}

func TestPlugConsProcessReceiveBlock(t *testing.T) {
	plugClear()
	defer plugClear()
	pc := plugPrepare(t)
	block := &pb.InternalBlock{
		Height: 50,
	}
	err := pc.ProcessConfirmBlock(block)
	if err != nil {
		t.Errorf("ProcessReceiveBlock failed, expect %v, actual %v", nil, err)
	}
	block = &pb.InternalBlock{
		Height: -1,
	}
	err = pc.ProcessConfirmBlock(block)
	if err != nil {
		t.Log("negative error ", err.Error())
	}
}

func TestPlugConsRun(t *testing.T) {
	plugClear()
	defer plugClear()
	pc := plugPrepare(t)
	testCases := map[string]struct {
		desc   *contract.TxDesc
		expect error
	}{
		"test default": {
			desc: &contract.TxDesc{
				Method: "default",
				Args: map[string]interface{}{
					"name": "single",
				},
			},
			expect: errors.New("PluggableConsensus not define this method"),
		},
		"test update_consensus": {
			desc: &contract.TxDesc{
				Method: updateConsensusMethod,
			},
			expect: errors.New("Consensus name can not be bull"),
		},
	}

	for k, v := range testCases {
		actual := pc.Run(v.desc)
		if v.expect.Error() != actual.Error() {
			t.Errorf("%s Run failed, expect %v, actual %v", k, v.expect, actual)
		}
	}
}

func TestPlugConsRollback(t *testing.T) {
	plugClear()
	defer plugClear()
	pc := plugPrepare(t)
	testCases := map[string]struct {
		desc   *contract.TxDesc
		expect error
	}{
		"test default": {
			desc: &contract.TxDesc{
				Method: "default",
			},
			expect: errors.New("PluggableConsensus not define this method"),
		},
		"test update_consensus": {
			desc: &contract.TxDesc{
				Method: updateConsensusMethod,
				Args:   map[string]interface{}{},
			},
			expect: errors.New("Consensus name can not be bull"),
		},
	}

	for k, v := range testCases {
		actual := pc.Rollback(v.desc)
		if v.expect.Error() != actual.Error() {
			t.Errorf("%s rollback failed, expect %v, actual %v", k, v.expect, actual)
		}
	}
}

func TestValidateUpdateConsensus(t *testing.T) {
	plugClear()
	defer plugClear()
	pc := plugPrepare(t)
	pc.context = &contract.TxContext{}
	testCases := map[string]struct {
		desc   *contract.TxDesc
		expect error
	}{
		"test1": {
			desc: &contract.TxDesc{
				Method: updateConsensusMethod,
				Args: map[string]interface{}{
					"name":   "dpos",
					"config": map[string]interface{}{},
				},
				Tx: &pb.Transaction{
					Txid: []byte("yyyy"),
				},
			},
			expect: nil,
		},
	}
	for k, v := range testCases {
		_, _, actual := pc.validateUpdateConsensus(v.desc.Args)
		if v.expect != actual {
			t.Errorf("%s ValidateUpdateConsensus failed, expect %v, actual %v", k, v.expect, actual)
		}
	}
}

func TestRollbackConsensus(t *testing.T) {
	plugClear()
	defer plugClear()
	pc := plugPrepare(t)
	plgMgr, _ := pluginmgr.GetPluginMgr()
	var ldb kvdb.Database
	soInst, _ := plgMgr.PluginMgr.CreatePluginInstance("kv", "default")
	ldb = soInst.(kvdb.Database)
	err := ldb.Open(testspace, map[string]interface{}{
		"cache":     512,
		"fds":       1024,
		"dataPaths": []string{},
	})

	pc.context = &contract.TxContext{
		UtxoBatch: ldb.NewBatch(),
	}
	second := &StepConsensus{}
	second.StartHeight = 100
	second.Conn = &tdpos.TDpos{}
	pc.cons = append(pc.cons, second)

	block := &pb.InternalBlock{
		Height: 100,
	}

	err = pc.rollbackConsensus("", nil, []byte("test"), block)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestUpdateConsensus(t *testing.T) {
	plugClear()
	defer plugClear()
	pc := plugPrepare(t)
	block := &pb.InternalBlock{
		Height: 100,
	}
	err := pc.updateConsensus("default", nil, []byte("test"), block)
	if err.Error() != "Consensus not support" {
		t.Error(err.Error())
	}
}

func TestUpdateSimpleDPosConsensus(t *testing.T) {
	plugClear()
	defer plugClear()
	pc := plugPrepare(t)
	pc.cfg.Miner.Keypath = "../data/keys"
	height := int64(100)
	timestamp := int64(12345678)
	consConf := make(map[string]interface{})
	consConf["proposer_num"] = "3"
	consConf["period"] = "3000"
	consConf["block_num"] = "20"
	consConf["vote_unit_price"] = "500"
	extParams := make(map[string]interface{})
	extParams["timestamp"] = timestamp
	_, err := pc.updateConsensusByName(ConsensusTypeTdpos, height, consConf, extParams)
	if err == nil {
		t.Error("err can not be null")
	}
}

func TestUpdateConsensusOther(t *testing.T) {
	plugClear()
	defer plugClear()
	pc := plugPrepareWithGensisBlock(t)
	pc.context = &contract.TxContext{}
	pc.context.UtxoBatch = pc.utxoVM.NewBatch()
	// 组装一个智能合约交易(没有input,output,只有desc)

	desc := contract.TxDesc{
		Module: "consensus",
		Method: updateConsensusMethod,
		Args: map[string]interface{}{
			"name": "tdpos",
			"config": map[string]interface{}{
				"proposer_num":       "3",
				"period":             "3000",
				"alternate_interval": "6000",
				"term_interval":      "9000",
				"block_num":          "1",
				"vote_unit_price":    "1",
				"init_proposer": map[string]interface{}{
					"1": []interface{}{
						"Y4TmpfV4pvhYT5W17J7TqHSLo6cqq23x3",
						"f3prTg9itaZY6m48wXXikXdcxiByW7zgk",
						"U9sKwFmgJVfzgWcfAG47dKn1kLQTqeZN3",
					},
				},
				"init_proposer_neturl": map[string]interface{}{
					"1": []interface{}{
						"/ip4/127.0.0.1/tcp/47101/p2p/QmVxeNubpg1ZQY4TmpfV4pvhYT5W17J7TqHSLo6cqq23x3",
						"/ip4/127.0.0.1/tcp/47102/p2p/QmVxeNubpg1ZQjQT8W5yZC9fD7ZB1ViArwvyGUB53sqf8e",
						"/ip4/127.0.0.1/tcp/47103/p2p/U9sKwFmgJVfzgWcfAG47dKn1kLQTqeZN3ZB1ViArwvTmpa",
					},
				},
			},
		},
	}

	txReq := &pb.TxData{}
	txReq.Bcname = "xuper-chain"
	txReq.Nonce = "nonce"
	txReq.Desc = []byte("")
	txReq.FromAddr = AliceAddress
	txReq.FromPubkey = AlicePubkey
	txReq.Timestamp = time.Now().UnixNano()
	txReq.FromScrkey = AlicePrivateKey
	tx, err := pc.utxoVM.GenerateTx(txReq)
	if err != nil {
		t.Fatal(err.Error())
	} else {
		t.Log("tx ", tx.Txid)
	}

	txList := []*pb.Transaction{}
	txList = append(txList, tx)
	ecdsaPk, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	preHash := pc.utxoVM.GetLatestBlockid()
	pendingBlock, _ := pc.ledger.FormatBlock(txList, []byte("miner-1"), ecdsaPk, 123456789, 0, 0, preHash, pc.utxoVM.GetTotal())

	name, consConf, validateErr := pc.validateUpdateConsensus(desc.Args)
	if validateErr != nil {
		t.Fatal(validateErr)
	} else {
		t.Log("name ", name)
		t.Log("consConf ", consConf)
	}
	updateErr := pc.updateConsensus(name, consConf, tx.Txid, pendingBlock)
	if updateErr != nil {
		t.Fatal(updateErr)
	}
}

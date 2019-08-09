package kernel

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"testing"
	"time"

	log "github.com/xuperchain/log15"
	"github.com/xuperchain/xuperunion/contract"
	"github.com/xuperchain/xuperunion/crypto/client"
	crypto_client "github.com/xuperchain/xuperunion/crypto/client"
	"github.com/xuperchain/xuperunion/global"
	"github.com/xuperchain/xuperunion/ledger"
	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/utxo"
)

const BobAddress = "dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN"
const BobPubkey = `{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571}`
const BobPrivateKey = `{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571,"D":29079635126530934056640915735344231956621504557963207107451663058887647996601}`
const AliceAddress = "WNWk3ekXeM5M2232dY2uCJmEqWhfQiDYT"
const defaultKVEngine = "default"

func bobToAlice(t *testing.T, utxovm *utxo.UtxoVM, ledger *ledger.Ledger, amount string, prehash []byte, desc string) ([]byte, error) {
	t.Logf("pre_hash of this block: %x", prehash)
	txreq := &pb.TxData{}
	txreq.Bcname = "xuper-chain"
	txreq.FromAddr = BobAddress
	txreq.FromPubkey = BobPubkey
	txreq.FromScrkey = BobPrivateKey
	txreq.Nonce = "nonce"
	txreq.Timestamp = time.Now().UnixNano()
	txreq.Desc = []byte(desc)
	//bob给alice转20
	txreq.Account = []*pb.TxDataAccount{
		{Address: AliceAddress, Amount: amount},
	}
	timer := global.NewXTimer()
	tx, err := utxovm.GenerateTx(txreq)
	if err != nil {
		t.Fatal(err)
		return nil, err
	}
	timer.Mark("GenerateTx")
	err = utxovm.DoTx(tx)
	timer.Mark("DoTx")
	if err != nil {
		t.Fatal(err)
		return nil, err
	}
	txlist, err := utxovm.GetUnconfirmedTx(true)
	timer.Mark("GetUnconfirmedTx")
	if err != nil {
		return nil, err
	}
	//奖励矿工
	awardtx, err := utxovm.GenerateAwardTx([]byte("miner-1"), "1000", []byte("award,onyeah!"))
	timer.Mark("GenerateAwardTx")
	if err != nil {
		return nil, err
	}
	txlist = append(txlist, awardtx)
	ecdsapk, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	timer.Mark("GenerateKey")
	block, _ := ledger.FormatBlock(txlist, []byte("miner-1"), ecdsapk, 123456789, 0, 0, prehash, utxovm.GetTotal())
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

func TestCreateBlockChain(t *testing.T) {
	workspace, workSpaceErr := ioutil.TempDir("/tmp", "")
	if workSpaceErr != nil {
		t.Error("create dir error ", workSpaceErr.Error())
	}
	defer os.RemoveAll(workspace)

	ledger, err := ledger.NewLedger(workspace+"xuper", nil, nil, defaultKVEngine, crypto_client.CryptoTypeDefault)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ledger)
	kl := &Kernel{}
	kLogger := log.New("module", "kernel")
	kLogger.SetHandler(log.StreamHandler(os.Stderr, log.LogfmtFormat()))
	kl.Init(workspace, kLogger, nil, "xuper")
	kl.SetNewChainWhiteList(map[string]bool{BobAddress: true})
	kl.SetMinNewChainAmount("0")
	utxovm, _ := utxo.MakeUtxoVM("xuper", ledger, workspace+"xuper", "", "", []byte(""), nil, 5000, 60, 500, nil, false, defaultKVEngine, crypto_client.CryptoTypeDefault)
	utxovm.RegisterVM("kernel", kl, global.VMPrivRing0)
	//创建链的时候分配财富
	tx, err := utxovm.GenerateRootTx([]byte(`
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
	err = utxovm.Play(block.Blockid)
	if err != nil {
		t.Fatal(err)
	}
	//通过tx创建一个基础链:Dog链
	nextBlockid, err := bobToAlice(t, utxovm, ledger, "1", block.Blockid, `{"module":"kernel", "method":"CreateBlockChain", "args": {"data": "{\n\t\"version\" : \"1\"\n\t, \"consensus\" : {\n\t\t\"miner\" : \"dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN\"\n\t}\n\t, \"predistribution\":[\n\t\t{\n\t\t\t\"address\" : \"dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN\"\n\t\t\t, \"quota\" : \"1000000000000000\"\n\t\t}\n\t]\n\t, \"maxblocksize\" : \"128\"\n\t, \"period\" : \"3000\"\n\t, \"award\" : \"1000000\"\n}\n", "name": "Dog"}}`)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Logf("next block id: %x", nextBlockid)
	}
	err = utxovm.Play(nextBlockid)
	if err != nil {
		t.Fatal(err)
	}
	//强行walk到根节点，触发createblockchain的回滚测试
	err = utxovm.Walk(ledger.GetMeta().RootBlockid)
	if err != nil {
		t.Fatal(err)
	}

	// test for GetVATWhiteList
	vatWhiteList := kl.GetVATWhiteList()
	t.Log("vatWhiteList ", vatWhiteList)

	// test for GetVerifiableAutogenTx
	txList, vatErr := kl.GetVerifiableAutogenTx(1, 1, 123456789)
	if vatErr != nil {
		t.Error("GetVerifiableAutogenTx error ", vatErr.Error())
	} else {
		t.Log("txList ", txList)
	}
	// test for Finalize
	finalErr := kl.Finalize(nextBlockid)
	if finalErr != nil {
		t.Error("Finalize error ", finalErr.Error())
	}
	// test for Stop
	kl.Stop()
}

func TestCreateBlockChainPermission(t *testing.T) {
	workspace, workSpaceErr := ioutil.TempDir("/tmp", "")
	if workSpaceErr != nil {
		t.Error("create dir error ", workSpaceErr.Error())
	}
	defer os.RemoveAll(workspace)

	chainName := "lovechain"
	ledger, err := ledger.NewLedger(workspace+chainName, nil, nil, defaultKVEngine, crypto_client.CryptoTypeDefault)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ledger)
	kl := &Kernel{}
	kLogger := log.New("module", "kernel")
	kLogger.SetHandler(log.StreamHandler(os.Stderr, log.LogfmtFormat()))
	kl.Init(workspace, kLogger, nil, chainName)
	kl.SetNewChainWhiteList(map[string]bool{BobAddress: true})
	utxovm, _ := utxo.MakeUtxoVM(chainName, ledger, workspace+chainName, "", "", []byte(""), nil, 5000, 60, 500, nil, false, defaultKVEngine, crypto_client.CryptoTypeDefault)
	utxovm.RegisterVM("kernel", kl, global.VMPrivRing0)
	//创建链的时候分配财富
	tx, err := utxovm.GenerateRootTx([]byte(`
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
	err = utxovm.Play(block.Blockid)
	if err != nil {
		t.Fatal(err)
	}
	//通过tx创建一个基础链:Dog链
	nextBlockid, err := bobToAlice(t, utxovm, ledger, "1", block.Blockid, `{"module":"kernel", "method":"CreateBlockChain", "args": {"data": "{\n\t\"version\" : \"1\"\n\t, \"consensus\" : {\n\t\t\"miner\" : \"dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN\"\n\t}\n\t, \"predistribution\":[\n\t\t{\n\t\t\t\"address\" : \"dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN\"\n\t\t\t, \"quota\" : \"1000000000000000\"\n\t\t}\n\t]\n\t, \"maxblocksize\" : \"128\"\n\t, \"period\" : \"3000\"\n\t, \"award\" : \"1000000\"\n}\n", "name": "Dog"}}`)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Logf("next block id: %x", nextBlockid)
	}
	err = utxovm.Play(nextBlockid)
	if err == nil {
		//t.Fatal("expected permission denied")
		//Play现在的机制如果能rollback成功就不向上返回err
	}
}

func TestGetKVEngineType(t *testing.T) {
	data := map[string]interface{}{
		"kvengine": "default",
		"crypto":   client.CryptoTypeDefault,
	}
	json, _ := json.Marshal(data)
	kl := &Kernel{}
	kvType, err := kl.GetKVEngineType(json)
	if err != nil {
		t.Error("GetKVEngineType error ", err.Error())
	} else {
		t.Log("KVEngineType ", kvType)
	}

}

func TestGetCryptoType(t *testing.T) {
	data := map[string]interface{}{
		"kvengine": "default",
		"crypto":   client.CryptoTypeDefault,
	}
	json, _ := json.Marshal(data)
	kl := &Kernel{}
	cryptoType, err := kl.GetCryptoType(json)
	if err != nil {
		t.Error("GetCryptoType error ", err.Error())
	} else {
		t.Log("CryptoType ", cryptoType)
	}
}

func TestRunUpdateMaxBlockSize(t *testing.T) {
	workspace, workSpaceErr := ioutil.TempDir("/tmp", "")
	if workSpaceErr != nil {
		t.Error("create dir error ", workSpaceErr.Error())
	}
	defer os.RemoveAll(workspace)

	L, err := ledger.NewLedger(workspace+"xuper", nil, nil, defaultKVEngine, crypto_client.CryptoTypeDefault)
	if err != nil {
		t.Error("new ledger error ", err.Error())
	}
	utxovm, _ := utxo.MakeUtxoVM("xuper", L, workspace+"xuper", "", "", []byte(""), nil, 5000, 60, 500, nil, false, defaultKVEngine, crypto_client.CryptoTypeDefault)
	tx, generateRootErr := utxovm.GenerateRootTx([]byte(`
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
	if generateRootErr != nil {
		t.Error("generate genesis tx error ", generateRootErr.Error())
	}
	block, _ := L.FormatRootBlock([]*pb.Transaction{tx})
	t.Logf("blockid %x", block.Blockid)
	confirmStatus := L.ConfirmBlock(block, true)
	if !confirmStatus.Succ {
		t.Error("confirm block fail")
	}
	playErr := utxovm.Play(block.Blockid)
	if playErr != nil {
		t.Error(playErr)
	}
	t.Log("L.GetMaxBlockSize:", L.GetMaxBlockSize())
	context := &contract.TxContext{
		LedgerObj: L,
		UtxoBatch: L.GetBaseDB().NewBatch(),
	}
	kl := &Kernel{}
	kLogger := log.New("module", "kernel")
	kLogger.SetHandler(log.StreamHandler(os.Stderr, log.LogfmtFormat()))
	kl.Init(workspace, kLogger, nil, "xuper")
	kl.SetContext(context)
	txDesc := &contract.TxDesc{
		Args: map[string]interface{}{
			"new_block_size": (127.0 << 20) * 1.0,
			"old_block_size": (128.0 << 20) * 1.0,
		},
	}
	runUpdateBlkChainErr := kl.runUpdateMaxBlockSize(txDesc)
	if runUpdateBlkChainErr != nil {
		t.Error("runUpdateMaxBlockSize error ", runUpdateBlkChainErr.Error())
	}
}

func TestRunUpdateReservedContracts(t *testing.T) {
	workspace, workSpaceErr := ioutil.TempDir("/tmp", "")
	if workSpaceErr != nil {
		t.Error("create dir error ", workSpaceErr.Error())
	}
	defer os.RemoveAll(workspace)

	L, err := ledger.NewLedger(workspace+"xuper", nil, nil, defaultKVEngine, crypto_client.CryptoTypeDefault)
	if err != nil {
		t.Error("new ledger error ", err.Error())
	}
	utxovm, _ := utxo.MakeUtxoVM("xuper", L, workspace+"xuper", "", "", []byte(""), nil, 5000, 60, 500, nil, false, defaultKVEngine, crypto_client.CryptoTypeDefault)
	tx, generateRootErr := utxovm.GenerateRootTx([]byte(`
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
		, "reserved_contracts": [
            {
                "module_name": "wasm",
                "contract_name": "banned",
                "method_name": "verify",
                "args": {
                    "contract": "{{.ContractNames}}"
                }
            }
        ]
    }
    `))
	if generateRootErr != nil {
		t.Error("generate genesis tx error ", generateRootErr.Error())
	}
	block, _ := L.FormatRootBlock([]*pb.Transaction{tx})
	t.Logf("blockid %x", block.Blockid)
	confirmStatus := L.ConfirmBlock(block, true)
	if !confirmStatus.Succ {
		t.Error("confirm block fail")
	}
	playErr := utxovm.Play(block.Blockid)
	if playErr != nil {
		t.Error(playErr)
	}
	reservedContracts := []*pb.InvokeRequest{}
	originalReservedContracts, err := L.GenesisBlock.GetConfig().GetReservedContract()
	if err != nil {
		t.Error("originalReservedContracts ", originalReservedContracts)
	}
	MetaReservedContracts := L.GetMeta().ReservedContracts
	t.Log("MetaReservedContracts: ", MetaReservedContracts)
	if MetaReservedContracts != nil {
		reservedContracts = MetaReservedContracts
	} else {
		reservedContracts = originalReservedContracts
	}
	t.Log("reservedContracts: ", reservedContracts)
	context := &contract.TxContext{
		LedgerObj: L,
		UtxoBatch: L.GetBaseDB().NewBatch(),
	}
	kl := &Kernel{}
	kLogger := log.New("module", "kernel")
	kLogger.SetHandler(log.StreamHandler(os.Stderr, log.LogfmtFormat()))
	kl.Init(workspace, kLogger, nil, "xuper")
	kl.SetContext(context)
	args := []byte(`
        {
            "args":{
                "old_reserved_contracts":[
                    {
                        "module_name":"wasm",
                        "contract_name":"banned",
                        "method_name":"verify",
                        "args":{
                            "contract":"{{.ContractNames}}"
                        }
                    }
                ],
                "reserved_contracts":[
                {
                    "module_name":"wasm",
                    "contract_name":"identity",
                    "method_name":"verify",
                        "args":{}
                }
                ]
            }
        }
	`)
	txDesc := &contract.TxDesc{}
	_ = json.Unmarshal(args, txDesc)
	runUpdateBlkChainErr := kl.runUpdateReservedContract(txDesc)
	if runUpdateBlkChainErr != nil {
		t.Error("runUpdateReservedContracts error: ", runUpdateBlkChainErr.Error())
	}

	rollbackUpdateBlkChainErr := kl.rollbackUpdateReservedContract(txDesc)
	if rollbackUpdateBlkChainErr != nil {
		t.Error("runUpdateReservedContracts error: ", rollbackUpdateBlkChainErr.Error())
	}
}

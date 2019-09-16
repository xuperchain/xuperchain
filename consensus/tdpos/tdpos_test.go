package tdpos

import (
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/xuperchain/xuperunion/common/config"
	cons_base "github.com/xuperchain/xuperunion/consensus/base"
	"github.com/xuperchain/xuperunion/contract"
	crypto_client "github.com/xuperchain/xuperunion/crypto/client"
	"github.com/xuperchain/xuperunion/global"
	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/utxo"
)

func close() {
	os.RemoveAll(workspace)
}

func TestTDpos(t *testing.T) {
	defer close()
	utxoObj := utxo.NewFakeUtxoVM(t, workspace, true)
	cfg := config.NewNodeConfig()
	cfg.Miner.Keypath = "../../data/keys"
	bcname := "xuper"
	consConf := `{
		"module": "consensus",
		"method": "update_consensus",
		"args": {
			"name": "tdpos",
			"config": {
				"proposer_num": "3",
				"period": "3000",
				"alternate_interval": "6000",
				"term_interval": "9000",
				"block_num": "10",
				"vote_unit_price": "1",
				"init_proposer": {
					"1": [
						"Y4TmpfV4pvhYT5W17J7TqHSLo6cqq23x3",
						"f3prTg9itaZY6m48wXXikXdcxiByW7zgk",
						"U9sKwFmgJVfzgWcfAG47dKn1kLQTqeZN3"
					]
				},
				"init_proposer_neturl": {
					"1": [
						"/ip4/127.0.0.1/tcp/47101/p2p/QmVxeNubpg1ZQY4TmpfV4pvhYT5W17J7TqHSLo6cqq23x3",
						"/ip4/127.0.0.1/tcp/47102/p2p/QmVxeNubpg1ZQjQT8W5yZC9fD7ZB1ViArwvyGUB53sqf8e",
						"/ip4/127.0.0.1/tcp/47103/p2p/U9sKwFmgJVfzgWcfAG47dKn1kLQTqeZN3ZB1ViArwvTmpa"
					]
				}
			}
		}
	}`
	desc, _ := contract.Parse(string([]byte(consConf)))
	conf := desc.Args["config"].(map[string]interface{})
	cryptoClient, err := crypto_client.CreateCryptoClient(crypto_client.CryptoTypeDefault)
	if err != nil {
		t.Fatal(err)
	}
	tdpos := &TDpos{}
	tdpos.Init()
	tdpos.config.initProposer = map[int64][]*cons_base.CandidateInfo{
		1: []*cons_base.CandidateInfo{
			&cons_base.CandidateInfo{
				Address:  "dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN",
				PeerAddr: "/ip4/127.0.0.1/tcp/47101/p2p/QmVxeNubpg1ZQY4TmpfV4pvhYT5W17J7TqHSLo6cqq23x3",
			},
			&cons_base.CandidateInfo{
				Address:  "Y4TmpfV4pvhYT5W17J7TqHSLo6cqq23x3",
				PeerAddr: "/ip4/127.0.0.1/tcp/47102/p2p/QmVxeNubpg1ZQjQT8W5yZC9fD7ZB1ViArwvyGUB53sqf8e",
			},
			&cons_base.CandidateInfo{
				Address:  "RUEMFGDEnLBpnYYggnXukpVfR9Skm59ph",
				PeerAddr: "/ip4/127.0.0.1/tcp/47103/p2p/U9sKwFmgJVfzgWcfAG47dKn1kLQTqeZN3ZB1ViArwvTmpa",
			},
			&cons_base.CandidateInfo{
				Address:  "bob",
				PeerAddr: "peerid4",
			},
		},
	}
	extParams := make(map[string]interface{})
	extParams["timestamp"] = time.Now().UnixNano()
	extParams["bcname"] = bcname
	extParams["ledger"] = utxoObj.L
	extParams["utxovm"] = utxoObj.U
	extParams["crypto_client"] = cryptoClient
	err = tdpos.Configure(nil, cfg, conf, extParams)
	if err != nil {
		t.Error("configure error ", err.Error())
	}
	utxoObj.U.RegisterVM(TYPE, tdpos, global.VMPrivRing0)
	utxoObj.UtxoWorkWithLedgerBasic(t)

	voteJSON := `{
  "module": "tdpos",
  "method": "vote",
  "args" : {
    "candidates":["Y4TmpfV4pvhYT5W17J7TqHSLo6cqq23x3","nEm7kPvfYpHn35EWFvh3VjGcwMZRuxtnJ","f3prTg9itaZY6m48wXXikXdcxiByW7zgk"]
    }
}`
	//投票
	_, _, tx, err := utxoObj.Transfer("alice", "bob", t, "1", utxoObj.L.GetMeta().TipBlockid, []byte(voteJSON), 10000)
	if tx.Txid == nil {
		t.Fatal("transfer failed", err)
	} else {
		t.Logf("tdpos vote tx id is %s", fmt.Sprintf("%x", tx))
	}
	utxoObj.U.Play(utxoObj.L.GetMeta().TipBlockid)
	isMaster, isSyncBlock := tdpos.CompeteMaster(100)
	t.Log("is master: ", isMaster, " isSyncBlock ", isSyncBlock)
	param, _ := tdpos.ProcessBeforeMiner(time.Now().UnixNano())
	t.Logf("params %v", param)
	pubJSON, _ := cryptoClient.GetEcdsaPrivateKeyFromFile("../../data/keys/private.key")

	// tx_list, proposer, ecdsa_pk, timestamp, curTerm, curBlockNum, pre_hash, 0, utxoTotal, false
	txList := []*pb.Transaction{}
	proposer := []byte("dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN")
	timestamp := time.Now().UnixNano()
	curTerm := int64(2)
	curBlockNum := int64(1)
	preHash := utxoObj.L.GetMeta().TipBlockid
	block1, _ := utxoObj.L.FormatBlock(txList, proposer, pubJSON, timestamp, curTerm, curBlockNum, preHash, big.NewInt(0))

	header := global.GHeader()

	ok, checkMinerErr := tdpos.CheckMinerMatch(header, block1)
	if checkMinerErr != nil {
		t.Error("CheckMinerMatch error ", checkMinerErr.Error())
	}
	t.Log(ok)

	t.Log(tdpos.Type())
	t.Log(tdpos.Version())
	t.Log(tdpos.InitCurrent(block1))
	tdpos.Stop()
	tdpos.ProcessConfirmBlock(block1)
	whiteList := tdpos.GetVATWhiteList()
	t.Log(whiteList)
	txVat, errVat := tdpos.GetVerifiableAutogenTx(5, 1, 12345678910)
	if errVat != nil {
		t.Error("GetVerifiableAutogenTx error ")
	} else {
		t.Log(txVat)
	}
	contractTxContent := &contract.TxContext{}
	errSetContent := tdpos.SetContext(contractTxContent)
	if errSetContent != nil {
		t.Error("SetContext error ", errSetContent.Error())
	}
	// test for rollback
	descRollBack := &contract.TxDesc{
		Method: "default",
	}
	rollbackErr := tdpos.Rollback(descRollBack)
	if rollbackErr != nil {
		t.Error("Rollback error ", rollbackErr.Error())
	}
}

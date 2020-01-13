package tdpos

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"sync"
	"testing"

	log "github.com/xuperchain/log15"
	"github.com/xuperchain/xuperunion/contract"
	crypto_client "github.com/xuperchain/xuperunion/crypto/client"
	"github.com/xuperchain/xuperunion/ledger"
	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/utxo"
)

func commonWork(t *testing.T) (U *utxo.UtxoVM, L *ledger.Ledger, T *TDpos) {
	workspace, _ := ioutil.TempDir("/tmp", "")
	defer os.RemoveAll(workspace)
	L, err1 := ledger.NewLedger(workspace, nil, nil, "default", crypto_client.CryptoTypeDefault)
	if err1 != nil {
		t.Fatal(err1)
	}

	rootTx, rootTxErr := utxo.GenerateRootTx([]byte(`
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
	if rootTxErr != nil {
		t.Error("GenerateRootTx error ", rootTxErr.Error())
	}
	rootBlock, formatErr := L.FormatRootBlock([]*pb.Transaction{rootTx})
	if formatErr != nil {
		t.Error("format genesis block error ", formatErr.Error())
	}
	L.ConfirmBlock(rootBlock, true)

	U, err2 := utxo.NewUtxoVM("xuper", L, workspace, minerPrivateKey, minerPublicKey, []byte(minerAddress), nil, false, "default", crypto_client.CryptoTypeDefault)
	if err2 != nil {
		t.Fatal(err2)
	}
	playErr := U.Play(rootBlock.Blockid)
	if playErr != nil {
		t.Error("play error ", playErr.Error())
	}

	xlog := log.New("module", "consensus")
	xlog.SetHandler(log.StreamHandler(os.Stderr, log.LogfmtFormat()))
	T = &TDpos{
		log:                   xlog,
		initTimestamp:         int64(1529390439100327238),
		ledger:                L,
		utxoVM:                U,
		candidateBallots:      &sync.Map{},
		candidateBallotsCache: &sync.Map{},
		revokeCache:           &sync.Map{},
		config: tDposConfig{
			period:            int64(3000 * 1e6),
			alternateInterval: int64(6000 * 1e6),
			termInterval:      int64(9000 * 1e6),
			proposerNum:       int64(3),
			blockNum:          int64(20),
			voteUnitPrice:     big.NewInt(12),
		},
	}

	return U, L, T
}

func makeTxWithDesc(strDesc []byte, U *utxo.UtxoVM, L *ledger.Ledger, t *testing.T) (*pb.Transaction, *pb.InternalBlock) {
	// make a tx
	txReq := &pb.TxData{}
	txReq.Bcname = "xuper-chain"
	txReq.Nonce = "nonce"
	txReq.Desc = []byte(strDesc)
	txReq.FromAddr = AliceAddress
	txReq.FromPubkey = AlicePubkey
	txReq.FromScrkey = AlicePrivateKey
	txDataAccount := &pb.TxDataAccount{
		Address: AliceAddress,
		Amount:  "1",
	}
	txReq.Account = append(txReq.Account, txDataAccount)
	txCons, errCons := U.GenerateTx(txReq)
	//fmt.Println("----------------------transaction:", txCons)
	if errCons != nil {
		t.Error("GenerateTx error ", errCons.Error())
	}
	errDo := U.DoTx(txCons)
	if errDo != nil {
		t.Error("Do tx error ", errDo.Error())
	}
	txList, errPack := U.GetUnconfirmedTx(true)
	if errPack != nil {
		t.Error("GetUnconfirmedTx error ", errPack.Error())
	}
	preHash := L.GetMeta().TipBlockid
	ecdsaPk, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	block, _ := L.FormatBlock(txList, []byte("miner-1"), ecdsaPk, 123456789, 0, 0, preHash, U.GetTotal())
	confirmStatus := L.ConfirmBlock(block, false)
	if !confirmStatus.Succ {
		t.Error("ConfirmBlock block error")
	}

	return txCons, block
}

func TestRunVote(t *testing.T) {
	desc := &contract.TxDesc{
		Module: "tdpos",
		Method: "vote",
		Args: map[string]interface{}{
			"candidates": []interface{}{"f3prTg9itaZY6m48wXXikXdcxiByW7zgk"},
		},
	}
	strDesc, _ := json.Marshal(desc)

	U, L, tdpos := commonWork(t)
	txCons, block := makeTxWithDesc(strDesc, U, L, t)
	t.Log("block id ", block.Blockid)

	desc2 := &contract.TxDesc{
		Module: "tdpos",
		Method: "vote",
		Args: map[string]interface{}{
			"txid":       fmt.Sprintf("%x", txCons.Txid),
			"candidates": []interface{}{"f3prTg9itaZY6m48wXXikXdcxiByW7zgk"},
		},
		Tx: txCons,
	}

	tdpos.candidateBallots.LoadOrStore("D_candidate_ballots_f3prTg9itaZY6m48wXXikXdcxiByW7zgk", int64(1))
	tdpos.context = &contract.TxContext{}
	tdpos.context.UtxoBatch = tdpos.utxoVM.NewBatch()
	err := tdpos.runVote(desc2, nil)
	if err != nil {
		t.Error("run vote error ", err.Error())
	}
	// 上面投过一次票,已经加缓存了,这次再投票走缓存路径
	err = tdpos.runVote(desc2, nil)
	if err != nil {
		t.Error("run vote error ", err.Error())
	}
}

func TestRevokeVote(t *testing.T) {
	desc := &contract.TxDesc{
		Module: "tdpos",
		Method: "revoke_vote",
		Args: map[string]interface{}{
			"candidates": []interface{}{"f3prTg9itaZY6m48wXXikXdcxiByW7zgk"},
		},
	}
	strDesc, _ := json.Marshal(desc)

	U, L, tdpos := commonWork(t)
	txCons, block := makeTxWithDesc(strDesc, U, L, t)

	tdpos.candidateBallots.LoadOrStore("D_candidate_ballots_f3prTg9itaZY6m48wXXikXdcxiByW7zgk", int64(1))
	tdpos.context = &contract.TxContext{}
	tdpos.context.UtxoBatch = tdpos.utxoVM.NewBatch()
	desc2 := &contract.TxDesc{
		Module: "tdpos",
		Method: "revoke_vote",
		Tx:     txCons,
		Args: map[string]interface{}{
			"txid": fmt.Sprintf("%x", txCons.Txid),
		},
	}
	revokeVoteErr := tdpos.runRevokeVote(desc2, block)
	if revokeVoteErr != nil {
		t.Error("runRevokeVote error ", revokeVoteErr.Error())
	}
}

func TestRunNominateCandidate(t *testing.T) {
	desc := &contract.TxDesc{
		Module: "tdpos",
		Method: "nominate_candidate",
		Args: map[string]interface{}{
			"candidates": []interface{}{"f3prTg9itaZY6m48wXXikXdcxiByW7zgk"},
			"neturls":    []interface{}{"/ip4/127.0.0.1/tcp/47101/p2p/QmVxeNubpg1ZQjQT8W5yZC9fD7ZB1ViArwvyGUB53sqf8e"},
		},
	}
	strDesc, _ := json.Marshal(desc)

	U, L, tdpos := commonWork(t)
	txCons, block := makeTxWithDesc(strDesc, U, L, t)

	tdpos.context = &contract.TxContext{}
	tdpos.context.UtxoBatch = tdpos.utxoVM.NewBatch()
	desc2 := &contract.TxDesc{
		Module: "tdpos",
		Method: "nominate_candidate",
		Tx:     txCons,
		Args: map[string]interface{}{
			"txid":      fmt.Sprintf("%x", txCons.Txid),
			"candidate": "f3prTg9itaZY6m48wXXikXdcxiByW7zgk",
			"neturl":    "/ip4/127.0.0.1/tcp/47101/p2p/QmVxeNubpg1ZQjQT8W5yZC9fD7ZB1ViArwvyGUB53sqf8e",
		},
	}
	nomCandErr := tdpos.runNominateCandidate(desc2, block)
	if nomCandErr == nil {
		//t.Error("runNominateCandidate error ", nomCandErr.Error())
		t.Error("candiate not auth")
	}
	/*
		nomCandErr = tdpos.runNominateCandidate(desc2, block)
		if nomCandErr == nil {
			t.Error("runNominateCandidate error ")
		}*/
}

func TestRunRevokeCandidate(t *testing.T) {
	desc := &contract.TxDesc{
		Module: "tdpos",
		Method: "nominate_candidate",
		Args: map[string]interface{}{
			"candidate": "f3prTg9itaZY6m48wXXikXdcxiByW7zgk",
			"neturl":    "/ip4/127.0.0.1/tcp/47101/p2p/QmVxeNubpg1ZQjQT8W5yZC9fD7ZB1ViArwvyGUB53sqf8e",
		},
	}
	strDesc, _ := json.Marshal(desc)

	U, L, tdpos := commonWork(t)
	txCons, block := makeTxWithDesc(strDesc, U, L, t)

	tdpos.context = &contract.TxContext{}
	tdpos.context.UtxoBatch = tdpos.utxoVM.NewBatch()
	desc2 := &contract.TxDesc{
		Module: "tdpos",
		Method: "revoke_candidate",
		Tx:     txCons,
		Args: map[string]interface{}{
			"txid":      fmt.Sprintf("%x", txCons.Txid),
			"candidate": "f3prTg9itaZY6m48wXXikXdcxiByW7zgk",
		},
	}
	key := "D_candidate_nominate_f3prTg9itaZY6m48wXXikXdcxiByW7zgk"
	value := fmt.Sprintf("%x", txCons.Txid) // txCons.Txid: bytes
	t.Log("run_test txNom ", value)
	tdpos.context.UtxoBatch.Put([]byte(key), txCons.Txid)
	tdpos.context.UtxoBatch.Write()

	val, errGetFromTable := tdpos.utxoVM.GetFromTable(nil, []byte(key))
	if errGetFromTable != nil {
		t.Error("GetFromTable error ", errGetFromTable.Error())
	} else {
		t.Log("val ", val)
	}
	revokeCandErr := tdpos.runRevokeCandidate(desc2, block)
	if revokeCandErr != nil {
		t.Error("runRevokeCandidate error ", revokeCandErr.Error(), desc2)
	}
}

func TestRunCheckValidater(t *testing.T) {
	desc := &contract.TxDesc{
		Module: "tdpos",
		Method: "check_validater",
		Args: map[string]interface{}{
			"version": "2",
			"term":    "90",
		},
	}
	strDesc, _ := json.Marshal(desc)

	U, L, tdpos := commonWork(t)
	txCons, block := makeTxWithDesc(strDesc, U, L, t)

	tdpos.context = &contract.TxContext{}
	tdpos.context.UtxoBatch = tdpos.utxoVM.NewBatch()
	tdpos.candidateBallots.LoadOrStore("D_candidate_ballots_f3prTg9itaZY6m48wXXikXdcxiByW7zgk", int64(1))
	tdpos.candidateBallots.LoadOrStore("D_candixdate_ballots_dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN", int64(1))
	tdpos.candidateBallots.LoadOrStore("D_candidate_ballots_Y4TmpfV4pvhYT5W17J7TqHSLo6cqq23x3", int64(1))
	desc2 := &contract.TxDesc{
		Module: "tdpos",
		Method: "check_validater",
		Args: map[string]interface{}{
			"version": "2",
			"term":    "90",
		},
		Tx: txCons,
	}
	checkValidErr := tdpos.runCheckValidater(desc2, block)
	if checkValidErr == nil {
		t.Error("runCheckValidater error ")
	}
}

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
	cons_base "github.com/xuperchain/xuperunion/consensus/base"
	"github.com/xuperchain/xuperunion/contract"
	crypto_client "github.com/xuperchain/xuperunion/crypto/client"
	"github.com/xuperchain/xuperunion/ledger"
	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/utxo"
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

func TestMinerScheduling(t *testing.T) {
	xlog := log.New("module", "consensus")
	xlog.SetHandler(log.StreamHandler(os.Stderr, log.LogfmtFormat()))

	sd := &TDpos{
		log:           xlog,
		initTimestamp: int64(1529390439100327238),
		config: tDposConfig{
			period:            int64(3000 * 1e6),
			alternateInterval: int64(6000 * 1e6),
			termInterval:      int64(9000 * 1e6),
			proposerNum:       int64(3),
			blockNum:          int64(20),
		},
	}

	timestamp := int64(1529400579100395794)
	term, pos, blpos := sd.minerScheduling(timestamp)
	if term != int64(53) && pos != int64(2) && blpos != int64(11) {
		t.Errorf("getTermPos error expect term=%d pos=%d blpos=%d, actual term=%d pos=%d blpos=%d ", 57, 1, 1,
			term, pos, blpos)
	}
}

func TestValidateCheckValidater(t *testing.T) {
	desc := &contract.TxDesc{
		Args: map[string]interface{}{
			"name":    "tdpos",
			"version": "2",
			"term":    "90",
		},
	}
	xlog := log.New("module", "consensus")
	xlog.SetHandler(log.StreamHandler(os.Stderr, log.LogfmtFormat()))
	tdpos := &TDpos{
		log:           xlog,
		initTimestamp: int64(1529390439100327238),
		config: tDposConfig{
			period:            int64(3000 * 1e6),
			alternateInterval: int64(6000 * 1e6),
			termInterval:      int64(9000 * 1e6),
			proposerNum:       int64(3),
			blockNum:          int64(20),
		},
	}
	version, term, err := tdpos.validateCheckValidater(desc)
	if err != nil {
		t.Error("TestValidateCheckValidater error", err.Error())
	} else {
		t.Log("version ", version, "term ", term)
	}
}

var (
	kvengine    = "default"
	tCryptoType = crypto_client.CryptoTypeDefault
)

var workspace, _ = ioutil.TempDir("/tmp", "")
var testspace, _ = ioutil.TempDir("/tmp", "")

func TestValidateRevokeVote(t *testing.T) {
	defer func() {
		os.RemoveAll(workspace)
		os.RemoveAll(testspace)
	}()
	ledger, err := ledger.NewLedger(workspace, nil, nil, kvengine, tCryptoType)
	if err != nil {
		t.Error("NewLedger error ", err.Error())
	}
	utxoVM, _ := utxo.NewUtxoVM("xuper", ledger, workspace, minerPrivateKey, minerPublicKey, []byte(minerAddress), nil, false, kvengine, tCryptoType)
	tx, gensisErr := utxoVM.GenerateRootTx([]byte(`
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
		t.Error("GenerateRootTx error ", gensisErr.Error())
	}
	block, err := ledger.FormatRootBlock([]*pb.Transaction{tx})
	if err != nil {
		t.Error("FormatRootBlock error ", err.Error())
	}
	confirmStatus := ledger.ConfirmBlock(block, true)
	if !confirmStatus.Succ {
		t.Error("ConfirmBlock error ")
	}
	playErr := utxoVM.Play(block.Blockid)
	if playErr != nil {
		t.Error("utxo vm paly error ", playErr.Error())
	}
	// 生成第二个一般交易并上链,其中的desc创建如下
	desc := contract.TxDesc{
		Module: "tdpos",
		Method: "revoke_vote",
		Args: map[string]interface{}{
			"candidates": []interface{}{"f3prTg9itaZY6m48wXXikXdcxiByW7zgk"},
		},
	}
	strDesc, _ := json.Marshal(desc)

	t.Log("strDesc ", strDesc)
	txReq := &pb.TxData{}
	txReq.Bcname = "xuper-chain"
	txReq.Nonce = "nonce"
	txReq.Desc = []byte(strDesc)
	txReq.FromAddr = AliceAddress
	txReq.FromPubkey = AlicePubkey
	txReq.FromScrkey = AlicePrivateKey
	txCons, errCons := utxoVM.GenerateTx(txReq)
	if errCons != nil {
		t.Error("GenerateTx error ", errCons.Error())
	}

	// do tx & genunconfirmedtx & formatblock & play
	errDo := utxoVM.DoTx(txCons)
	if errDo != nil {
		t.Error("Do tx error ", errDo.Error())
	}
	txList, errPack := utxoVM.GetUnconfirmedTx(true)
	if errPack != nil {
		t.Error("GetUnconfirmedTx error ", errPack.Error())
	}
	preHash := block.Blockid
	ecdsaPk, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	block2, _ := ledger.FormatBlock(txList, []byte("miner-1"), ecdsaPk, 123456789, 0, 0, preHash, utxoVM.GetTotal())
	confirmStatus = ledger.ConfirmBlock(block2, false)
	if !confirmStatus.Succ {
		t.Error("ConfirmBlock block2 error")
	}
	desc2 := contract.TxDesc{
		Module: "tdpos",
		Method: "revoke_vote",
		Args: map[string]interface{}{
			"txid":       fmt.Sprintf("%x", txCons.Txid),
			"candidates": []interface{}{"f3prTg9itaZY6m48wXXikXdcxiByW7zgk"},
		},
	}

	xlog := log.New("module", "consensus")
	xlog.SetHandler(log.StreamHandler(os.Stderr, log.LogfmtFormat()))
	tdpos := &TDpos{
		log:                   xlog,
		initTimestamp:         int64(1529390439100327238),
		ledger:                ledger,
		utxoVM:                utxoVM,
		candidateBallots:      &sync.Map{},
		candidateBallotsCache: &sync.Map{},
		config: tDposConfig{
			period:            int64(3000 * 1e6),
			alternateInterval: int64(6000 * 1e6),
			termInterval:      int64(9000 * 1e6),
			proposerNum:       int64(3),
			blockNum:          int64(20),
			voteUnitPrice:     big.NewInt(12),
		},
	}
	tdpos.candidateBallots.LoadOrStore("D_candidate_ballots_f3prTg9itaZY6m48wXXikXdcxiByW7zgk", 1)

	desc3, validErr := tdpos.validRevoke(&desc2)
	if validErr != nil {
		t.Error("validRevoke error ", validErr.Error())
	}

	voteInfo, errValid := tdpos.validateVote(desc3)
	if errValid != nil {
		t.Error("validateVote error ", errValid.Error())
	} else {
		t.Log("voteInfo ", voteInfo)
	}

	voteInfo, txID, errValid := tdpos.validateRevokeVote(&desc2)
	if errValid != nil {
		t.Error("validateRevokeVote error ")
	} else {
		t.Log("voteInfo ", voteInfo)
		t.Log("txID ", txID)
	}
}

func TestTermProposerBasic(t *testing.T) {
	defer close()
	ledger, err := ledger.NewLedger(workspace, nil, nil, kvengine, tCryptoType)
	if err != nil {
		t.Error("NewLedger error ", err.Error())
	}
	utxoVM, _ := utxo.NewUtxoVM("xuper", ledger, workspace, minerPrivateKey, minerPublicKey, []byte(minerAddress), nil, false, kvengine, tCryptoType)
	tx, gensisErr := utxoVM.GenerateRootTx([]byte(`
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
		t.Error("generate genesis tx error ", gensisErr.Error())
	}
	block, err := ledger.FormatRootBlock([]*pb.Transaction{tx})
	if err != nil {
		t.Error("format genesis block error ", err.Error())
	}
	confirmStatus := ledger.ConfirmBlock(block, true)
	if !confirmStatus.Succ {
		t.Error("ledger confirm block error ")
	}
	playErr := utxoVM.Play(block.Blockid)
	if playErr != nil {
		t.Error("utxo play error ", playErr.Error())
	}

	xlog := log.New("module", "consensus")
	xlog.SetHandler(log.StreamHandler(os.Stderr, log.LogfmtFormat()))
	tdpos := &TDpos{
		log:                   xlog,
		initTimestamp:         int64(1529390439100327238),
		candidateBallots:      &sync.Map{},
		candidateBallotsCache: &sync.Map{},
		utxoVM:                utxoVM,
		ledger:                ledger,
		config: tDposConfig{
			period:            int64(3000 * 1e6),
			alternateInterval: int64(6000 * 1e6),
			termInterval:      int64(9000 * 1e6),
			proposerNum:       int64(1),
			blockNum:          int64(20),
			voteUnitPrice:     big.NewInt(12),
			initProposer: map[int64][]*cons_base.CandidateInfo{
				1: []*cons_base.CandidateInfo{
					&cons_base.CandidateInfo{
						Address:  "Y4TmpfV4pvhYT5W17J7TqHSLo6cqq23x3",
						PeerAddr: "peerid1",
					},
					&cons_base.CandidateInfo{
						Address:  "RUEMFGDEnLBpnYYggnXukpVfR9Skm59ph",
						PeerAddr: "peerid2",
					},
					&cons_base.CandidateInfo{
						Address:  "bob",
						PeerAddr: "peerid3",
					},
				},
			},
		},
	}
	tdpos.context = &contract.TxContext{}
	tdpos.context.UtxoBatch = tdpos.utxoVM.NewBatch()
	canInfo := &cons_base.CandidateInfo{
		Address:  "f3prTg9itaZY6m48wXXikXdcxiByW7zgk",
		PeerAddr: "peerid4",
	}
	canInfoData, _ := json.Marshal(canInfo)
	tdpos.candidateBallots.LoadOrStore("D_candidate_ballots_f3prTg9itaZY6m48wXXikXdcxiByW7zgk", int64(1))
	tdpos.context.UtxoBatch.Put([]byte("D_candidate_info_f3prTg9itaZY6m48wXXikXdcxiByW7zgk"),
		canInfoData)
	tdpos.context.UtxoBatch.Write()
	// test for genTermProposer
	strList, genTermProposerErr := tdpos.genTermProposer()
	if genTermProposerErr != nil {
		t.Error("genTermProposer error")
	}

	// test for getTermProposer
	strList = tdpos.getTermProposer(1)
	t.Log("term 1 proposer ", strList)
	strList = tdpos.getTermProposer(2)
	t.Log("term 2 ", strList)
	strList = tdpos.getTermProposer(100)
	t.Log("term 100 ", strList)

	// test for isProposer
	isProposer := tdpos.isProposer(2, 2, []byte("f3prTg9itaZY6m48wXXikXdcxiByW7zgk"))
	if isProposer != false {
		t.Error("expect false, but got ", isProposer)
	}
}

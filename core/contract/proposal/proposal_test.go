package proposal

import (
	//"fmt"
	"io/ioutil"
	//"os"
	//"testing"
	//log "github.com/xuperchain/log15"
	//"github.com/xuperchain/xuperchain/core/utxo"
	//"encoding/json"
	//"github.com/xuperchain/xuperchain/core/contract"
	//"github.com/xuperchain/xuperchain/core/global"
	//"github.com/xuperchain/xuperchain/core/pb"
)

var workspace, workSpaceErr = ioutil.TempDir("/tmp", "")

var proposeJSON = `
{
 "module": "proposal",
 "method": "Propose",
 "args" : {
   "min_vote_percent": 51,
   "stop_vote_height": 25
 },
 "trigger": {
      "height": 30,
      "module": "consensus",
      "method": "update_consensus",
      "args" : {
          "name": "tdpos",
           "config": {
              "proposer_num":"3",
              "period":"3000",
              "term_gap":"60000",
              "block_num":"10",
              "vote_unit_price":"1",
              "init_proposer": {
                "1":["Y4TmpfV4pvhYT5W17J7TqHSLo6cqq23x3","f3prTg9itaZY6m48wXXikXdcxiByW7zgk","U9sKwFmgJVfzgWcfAG47dKn1kLQTqeZN3"],
                "2":["Y4TmpfV4pvhYT5W17J7TqHSLo6cqq23x3","f3prTg9itaZY6m48wXXikXdcxiByW7zgk","U9sKwFmgJVfzgWcfAG47dKn1kLQTqeZN3"]
              }
            }

      }
  }
}
`

/*
func getLogger() log.Logger {
	ylog := log.New("module", "proposal_test")
	ylog.SetHandler(log.StreamHandler(os.Stderr, log.LogfmtFormat()))
	return ylog
}

func TestPropose(t *testing.T) {
	defer os.RemoveAll(workspace)
	xlog := getLogger()
	utxoObj := utxo.NewFakeUtxoVM(t, workspace, true)
	prp := NewProposal(xlog, utxoObj.L, utxoObj.U)
	utxoObj.U.RegisterVM("proposal", prp, global.VMPrivRing0)
	// 创世块
	utxoObj.UtxoWorkWithLedgerBasic(t)
	xlog.Info("total money", "total", utxoObj.U.GetTotal())
	bobBalance, _ := utxoObj.U.GetBalance(utxoObj.BobAddress)
	aliceBalance, _ := utxoObj.U.GetBalance(utxoObj.AliceAddress)
	minerBalance, _ := utxoObj.U.GetBalance("miner-1")
	xlog.Info("alice has:", "balance", aliceBalance)
	xlog.Info("bob has:", "balance", bobBalance)
	xlog.Info("miner has:", "balance", minerBalance)

	//提案
	_, _, proposalTxid, _ := utxoObj.Transfer("alice", "bob", t, "1", utxoObj.L.GetMeta().TipBlockid, []byte(proposeJSON), 1000)
	if proposalTxid == nil {
		t.Fatal("transfer failed")
	} else {
		xlog.Info("proposal tx id is ", "txid", fmt.Sprintf("%x", proposalTxid.Txid))
	}

	// test for IsPropose
	isPropose := prp.IsPropose(proposalTxid)
	if isPropose != true {
		t.Error("expect true, but got ", isPropose)
	}

	// test for getDescArg
	val, err := prp.getDescArg(proposalTxid, "min_vote_percent")
	if err != nil {
		t.Error("get desc arg error ", err.Error())
	} else if val != 51 {
		t.Error("expect 51 but got ", val)
	}

	voteJSON := `{
    "module":"proposal",
    "method": "Vote",
    "args" : {
        "txid":"` + fmt.Sprintf("%x", proposalTxid.Txid) + `"
     }
}`

	utxoObj.U.Play(utxoObj.L.GetMeta().TipBlockid)
	// 生效:30 投票截止:25 最少投票:51% 冻结高度1000
	for i := 0; i <= 30; i++ {
		//投票，价值500, 冻结到1000 Height
		utxoObj.Transfer("alice", "alice", t, "15000000000000000", utxoObj.L.GetMeta().TipBlockid, []byte(voteJSON), 1000)
		triggeredTxList, err := prp.GetVerifiableAutogenTx(utxoObj.L.GetMeta().TrunkHeight+1, -1, 0)
		if utxoObj.L.GetMeta().TrunkHeight == 29 && len(triggeredTxList) != 1 {
			t.Fatal("not triggered, unexpecated")
		}
		xlog.Info("Height", "H", utxoObj.L.GetMeta().TrunkHeight, "tx", triggeredTxList, "err", err)
		utxoObj.U.Play(utxoObj.L.GetMeta().TipBlockid)
	}
}
func TestParseTriggerKey(t *testing.T) {
	prp := NewProposal(nil, nil, nil)
	height := int64(20)
	txid := []byte("123")
	triggerKey := prp.makeTriggerKey(height, txid)
	if triggerKey != "T00000000000000000020_313233" {
		t.Error("expect triggerKey T00000000000000000020_313233, but got ", triggerKey)
	}

	height2, txid2, err := prp.parseTriggerKey([]byte(triggerKey))
	if err != nil {
		t.Error("parse trigger key error ", err.Error())
	}
	if height2 != height {
		t.Error("expect height2 ", height, "but got ", height2)
	}
	if string(txid2) != "123" {
		t.Error("expect string(txid2) 123 but got ", string(txid2))
	}
}

func TestMakeVoteKey(t *testing.T) {
	prp := NewProposal(nil, nil, nil)
	proposalTxid := []byte("123")
	voteTxid := []byte("456")
	voteKey := prp.makeVoteKey(proposalTxid, voteTxid)
	if voteKey != "V313233_343536" {
		t.Error("expect voteKey is V313233_343536 but got ", voteKey)
	}
}

func TestSaveTrigger(t *testing.T) {
	defer os.RemoveAll(workspace)
	xlog := getLogger()
	utxoObj := utxo.NewFakeUtxoVM(t, workspace, true)
	prp := NewProposal(xlog, utxoObj.L, utxoObj.U)
	utxoObj.U.RegisterVM("proposal", prp, global.VMPrivRing0)
	utxoObj.UtxoWorkWithLedgerBasic(t)
	triggerDesc := &contract.TriggerDesc{
		Height: 5,
	}
	proposeJSON2 := `{
 "module": "proposal",
 "method": "Propose",
 "args" : {
   "min_vote_percent": 51,
   "stop_vote_height": 3
 },
 "trigger": {
      "height": 5,
      "module": "consensus",
      "method": "update_consensus",
      "args" : {
          "name": "tdpos",
           "config": {
              "proposer_num":"3",
              "period":"3000",
              "term_gap":"60000",
              "block_num":"10",
              "vote_unit_price":"1",
              "init_proposer": {
                "1":["Y4TmpfV4pvhYT5W17J7TqHSLo6cqq23x3","f3prTg9itaZY6m48wXXikXdcxiByW7zgk","U9sKwFmgJVfzgWcfAG47dKn1kLQTqeZN3"],
                "2":["Y4TmpfV4pvhYT5W17J7TqHSLo6cqq23x3","f3prTg9itaZY6m48wXXikXdcxiByW7zgk","U9sKwFmgJVfzgWcfAG47dKn1kLQTqeZN3"]
              }
            }

      }
  }
}`
	// 提案
	_, _, proposalTx, _ := utxoObj.Transfer("alice", "alice", t, "19999990000000000", utxoObj.L.GetMeta().TipBlockid, []byte(proposeJSON2), 3)
	// 投票
	voteJSON := `{
    "module":"proposal",
    "method": "Vote",
    "args" : {
        "txid":"` + fmt.Sprintf("%x", proposalTx.Txid) + `"
    }
    }`
	t.Log("before vote ", utxoObj.L.GetMeta().TrunkHeight)
	_, _, voteTx, _ := utxoObj.Transfer("alice", "alice", t, "19999990000000000", utxoObj.L.GetMeta().TipBlockid, []byte(voteJSON), 3)
	t.Log("after vote ", utxoObj.L.GetMeta().TrunkHeight)
	t.Log("vote tx id ", voteTx.Txid)
	voteTx, err := prp.ledger.QueryTransaction(voteTx.Txid)
	if err != nil {
		t.Error("query transaction error ", err.Error())
	} else {
		t.Log("raw vote desc ", []byte(voteJSON))
		t.Log("query transaction succ ", voteTx.Desc)
	}
	t.Log("proposal tx id ", fmt.Sprintf("%x", proposalTx.Txid))
	referTxid := proposalTx.Txid
	err = prp.saveTrigger(referTxid, triggerDesc)
	if err != nil {
		t.Error("save trigger error ", err.Error())
	}
	t.Log("utxoObj.L.GetMeta().TrunkHeight+1 ", utxoObj.L.GetMeta().TrunkHeight+1)
	prp.context.UtxoBatch.Write()
	triggeredTxList, err := prp.GetVerifiableAutogenTx(utxoObj.L.GetMeta().TrunkHeight+1, -1, 0)
	if err != nil {
		t.Error("GetVerifiableAutogenTx error ", err.Error())
	} else {
		t.Log(triggeredTxList)
	}
	err = prp.removeTrigger(123, referTxid)
	if err != nil {
		t.Error("remove trigger error ", err.Error())
	}
}
func TestGetTxidFromArgs(t *testing.T) {
	prp := NewProposal(nil, nil, nil)
	txDesc := &contract.TxDesc{
		Module: "kernel",
		Method: "vote",
		Args: map[string]interface{}{
			"txid": "8bec1a342f5bafb389193610b5ea7e4a58b02d09429902823ac696d4b6e5c822",
		},
	}
	txID, err := prp.getTxidFromArgs(txDesc)
	if err != nil {
		t.Error("get txid from args error ", err.Error())
	} else {
		t.Log(txID)
	}
}

func TestRunVote(t *testing.T) {
	defer os.RemoveAll(workspace)
	xlog := getLogger()
	utxoObj := utxo.NewFakeUtxoVM(t, workspace, true)
	prp := NewProposal(xlog, utxoObj.L, utxoObj.U)
	utxoObj.U.RegisterVM("proposal", prp, global.VMPrivRing0)
	utxoObj.UtxoWorkWithLedgerBasic(t)
	// 发起一个提案
	_, _, proposalTx, _ := utxoObj.Transfer("alice", "bob", t, "1", utxoObj.L.GetMeta().TipBlockid, []byte(proposeJSON), 1000)
	if proposalTx == nil {
		t.Fatal("transfer failed")
	} else {
		xlog.Info("proposal tx id is ", "txid", fmt.Sprintf("%x", proposalTx.Txid))
	}

	// 生成vote交易并打包上链
	voteJSON := `{
    "module":"proposal",
    "method": "Vote",
    "args" : {
        "txid":"` + fmt.Sprintf("%x", proposalTx.Txid) + `"
    }
    }`
	_, _, voteTx, _ := utxoObj.Transfer("alice", "alice", t, "1", utxoObj.L.GetMeta().TipBlockid, []byte(voteJSON), 1000)
	if voteTx == nil {
		t.Fatal("transfer failed")
	} else {
		xlog.Info("vote tx id is ", "txid", fmt.Sprintf("%x", voteTx.Txid))
	}

	voteDesc := &contract.TxDesc{
		Module: "proposal",
		Method: "vote",
		Args: map[string]interface{}{
			"txid": fmt.Sprintf("%x", proposalTx.Txid),
		},
		// Tx: 应该是投票的tx
		Tx: voteTx,
	}

	utxoObj.U.Play(utxoObj.L.GetMeta().TipBlockid)
	// 投票提案
	prp.runVote(voteDesc)
	prp.context.UtxoBatch.Write()
	voteAmount, voteErr := prp.sumVoteAmount(proposalTx.Txid)
	if voteErr != nil {
		t.Error("sumVoteAmount error ", voteErr.Error())
	}
	if fmt.Sprintf("%s", voteAmount) != "1" {
		t.Error("expect 1 vote, but got ", voteAmount)
	}

	// test for IsVoteOk
	isVoteOk := prp.IsVoteOk(proposalTx)
	if isVoteOk == true {
		t.Error("expect false, but got ", isVoteOk)
	}

	// 回滚投票

	rbVoteDesc := &contract.TxDesc{
		Module: "proposal",
		Method: "rollback_vote",
		Args: map[string]interface{}{
			"txid": fmt.Sprintf("%x", proposalTx.Txid),
		},
		Tx: voteTx,
	}
	err := prp.rollbackVote(rbVoteDesc)
	if err != nil {
		t.Error("roll back vote error ", err.Error())
	}
	prp.context.UtxoBatch.Write()
	voteAmount, voteErr = prp.sumVoteAmount(proposalTx.Txid)
	if voteErr != nil {
		t.Error("sumVoteAmount error ", voteErr.Error())
	}
	if fmt.Sprintf("%s", voteAmount) != "0" {
		t.Error("expect 0 vote, but got ", voteAmount)
	}

	// 回滚proposal
	descTrigger := &contract.TriggerDesc{
		Height: 30,
	}
	desc := &contract.TxDesc{
		Trigger: descTrigger,
		Tx:      proposalTx,
	}
	err = prp.rollbackPropose(desc)
	if err != nil {
		t.Error("rollback propose error ", err.Error())
	}

	fakeDesc := &contract.TxDesc{
		Method: "default",
	}
	err = prp.Run(fakeDesc)
	if err != nil {
		t.Log(err)
	}
	err = prp.Rollback(fakeDesc)
	if err != nil {
		t.Log(err)
	}

	// test for fillOldState
	tmpDesc := &contract.TxDesc{
		Module: "kernel",
		Method: "UpdateMaxBlockSize",
		Args: map[string]interface{}{
			"old_block_size": 123.0,
		},
	}
	tmp, _ := json.Marshal(tmpDesc)
	ret, err2 := prp.fillOldState(tmp)
	if err2 != nil {
		t.Error("fill Old State error ", err2.Error())
	} else {
		t.Log("ret ", ret)
	}
	tmpDesc2 := &contract.TxDesc{
		Module: "consensus",
		Method: "update_consensus",
	}
	tmp, _ = json.Marshal(tmpDesc2)
	ret, err2 = prp.fillOldState(tmp)
	if err2 != nil {
		t.Error("fill Old State error ", err2.Error())
	} else {
		t.Log("ret ", ret)
	}
}

func TestRunThaw(t *testing.T) {
	defer os.RemoveAll(workspace)
	xlog := getLogger()
	utxoObj := utxo.NewFakeUtxoVM(t, workspace, true)
	prp := NewProposal(xlog, utxoObj.L, utxoObj.U)
	utxoObj.U.RegisterVM("proposal", prp, global.VMPrivRing0)
	utxoObj.UtxoWorkWithLedgerBasic(t)
	_, _, proposalTx, _ := utxoObj.Transfer("alice", "bob", t, "1", utxoObj.L.GetMeta().TipBlockid, []byte(proposeJSON), 1000)
	if proposalTx == nil {
		t.Fatal("transfer failed")
	}
	utxoObj.U.Play(utxoObj.L.GetMeta().TipBlockid)
	desc := &contract.TxDesc{
		Args: map[string]interface{}{
			"txid": fmt.Sprintf("%x", proposalTx.Txid),
		},
		Tx: proposalTx,
	}
	block := &pb.InternalBlock{
		Height: 1,
	}
	err := prp.runThaw(desc, block)
	if err != nil {
		t.Error("run thaw error ", err.Error())
	}

	err = prp.rollbackThaw(desc, block)
	if err != nil {
		t.Error("rollback thaw error ", err.Error())
	}

}
*/

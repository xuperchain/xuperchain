package tdpos

import (
	"encoding/json"
	"fmt"
	"github.com/xuperchain/xuperunion/contract"
	"testing"
)

func TestRollBackVote(t *testing.T) {
	desc := &contract.TxDesc{
		Module: "tdpos",
		Method: "rollback_vote",
	}
	strDesc, _ := json.Marshal(desc)

	U, L, tdpos := commonWork(t)
	txCons, block := makeTxWithDesc(strDesc, U, L, t)

	tdpos.context = &contract.TxContext{}
	tdpos.context.UtxoBatch = tdpos.utxoVM.NewBatch()
	tdpos.candidateBallots.LoadOrStore("D_candidate_ballots_f3prTg9itaZY6m48wXXikXdcxiByW7zgk", int64(1))
	desc2 := &contract.TxDesc{
		Module: "tdpos",
		Method: "rollback_vote",
		Tx:     txCons,
		Args: map[string]interface{}{
			"candidates": []interface{}{"f3prTg9itaZY6m48wXXikXdcxiByW7zgk"},
		},
	}
	rollBackVoteErr := tdpos.rollbackVote(desc2, block)
	if rollBackVoteErr != nil {
		t.Error("roll back vote error ", rollBackVoteErr.Error())
	}
	// add cache
	canBal := &candidateBallotsCacheValue{
		ballots: int64(1),
		isDel:   false,
	}
	tdpos.candidateBallotsCache.LoadOrStore("D_candidate_ballots_f3prTg9itaZY6m48wXXikXdcxiByW7zgk", canBal)
	rollBackVoteErr = tdpos.rollbackVote(desc2, block)
	if rollBackVoteErr != nil {
		t.Error("roll back vote error ", rollBackVoteErr.Error())
	}
}

func TestRollBackRevokeVote(t *testing.T) {
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
	key := "D_revoke_0fb3281778163a2c66b1bfcddbb1866b1f10b47efd768fc6065e404042f63009"
	tdpos.context.UtxoBatch.Put([]byte(key), []byte(fmt.Sprintf("%x", txCons.Txid)))
	tdpos.context.UtxoBatch.Write()

	desc2 := &contract.TxDesc{
		Module: "tdpos",
		Method: "rollback_revoke_vote",
		Tx:     txCons,
		Args: map[string]interface{}{
			"txid": fmt.Sprintf("%x", txCons.Txid),
		},
	}
	tdpos.context = &contract.TxContext{}
	tdpos.context.UtxoBatch = tdpos.utxoVM.NewBatch()
	rollBackRevokeVoteErr := tdpos.rollbackRevokeVote(desc2, block)
	if rollBackRevokeVoteErr != nil {
		t.Error("rollbackRevokeVote error ", rollBackRevokeVoteErr.Error())
	}
}

func TestRollBackNominateCandidate(t *testing.T) {
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

	tdpos.candidateBallots.LoadOrStore("D_candidate_ballots_f3prTg9itaZY6m48wXXikXdcxiByW7zgk", int64(1))
	tdpos.context = &contract.TxContext{}
	tdpos.context.UtxoBatch = tdpos.utxoVM.NewBatch()
	key := "D_candidate_nominate_f3prTg9itaZY6m48wXXikXdcxiByW7zgk"
	tdpos.context.UtxoBatch.Put([]byte(key), []byte(txCons.Txid))
	tdpos.context.UtxoBatch.Write()
	desc2 := &contract.TxDesc{
		Module: "tdpos",
		Method: "rollback_nominate_candidate",
		Tx:     txCons,
		Args: map[string]interface{}{
			"candidate": "f3prTg9itaZY6m48wXXikXdcxiByW7zgk",
			"neturl":    "/ip4/127.0.0.1/tcp/47101/p2p/QmVxeNubpg1ZQjQT8W5yZC9fD7ZB1ViArwvyGUB53sqf8e",
		},
	}
	rollBackNomCandErr := tdpos.rollbackNominateCandidate(desc2, block)
	if rollBackNomCandErr != nil {
		t.Error("roll back nominate candidate error ", rollBackNomCandErr.Error())
	}
}

func TestRollBackRevokeCandidate(t *testing.T) {
	desc := &contract.TxDesc{
		Module: "tdpos",
		Method: "nominate_candidate",
		Args: map[string]interface{}{
			"candidate": "f3prTg9itaZY6m48wXXikXdcxiByW7zgk",
		},
	}
	strDesc, _ := json.Marshal(desc)

	U, L, tdpos := commonWork(t)
	txCons, block := makeTxWithDesc(strDesc, U, L, t)

	tdpos.context = &contract.TxContext{}
	tdpos.context.UtxoBatch = tdpos.utxoVM.NewBatch()

	key := "D_revoke_18786b9f4898a3ef375049efe589c3adaa170339c057edab9bdd863860def5b2"
	tdpos.context.UtxoBatch.Put([]byte(key), []byte(txCons.Txid))
	tdpos.context.UtxoBatch.Write()
	desc2 := &contract.TxDesc{
		Module: "tdpos",
		Method: "rollback_revoke_candidate",
		Tx:     txCons,
		Args: map[string]interface{}{
			"candidate": "f3prTg9itaZY6m48wXXikXdcxiByW7zgk",
			"txid":      fmt.Sprintf("%x", txCons.Txid),
		},
	}

	rollBackRevokeCandErr := tdpos.rollbackRevokeCandidate(desc2, block)
	if rollBackRevokeCandErr != nil {
		t.Error("rollbackRevokeCandidate error ", rollBackRevokeCandErr.Error())
	}
}

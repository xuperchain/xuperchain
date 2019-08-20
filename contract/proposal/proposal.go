/*
Copyright Baidu Inc. All Rights Reserved.
*/

package proposal

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"

	log "github.com/xuperchain/log15"

	"github.com/xuperchain/xuperunion/contract"
	"github.com/xuperchain/xuperunion/ledger"
	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/utxo"
)

const proposeMethod = "Propose"
const voteMethod = "Vote"
const createTriggerMethod = "CreateTrigger"
const thawMethod = "Thaw"
const systemContractMinPercent = 50 //系统级合约要通过至少需要多少票比例

// Proposal implements the proposal function on xchain
type Proposal struct {
	log        log.Logger
	ledger     *ledger.Ledger
	utxoVM     *utxo.UtxoVM
	context    *contract.TxContext
	rightsPair map[string]map[string]string
}

// NewProposal instances a new Proposal
func NewProposal(log log.Logger, lg *ledger.Ledger, uvm *utxo.UtxoVM) *Proposal {
	prp := &Proposal{
		log:    log,
		ledger: lg,
		utxoVM: uvm,
	}
	prp.rightsPair = make(map[string]map[string]string)
	prp.rightsPair["tdpos"] = map[string]string{
		"vote":               "revoke_vote",
		"nominate_candidate": "revoke_candidate",
	}
	return prp
}

// Run implements ContractInterface
func (prp *Proposal) Run(desc *contract.TxDesc) error {
	switch desc.Method {
	case proposeMethod:
		return prp.runPropose(desc)
	case voteMethod:
		return prp.runVote(desc)
	case createTriggerMethod:
		return prp.saveTrigger(desc.Tx.Txid, desc.Trigger)
	case thawMethod:
		return prp.runThaw(desc, prp.context.Block)
	default:
		prp.log.Warn("method not implemented", "method", desc.Method)
		return fmt.Errorf("%s not implemented", desc.Method)
	}
}

// ReadOutput implements ContractInterface
func (prp *Proposal) ReadOutput(desc *contract.TxDesc) (contract.ContractOutputInterface, error) {
	return nil, nil
}

// Rollback implements ContractInterface
func (prp *Proposal) Rollback(desc *contract.TxDesc) error {
	switch desc.Method {
	case proposeMethod:
		return prp.rollbackPropose(desc)
	case voteMethod:
		return prp.rollbackVote(desc)
	case createTriggerMethod:
		return prp.removeTrigger(desc.Trigger.Height, desc.Tx.Txid)
	case thawMethod:
		return prp.rollbackThaw(desc, prp.context.Block)
	default:
		prp.log.Warn("method not implemented", "method", desc.Method)
		return fmt.Errorf("%s not implemented", desc.Method)
	}
}

// Finalize implements ContractInterface
func (prp *Proposal) Finalize(blockid []byte) error {
	return nil
}

// SetContext implements ContractInterface
func (prp *Proposal) SetContext(context *contract.TxContext) error {
	prp.context = context
	return nil
}

func (prp *Proposal) runPropose(desc *contract.TxDesc) error {
	if desc.Trigger != nil {
		err := prp.saveTrigger(desc.Tx.Txid, desc.Trigger)
		if err != nil {
			return err
		}
	}
	return nil
}

func (prp *Proposal) rollbackPropose(desc *contract.TxDesc) error {
	if desc.Trigger != nil {
		err := prp.removeTrigger(desc.Trigger.Height, desc.Tx.Txid)
		if err != nil {
			return err
		}
	}
	return nil
}

func (prp *Proposal) makeTriggerKey(height int64, txid []byte) string {
	return fmt.Sprintf("%s%020d_%x", pb.TriggerPrefix, height, txid)
}

func (prp *Proposal) parseTriggerKey(ldbKey []byte) (int64, []byte, error) {
	var height int64
	var txid = []byte{}
	var prefix string
	n, err := fmt.Sscanf(string(ldbKey), "%1s%020d_%x", &prefix, &height, &txid)
	if n != 3 || err != nil {
		return 0, nil, fmt.Errorf("failed to parse ldb key %s, %v", string(ldbKey), err)
	}
	return height, txid, nil
}

func (prp *Proposal) makeVoteKey(proposalTxid []byte, voteTxid []byte) string {
	return fmt.Sprintf("%s%x_%x", pb.VoteProposalPrefix, proposalTxid, voteTxid)
}

func (prp *Proposal) saveTrigger(txid []byte, trigger *contract.TriggerDesc) error {
	if txid == nil || trigger == nil {
		return errors.New("save trigger failed, invalid trigger definition")
	}
	trigger.RefTxid = txid
	key := prp.makeTriggerKey(trigger.Height, txid)
	prp.log.Info("save trigger", "key", key)
	batch := prp.context.UtxoBatch
	buf, err := json.Marshal(trigger)
	if err != nil {
		prp.log.Warn("marshal trigger failed", "buf_err", err)
		return err
	}
	return batch.Put([]byte(key), buf)
}

func (prp *Proposal) removeTrigger(height int64, txid []byte) error {
	if txid == nil {
		return errors.New("remove trigger failed, invalid trigger definition")
	}
	key := prp.makeTriggerKey(height, txid)
	prp.log.Info("undo tx, so remove trigger", "key", key)
	batch := prp.context.UtxoBatch
	return batch.Delete([]byte(key))
}

func (prp Proposal) getTxidFromArgs(desc *contract.TxDesc) ([]byte, error) {
	if desc.Args["txid"] == nil {
		return nil, errors.New("txid can not be found in args")
	}
	argTxid, ok := desc.Args["txid"].(string)
	if !ok {
		return nil, fmt.Errorf("getTxidFromArgs failed, txid should be string. but got %v", desc.Args["txid"])
	}
	proposalTxid, err := hex.DecodeString(argTxid)
	if err != nil {
		return nil, err
	}
	return proposalTxid, nil
}

func (prp *Proposal) getDescArg(proposalTx *pb.Transaction, argName string) (float64, error) {
	proposalDesc, err := contract.Parse(string(proposalTx.Desc))
	if err != nil {
		return 0, err
	}
	if proposalDesc.Args[argName] == nil {
		return 0, fmt.Errorf("%s not found in contract args", argName)
	}
	argValue := proposalDesc.Args[argName].(float64)
	return argValue, nil
}

// IsPropose return true if tx has Propose method
func (prp *Proposal) IsPropose(proposalTx *pb.Transaction) bool {
	proposalDesc, err := contract.Parse(string(proposalTx.Desc))
	if err != nil {
		return false
	}
	return proposalDesc.Method == proposeMethod
}

func (prp *Proposal) runVote(desc *contract.TxDesc) error {
	prp.log.Debug("start run vote")
	proposalTxid, err := prp.getTxidFromArgs(desc)
	if err != nil {
		return err
	}
	proposalTx, err := prp.ledger.QueryTransaction(proposalTxid)
	if err != nil {
		proposalTx = prp.context.Block.GetTx(proposalTxid)
		if proposalTx == nil {
			prp.log.Warn("vote fail, because proposal tx not found", "proposalTxid", fmt.Sprintf("%x", proposalTxid))
			return err
		}
	}
	argValue, err := prp.getDescArg(proposalTx, "stop_vote_height")
	if err != nil {
		return err
	}
	stopVoteHeight := int64(argValue)
	ledgerHeight := prp.ledger.GetMeta().TrunkHeight
	if ledgerHeight > stopVoteHeight {
		prp.log.Warn(fmt.Sprintf("this propposal is expired for voting, %d > %d", ledgerHeight, stopVoteHeight))
		return nil
	}
	key := prp.makeVoteKey(proposalTxid, desc.Tx.Txid)
	amount := desc.Tx.GetFrozenAmount(stopVoteHeight)
	if amount.Cmp(big.NewInt(0)) <= 0 {
		return errors.New("no frozen money for vote, so the system ignore the vote")
	}
	batch := prp.context.UtxoBatch
	return batch.Put([]byte(key), amount.Bytes())
}

func (prp *Proposal) rollbackVote(desc *contract.TxDesc) error {
	proposalTxid, err := prp.getTxidFromArgs(desc)
	if err != nil {
		return err
	}
	key := prp.makeVoteKey(proposalTxid, desc.Tx.Txid)
	batch := prp.context.UtxoBatch
	return batch.Delete([]byte(key))
}

//统计票数
func (prp *Proposal) sumVoteAmount(proposalTxid []byte) (*big.Int, error) {
	sumPrefix := fmt.Sprintf("%s%x_", pb.VoteProposalPrefix, proposalTxid)
	it := prp.utxoVM.ScanWithPrefix([]byte(sumPrefix))
	total := big.NewInt(0)
	defer it.Release()
	for it.Next() {
		amount := big.NewInt(0)
		amount.SetBytes(it.Value())
		total = total.Add(total, amount)
	}
	if it.Error() != nil {
		return nil, it.Error()
	}
	return total, nil
}

// IsVoteOk check vote result of tx
func (prp *Proposal) IsVoteOk(proposalTx *pb.Transaction) bool {
	minVotePercent, err := prp.getDescArg(proposalTx, "min_vote_percent")
	prp.log.Debug("check contract arg", "mint_vote_percent", minVotePercent, "err", err)
	if err != nil {
		return false
	}
	if minVotePercent > 100 || minVotePercent < systemContractMinPercent {
		prp.log.Debug("invalid minVotePercent", "percent", minVotePercent)
		return false
	}
	curVoteAmount, err := prp.sumVoteAmount(proposalTx.Txid)
	if err != nil {
		prp.log.Warn("calc vote sum failed", "err", err)
		return false
	}
	voteNeeded := prp.utxoVM.GetTotal()
	voteNeeded.Mul(voteNeeded, big.NewInt(int64(minVotePercent)))
	voteNeeded.Div(voteNeeded, big.NewInt(100))
	prp.log.Debug("vote result", "need", voteNeeded, "got", curVoteAmount)
	voteOK := curVoteAmount.Cmp(voteNeeded) >= 0
	if !voteOK {
		prp.log.Warn("no enough vote collected", "txid", proposalTx.HexTxid())
	}
	return voteOK
}

//填充一些必要的当前状态
//例如，更新block的大小，那么当合约回滚的时候需要知道之前是多大
func (prp *Proposal) fillOldState(desc []byte) ([]byte, error) {
	descObj, err := contract.Parse(string(desc))
	if err != nil {
		prp.log.Warn("contract desc cannot be parsed", "desc", desc)
		return nil, err
	}
	contractKey := descObj.Module + "." + descObj.Method
	switch contractKey {
	case "kernel.UpdateMaxBlockSize":
		prp.log.Trace("contract desc need to process", "contractKey", "kernel.UpdateMaxBlockSize")
		descObj.Args["old_block_size"] = prp.ledger.GetMaxBlockSize()
	case "kernel.UpdateReservedContract":
		prp.log.Trace("contract desc need to process", "contractKey", "kernel.UpdateReservedContract")
		reservedContracts := []ledger.InvokeRequest{}
		for _, rc := range prp.ledger.GetMeta().GetReservedContracts() {
			args := map[string]string{}
			for k, v := range rc.GetArgs() {
				args[k] = string(v)
			}
			param := ledger.InvokeRequest{
				ModuleName:   rc.GetModuleName(),
				ContractName: rc.GetContractName(),
				MethodName:   rc.GetMethodName(),
				Args:         args,
			}
			reservedContracts = append(reservedContracts, param)
		}
		descObj.Args["old_reserved_contracts"] = reservedContracts
	case "kernel.UpdateForbiddenContract":
		prp.log.Trace("contract desc need to process", "contractKey", "kernel.UpdateForbiddenContract")
		forbiddenContract := prp.ledger.GetMeta().GetForbiddenContract()
		forbiddenContractMap := map[string]interface{}{}
		forbiddenContractMap["module_name"] = forbiddenContract.GetModuleName()
		forbiddenContractMap["contract_name"] = forbiddenContract.GetContractName()
		forbiddenContractMap["method_name"] = forbiddenContract.GetMethodName()
		forbiddenContractMap["args"] = forbiddenContract.GetArgs()

		descObj.Args["old_forbidden_contract"] = forbiddenContractMap
	default:
		prp.log.Trace("contract desc do not need to process")
	}
	enhancedDesc, jsErr := json.Marshal(descObj)
	if jsErr != nil {
		prp.log.Warn("failed to marshal new descObj", "err", jsErr)
		return nil, jsErr
	}
	return enhancedDesc, nil
}

// GetVerifiableAutogenTx 实现VAT接口
func (prp *Proposal) GetVerifiableAutogenTx(blockHeight int64, maxCount int, timestamp int64) ([]*pb.Transaction, error) {
	triggeredTxList := []*pb.Transaction{}
	//从trigger表获取应到触发形成的新tx
	heightPrefix := fmt.Sprintf("%s%020d_", pb.TriggerPrefix, blockHeight)
	it := prp.utxoVM.ScanWithPrefix([]byte(heightPrefix))
	defer it.Release()
	for it.Next() {
		//it.Key() 是高度_提案txid
		height, proposalTxid, err := prp.parseTriggerKey(it.Key())
		prp.log.Debug("check proposal", "tx", fmt.Sprintf("%x", proposalTxid), "height", height)
		if err != nil {
			prp.log.Warn("invalid trigger", "err", err)
			continue
		}
		proposalTx, err := prp.ledger.QueryTransaction(proposalTxid)
		if err != nil {
			prp.log.Warn("check trigger failed, because proposal tx not found", "proposalTxid", fmt.Sprintf("%x", proposalTxid))
			continue
		}
		if prp.IsPropose(proposalTx) { //如果是个标准提案，需要看看票数OK了么
			if !prp.IsVoteOk(proposalTx) {
				continue
			}
		}
		desc := it.Value() //value 是即将触发的合约
		enhancedDesc, dErr := prp.fillOldState(desc)
		if dErr != nil {
			return nil, dErr
		}
		tx, err := prp.utxoVM.GenerateEmptyTx(enhancedDesc)
		if err != nil {
			prp.log.Warn("failed to generate triggered tx", "err", err)
		}
		prp.log.Debug("tirgger new tx", "txid", tx.HexTxid(), "desc", string(desc), "tx", tx, "tx.desc", string(tx.Desc))
		triggeredTxList = append(triggeredTxList, tx)
	}
	if it.Error() != nil {
		return nil, it.Error()
	}
	return triggeredTxList, nil
}

// GetVATWhiteList 实现VAT接口
func (prp *Proposal) GetVATWhiteList() map[string]bool {
	return nil
}

func (prp *Proposal) runThaw(desc *contract.TxDesc, block *pb.InternalBlock) error {
	prp.log.Trace("start to runThaw", "desc", desc, "txid", fmt.Sprintf("%x", desc.Tx.Txid))
	// 获取高度
	height := int64(0)
	if block.Height == 0 {
		height = prp.ledger.GetMeta().TrunkHeight + 1
	} else {
		height = block.Height
	}

	txidThaw, err := prp.getTxidFromArgs(desc)
	if err != nil {
		prp.log.Warn("runThaw getTxidFromArgs error")
		return nil
	}
	prp.log.Trace("runThaw", "txidThaw", fmt.Sprintf("%x", txidThaw))

	tx, err := prp.ledger.QueryTransaction(txidThaw)
	if err != nil {
		prp.log.Warn("runThaw failed, because thaw tx not found", "thaw_txid", fmt.Sprintf("%x", txidThaw))
		return nil
	}

	if len(tx.TxInputs) == 0 {
		prp.log.Warn("rollbackThaw failed, tx.TxInputs can not be null")
		return nil
	}

	fromAddress := tx.TxInputs[0].FromAddr
	fromAddressThaw := desc.Tx.TxInputs[0].FromAddr
	if string(fromAddress) != string(fromAddressThaw) {
		prp.log.Warn("runThaw failed, fromAddress and fromAddressThaw not equal", "fromAddress", string(fromAddress),
			"fromAddressThaw", string(fromAddressThaw))
		return nil
	}

	// 解冻utxo
	for offset, txOutput := range tx.TxOutputs {
		if txOutput.FrozenHeight == -1 {
			utxoKey := utxo.GenUtxoKeyWithPrefix(txOutput.ToAddr, tx.Txid, int32(offset))
			thawUtxoKey := fmt.Sprintf("%s_thaw_%x_%s", pb.VoteProposalPrefix, desc.Tx.Txid, utxoKey)

			val, err := prp.utxoVM.GetFromTable(nil, []byte(utxoKey))
			if err != nil {
				prp.log.Warn("runThaw failed, because thaw tx not found", "thaw_txid", fmt.Sprintf("%x", txidThaw))
				return err
			}
			uItem := &utxo.UtxoItem{}
			err = uItem.Loads(val)
			if err != nil {
				return err
			}
			uItem.FrozenHeight = 0
			uItemBinary, err := uItem.Dumps()
			if err != nil {
				return err
			}
			prp.log.Trace("runThaw thaw utxoKey", "utxoKey", utxoKey)
			// 记录修改了哪些utxo
			prp.context.UtxoBatch.Put([]byte(thawUtxoKey), []byte{})
			prp.context.UtxoBatch.Put([]byte(utxoKey), uItemBinary)
			// 清理utxo_cache缓存
			prp.utxoVM.RemoveUtxoCache(string(txOutput.ToAddr), string(utxoKey))
		}
	}

	descThaw, _ := contract.Parse(string(tx.Desc))
	prp.log.Trace("runThaw Parse", "descThaw", descThaw)
	if descThaw != nil {
		if prp.rightsPair[descThaw.Module] != nil && prp.rightsPair[descThaw.Module][descThaw.Method] != "" {
			prp.log.Trace("runThaw Parse", "method", prp.rightsPair[descThaw.Module][descThaw.Method])
			descTrigger := contract.TxDesc{
				Module: descThaw.Module,
				Method: prp.rightsPair[descThaw.Module][descThaw.Method],
				Args:   make(map[string]interface{}),
			}
			descTrigger.Args["txid"] = fmt.Sprintf("%x", txidThaw)
			trigger := &contract.TriggerDesc{
				TxDesc:  descTrigger,
				RefTxid: desc.Tx.Txid,
				Height:  height + 3,
			}
			key := prp.makeTriggerKey(trigger.Height, desc.Tx.Txid)
			prp.log.Info("save trigger", "key", key, "trigger", trigger)
			buf, err := json.Marshal(trigger)
			if err != nil {
				prp.log.Warn("marshal trigger failed", "buf_err", err)
				return err
			}
			return prp.context.UtxoBatch.Put([]byte(key), buf)
		}
	}
	return nil
}

func (prp *Proposal) rollbackThaw(desc *contract.TxDesc, block *pb.InternalBlock) error {
	prp.log.Trace("start to rollbackThaw", "desc", desc, "txid", fmt.Sprintf("%x", desc.Tx.Txid))
	txid, err := prp.getTxidFromArgs(desc)
	if err != nil {
		return nil
	}

	tx, err := prp.ledger.QueryTransaction(txid)
	if err != nil {
		prp.log.Warn("rollbackThaw failed, because thaw tx not found", "thaw_txid", fmt.Sprintf("%x", txid))
		return nil
	}

	if len(tx.TxInputs) == 0 {
		prp.log.Warn("rollbackThaw failed, tx.TxInputs can not be null")
		return nil
	}

	fromAddress := tx.TxInputs[0].FromAddr
	fromAddressThaw := desc.Tx.TxInputs[0].FromAddr
	if string(fromAddress) != string(fromAddressThaw) {
		prp.log.Warn("runThaw failed, fromAddress and fromAddressThaw not equal", "fromAddress", string(fromAddress),
			"fromAddressThaw", string(fromAddressThaw))
		return nil
	}

	// 冻结utxo
	thawUtxoKeyPrefix := fmt.Sprintf("%s_thaw_%x_", pb.VoteProposalPrefix, desc.Tx.Txid)
	it := prp.utxoVM.ScanWithPrefix([]byte(thawUtxoKeyPrefix))
	defer it.Release()
	for it.Next() {
		key := string(it.Key())
		utxoKey := strings.TrimPrefix(key, thawUtxoKeyPrefix)
		addr := strings.Split(utxoKey, "_")[0]
		val, err := prp.utxoVM.GetFromTable(nil, []byte(utxoKey))
		if err != nil {
			prp.log.Warn("rollbackThaw failed, because thaw tx not found", "thaw_txid", fmt.Sprintf("%x", txid))
			return errors.New("run thaw utxo error, utxo not found")
		}
		uItem := &utxo.UtxoItem{}
		err = uItem.Loads(val)
		if err != nil {
			return err
		}
		uItem.FrozenHeight = -1
		uItemBinary, err := uItem.Dumps()
		if err != nil {
			return err
		}
		prp.context.UtxoBatch.Delete([]byte(key))
		prp.context.UtxoBatch.Put([]byte(utxoKey), uItemBinary)
		prp.utxoVM.RemoveUtxoCache(addr, string(utxoKey))
	}
	return nil
}

// Stop implements ContractInterface
func (prp *Proposal) Stop() {
}

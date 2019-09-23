package utxo

import (
	"bytes"
	"html/template"

	"github.com/xuperchain/xuperunion/pb"
)

// reservedArgs used to get contractnames from InvokeRPCRequest
type reservedArgs struct {
	ContractNames string
}

func genArgs(req []*pb.InvokeRequest) *reservedArgs {
	ra := &reservedArgs{}
	for i, v := range req {
		ra.ContractNames += v.GetContractName()
		if i < len(req)-1 {
			ra.ContractNames += ","
		}
	}
	return ra
}

// It will check whether the transaction in reserved whitelist
// if the config of chain contains reserved contracts
// but the transaction does not contains reserved requests.
func (uv *UtxoVM) verifyReservedWhitelist(tx *pb.Transaction) bool {
	// verify reservedContracts len
	reservedContracts := uv.ledger.GetMeta().ReservedContracts
	if len(reservedContracts) == 0 {
		uv.xlog.Info("verifyReservedWhitelist false reservedReqs is nil")
		return false
	}

	// get white list account
	accountName := uv.ledger.GetGenesisBlock().GetConfig().GetReservedWhitelistAccount()
	uv.xlog.Trace("verifyReservedWhitelist", "accountName", accountName)
	if accountName == "" {
		uv.xlog.Info("verifyReservedWhitelist false, the chain does not have reserved whitelist", "accountName", accountName)
		return false
	}
	acl, isConfirmed, err := uv.aclMgr.GetAccountACLWithConfirmed(accountName)
	if err != nil || acl == nil || !isConfirmed {
		uv.xlog.Info("verifyReservedWhitelist false, get reserved whitelist acl failed",
			"err", err, "acl", acl, "isConfirmed", isConfirmed)
		return false
	}

	// verify storage
	if tx.GetDesc() != nil ||
		tx.GetContractRequests() != nil ||
		tx.GetTxInputsExt() != nil ||
		tx.GetTxOutputsExt() != nil {
		uv.xlog.Info("verifyReservedWhitelist false the storage info should be nil")
		return false
	}

	// verify utxo input
	if len(tx.GetTxInputs()) == 0 && len(tx.GetTxOutputs()) == 0 {
		uv.xlog.Info("verifyReservedWhitelist true the utxo list is nil")
		return true
	}
	fromAddr := string(tx.GetTxInputs()[0].GetFromAddr())
	for _, v := range tx.GetTxInputs() {
		if string(v.GetFromAddr()) != fromAddr {
			uv.xlog.Info("verifyReservedWhitelist false fromAddr should no more than one")
			return false
		}
	}

	// verify utxo output
	toAddrs := make(map[string]bool)
	for _, v := range tx.GetTxOutputs() {
		if bytes.Equal(v.GetToAddr(), []byte(FeePlaceholder)) {
			continue
		}
		toAddrs[string(v.GetToAddr())] = true
		if len(toAddrs) > 2 {
			uv.xlog.Info("verifyReservedWhitelist false toAddrs should no more than two")
			return false
		}
	}

	// verify utxo output whitelist
	for k := range toAddrs {
		if k == fromAddr {
			continue
		}
		if _, ok := acl.GetAksWeight()[k]; !ok {
			uv.xlog.Info("verifyReservedWhitelist false the toAddr should in whitelist acl")
			return false
		}
	}
	return true
}

func (uv *UtxoVM) verifyReservedContractRequests(reservedReqs, txReqs []*pb.InvokeRequest) bool {
	if len(reservedReqs) > len(txReqs) {
		return false
	}
	for i := 0; i < len(reservedReqs); i++ {
		if (reservedReqs[i].GetModuleName() != txReqs[i].GetModuleName()) || (reservedReqs[i].GetContractName() != txReqs[i].GetContractName()) ||
			(reservedReqs[i].GetMethodName() != txReqs[i].GetMethodName()) {
			return false
		}
		for k, v := range txReqs[i].Args {
			if !bytes.Equal(reservedReqs[i].GetArgs()[k], v) {
				return false
			}
		}
	}
	return true
}

// geReservedContractRequest get reserved contract requests from system params, it doesn't consume gas.
func (uv *UtxoVM) getReservedContractRequests(req []*pb.InvokeRequest, isPreExec bool) ([]*pb.InvokeRequest, error) {
	MetaReservedContracts := uv.ledger.GetMeta().ReservedContracts
	if MetaReservedContracts == nil {
		return nil, nil
	}
	reservedContractstpl := MetaReservedContracts
	uv.xlog.Info("MetaReservedContracts", "reservedContracts", reservedContractstpl)

	// if all reservedContracts have not been updated, return nil, nil
	ra := &reservedArgs{}
	if isPreExec || len(reservedContractstpl) == 0 {
		ra = genArgs(req)
	} else {
		// req should contrain reservedContracts, so the len of req should no less than reservedContracts
		if len(req) < len(reservedContractstpl) {
			uv.xlog.Warn("req should contain reservedContracts")
			return nil, ErrGetReservedContracts
		} else if len(req) > len(reservedContractstpl) {
			ra = genArgs(req[len(reservedContractstpl):])
		}
	}

	reservedContracts := []*pb.InvokeRequest{}
	for _, rc := range reservedContractstpl {
		rctmp := *rc
		rctmp.Args = make(map[string][]byte)
		for k, v := range rc.GetArgs() {
			buf := new(bytes.Buffer)
			tpl := template.Must(template.New("value").Parse(string(v)))
			tpl.Execute(buf, ra)
			rctmp.Args[k] = buf.Bytes()
		}
		reservedContracts = append(reservedContracts, &rctmp)
	}
	return reservedContracts, nil
}

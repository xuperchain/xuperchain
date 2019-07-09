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

func verifyReservedContractRequests(reservedReqs, txReqs []*pb.InvokeRequest) bool {
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
	reservedContracts, err := uv.ledger.GenesisBlock.GetConfig().GetReservedContract()
	if err != nil {
		return nil, err
	}
	// if all reservedContracts have not been updated, return nil, nil
	ra := &reservedArgs{}
	if isPreExec || len(reservedContracts) == 0 {
		ra = genArgs(req)
	} else {
		// req should contrain reservedContracts, so the len of req should no less than reservedContracts
		if len(req) < len(reservedContracts) {
			uv.xlog.Warn("req should contain reservedContracts")
			return nil, ErrGetReservedContracts
		} else if len(req) > len(reservedContracts) {
			ra = genArgs(req[len(reservedContracts):])
		}
	}

	for _, rc := range reservedContracts {
		for k, v := range rc.GetArgs() {
			buf := new(bytes.Buffer)
			tpl := template.Must(template.New("value").Parse(string(v)))
			tpl.Execute(buf, ra)
			rc.Args[k] = buf.Bytes()
		}
	}
	return reservedContracts, nil
}

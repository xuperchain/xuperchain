package utxo

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/xuperchain/xuperunion/pb"
)

func verifyReservedContractRequests(reservedReqs, txReqs []*pb.InvokeRequest) bool {
	if len(reservedReqs) > len(txReqs) {
		return false
	}
	for i := 0; i < len(reservedReqs); i++ {
		if (reservedReqs[i].ModuleName != txReqs[i].ModuleName) || (reservedReqs[i].ContractName != txReqs[i].ContractName) ||
			(reservedReqs[i].MethodName != txReqs[i].MethodName) {
			return false
		}
		for k, v := range txReqs[i].Args {
			if !bytes.Equal(reservedReqs[i].Args[k], v) {
				return false
			}
		}
	}
	return true
}

// geReservedContractRequest get reserved contract requests from system params, it doesn't consume gas.
func (uv *UtxoVM) getReservedContractRequests(req *pb.InvokeRPCRequest, tx *pb.Transaction) ([]*pb.InvokeRequest, error) {
	reservedContractCfgs, err := uv.ledger.GenesisBlock.GetConfig().GetReservedContract()
	if err != nil {
		return nil, err
	}

	reservedContracts := []*pb.InvokeRequest{}
	// FIXME zq: need to suport contract args
	for _, v := range reservedContractCfgs {
		req, err := parseReservedContractCfg(v)
		if err != nil {
			return nil, err
		}
		reservedContracts = append(reservedContracts, req)
	}
	uv.xlog.Trace("geReservedContractRequest results", "results", reservedContracts)
	return reservedContracts, nil
}

func parseReservedContractCfg(contract string) (*pb.InvokeRequest, error) {
	subContract := strings.Split(contract, ".")
	if len(subContract) != 2 {
		return nil, fmt.Errorf("parseReservedContractCfg error!")
	}

	req := &pb.InvokeRequest{
		ModuleName:   "wasm",
		ContractName: subContract[0],
		MethodName:   subContract[1],
		Args:         map[string][]byte{},
	}
	return req, nil
}

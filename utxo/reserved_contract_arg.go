package utxo

import (
	"github.com/xuperchain/xuperunion/pb"
)

func setReservedContractArg(reservedReq *pb.InvokeRequest, customReq *pb.InvokeRPCRequest) (*pb.InvokeRequest, error) {
	// if reservedReq or customReq is nil, return directly
	if reservedReq == nil || customReq == nil {
		return nil, nil
	}
	reservedContractName := reservedReq.GetContractName()
	switch reservedContractName {
	case "banned":
		return setBannedContractArg(reservedReq, customReq)
	default:
		return reservedReq, nil
	}
}

func setBannedContractArg(reservedReq *pb.InvokeRequest, customReq *pb.InvokeRPCRequest) (*pb.InvokeRequest, error) {
	customRequests := customReq.GetRequests()
	contractNames := ""
	for _, v := range customRequests {
		contractName := v.GetContractName()
		contractNames += "," + contractName
	}
	reservedReq.Args["contracts"] = []byte(contractNames)
	return reservedReq, nil
}

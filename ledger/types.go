package ledger

import (
	"github.com/xuperchain/xuperunion/pb"
)

type InvokeRequest struct {
	ModuleName   string            `json:"module_name"`
	ContractName string            `json:"contract_name"`
	MethodName   string            `json:"method_name"`
	Args         map[string]string `json:"args"`
}

func invokeRequestFromJson2Pb(jsonRequest []InvokeRequest) ([]*pb.InvokeRequest, error) {
	requestsWithPb := []*pb.InvokeRequest{}
	for _, request := range jsonRequest {
		tmpReqWithPB := &pb.InvokeRequest{
			ModuleName:   request.ModuleName,
			ContractName: request.ContractName,
			MethodName:   request.MethodName,
			Args:         make(map[string][]byte),
		}
		for k, v := range request.Args {
			tmpReqWithPB.Args[k] = []byte(v)
		}
		requestsWithPb = append(requestsWithPb, tmpReqWithPB)
	}
	return requestsWithPb, nil
}

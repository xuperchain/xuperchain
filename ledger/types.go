package ledger

import (
	"github.com/xuperchain/xuperunion/pb"
)

type InvokeRequest struct {
	ModuleName    string            `json:"module_name"`
	ContractName  string            `json:"contract_name"`
	MethodName    string            `json:"method_name"`
	Args          map[string]string `json:"args"`
	ResouceLimits []ResourceLimit   `json:"resource_limits"`
}

type ResourceLimit struct {
	Type  string `json:"type"`
	Limit int64  `json:"limit"`
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
		for _, v := range request.ResouceLimits {
			tmp := &pb.ResourceLimit{
				Type:  pb.ResourceType(pb.ResourceType_value[v.Type]),
				Limit: v.Limit,
			}
			tmpReqWithPB.ResourceLimits = append(tmpReqWithPB.ResourceLimits, tmp)
		}
		for k, v := range request.Args {
			tmpReqWithPB.Args[k] = []byte(v)
		}
		requestsWithPb = append(requestsWithPb, tmpReqWithPB)
	}
	return requestsWithPb, nil
}

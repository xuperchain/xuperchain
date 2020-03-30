package xmodel

import (
	"testing"

	"github.com/xuperchain/xuperchain/core/pb"
)

func TestIsCossQueryValid(t *testing.T) {
	testCases := map[string]struct {
		request   *pb.CrossQueryRequest
		queryMeta *pb.CrossQueryMeta
		queryInfo *pb.CrossQueryInfo
		result    bool
	}{
		"test Request not equal": {
			request: &pb.CrossQueryRequest{
				Bcname:    "xuper",
				Initiator: "12345",
				AuthRequire: []string{
					"bob",
					"alice",
				},
				Request: &pb.InvokeRequest{
					ModuleName:   "wasm",
					ContractName: "contract1",
					MethodName:   "method1",
					Args: map[string][]byte{
						"agr1": []byte("arg1"),
						"agr2": []byte("arg2"),
					},
				},
			},
			queryMeta: &pb.CrossQueryMeta{},
			queryInfo: &pb.CrossQueryInfo{
				Request: &pb.CrossQueryRequest{
					Bcname:    "xuper",
					Initiator: "12345",
					AuthRequire: []string{
						"bob",
						"alice",
					},
					Request: &pb.InvokeRequest{
						ModuleName:   "wasm",
						ContractName: "contract1",
						MethodName:   "method1",
						Args: map[string][]byte{
							"agr1": []byte("arg2"),
							"agr2": []byte("arg1"),
						},
					},
				},
			},
			result: false,
		},
	}

	for k, v := range testCases {
		res := isCossQueryValid(v.request, v.queryMeta, v.queryInfo)
		if res != v.result {
			t.Error("test failed", "case", k, "expect=", v.result, "actual=", res)
		}
	}
}

func TestIsEndorsorInfoValid(t *testing.T) {
	testCases := map[string]struct {
		queryMeta *pb.CrossQueryMeta
		signs     []*pb.SignatureInfo
		result    bool
	}{
		"test IsEndorsorInfoValid succeed": {
			queryMeta: &pb.CrossQueryMeta{
				ChainMeta: &pb.CrossChainMeta{
					Type:           "xuper",
					MinEndorsorNum: 2,
				},
				Endorsors: []*pb.CrossEndorsor{
					&pb.CrossEndorsor{
						Address: "12345",
						PubKey:  "12345",
					},
					&pb.CrossEndorsor{
						Address: "23456",
						PubKey:  "23456",
					},
					&pb.CrossEndorsor{
						Address: "34567",
						PubKey:  "34567",
					},
				},
			},
			signs: []*pb.SignatureInfo{
				&pb.SignatureInfo{
					PublicKey: "23456",
				},
				&pb.SignatureInfo{
					PublicKey: "34567",
				},
				&pb.SignatureInfo{
					PublicKey: "34568",
				},
			},
			result: true,
		},
		"test IsEndorsorInfoValid failed": {
			queryMeta: &pb.CrossQueryMeta{
				ChainMeta: &pb.CrossChainMeta{
					Type:           "xuper",
					MinEndorsorNum: 2,
				},
				Endorsors: []*pb.CrossEndorsor{
					&pb.CrossEndorsor{
						Address: "12345",
						PubKey:  "12345",
					},
					&pb.CrossEndorsor{
						Address: "23456",
						PubKey:  "23456",
					},
					&pb.CrossEndorsor{
						Address: "34567",
						PubKey:  "34567",
					},
				},
			},
			signs: []*pb.SignatureInfo{
				&pb.SignatureInfo{
					PublicKey: "daerrf",
				},
				&pb.SignatureInfo{
					PublicKey: "34567",
				},
				&pb.SignatureInfo{
					PublicKey: "34568",
				},
			},
			result: false,
		},
	}
	for k, v := range testCases {
		_, res := isEndorsorInfoValid(v.queryMeta, v.signs)
		if res != v.result {
			t.Error("test failed", "case", k, "expect=", v.result, "actual=", res)
		}
	}
}

func TestIsCrossQueryResponseEqual(t *testing.T) {
	testCases := map[string]struct {
		a      *pb.CrossQueryResponse
		b      *pb.CrossQueryResponse
		result bool
	}{
		"test equal": {
			a: &pb.CrossQueryResponse{
				Response: &pb.ContractResponse{
					Status:  1,
					Message: "ok",
					Body:    []byte("ok"),
				},
			},
			b: &pb.CrossQueryResponse{
				Response: &pb.ContractResponse{
					Status:  1,
					Message: "ok",
					Body:    []byte("ok"),
				},
			},
			result: true,
		},
		"test body not equal": {
			a: &pb.CrossQueryResponse{
				Response: &pb.ContractResponse{
					Status:  1,
					Message: "ok",
					Body:    []byte("ok"),
				},
			},
			b: &pb.CrossQueryResponse{
				Response: &pb.ContractResponse{
					Status:  1,
					Message: "ok",
					Body:    []byte("false"),
				},
			},
			result: false,
		},
	}
	for k, v := range testCases {
		res := isCrossQueryResponseEqual(v.a, v.b)
		if res != v.result {
			t.Error("test failed", "case", k, "expect=", v.result, "actual=", res)
		}
	}
}

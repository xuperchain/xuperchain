package main

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/xuperchain/xuperchain/core/pb"
)

var baseDir = os.Getenv("XCHAIN_ROOT")

type MockXEndorserServer struct {
	Txs map[string]*pb.TxStatus
}

func (s *MockXEndorserServer) PostTx(context.Context, *pb.TxStatus) (*pb.CommonReply, error) {
	return nil, nil
}
func (s *MockXEndorserServer) QueryTx(context context.Context, txStatus *pb.TxStatus) (*pb.TxStatus, error) {
	return s.Txs[string(txStatus.Txid)], nil
}
func (s *MockXEndorserServer) PreExecWithSelectUTXO(context.Context, *pb.PreExecWithSelectUTXORequest) (*pb.PreExecWithSelectUTXOResponse, error) {
	return nil, nil
}
func (s *MockXEndorserServer) PreExec(context.Context, *pb.InvokeRPCRequest) (*pb.InvokeRPCResponse, error) {
	return nil, nil
}

func TestGetTx(t *testing.T) {
	if baseDir == "" {
		return
	}
	if err := os.Chdir(baseDir); err != nil {
		t.Fatal(err)
	}

	endorser := NewDefaultXEndorser()
	params := map[string]interface{}{
		"server": &MockXEndorserServer{
			Txs: map[string]*pb.TxStatus{
				"test123": {
					Tx: &pb.Transaction{
						Txid: []byte("test123"),
					},
				},
			},
		},
	}

	if err := endorser.Init("", params); err != nil {
		t.Fatal(err)
	}

	request := &pb.TxStatus{
		Bcname: "xuper",
		Txid:   []byte("test123"),
	}
	reqData, err := json.Marshal(request)
	if err != nil {
		t.Fatal("unmarshall reqData error", "err", err.Error())
	}
	req := &pb.EndorserRequest{
		RequestName: "TxQuery",
		BcName:      "xuper",
		RequestData: reqData,
	}

	ctx, _ := context.WithTimeout(context.TODO(), 6*time.Second)
	endorsorRes, err := endorser.EndorserCall(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	res := &pb.Transaction{}
	err = json.Unmarshal(endorsorRes.GetResponseData(), res)
	if err != nil {
		t.Fatal("endorsorQuery Unmarshal error", "err", err)
	}

	if string(res.Txid) != "test123" {
		t.Fatal("endorser query tx res error")
	}
}

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

type MorkXEndorserServer struct {
	Txs map[string]*pb.TxStatus
}

func (s *MorkXEndorserServer) PostTx(context.Context, *pb.TxStatus) (*pb.CommonReply, error)  { return nil, nil }
func (s *MorkXEndorserServer) QueryTx(context context.Context, txStatus *pb.TxStatus) (*pb.TxStatus, error) {
	return s.Txs[string(txStatus.Txid)], nil
}
func (s *MorkXEndorserServer) PreExecWithSelectUTXO(context.Context, *pb.PreExecWithSelectUTXORequest) (*pb.PreExecWithSelectUTXOResponse, error) {
	return nil, nil
}
func (s *MorkXEndorserServer) PreExec(context.Context, *pb.InvokeRPCRequest) (*pb.InvokeRPCResponse, error) {
	return nil, nil
}

func TestGetTx(t *testing.T) {
	morkServer := &MorkXEndorserServer{
		Txs: make(map[string]*pb.TxStatus),
	}
	morkServer.Txs["test123"] = &pb.TxStatus{
		Tx: &pb.Transaction{
			Txid: []byte("test123"),
		},
	}
	endorser := NewDefaultXEndorser()

	params := make(map[string]interface{})
	params["server"] = morkServer

	if err := endorser.Init("", params); err != nil {
		t.Error(err)
	}

	request := &pb.TxStatus{
		Bcname: "xuper",
		Txid:   []byte("test123"),
	}
	reqData, err := json.Marshal(request)
	if err != nil {
		t.Error("unmarshall reqData error", "err", err.Error())
	}

	req := &pb.EndorserRequest{
		RequestName: "TxQuery",
		BcName:      "xuper",
		RequestData: reqData,
	}

	ctx, _ := context.WithTimeout(context.TODO(), 6*time.Second)

	os.Chdir(baseDir)

	endorsorRes, err := endorser.EndorserCall(ctx, req)
	if err != nil {
		t.Error(err)
	}
	res := &pb.Transaction{}
	err = json.Unmarshal(endorsorRes.GetResponseData(), res)
	if err != nil {
		t.Error("endorsorQuery Unmarshal error", "err", err)
	}

	if string(res.Txid) != "test123"{
		t.Error("endorser query tx res error")
	}
}

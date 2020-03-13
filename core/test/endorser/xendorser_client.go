package main

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc"
	"os"
	"time"

	"github.com/xuperchain/log15"
	"github.com/xuperchain/xuperchain/core/pb"
)

var (
	address    = "dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN"
	publickey  = "{\"Curvname\":\"P-256\",\"X\":74695617477160058757747208220371236837474210247114418775262229497812962582435,\"Y\":51348715319124770392993866417088542497927816017012182211244120852620959209571}"
	privatekey = "{\"Curvname\":\"P-256\",\"X\":74695617477160058757747208220371236837474210247114418775262229497812962582435,\"Y\":51348715319124770392993866417088542497927816017012182211244120852620959209571,\"D\":29079635126530934056640915735344231956621504557963207107451663058887647996601}"
)

func main() {
	fmt.Println("Start test")
	invokeReq := make([]*pb.InvokeRequest, 0)
	invoke := &pb.InvokeRequest{
		ModuleName:   "wasm",
		ContractName: "counter",
		MethodName:   "increase",
		Args:         map[string][]byte{"key": []byte("test")},
	}
	invokeReq = append(invokeReq, invoke)
	req := &pb.CrossQueryRequest{
		Bcname:    "xuper",
		Timestamp: time.Now().Unix(),
		Initiator: address,
		Request:   invoke,
	}
	reqJSON, _ := json.Marshal(req)
	xreq := &pb.EndorserRequest{
		RequestName: "CrossQueryPreExec",
		BcName:      "xuper",
		Fee:         nil,
		RequestData: reqJSON,
	}
	// read conf
	xlog := log.New("app", "xendorser")
	xlog.SetHandler(log.StreamHandler(os.Stdout, log.LogfmtFormat()))
	conn, err := grpc.Dial("127.0.0.1:37101",
		grpc.WithInsecure())
	if err != nil {
		xlog.Warn("connect to xchain service failed", "error", err)
		return
	}
	defer conn.Close()
	client := pb.NewXendorserClient(conn)
	ctx := context.Background()
	res, err := client.EndorserCall(ctx, xreq)
	if err != nil {
		xlog.Warn("PreExecWithSelectUTXO failed", "error", err)
		return
	}
	xlog.Info("get endorser result", "result", res)
}

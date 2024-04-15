package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/xuperchain/xuperchain/data/mock"
	scom "github.com/xuperchain/xuperchain/service/common"
	"github.com/xuperchain/xuperchain/service/pb"

	// import要使用的内核核心组件驱动
	_ "github.com/xuperchain/xupercore/bcs/consensus/pow"
	_ "github.com/xuperchain/xupercore/bcs/consensus/single"
	_ "github.com/xuperchain/xupercore/bcs/consensus/tdpos"
	_ "github.com/xuperchain/xupercore/bcs/consensus/xpoa"
	_ "github.com/xuperchain/xupercore/bcs/contract/evm"
	_ "github.com/xuperchain/xupercore/bcs/contract/native"
	_ "github.com/xuperchain/xupercore/bcs/contract/xvm"
	txn "github.com/xuperchain/xupercore/bcs/ledger/xledger/tx"
	xledger "github.com/xuperchain/xupercore/bcs/ledger/xledger/utils"
	_ "github.com/xuperchain/xupercore/bcs/network/p2pv1"
	_ "github.com/xuperchain/xupercore/bcs/network/p2pv2"
	xconf "github.com/xuperchain/xupercore/kernel/common/xconfig"
	_ "github.com/xuperchain/xupercore/kernel/contract/kernel"
	_ "github.com/xuperchain/xupercore/kernel/contract/manager"
	"github.com/xuperchain/xupercore/kernel/engines/xuperos"
	"github.com/xuperchain/xupercore/kernel/engines/xuperos/common"
	_ "github.com/xuperchain/xupercore/lib/crypto/client"
	"github.com/xuperchain/xupercore/lib/logs"
	_ "github.com/xuperchain/xupercore/lib/storage/kvdb/leveldb"
)

var (
	address   = "dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN"
	publickey = "{\"Curvname\":\"P-256\",\"X\":74695617477160058757747208220371236837474210247114418775262229497812962582435,\"Y\":51348715319124770392993866417088542497927816017012182211244120852620959209571}"
)

func TestEndorserCall(t *testing.T) {
	workspace := os.TempDir()
	os.RemoveAll(workspace)
	defer os.RemoveAll(workspace)
	conf, _ := mock.NewEnvConfForTest()
	defer removeLedger(conf)

	engine, err := MockEngine()
	if err != nil {
		t.Fatal(err)
	}
	log, _ := logs.NewLogger("", scom.SubModName)
	rpcServ := NewRpcServ(engine, log)

	endor := NewDefaultXEndorser(rpcServ, engine)
	awardTx, err := txn.GenerateAwardTx("miner", "1000", []byte("award"))
	if err != nil {
		t.Fatalf("txn.GenerateAwardTx() err: %s", err)
	}

	txStatus := &pb.TxStatus{
		Bcname: "xuper",
		Tx:     scom.TxToXchain(awardTx),
	}
	requestData, err := json.Marshal(txStatus)
	if err != nil {
		fmt.Printf("json encode txStatus failed: %v", err)
		t.Fatal(err)
	}
	ctx := context.TODO()
	req := &pb.EndorserRequest{
		RequestName: "ComplianceCheck",
		BcName:      "xuper",
		Fee:         nil,
		RequestData: requestData,
	}
	resp, err := endor.EndorserCall(ctx, req)
	if err != nil {
		t.Log(err)
	}
	t.Log(resp)
	invokeReq := make([]*pb.InvokeRequest, 0)
	invoke := &pb.InvokeRequest{
		ModuleName:   "wasm",
		ContractName: "counter",
		MethodName:   "increase",
		Args:         map[string][]byte{"key": []byte("test")},
	}
	invokeReq = append(invokeReq, invoke)
	preq := &pb.PreExecWithSelectUTXORequest{
		Bcname:      "xuper",
		Address:     address,
		TotalAmount: 100,
		SignInfo: &pb.SignatureInfo{
			PublicKey: publickey,
			Sign:      []byte("sign"),
		},
		NeedLock: false,
		Request: &pb.InvokeRPCRequest{
			Bcname:      "xuper",
			Requests:    invokeReq,
			Initiator:   address,
			AuthRequire: []string{address},
		},
	}

	reqJSON, _ := json.Marshal(preq)
	xreq := &pb.EndorserRequest{
		RequestName: "PreExecWithFee",
		BcName:      "xuper",
		Fee:         nil,
		RequestData: reqJSON,
	}
	resp, err = endor.EndorserCall(ctx, xreq)
	if err != nil {
		//pass
		t.Log(err)
	}
	t.Log(resp)
	qtxTxStatus := &pb.TxStatus{
		Bcname: "xuper",
		Txid:   []byte("70c64d6cb9b5647048d067c6775575fc52e3c51c6425cec3881d8564ad8e887c"),
	}
	requestData, err = json.Marshal(qtxTxStatus)
	if err != nil {
		fmt.Printf("json encode txStatus failed: %v", err)
		t.Fatal(err)
	}
	req = &pb.EndorserRequest{
		RequestName: "TxQuery",
		BcName:      "xuper",
		RequestData: requestData,
	}
	resp, err = endor.EndorserCall(ctx, req)
	if err != nil {
		t.Log(err)
	}
	t.Log(resp)
}

func MockEngine() (common.Engine, error) {
	conf, err := mock.NewEnvConfForTest()
	if err != nil {
		return nil, fmt.Errorf("new env conf error: %v", err)
	}

	if err = createLedger(conf); err != nil {
		return nil, err
	}

	engine := xuperos.NewEngine()
	if err := engine.Init(conf); err != nil {
		return nil, fmt.Errorf("init engine error: %v", err)
	}

	eng, err := xuperos.EngineConvert(engine)
	if err != nil {
		return nil, fmt.Errorf("engine convert error: %v", err)
	}

	return eng, nil
}

func removeLedger(conf *xconf.EnvConf) {
	path := conf.GenDataAbsPath("blockchain")
	if err := os.RemoveAll(path); err != nil {
		log.Printf("remove ledger failed.err:%v\n", err)
	}
}

func createLedger(conf *xconf.EnvConf) error {
	// init env
	removeLedger(conf)

	mockConf, err := mock.NewEnvConfForTest()
	if err != nil {
		return fmt.Errorf("new mock env conf error: %v", err)
	}

	genesisPath := mockConf.GenDataAbsPath("genesis/xuper.json")
	err = xledger.CreateLedger("xuper", genesisPath, conf)
	if err != nil {
		log.Printf("create ledger failed.err:%v\n", err)
		return fmt.Errorf("create ledger failed")
	}
	return nil
}

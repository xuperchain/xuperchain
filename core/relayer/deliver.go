package relayer

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc"

	"github.com/xuperchain/xuperchain/core/contract"
	"github.com/xuperchain/xuperchain/core/global"
	"github.com/xuperchain/xuperchain/core/pb"
	"github.com/xuperchain/xuperchain/core/utxo"
	"github.com/xuperchain/xuperchain/core/utxo/txhash"
)

type DeliverBlockCommand struct {
	client pb.XchainClient
	Cfg    ChainConfig
}

func (cmd *DeliverBlockCommand) InitXchainClient() error {
	conn, err := grpc.Dial(cmd.Cfg.RPCAddr, grpc.WithInsecure(), grpc.WithMaxMsgSize(64<<20-1))
	if err != nil {
		return err
	}
	cmd.client = pb.NewXchainClient(conn)
	return nil
}

func (cmd *DeliverBlockCommand) DeliverAnchorBlockHeader(blockBuf []byte) error {
	args := make(map[string][]byte)
	args["blockHeader"] = blockBuf
	// set preExe parameter
	moduleName := cmd.Cfg.ContractConfig.ModuleName
	contractName := cmd.Cfg.ContractConfig.ContractName
	methodName := cmd.Cfg.ContractConfig.AnchorMethod
	tx, err := cmd.PreExe(moduleName, contractName, methodName, args)
	if err != nil {
		return err
	}
	txid, err := cmd.SendTx(tx)
	if err != nil {
		return err
	}
	fmt.Println("txid:", txid)
	return nil
}

func (cmd *DeliverBlockCommand) DeliverBlockHeader(blockBuf []byte) error {
	// step1: fetch block header
	// TODO
	// step2: prepare exec
	args := make(map[string][]byte)
	args["blockHeader"] = blockBuf
	// set preExe parameter
	moduleName := cmd.Cfg.ContractConfig.ModuleName
	contractName := cmd.Cfg.ContractConfig.ContractName
	methodName := cmd.Cfg.ContractConfig.UpdateMethod
	tx, err := cmd.PreExe(moduleName, contractName, methodName, args)
	if err != nil {
		return err
	}
	// step3: postTx
	txid, err := cmd.SendTx(tx)
	if err != nil {
		return err
	}
	fmt.Println("txid:", txid)
	return nil
}

func (cmd *DeliverBlockCommand) PreExe(moduleName, contractName, methodName string,
	args map[string][]byte) (*pb.Transaction, error) {
	preExeRPCRes, preExeReqs, err := cmd.GenPreExeRes(moduleName, contractName, methodName, args)
	if err != nil {
		return nil, err
	}

	return cmd.GenRawTx(preExeRPCRes.GetResponse(), preExeReqs)
}

func (cmd *DeliverBlockCommand) GenPreExeRes(moduleName, contractName, methodName string,
	args map[string][]byte) (
	*pb.InvokeRPCResponse, []*pb.InvokeRequest, error) {
	preExeReqs := []*pb.InvokeRequest{}
	preExeReqs = append(preExeReqs, &pb.InvokeRequest{
		ModuleName:   moduleName,
		ContractName: contractName,
		MethodName:   methodName,
		Args:         args,
	})

	preExeRPCReq := &pb.InvokeRPCRequest{
		Bcname:   cmd.Cfg.Bcname,
		Header:   global.GHeader(),
		Requests: preExeReqs,
	}

	// 根据配置读取initiator以及authrequire
	initiator, err := readAddress(cmd.Cfg.Keys)
	if err != nil {
		panic("read address error")
	}
	preExeRPCReq.Initiator = initiator
	preExeRPCReq.AuthRequire = []string{initiator}

	preExeRPCRes, err := cmd.client.PreExec(context.TODO(), preExeRPCReq)
	if err != nil {
		return nil, nil, fmt.Errorf("PreExe contract response : %v, logid:%s", err, preExeRPCReq.Header.Logid)
	}
	for _, res := range preExeRPCRes.Response.Responses {
		if res.Status >= contract.StatusErrorThreshold {
			return nil, nil, fmt.Errorf("contract error status:%d message:%s", res.Status, res.Message)
		}
		fmt.Printf("contract response: %s\n", string(res.Body))
	}

	return preExeRPCRes, preExeRPCRes.Response.Requests, nil
}

func (cmd *DeliverBlockCommand) GenRawTx(preExeRes *pb.InvokeResponse, preExeReqs []*pb.InvokeRequest) (
	*pb.Transaction, error) {
	tx := &pb.Transaction{
		Coinbase:  false,
		Nonce:     global.GenNonce(),
		Timestamp: time.Now().UnixNano(),
		Version:   utxo.TxVersion,
	}

	var gasUsed int64
	if preExeRes != nil {
		gasUsed = preExeRes.GasUsed
		fmt.Printf("The gas you consume is: %v\n", gasUsed)
	}

	txOutputs, totalNeed, err := cmd.GenTxOutputs(gasUsed)
	if err != nil {
		return nil, err
	}
	tx.TxOutputs = append(tx.TxOutputs, txOutputs...)
	txInputs, deltaTxOutput, err := cmd.GenTxInputs(totalNeed)
	if err != nil {
		return nil, err
	}
	tx.TxInputs = append(tx.TxInputs, txInputs...)
	if deltaTxOutput != nil {
		tx.TxOutputs = append(tx.TxOutputs, deltaTxOutput)
	}

	// 填充contract预执行结果
	if preExeRes != nil {
		tx.TxInputsExt = preExeRes.GetInputs()
		tx.TxOutputsExt = preExeRes.GetOutputs()
		tx.ContractRequests = preExeReqs
	}

	// 填充交易发起者的addr
	initiator, err := readAddress(cmd.Cfg.Keys)
	if err != nil {
		panic("read address error")
	}
	tx.Initiator = initiator

	return tx, nil
}

func (cmd *DeliverBlockCommand) SendTx(tx *pb.Transaction) (string, error) {
	initiator, err := readAddress(cmd.Cfg.Keys)
	if err != nil {
		panic("read address error")
	}
	tx.AuthRequire = []string{initiator}
	signInfos, err := cmd.GenInitSign(tx)
	if err != nil {
		return "", err
	}
	tx.InitiatorSigns = signInfos
	tx.AuthRequireSigns = signInfos

	tx.Txid, err = txhash.MakeTransactionID(tx)
	if err != nil {
		return "", errors.New("MakeTxDigestHash txid error")
	}

	txStatus := &pb.TxStatus{
		Bcname: cmd.Cfg.Bcname,
		Status: pb.TransactionStatus_UNCONFIRM,
		Tx:     tx,
		Header: &pb.Header{
			Logid: global.Glogid(),
		},
		Txid: tx.Txid,
	}

	reply, err := cmd.client.PostTx(context.TODO(), txStatus)
	if err != nil {
		return "", err
	}
	if reply.Header.Error != pb.XChainErrorEnum_SUCCESS {
		return "", fmt.Errorf("failed to post tx:%s, logid:%s", reply.Header.Error.String(), reply.Header.Logid)
	}

	return hex.EncodeToString(txStatus.Txid), nil
}

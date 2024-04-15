/*
 * Copyright (c) 2021. Baidu Inc. All Rights Reserved.
 */

package cmd

import (
	"bufio"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"time"

	"github.com/golang/protobuf/proto" //nolint:staticcheck
	"github.com/xuperchain/xupercore/bcs/ledger/xledger/state/utxo"
	"github.com/xuperchain/xupercore/kernel/contract"
	"github.com/xuperchain/xupercore/lib/crypto/client"
	"github.com/xuperchain/xupercore/lib/utils"
	"google.golang.org/grpc"

	"github.com/xuperchain/xuperchain/models"
	"github.com/xuperchain/xuperchain/service/common"
	"github.com/xuperchain/xuperchain/service/pb"
)

const (
	defaultDesc = "Maybe common transfer transaction"
)

// CommTrans base method
// a. generate tx
// b. sign it for default single and send it
type CommTrans struct {
	To           string
	Amount       string
	Descfile     string
	Fee          string
	FrozenHeight int64
	Version      int32
	From         string
	ModuleName   string
	ContractName string
	MethodName   string
	Args         map[string][]byte
	// 走mulitisig gen流程
	MultiAddrs string
	Output     string
	IsQuick    bool
	IsPrint    bool

	ChainName    string
	Keys         string
	XchainClient pb.XchainClient
	CryptoType   string
	RootOptions  RootOptions

	// DebugTx if enabled, tx will be printed instead of being posted
	DebugTx bool
}

// GenerateTx generate raw tx
func (t *CommTrans) GenerateTx(ctx context.Context) (*pb.Transaction, error) {
	preExeRPCRes, preExeReqs, err := t.GenPreExeRes(ctx)
	if err != nil {
		return nil, err
	}

	desc, _ := t.GetDesc()

	tx, err := t.GenRawTx(ctx, desc, preExeRPCRes.GetResponse(), preExeReqs)
	return tx, err
}

// GenPreExeRes 得到预执行的结果
func (t *CommTrans) GenPreExeRes(ctx context.Context) (
	*pb.InvokeRPCResponse, []*pb.InvokeRequest, error) {
	preExeReqs := []*pb.InvokeRequest{}
	if t.ModuleName != "" {
		if t.ModuleName == "xkernel" {
			preExeReqs = append(preExeReqs, &pb.InvokeRequest{
				ModuleName:   t.ModuleName,
				ContractName: t.ContractName,
				MethodName:   t.MethodName,
				Args:         t.Args,
			})
		} else {
			invokeReq := &pb.InvokeRequest{
				ModuleName:   t.ModuleName,
				ContractName: t.ContractName,
				MethodName:   t.MethodName,
				Args:         t.Args,
			}
			// transfer to contract
			if t.To == t.ContractName {
				invokeReq.Amount = t.Amount
			}
			preExeReqs = append(preExeReqs, invokeReq)
		}
	} else {
		tmpReq, err := t.GetInvokeRequestFromDesc()
		if err != nil {
			return nil, nil, fmt.Errorf("Get pb.InvokeRPCRequest error:%s", err)
		}
		if tmpReq != nil {
			preExeReqs = append(preExeReqs, tmpReq)
		}
	}

	preExeRPCReq := &pb.InvokeRPCRequest{
		Bcname: t.ChainName,
		Header: &pb.Header{
			Logid: utils.GenLogId(),
		},
		Requests: preExeReqs,
	}

	initiator, err := t.genInitiator()
	if err != nil {
		return nil, nil, fmt.Errorf("Get initiator error: %s", err.Error())
	}

	preExeRPCReq.Initiator = initiator
	if !t.IsQuick {
		preExeRPCReq.AuthRequire, err = t.genAuthRequireQuick()
		if err != nil {
			return nil, nil, fmt.Errorf("Get auth require quick error: %s", err.Error())
		}
	} else {
		preExeRPCReq.AuthRequire, err = t.GenAuthRequire(t.MultiAddrs)
		if err != nil {
			return nil, nil, fmt.Errorf("Get auth require error: %s", err.Error())
		}
	}
	preExeRPCRes, err := t.XchainClient.PreExec(ctx, preExeRPCReq)
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

// GetInvokeRequestFromDesc get invokerequest from desc file
func (t *CommTrans) GetInvokeRequestFromDesc() (*pb.InvokeRequest, error) {
	desc, err := t.GetDesc()
	if err != nil {
		return nil, fmt.Errorf("get desc error:%s", err)
	}

	var preExeReq *pb.InvokeRequest
	preExeReq, err = t.ReadPreExeReq(desc)
	if err != nil {
		return nil, err
	}

	return preExeReq, nil
}

// GetDesc 解析desc字段，主要是针对合约
func (t *CommTrans) GetDesc() ([]byte, error) {
	if t.Descfile == "" {
		return []byte(defaultDesc), nil
	}
	return os.ReadFile(t.Descfile)
}

// ReadPreExeReq 从desc中填充出发起合约调用的结构体
func (t *CommTrans) ReadPreExeReq(buf []byte) (*pb.InvokeRequest, error) {
	params := new(invokeRequestWrapper)
	err := json.Unmarshal(buf, params)
	if err != nil {
		return nil, nil
	}

	if params.InvokeRequest.ModuleName == "" {
		return nil, nil
	}

	params.InvokeRequest.Args = make(map[string][]byte)
	for k, v := range params.Args {
		params.InvokeRequest.Args[k] = []byte(v)
	}
	return &params.InvokeRequest, nil
}

// GenRawTx 生成一个完整raw的交易
func (t *CommTrans) GenRawTx(ctx context.Context, desc []byte, preExeRes *pb.InvokeResponse,
	preExeReqs []*pb.InvokeRequest) (*pb.Transaction, error) {
	tx := &pb.Transaction{
		Desc:      desc,
		Coinbase:  false,
		Nonce:     utils.GenNonce(),
		Timestamp: time.Now().UnixNano(),
		Version:   utxo.TxVersion,
	}

	var gasUsed int64
	if preExeRes != nil {
		gasUsed = preExeRes.GasUsed
		fmt.Printf("The gas you cousume is: %v\n", gasUsed)
	}

	if preExeRes.GetUtxoInputs() != nil {
		tx.TxInputs = append(tx.TxInputs, preExeRes.GetUtxoInputs()...)
	}
	if preExeRes.GetUtxoOutputs() != nil {
		tx.TxOutputs = append(tx.TxOutputs, preExeRes.GetUtxoOutputs()...)
	}

	txOutputs, totalNeed, err := t.GenTxOutputs(gasUsed)
	if err != nil {
		return nil, err
	}
	tx.TxOutputs = append(tx.TxOutputs, txOutputs...)

	txInputs, deltaTxOutput, err := t.GenTxInputs(ctx, totalNeed)
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
	fromAddr, err := t.genInitiator()
	if err != nil {
		return nil, err
	}
	tx.Initiator = fromAddr

	return tx, nil
}

// genInitiator generate initiator of transaction
func (t *CommTrans) genInitiator() (string, error) {
	if t.From != "" {
		return t.From, nil
	}
	fromAddr, err := readAddress(t.Keys)
	if err != nil {
		return "", err
	}
	return fromAddr, nil
}

// GenTxOutputs 填充得到transaction的repeated TxOutput tx_outputs
func (t *CommTrans) GenTxOutputs(gasUsed int64) ([]*pb.TxOutput, *big.Int, error) {
	// 组装转账的账户信息
	account := &pb.TxDataAccount{
		Address:      t.To,
		Amount:       t.Amount,
		FrozenHeight: t.FrozenHeight,
	}

	accounts := []*pb.TxDataAccount{}
	if t.To != "" {
		accounts = append(accounts, account)
	}

	// 如果有小费,增加转个小费地址的账户
	// 如果有合约，需要支付gas
	if gasUsed > 0 {
		if t.Fee != "" && t.Fee != "0" {
			fee, _ := strconv.ParseInt(t.Fee, 10, 64)
			if fee < gasUsed {
				return nil, nil, errors.New("Fee not enough")
			}
		} else {
			return nil, nil, errors.New("You need add fee")
		}
		fmt.Printf("The fee you pay is: %v\n", t.Fee)
		accounts = append(accounts, newFeeAccount(t.Fee))
	} else if t.Fee != "" && t.Fee != "0" && gasUsed <= 0 {
		fmt.Printf("The fee you pay is: %v\n", t.Fee)
		accounts = append(accounts, newFeeAccount(t.Fee))
	}

	// 组装txOutputs
	bigZero := big.NewInt(0)
	totalNeed := big.NewInt(0)
	txOutputs := []*pb.TxOutput{}
	for _, acc := range accounts {
		amount, ok := big.NewInt(0).SetString(acc.Amount, 10)
		if !ok {
			return nil, nil, ErrInvalidAmount
		}
		cmpRes := amount.Cmp(bigZero)
		if cmpRes < 0 {
			return nil, nil, ErrNegativeAmount
		} else if cmpRes == 0 {
			// trim 0 output
			continue
		}
		// 得到总的转账金额
		totalNeed.Add(totalNeed, amount)

		txOutput := &pb.TxOutput{}
		txOutput.Amount = amount.Bytes()
		txOutput.ToAddr = []byte(acc.Address)
		txOutput.FrozenHeight = acc.FrozenHeight
		txOutputs = append(txOutputs, txOutput)
	}

	return txOutputs, totalNeed, nil
}

// GenTxInputs 填充得到transaction的repeated TxInput tx_inputs,
// 如果输入大于输出，增加一个转给自己(data/keys/)的输入-输出的交易
func (t *CommTrans) GenTxInputs(ctx context.Context, totalNeed *big.Int) (
	[]*pb.TxInput, *pb.TxOutput, error) {
	var fromAddr string
	var err error
	if t.From != "" {
		fromAddr = t.From
	} else {
		fromAddr, err = readAddress(t.Keys)
		if err != nil {
			return nil, nil, err
		}
	}

	utxoInput := &pb.UtxoInput{
		Bcname:    t.ChainName,
		Address:   fromAddr,
		TotalNeed: totalNeed.String(),
		NeedLock:  false,
	}

	utxoOutputs, err := t.XchainClient.SelectUTXO(ctx, utxoInput)
	if err != nil {
		return nil, nil, fmt.Errorf("%v, details:%v", ErrSelectUtxo, err)
	}
	if utxoOutputs.Header.Error != pb.XChainErrorEnum_SUCCESS {
		return nil, nil, fmt.Errorf("%v, details:%v", ErrSelectUtxo, utxoOutputs.Header.Error)
	}

	// 组装txInputs
	var txInputs []*pb.TxInput
	var txOutput *pb.TxOutput
	for _, utxo := range utxoOutputs.UtxoList {
		txInput := &pb.TxInput{}
		txInput.RefTxid = utxo.RefTxid
		txInput.RefOffset = utxo.RefOffset
		txInput.FromAddr = utxo.ToAddr
		txInput.Amount = utxo.Amount
		txInputs = append(txInputs, txInput)
	}

	utxoTotal, ok := big.NewInt(0).SetString(utxoOutputs.TotalSelected, 10)
	if !ok {
		return nil, nil, ErrSelectUtxo
	}

	// 通过selectUTXO选出的作为交易的输入大于输出,
	// 则多出来再生成一笔交易转给自己
	if utxoTotal.Cmp(totalNeed) > 0 {
		delta := utxoTotal.Sub(utxoTotal, totalNeed)
		txOutput = &pb.TxOutput{
			ToAddr: []byte(fromAddr),
			Amount: delta.Bytes(),
		}
	}

	return txInputs, txOutput, nil
}

// Transfer quick access to transfer
func (t *CommTrans) Transfer(ctx context.Context) error {
	if t.RootOptions.ComplianceCheck.IsNeedComplianceCheck {
		preSelectUTXORes, err := t.GenPreExeWithSelectUtxoRes(ctx)
		if err != nil {
			return err
		}
		return t.GenCompleteTxAndPost(ctx, preSelectUTXORes)
	} else {
		tx, err := t.GenerateTx(ctx)
		if err != nil {
			return err
		}

		if t.DebugTx {
			ttx := FromPBTx(tx)
			out, _ := json.MarshalIndent(ttx, "", "  ")
			fmt.Println(string(out))
			return nil
		}

		return t.SendTx(ctx, tx)
	}
}

// SendTx post tx
func (t *CommTrans) SendTx(ctx context.Context, tx *pb.Transaction) error {
	fromAddr, err := readAddress(t.Keys)
	if err != nil {
		return err
	}

	var authRequire string
	if t.From != "" {
		authRequire = t.From + "/" + fromAddr
	} else {
		authRequire = fromAddr
	}
	tx.AuthRequire = append(tx.AuthRequire, authRequire)

	signInfos, err := t.signTxForInitiator(tx)
	if err != nil {
		return err
	}
	tx.InitiatorSigns = signInfos
	tx.AuthRequireSigns = signInfos

	tx.Txid, err = common.MakeTxId(tx)
	if err != nil {
		return errors.New("MakeTxId error")
	}

	txID, err := t.postTx(ctx, tx)
	if err != nil {
		return err
	}
	fmt.Printf("Tx id: %s\n", txID)

	return nil
}

// signTx generates auth required signatures for transaction according to path
func (t *CommTrans) signTx(tx *pb.Transaction, path string) ([]*pb.SignatureInfo, error) {

	// generate signature for AK: signed by initiator
	if path == "" {
		return t.signTxForInitiator(tx)
	}

	// generate signatures for account: signed by all AK in account path
	return t.signTxForAccount(tx, path)
}

// signTxForInitiator generates initiator's signature for this Tx
func (t *CommTrans) signTxForInitiator(tx *pb.Transaction) ([]*pb.SignatureInfo, error) {
	initiator := newAK(t.Keys)
	keyPair, err := initiator.keyPair()
	if err != nil {
		return nil, err
	}

	// create crypto client
	crypto, err := client.CreateCryptoClient(t.CryptoType)
	if err != nil {
		return nil, errors.New("Create crypto client error")
	}

	// sign by initiator
	return signTxForAK(tx, keyPair, crypto)
}

// signTxForAccount generates transaction signatures for account
func (t *CommTrans) signTxForAccount(tx *pb.Transaction, path string) ([]*pb.SignatureInfo, error) {

	// create crypto client
	crypto, err := client.CreateCryptoClient(t.CryptoType)
	if err != nil {
		return nil, errors.New("Create crypto client error")
	}

	return signTxForAccount(tx, path, crypto)
}

func (t *CommTrans) postTx(ctx context.Context, tx *pb.Transaction) (string, error) {
	txStatus := &pb.TxStatus{
		Bcname: t.ChainName,
		Status: pb.TransactionStatus_UNCONFIRM,
		Tx:     tx,
		Header: &pb.Header{
			Logid: utils.GenLogId(),
		},
		Txid: tx.Txid,
	}

	reply, err := t.XchainClient.PostTx(ctx, txStatus)
	if err != nil {
		return "", err
	}
	if reply.Header.Error != pb.XChainErrorEnum_SUCCESS {
		return "", fmt.Errorf("Failed to post tx:%s, logid:%s", reply.Header.Error.String(), reply.Header.Logid)
	}

	return hex.EncodeToString(txStatus.Txid), nil
}

// GenerateMultisigGenRawTx for multisig gen cmd
func (t *CommTrans) GenerateMultisigGenRawTx(ctx context.Context) error {
	tx, err := t.GenerateTx(ctx)
	if err != nil {
		return err
	}

	// 填充需要多重签名的addr
	multiAddrs, err := t.GenAuthRequire(t.MultiAddrs)
	if err != nil {
		return err
	}
	tx.AuthRequire = multiAddrs

	return t.GenTxFile(tx)
}

func (t *CommTrans) genAuthRequireQuick() ([]string, error) {

	fromAddr, err := readAddress(t.Keys)
	if err != nil {
		return nil, err
	}
	authRequires := []string{}
	if t.From != "" {
		authRequires = append(authRequires, t.From+"/"+fromAddr)
	} else {
		authRequires = append(authRequires, fromAddr)
	}
	return authRequires, nil
}

// GenAuthRequire get auth require aks from file
func (t *CommTrans) GenAuthRequire(filename string) ([]string, error) {
	var addrs []string

	fileIn, err := os.Open(filename)
	if err != nil {
		return addrs, err
	}
	defer fileIn.Close()

	scanner := bufio.NewScanner(fileIn)
	for scanner.Scan() {
		addr := scanner.Text()
		if addr == "" {
			continue
		}
		addrs = append(addrs, addr)
	}

	if err := scanner.Err(); err != nil {
		return addrs, err
	}

	return addrs, nil
}

// GenTxFile generate raw tx file
func (t *CommTrans) GenTxFile(tx *pb.Transaction) error {
	data, err := proto.Marshal(tx)
	if err != nil {
		return errors.New("Tx marshal error")
	}
	err = os.WriteFile(t.Output, data, 0755)
	if err != nil {
		return errors.New("WriteFile error")
	}

	if t.IsPrint {
		return printTx(tx)
	}

	return nil
}

func printTx(tx *pb.Transaction) error {
	// print tx
	t := FromPBTx(tx)
	output, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))

	return nil
}

// GenTxInputsWithMergeUTXO generate tx with merge utxo
func (t *CommTrans) GenTxInputsWithMergeUTXO(ctx context.Context) ([]*pb.TxInput, *pb.TxOutput, error) {
	var fromAddr string
	var err error
	if t.From != "" {
		fromAddr = t.From
	} else {
		fromAddr, err = readAddress(t.Keys)
		if err != nil {
			return nil, nil, err
		}
	}

	utxoInput := &pb.UtxoInput{
		Bcname:   t.ChainName,
		Address:  fromAddr,
		NeedLock: true,
	}
	signature, err := t.signLockUtxo(utxoInput)
	if err != nil {
		return nil, nil, err
	}
	utxoInput.Publickey = signature.PublicKey
	utxoInput.UserSign = signature.Sign
	utxoOutputs, err := t.XchainClient.SelectUTXOBySize(ctx, utxoInput)
	if err != nil {
		return nil, nil, fmt.Errorf("%v, details:%v", ErrSelectUtxo, err)
	}
	if utxoOutputs.Header.Error != pb.XChainErrorEnum_SUCCESS {
		return nil, nil, fmt.Errorf("%v, details:%v", ErrSelectUtxo, utxoOutputs.Header.Error)
	}

	// 组装txInputs
	var txInputs []*pb.TxInput
	var txOutput *pb.TxOutput
	for _, utxo := range utxoOutputs.UtxoList {
		txInput := &pb.TxInput{
			RefTxid:   utxo.RefTxid,
			RefOffset: utxo.RefOffset,
			FromAddr:  utxo.ToAddr,
			Amount:    utxo.Amount,
		}
		txInputs = append(txInputs, txInput)
	}

	utxoTotal, ok := big.NewInt(0).SetString(utxoOutputs.TotalSelected, 10)
	if !ok {
		return nil, nil, ErrSelectUtxo
	}
	txOutput = &pb.TxOutput{
		ToAddr: []byte(fromAddr),
		Amount: utxoTotal.Bytes(),
	}

	return txInputs, txOutput, nil
}

// invokeReq generates invoke request by module name
func (t *CommTrans) invokeReq() (*pb.InvokeRequest, error) {
	if t.ModuleName == "" {
		req, err := t.GetInvokeRequestFromDesc()
		if err != nil {
			return nil, fmt.Errorf("Get pb.InvokeRPCRequest error:%s", err)
		}
		return req, nil
	}

	// TODO: extract to constant
	if t.ModuleName == "xkernel" {
		req := &pb.InvokeRequest{
			ModuleName:   t.ModuleName,
			ContractName: t.ContractName,
			MethodName:   t.MethodName,
			Args:         t.Args,
		}
		return req, nil
	}

	req := &pb.InvokeRequest{
		ModuleName:   t.ModuleName,
		ContractName: t.ContractName,
		MethodName:   t.MethodName,
		Args:         t.Args,
	}
	// transfer to contract
	if t.To == t.ContractName {
		req.Amount = t.Amount
	}
	return req, nil
}

// preExecWithSelectUTXOReq generates preExecWithSelectUTXO request
func (t *CommTrans) preExecWithSelectUTXOReq() (*pb.PreExecWithSelectUTXORequest, error) {

	// prepare request
	preExeReqs := []*pb.InvokeRequest{}
	preExeReq, err := t.invokeReq()
	if err != nil {
		return nil, err
	}
	if preExeReq != nil {
		preExeReqs = append(preExeReqs, preExeReq)
	}

	preExeRPCReq := &pb.InvokeRPCRequest{
		Bcname: t.ChainName,
		Header: &pb.Header{
			Logid: utils.GenLogId(),
		},
		Requests: preExeReqs,
	}

	// prepare address
	initiator, err := t.genInitiator()
	if err != nil {
		return nil, fmt.Errorf("Get initiator error: %s", err.Error())
	}

	preExeRPCReq.Initiator = initiator
	if !t.IsQuick {
		preExeRPCReq.AuthRequire, err = t.genAuthRequireQuick()
		if err != nil {
			return nil, fmt.Errorf("Get auth require quick error: %s", err.Error())
		}
	} else {
		preExeRPCReq.AuthRequire, err = t.GenAuthRequire(t.MultiAddrs)
		if err != nil {
			return nil, fmt.Errorf("Get auth require error: %s", err.Error())
		}
	}

	// prepare total amount
	extraAmount := int64(t.RootOptions.ComplianceCheck.ComplianceCheckEndorseServiceFee)
	if t.Fee != "" && t.Fee != "0" {
		fee, err := strconv.ParseInt(t.Fee, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("Invalid Fee: %s", t.Fee)
		}
		extraAmount += fee
	}
	if t.Amount != "" {
		amount, err := strconv.ParseInt(t.Amount, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid Amount: %s", t.Amount)
		}
		extraAmount += amount
	}
	preExeRPCReq.AuthRequire = append(preExeRPCReq.AuthRequire, t.RootOptions.ComplianceCheck.ComplianceCheckEndorseServiceAddr)

	// pack
	preSelUTXOReq := &pb.PreExecWithSelectUTXORequest{
		Bcname:      t.ChainName,
		Address:     initiator,
		TotalAmount: extraAmount,
		Request:     preExeRPCReq,
	}
	return preSelUTXOReq, err
}

func (t *CommTrans) GenPreExeWithSelectUtxoRes(ctx context.Context) (
	*pb.PreExecWithSelectUTXOResponse, error) {

	preSelUTXOReq, err := t.preExecWithSelectUTXOReq()
	if err != nil {
		return nil, err
	}

	// preExe
	preExecWithSelectUTXOResponse, err := t.XchainClient.PreExecWithSelectUTXO(ctx, preSelUTXOReq)
	if err != nil {
		return nil, err
	}

	for _, res := range preExecWithSelectUTXOResponse.GetResponse().GetResponses() {
		if res.Status >= contract.StatusErrorThreshold {
			return nil, fmt.Errorf("contract error status:%d message:%s", res.Status, res.Message)
		}
		fmt.Printf("contract response: %s\n", string(res.Body))
	}

	gasUsed := preExecWithSelectUTXOResponse.GetResponse().GetGasUsed()
	fmt.Printf("The gas you cousume is: %v\n", gasUsed)
	if gasUsed > 0 {
		if t.Fee == "" || t.Fee == "0" {
			return nil, errors.New("You need add fee")
		}

		fee, _ := strconv.ParseInt(t.Fee, 10, 64)
		if fee < gasUsed {
			return nil, errors.New("Fee not enough")
		}
		fmt.Printf("The fee you pay is: %v\n", t.Fee)
	} else if t.Fee != "" && t.Fee != "0" && gasUsed <= 0 {
		fmt.Printf("The fee you pay is: %v\n", t.Fee)
	}

	return preExecWithSelectUTXOResponse, nil
}

func (t *CommTrans) GenCompleteTxAndPost(ctx context.Context, preExeResp *pb.PreExecWithSelectUTXOResponse) error {
	complianceCheckTx, err := t.GenComplianceCheckTx(preExeResp.GetUtxoOutput())
	if err != nil {
		fmt.Printf("GenCompleteTxAndPost GenComplianceCheckTx failed, err: %v", err)
		return err
	}
	fmt.Printf("ComplianceCheck txid: %v\n", hex.EncodeToString(complianceCheckTx.Txid))

	tx, err := t.GenRealTx(preExeResp, complianceCheckTx)
	if err != nil {
		fmt.Printf("GenRealTx failed, err: %v", err)
		return err
	}
	endorserSign, err := t.ComplianceCheck(tx, complianceCheckTx)
	if err != nil {
		return err
	}
	tx.AuthRequireSigns = append(tx.AuthRequireSigns, endorserSign)
	tx.Txid, _ = common.MakeTxId(tx)

	txid, err := t.postTx(ctx, tx)
	if err != nil {
		return err
	}
	fmt.Printf("Tx id: %s\n", txid)

	return nil
}

func (t *CommTrans) GenRealTx(response *pb.PreExecWithSelectUTXOResponse,
	complianceCheckTx *pb.Transaction) (*pb.Transaction, error) {
	utxolist := []*pb.Utxo{}
	totalSelected := big.NewInt(0)
	initiator, err := t.genInitiator()
	if err != nil {
		return nil, err
	}
	for index, txOutput := range complianceCheckTx.TxOutputs {
		if string(txOutput.ToAddr) == initiator {
			utxo := &pb.Utxo{
				Amount:    txOutput.Amount,
				ToAddr:    txOutput.ToAddr,
				RefTxid:   complianceCheckTx.Txid,
				RefOffset: int32(index),
			}
			utxolist = append(utxolist, utxo)
			utxoAmount := big.NewInt(0).SetBytes(utxo.Amount)
			totalSelected.Add(totalSelected, utxoAmount)
		}
	}
	utxoOutput := &pb.UtxoOutput{
		UtxoList:      utxolist,
		TotalSelected: totalSelected.String(),
	}

	totalNeed := big.NewInt(0)

	// no need to double-check
	amount, ok := big.NewInt(0).SetString("0", 10)
	if !ok {
		return nil, ErrInvalidAmount
	}
	fee, ok := big.NewInt(0).SetString(t.Fee, 10)
	if !ok {
		return nil, ErrInvalidAmount
	}
	amount.Add(amount, fee)
	totalNeed.Add(totalNeed, amount)
	if t.Amount != "" {
		amount, ok := big.NewInt(0).SetString(t.Amount, 10)
		if !ok {
			return nil, ErrInvalidAmount
		}
		totalNeed.Add(totalNeed, amount)
	}

	selfAmount := totalSelected.Sub(totalSelected, totalNeed)
	txOutputs, err := t.GenerateMultiTxOutputs(selfAmount.String(), t.Fee)
	if err != nil {
		fmt.Printf("GenRealTx GenerateTxOutput failed.")
		return nil, fmt.Errorf("GenRealTx GenerateTxOutput err: %v", err)
	}

	txInputs, err := t.GeneratePureTxInputs(utxoOutput)
	if err != nil {
		fmt.Printf("GenRealTx GenerateTxInput failed.")
		return nil, fmt.Errorf("GenRealTx GenerateTxInput err: %v", err)
	}

	tx := &pb.Transaction{
		Version:   utxo.TxVersion,
		Coinbase:  false,
		Timestamp: time.Now().UnixNano(),
		TxInputs:  txInputs,
		TxOutputs: txOutputs,
		Initiator: initiator,
		Nonce:     utils.GenNonce(),
	}

	if response.Response.GetUtxoInputs() != nil {
		tx.TxInputs = append(tx.TxInputs, response.GetResponse().GetUtxoInputs()...)
	}
	if response.Response.GetUtxoOutputs() != nil {
		tx.TxOutputs = append(tx.TxOutputs, response.GetResponse().GetUtxoOutputs()...)
	}

	desc, _ := t.GetDesc()
	tx.Desc = desc
	tx.TxInputsExt = response.GetResponse().GetInputs()
	tx.TxOutputsExt = response.GetResponse().GetOutputs()
	tx.ContractRequests = response.GetResponse().GetRequests()

	fromAddr, err := readAddress(t.Keys)
	if err != nil {
		return nil, err
	}
	var authRequire string
	if t.From != "" {
		authRequire = t.From + "/" + fromAddr
	} else {
		authRequire = fromAddr
	}
	tx.AuthRequire = append(tx.AuthRequire, authRequire)
	tx.AuthRequire = append(tx.AuthRequire, t.RootOptions.ComplianceCheck.ComplianceCheckEndorseServiceAddr)

	signInfos, err := t.signTxForInitiator(tx)
	if err != nil {
		return nil, err
	}
	tx.InitiatorSigns = signInfos
	tx.AuthRequireSigns = signInfos

	// make txid
	tx.Txid, _ = common.MakeTxId(tx)
	return tx, nil
}

func (t *CommTrans) GenerateMultiTxOutputs(selfAmount string, gasUsed string) ([]*pb.TxOutput, error) {
	selfAddr, err := t.genInitiator()
	if err != nil {
		return nil, err
	}
	feeAmount := gasUsed

	var txOutputs []*pb.TxOutput
	txOutputSelf := new(pb.TxOutput)
	txOutputSelf.ToAddr = []byte(selfAddr)
	realSelfAmount, isSuccess := new(big.Int).SetString(selfAmount, 10)
	if !isSuccess {
		fmt.Printf("selfAmount convert to bigint failed")
		return nil, ErrInvalidAmount
	}
	txOutputSelf.Amount = realSelfAmount.Bytes()
	txOutputs = append(txOutputs, txOutputSelf)
	if feeAmount != "" && feeAmount != "0" {
		realFeeAmount, isSuccess := new(big.Int).SetString(feeAmount, 10)
		if !isSuccess {
			fmt.Printf("feeAmount convert to bigint failed")
			return nil, ErrInvalidAmount
		}
		if realFeeAmount.Cmp(big.NewInt(0)) < 0 {
			return nil, ErrInvalidAmount
		}
		txOutputFee := new(pb.TxOutput)
		txOutputFee.ToAddr = []byte("$")
		txOutputFee.Amount = realFeeAmount.Bytes()
		txOutputs = append(txOutputs, txOutputFee)
	}

	if t.Amount != "" {
		amount, ok := new(big.Int).SetString(t.Amount, 10)
		if !ok {
			return nil, ErrInvalidAmount
		}
		txOutputCountractAmount := new(pb.TxOutput)
		txOutputCountractAmount.ToAddr = []byte(t.ContractName)
		txOutputCountractAmount.Amount = amount.Bytes()
		txOutputs = append(txOutputs, txOutputCountractAmount)
	}

	return txOutputs, nil
}

func (t *CommTrans) GeneratePureTxInputs(utxoOutputs *pb.UtxoOutput) (
	[]*pb.TxInput, error) {
	// gen txInputs
	var txInputs []*pb.TxInput
	for _, utxo := range utxoOutputs.UtxoList {
		txInput := &pb.TxInput{}
		txInput.RefTxid = utxo.RefTxid
		txInput.RefOffset = utxo.RefOffset
		txInput.FromAddr = utxo.ToAddr
		txInput.Amount = utxo.Amount
		txInputs = append(txInputs, txInput)
	}

	return txInputs, nil
}

func (t *CommTrans) ComplianceCheck(tx *pb.Transaction, fee *pb.Transaction) (
	*pb.SignatureInfo, error) {
	txStatus := &pb.TxStatus{
		Bcname: t.ChainName,
		Tx:     tx,
	}

	requestData, err := json.Marshal(txStatus)
	if err != nil {
		fmt.Printf("json encode txStatus failed: %v", err)
		return nil, err
	}

	endorserRequest := &pb.EndorserRequest{
		RequestName: "ComplianceCheck",
		BcName:      t.ChainName,
		Fee:         fee,
		RequestData: requestData,
	}

	//nolint:staticcheck
	conn, err := grpc.Dial(t.RootOptions.EndorseServiceHost, grpc.WithInsecure(), grpc.WithMaxMsgSize(64<<20-1))
	if err != nil {
		fmt.Printf("ComplianceCheck connect EndorseServiceHost err: %v", err)
		return nil, err
	}

	defer conn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 15000*time.Millisecond)
	defer cancel()

	client := pb.NewXendorserClient(conn)
	endorserResponse, err := client.EndorserCall(ctx, endorserRequest)
	if err != nil {
		fmt.Printf("EndorserCall failed and err is: %v", err)
		return nil, fmt.Errorf("EndorserCall error! Response is: %v", err)
	}

	return endorserResponse.GetEndorserSign(), nil
}

func (t *CommTrans) GenComplianceCheckTx(utxoOutput *pb.UtxoOutput) (*pb.Transaction, error) {
	totalNeed := new(big.Int).SetInt64(int64(t.RootOptions.ComplianceCheck.ComplianceCheckEndorseServiceFee))
	txInputs, deltaTxOutput, err := t.GenerateTxInput(utxoOutput, totalNeed)
	if err != nil {
		fmt.Printf("GenerateComplianceTx GenerateTxInput failed.")
		return nil, fmt.Errorf("GenerateComplianceTx GenerateTxInput err: %v", err)
	}

	checkAmount := strconv.Itoa(t.RootOptions.ComplianceCheck.ComplianceCheckEndorseServiceFee)
	txOutputs, err := t.GenerateTxOutput(t.RootOptions.ComplianceCheck.ComplianceCheckEndorseServiceFeeAddr, checkAmount, "0")
	if err != nil {
		fmt.Printf("GenerateComplianceTx GenerateTxOutput failed.")
		return nil, fmt.Errorf("GenerateComplianceTx GenerateTxOutput err: %v", err)
	}
	if deltaTxOutput != nil {
		txOutputs = append(txOutputs, deltaTxOutput)
	}
	// populate fields
	tx := &pb.Transaction{
		Desc:      []byte(""),
		Version:   utxo.TxVersion,
		Coinbase:  false,
		Timestamp: time.Now().UnixNano(),
		TxInputs:  txInputs,
		TxOutputs: txOutputs,
		Nonce:     utils.GenNonce(),
	}
	initiator, err := t.genInitiator()
	if err != nil {
		return nil, err
	}
	tx.Initiator = initiator

	fromAddr, err := readAddress(t.Keys)
	if err != nil {
		return nil, err
	}

	var authRequire string
	if t.From != "" {
		authRequire = t.From + "/" + fromAddr
	} else {
		authRequire = fromAddr
	}
	tx.AuthRequire = append(tx.AuthRequire, authRequire)

	signInfos, err := t.signTxForInitiator(tx)
	if err != nil {
		return nil, err
	}
	tx.InitiatorSigns = signInfos
	tx.AuthRequireSigns = signInfos

	// make Tx ID
	tx.Txid, _ = common.MakeTxId(tx)
	return tx, nil
}

func (t *CommTrans) GenerateTxInput(utxoOutputs *pb.UtxoOutput, totalNeed *big.Int) (
	[]*pb.TxInput, *pb.TxOutput, error) {
	var txInputs []*pb.TxInput
	var txOutput *pb.TxOutput
	for _, utxo := range utxoOutputs.UtxoList {
		txInput := &pb.TxInput{}
		txInput.RefTxid = utxo.RefTxid
		txInput.RefOffset = utxo.RefOffset
		txInput.FromAddr = utxo.ToAddr
		txInput.Amount = utxo.Amount
		txInputs = append(txInputs, txInput)
	}

	utxoTotal, ok := big.NewInt(0).SetString(utxoOutputs.TotalSelected, 10)
	if !ok {
		return nil, nil, fmt.Errorf("GenerateTxInput totalSelected err: %v", ok)
	}

	var fromAddr string
	var err error
	if t.From != "" {
		fromAddr = t.From
	} else {
		fromAddr, err = readAddress(t.Keys)
		if err != nil {
			return nil, nil, err
		}
	}
	// input > output, generate output-input to me
	if utxoTotal.Cmp(totalNeed) > 0 {
		delta := utxoTotal.Sub(utxoTotal, totalNeed)
		txOutput = &pb.TxOutput{
			ToAddr: []byte(fromAddr),
			Amount: delta.Bytes(),
		}
	}

	return txInputs, txOutput, nil
}

func (t *CommTrans) GenerateTxOutput(to, amount, fee string) ([]*pb.TxOutput, error) {
	accounts := []*pb.TxDataAccount{}
	if to != "" {
		account := &pb.TxDataAccount{
			Address:      to,
			Amount:       amount,
			FrozenHeight: 0,
		}
		accounts = append(accounts, account)
	}
	if fee != "0" {
		feeAccount := &pb.TxDataAccount{
			Address: "$",
			Amount:  fee,
		}
		accounts = append(accounts, feeAccount)
	}

	bigZero := big.NewInt(0)
	txOutputs := []*pb.TxOutput{}
	for _, acc := range accounts {
		amount, ok := big.NewInt(0).SetString(acc.Amount, 10)
		if !ok {
			return nil, ErrInvalidAmount
		}
		cmpRes := amount.Cmp(bigZero)
		if cmpRes < 0 {
			return nil, errors.New("Invalid negative number")
		} else if cmpRes == 0 {
			continue
		}
		txOutput := &pb.TxOutput{}
		txOutput.Amount = amount.Bytes()
		txOutput.ToAddr = []byte(acc.Address)
		txOutput.FrozenHeight = acc.FrozenHeight
		txOutputs = append(txOutputs, txOutput)
	}

	return txOutputs, nil
}

func (t *CommTrans) signLockUtxo(utxo *pb.UtxoInput) (pb.SignatureInfo, error) {

	ak := newAK(t.Keys)
	keyPair, err := ak.keyPair()
	if err != nil {
		return pb.SignatureInfo{}, err
	}

	crypto, err := client.CreateCryptoClient(t.CryptoType)
	if err != nil {
		return pb.SignatureInfo{}, errors.New("Create crypto client error")
	}

	lockedUtxoAll := models.NewLockedUtxoAll(utxo.Bcname, utxo.Address)
	return keyPair.SignUtxo(lockedUtxoAll, crypto)
}

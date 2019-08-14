/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package main

import (
	"bufio"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strconv"
	"time"

	"github.com/golang/protobuf/proto"

	crypto_client "github.com/xuperchain/xuperunion/crypto/client"
	"github.com/xuperchain/xuperunion/global"
	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/utxo"
	"github.com/xuperchain/xuperunion/utxo/txhash"
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
}

// GenerateTx generate raw tx
func (c *CommTrans) GenerateTx(ctx context.Context) (*pb.Transaction, error) {
	preExeRPCRes, preExeReqs, err := c.GenPreExeRes(ctx)
	if err != nil {
		return nil, err
	}

	desc, _ := c.GetDesc()

	tx, err := c.GenRawTx(ctx, desc, preExeRPCRes.GetResponse(), preExeReqs)
	return tx, err
}

// GenPreExeRes 得到预执行的结果
func (c *CommTrans) GenPreExeRes(ctx context.Context) (
	*pb.InvokeRPCResponse, []*pb.InvokeRequest, error) {
	preExeReqs := []*pb.InvokeRequest{}
	if c.ModuleName != "" {
		if c.ModuleName == "xkernel" {
			preExeReqs = append(preExeReqs, &pb.InvokeRequest{
				ModuleName: c.ModuleName,
				MethodName: c.MethodName,
				Args:       c.Args,
			})
		} else {
			preExeReqs = append(preExeReqs, &pb.InvokeRequest{
				ModuleName:   c.ModuleName,
				ContractName: c.ContractName,
				MethodName:   c.MethodName,
				Args:         c.Args,
			})
		}
	} else {
		tmpReq, err := c.GetInvokeRequestFromDesc()
		if err != nil {
			return nil, nil, fmt.Errorf("Get pb.InvokeRPCRequest error:%s", err)
		}
		if tmpReq != nil {
			preExeReqs = append(preExeReqs, tmpReq)
		}
	}

	preExeRPCReq := &pb.InvokeRPCRequest{
		Bcname:   c.ChainName,
		Header:   global.GHeader(),
		Requests: preExeReqs,
	}
	initiator, err := c.genInitiator()
	if err != nil {
		return nil, nil, fmt.Errorf("Get initiator error: %s", err.Error())
	}

	preExeRPCReq.Initiator = initiator
	if !c.IsQuick {
		preExeRPCReq.AuthRequire, err = c.genAuthRequireQuick()
		if err != nil {
			return nil, nil, fmt.Errorf("Get auth require quick error: %s", err.Error())
		}
	} else {
		preExeRPCReq.AuthRequire, err = c.genAuthRequire(c.MultiAddrs)
		if err != nil {
			return nil, nil, fmt.Errorf("Get auth require error: %s", err.Error())
		}
	}
	preExeRPCRes, err := c.XchainClient.PreExec(ctx, preExeRPCReq)
	if err != nil {
		return nil, nil, fmt.Errorf("PreExe contract response : %v, logid:%s", err, preExeRPCReq.Header.Logid)
	}
	for _, res := range preExeRPCRes.Response.Response {
		fmt.Printf("contract response: %s\n", string(res))
	}
	return preExeRPCRes, preExeRPCRes.Response.Requests, nil
}

// GetInvokeRequestFromDesc get invokerequest from desc file
func (c *CommTrans) GetInvokeRequestFromDesc() (*pb.InvokeRequest, error) {
	desc, err := c.GetDesc()
	if err != nil {
		return nil, fmt.Errorf("get desc error:%s", err)
	}

	var preExeReq *pb.InvokeRequest
	preExeReq, err = c.ReadPreExeReq(desc)
	if err != nil {
		return nil, err
	}

	return preExeReq, nil
}

// GetDesc 解析desc字段，主要是针对合约
func (c *CommTrans) GetDesc() ([]byte, error) {
	if c.Descfile == "" {
		return []byte("Maybe common transfer transaction"), nil
	}
	return ioutil.ReadFile(c.Descfile)
}

// ReadPreExeReq 从desc中填充出发起合约调用的结构体
func (c *CommTrans) ReadPreExeReq(buf []byte) (*pb.InvokeRequest, error) {
	params := new(invokeRequestWraper)
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
func (c *CommTrans) GenRawTx(ctx context.Context, desc []byte, preExeRes *pb.InvokeResponse,
	preExeReqs []*pb.InvokeRequest) (*pb.Transaction, error) {
	tx := &pb.Transaction{
		Desc:      desc,
		Coinbase:  false,
		Nonce:     global.GenNonce(),
		Timestamp: time.Now().UnixNano(),
		Version:   utxo.TxVersion,
	}

	var gasUsed int64
	if preExeRes != nil {
		gasUsed = preExeRes.GasUsed
		fmt.Printf("The gas you cousume is: %v\n", gasUsed)
	}
	txOutputs, totalNeed, err := c.GenTxOutputs(gasUsed)
	if err != nil {
		return nil, err
	}
	tx.TxOutputs = txOutputs

	txInputs, deltaTxOutput, err := c.GenTxInputs(ctx, totalNeed)
	if err != nil {
		return nil, err
	}

	tx.TxInputs = txInputs
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
	fromAddr, err := c.genInitiator()
	if err != nil {
		return nil, err
	}
	tx.Initiator = fromAddr

	return tx, nil
}

// genInitiator generate initiator of transaction
func (c *CommTrans) genInitiator() (string, error) {
	fromAddr, err := readAddress(c.Keys)
	if err != nil {
		return "", err
	}
	return fromAddr, nil
}

// GenTxOutputs 填充得到transaction的repeated TxOutput tx_outputs
func (c *CommTrans) GenTxOutputs(gasUsed int64) ([]*pb.TxOutput, *big.Int, error) {
	// 组装转账的账户信息
	account := &pb.TxDataAccount{
		Address:      c.To,
		Amount:       c.Amount,
		FrozenHeight: c.FrozenHeight,
	}

	accounts := []*pb.TxDataAccount{}
	if c.To != "" {
		accounts = append(accounts, account)
	}

	// 如果有小费,增加转个小费地址的账户
	// 如果有合约，需要支付gas
	if gasUsed > 0 {
		if c.Fee != "" && c.Fee != "0" {
			fee, _ := strconv.ParseInt(c.Fee, 10, 64)
			if fee < gasUsed {
				return nil, nil, errors.New("Fee not enough")
			}
		} else {
			return nil, nil, errors.New("You need add fee")
		}
		fmt.Printf("The fee you pay is: %v\n", c.Fee)
		accounts = append(accounts, newFeeAccount(c.Fee))
	} else if c.Fee != "" && c.Fee != "0" && gasUsed <= 0 {
		fmt.Printf("The fee you pay is: %v\n", c.Fee)
		accounts = append(accounts, newFeeAccount(c.Fee))
	}

	// 组装txOutputs
	bigZero := big.NewInt(0)
	totalNeed := bigZero
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
func (c *CommTrans) GenTxInputs(ctx context.Context, totalNeed *big.Int) (
	[]*pb.TxInput, *pb.TxOutput, error) {
	var fromAddr string
	var err error
	if c.From != "" {
		fromAddr = c.From
	} else {
		fromAddr, err = readAddress(c.Keys)
		if err != nil {
			return nil, nil, err
		}
	}

	utxoInput := &pb.UtxoInput{
		Bcname:    c.ChainName,
		Address:   fromAddr,
		TotalNeed: totalNeed.String(),
		NeedLock:  false,
	}

	utxoOutputs, err := c.XchainClient.SelectUTXO(ctx, utxoInput)
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
func (c *CommTrans) Transfer(ctx context.Context) error {
	tx, err := c.GenerateTx(ctx)
	if err != nil {
		return err
	}

	return c.SendTx(ctx, tx)
}

// SendTx post tx
func (c *CommTrans) SendTx(ctx context.Context, tx *pb.Transaction) error {
	fromAddr, err := readAddress(c.Keys)
	if err != nil {
		return err
	}

	var authRequire string
	if c.From != "" {
		authRequire = c.From + "/" + fromAddr
	} else {
		authRequire = fromAddr
	}
	tx.AuthRequire = append(tx.AuthRequire, authRequire)

	signInfos, err := c.genInitSign(tx)
	if err != nil {
		return err
	}
	tx.InitiatorSigns = signInfos
	tx.AuthRequireSigns = signInfos

	tx.Txid, err = txhash.MakeTransactionID(tx)
	if err != nil {
		return errors.New("MakeTxDigesthash txid error")
	}
	txid, err := c.postTx(ctx, tx)
	if err != nil {
		return err
	}
	fmt.Printf("Tx id: %s\n", txid)

	return nil
}

func (c *CommTrans) genInitSign(tx *pb.Transaction) ([]*pb.SignatureInfo, error) {
	fromPubkey, err := readPublicKey(c.Keys)
	if err != nil {
		return nil, err
	}

	cryptoClient, err := crypto_client.CreateCryptoClient(c.CryptoType)
	if err != nil {
		return nil, errors.New("Create crypto client error")
	}
	fromScrkey, err := readPrivateKey(c.Keys)
	if err != nil {
		return nil, err
	}
	signTx, err := txhash.ProcessSignTx(cryptoClient, tx, []byte(fromScrkey))
	if err != nil {
		return nil, err
	}

	signInfo := &pb.SignatureInfo{
		PublicKey: fromPubkey,
		Sign:      signTx,
	}

	signInfos := []*pb.SignatureInfo{}
	signInfos = append(signInfos, signInfo)

	return signInfos, nil
}

func (c *CommTrans) genAuthRequireSignsFromPath(tx *pb.Transaction, path string) ([]*pb.SignatureInfo, error) {
	cryptoClient, err := crypto_client.CreateCryptoClient(c.CryptoType)
	if err != nil {
		return nil, errors.New("Create crypto client error")
	}

	authRequireSigns := []*pb.SignatureInfo{}
	if path == "" {
		initPubkey, err := readPublicKey(c.Keys)
		if err != nil {
			return nil, err
		}

		initScrkey, err := readPrivateKey(c.Keys)
		if err != nil {
			return nil, err
		}
		signTx, err := txhash.ProcessSignTx(cryptoClient, tx, []byte(initScrkey))
		if err != nil {
			return nil, err
		}
		signInfo := &pb.SignatureInfo{
			PublicKey: initPubkey,
			Sign:      signTx,
		}
		authRequireSigns = append(authRequireSigns, signInfo)
		return authRequireSigns, nil
	}
	dir, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, fi := range dir {
		if fi.IsDir() {
			sk, err := readPrivateKey(path + "/" + fi.Name())
			if err != nil {
				return nil, err
			}
			pk, err := readPublicKey(path + "/" + fi.Name())
			if err != nil {
				return nil, err
			}
			signTx, err := txhash.ProcessSignTx(cryptoClient, tx, []byte(sk))
			if err != nil {
				return nil, err
			}
			signInfo := &pb.SignatureInfo{
				PublicKey: pk,
				Sign:      signTx,
			}
			authRequireSigns = append(authRequireSigns, signInfo)
		}
	}
	return authRequireSigns, nil
}

func (c *CommTrans) postTx(ctx context.Context, tx *pb.Transaction) (string, error) {
	txStatus := &pb.TxStatus{
		Bcname: c.ChainName,
		Status: pb.TransactionStatus_UNCONFIRM,
		Tx:     tx,
		Header: &pb.Header{
			Logid: global.Glogid(),
		},
		Txid: tx.Txid,
	}

	reply, err := c.XchainClient.PostTx(ctx, txStatus)
	if err != nil {
		return "", err
	}
	if reply.Header.Error != pb.XChainErrorEnum_SUCCESS {
		return "", fmt.Errorf("Failed to post tx:%s, logid:%s", reply.Header.Error.String(), reply.Header.Logid)
	}

	return hex.EncodeToString(txStatus.Txid), nil
}

// GenerateMultisigGenRawTx for mulitisig gen cmd
func (c *CommTrans) GenerateMultisigGenRawTx(ctx context.Context) error {
	tx, err := c.GenerateTx(ctx)
	if err != nil {
		return err
	}

	// 填充需要多重签名的addr
	multiAddrs, err := c.genAuthRequire(c.MultiAddrs)
	if err != nil {
		return err
	}
	tx.AuthRequire = multiAddrs

	return c.genTxFile(tx)
}

func (c *CommTrans) genAuthRequireQuick() ([]string, error) {

	fromAddr, err := readAddress(c.Keys)
	if err != nil {
		return nil, err
	}
	authRequires := []string{}
	if c.From != "" {
		authRequires = append(authRequires, c.From+"/"+fromAddr)
	} else {
		authRequires = append(authRequires, fromAddr)
	}
	return authRequires, nil
}

func (c *CommTrans) genAuthRequire(filename string) ([]string, error) {
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

func (c *CommTrans) genTxFile(tx *pb.Transaction) error {
	data, err := proto.Marshal(tx)
	if err != nil {
		return errors.New("Tx marshal error")
	}
	err = ioutil.WriteFile(c.Output, data, 0755)
	if err != nil {
		return errors.New("WriteFile error")
	}

	if c.IsPrint {
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

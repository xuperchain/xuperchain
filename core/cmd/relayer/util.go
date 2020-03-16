package relayer

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"path/filepath"
	"strconv"
	"time"

	"github.com/xuperchain/xuperchain/core/contract"
	crypto_client "github.com/xuperchain/xuperchain/core/crypto/client"
	"github.com/xuperchain/xuperchain/core/global"
	"github.com/xuperchain/xuperchain/core/pb"
	"github.com/xuperchain/xuperchain/core/utxo"
	"github.com/xuperchain/xuperchain/core/utxo/txhash"
)

// 充得到transaction的repeated TxOutput tx_outputs
func (cmd *DeliverBlockCommand) GenTxOutputs(gasUsed int64) ([]*pb.TxOutput, *big.Int, error) {
	accounts := []*pb.TxDataAccount{}

	// 如果有消费, 增加转个消费地址的账户
	// 如果有合约, 需要支付gas
	if gasUsed > 0 {
		accounts = append(accounts, newFeeAccount(strconv.FormatInt(gasUsed, 10)))
	}
	// 组装txOutputs
	bigZero := big.NewInt(0)
	totalNeed := big.NewInt(0)
	txOutputs := []*pb.TxOutput{}
	for _, acc := range accounts {
		amount, ok := big.NewInt(0).SetString(acc.Amount, 10)
		if !ok {
			return nil, nil, errors.New("invalid amount")
		}
		cmpRes := amount.Cmp(bigZero)
		if cmpRes < 0 {
			return nil, nil, errors.New("negative amount")
		} else if cmpRes == 0 {
			// trim 0 output
			continue
		}
		// 得到总的转账金额
		totalNeed.Add(totalNeed, amount)

		txOutput := &pb.TxOutput{}
		txOutput.Amount = amount.Bytes()
		txOutput.ToAddr = []byte(acc.Address)
		txOutputs = append(txOutputs, txOutput)
	}

	return txOutputs, totalNeed, nil
}

// 填充得到transaction的repeated TxInput tx_inputs,
// 如果输入大于输出，增加一个转给自己(data/keys/)的输入-输出的交易
func (cmd *DeliverBlockCommand) GenTxInputs(totalNeed *big.Int) ([]*pb.TxInput, *pb.TxOutput, error) {
	initiator, err := readAddress(cmd.Cfg.Keys)
	if err != nil {
		panic("read address error")
	}
	utxoInput := &pb.UtxoInput{
		Bcname:    cmd.Cfg.Bcname,
		Address:   initiator,
		TotalNeed: totalNeed.String(),
		NeedLock:  false,
	}

	utxoOutputs, err := cmd.client.SelectUTXO(context.TODO(), utxoInput)
	if err != nil {
		return nil, nil, err
	}
	if utxoOutputs.Header.Error != pb.XChainErrorEnum_SUCCESS {
		return nil, nil, err
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
		return nil, nil, errors.New("select utxo error")
	}

	// 通过selectUTXO选出的作为交易的输入大于输出,
	// 则多出来再生成一笔交易转给自己
	if utxoTotal.Cmp(totalNeed) > 0 {
		delta := utxoTotal.Sub(utxoTotal, totalNeed)
		txOutput = &pb.TxOutput{
			ToAddr: []byte(initiator),
			Amount: delta.Bytes(),
		}
	}

	return txInputs, txOutput, nil
}

func (cmd *DeliverBlockCommand) GenInitSign(tx *pb.Transaction) ([]*pb.SignatureInfo, error) {
	fromPubKey, err := readPublicKey("./data/keys")
	if err != nil {
		return nil, err
	}

	cryptoClient, err := crypto_client.CreateCryptoClient("default")
	if err != nil {
		return nil, errors.New("Create crypto client error")
	}
	fromScrKey, err := readPrivateKey("./data/keys")
	if err != nil {
		return nil, err
	}
	signTx, err := txhash.ProcessSignTx(cryptoClient, tx, []byte(fromScrKey))
	if err != nil {
		return nil, err
	}

	signInfo := &pb.SignatureInfo{
		PublicKey: fromPubKey,
		Sign:      signTx,
	}

	signInfos := []*pb.SignatureInfo{}
	signInfos = append(signInfos, signInfo)

	return signInfos, nil
}

func readAddress(keypath string) (string, error) {
	return readKeys(filepath.Join(keypath, "address"))
}

func readPublicKey(keypath string) (string, error) {
	return readKeys(filepath.Join(keypath, "public.key"))
}

func readPrivateKey(keypath string) (string, error) {
	return readKeys(filepath.Join(keypath, "private.key"))
}

func readKeys(file string) (string, error) {
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	buf = bytes.TrimSpace(buf)
	return string(buf), nil
}

func newFeeAccount(fee string) *pb.TxDataAccount {
	return &pb.TxDataAccount{
		Address: utxo.FeePlaceholder,
		Amount:  fee,
	}
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
	if tx == nil {
		panic("tx should not be nil")
	}
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

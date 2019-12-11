// Copyright (c) 2019. Baidu Inc. All Rights Reserved.

package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/xuperchain/xuperunion/global"
	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/permission/acl"
	"github.com/xuperchain/xuperunion/utxo"
	"github.com/xuperchain/xuperunion/utxo/txhash"
)

// AccountMergeUtxoCommand necessary parameter for merge utxo
type AccountMergeUtxoCommand struct {
	cli *Cli
	cmd *cobra.Command
	// account will be merged
	account string
	// white merge an contract account, it can not be null
	accountPath string
	num         int64
}

// NewAccountMergeUtxoCommand new an instance of merge utxo command
func NewAccountMergeUtxoCommand(cli *Cli) *cobra.Command {
	c := new(AccountMergeUtxoCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "merge ",
		Short: "merge the utxo of an account or address.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return c.mergeUtxo(ctx)
		},
	}
	c.addFlags()
	return c.cmd
}

func (c *AccountMergeUtxoCommand) addFlags() {
	c.cmd.Flags().StringVarP(&c.account, "account", "A", "", "The account/address to be merged (default ./data/keys/address).")
	c.cmd.Flags().StringVarP(&c.accountPath, "accountPath", "P", "", "The account path, which is required for an account.")
	c.cmd.Flags().Int64VarP(&c.num, "num", "N", 10, "final open utxo count")
}

func (c *AccountMergeUtxoCommand) mergeUtxo(ctx context.Context) error {
Again:
	if acl.IsAccount(c.account) == 1 && c.accountPath == "" {
		return errors.New("accountPath can not be null because account is an Account name")
	}

	initAk, _ := readAddress(c.cli.RootOptions.Keys)
	if c.account == "" {
		c.account = initAk
	}

	tx := &pb.Transaction{
		Version:   utxo.TxVersion,
		Coinbase:  false,
		Nonce:     global.GenNonce(),
		Timestamp: time.Now().UnixNano(),
		Initiator: initAk,
	}

	ct := &CommTrans{
		FrozenHeight: 0,
		Version:      utxo.TxVersion,
		From:         c.account,
		Args:         make(map[string][]byte),
		IsQuick:      false,
		ChainName:    c.cli.RootOptions.Name,
		Keys:         c.cli.RootOptions.Keys,
		XchainClient: c.cli.XchainClient(),
		CryptoType:   c.cli.RootOptions.CryptoType,
	}

	txInputs, txOutput, err := ct.GenTxInputsWithMergeUTXO(context.Background())
	tx.TxInputs = txInputs

	txOutputs := []*pb.TxOutput{}
	txOutputs = append(txOutputs, txOutput)
	tx.TxOutputs = txOutputs

	tx.AuthRequire, err = genAuthRequire(c.account, c.accountPath)
	if err != nil {
		return errors.New("genAuthRequire error")
	}

	// preExe
	preExeRPCReq := &pb.InvokeRPCRequest{
		Bcname:      c.cli.RootOptions.Name,
		Requests:    []*pb.InvokeRequest{},
		Header:      global.GHeader(),
		Initiator:   initAk,
		AuthRequire: tx.AuthRequire,
	}
	preExeRes, err := ct.XchainClient.PreExec(context.Background(), preExeRPCReq)
	if err != nil {
		return err
	}
	tx.ContractRequests = preExeRes.GetResponse().GetRequests()
	tx.TxInputsExt = preExeRes.GetResponse().GetInputs()
	tx.TxOutputsExt = preExeRes.GetResponse().GetOutputs()

	tx.InitiatorSigns, err = ct.genInitSign(tx)
	if err != nil {
		return err
	}
	tx.AuthRequireSigns, err = ct.genAuthRequireSignsFromPath(tx, c.accountPath)
	if err != nil {
		return err
	}

	// calculate txid
	tx.Txid, err = txhash.MakeTransactionID(tx)
	if err != nil {
		return err
	}
	txid, err := ct.postTx(context.Background(), tx)
	if err != nil {
		return err
	}
	fmt.Println(txid)
	request := &pb.UtxoRecordDetail{
		Bcname:      c.cli.RootOptions.Name,
		AccountName: c.account,
	}
	response, err := ct.XchainClient.QueryUtxoRecord(ctx, request)
	if err != nil {
		return err
	}
	openUtxoCountStr := response.GetOpenUtxoRecord().GetUtxoCount()
	openUtxoCount, err := strconv.ParseInt(openUtxoCountStr, 10, 64)
	if err != nil {
		return err
	}
	if openUtxoCount > c.num {
		goto Again
	}
	return nil
}

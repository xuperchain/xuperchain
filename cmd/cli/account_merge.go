// Copyright (c) 2019. Baidu Inc. All Rights Reserved.

package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/xuperchain/xuperunion/global"
	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/permission/acl"
	"github.com/xuperchain/xuperunion/utxo"
	"github.com/xuperchain/xuperunion/utxo/txhash"
)

// AccountSplitUtxoCommand split utxo of ak or account
type AccountMergeUtxoCommand struct {
	cli *Cli
	cmd *cobra.Command
	// account will be splited
	account string
	// while spliting a Account, it can not be null
	accountPath string
}

// NewAccountSplitUtxoCommand return
func NewAccountMergeUtxoCommand(cli *Cli) *cobra.Command {
	c := new(AccountMergeUtxoCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "merge ",
		Short: "Merge the utxo of an account or address.",
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
}

func (c *AccountMergeUtxoCommand) mergeUtxo(ctx context.Context) error {
	if acl.IsAccount(c.account) == 0 && c.accountPath == "" {
		return errors.New("accountPath can not be null because account is an Account name")
	}

	initAk, err := readAddress(c.cli.RootOptions.Keys)
	if c.account == "" {
		c.account = initAk
	}

	if acl.IsAccount(c.account) == 1 && c.account != initAk {
		return errors.New("parse account error")
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
	// preExec
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
	fmt.Println(txid)
	return err
}

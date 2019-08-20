// Copyright (c) 2019. Baidu Inc. All Rights Reserved.

package main

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/spf13/cobra"
	"github.com/xuperchain/xuperunion/global"
	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/permission/acl"
	"github.com/xuperchain/xuperunion/utxo"
	"github.com/xuperchain/xuperunion/utxo/txhash"
)

// AccountSplitUtxoCommand split utxo of ak or account
type AccountSplitUtxoCommand struct {
	cli *Cli
	cmd *cobra.Command
	// account will be splited
	account string
	num     int64
	// while spliting a Account, it can not be null
	accountPath string
}

// NewAccountSplitUtxoCommand return
func NewAccountSplitUtxoCommand(cli *Cli) *cobra.Command {
	c := new(AccountSplitUtxoCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "split ",
		Short: "Split the utxo of an account or address.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return c.splitUtxo(ctx)
		},
	}
	c.addFlags()
	return c.cmd
}

func (c *AccountSplitUtxoCommand) addFlags() {
	c.cmd.Flags().StringVarP(&c.account, "account", "A", "", "The account/address to be splited (default ./data/keys/address).")
	c.cmd.Flags().Int64VarP(&c.num, "num", "N", 1, "The number to split.")
	c.cmd.Flags().StringVarP(&c.accountPath, "accountPath", "P", "", "The account path, which is required for an account.")
}

func (c *AccountSplitUtxoCommand) splitUtxo(ctx context.Context) error {
	if c.num <= 0 {
		return errors.New("illegal splitutxo num, num > 0 required")
	}
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

	amount, err := c.getBalanceHelper()
	if err != nil {
		return err
	}
	ct := &CommTrans{
		Amount:       amount,
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

	totalNeed, ok := big.NewInt(0).SetString(amount, 10)
	if !ok {
		return errors.New("get totalNeed error")
	}

	txInputs, txOutput, err := ct.GenTxInputs(context.Background(), totalNeed)
	tx.TxInputs = txInputs

	txOutputs, err := c.genSplitOutputs(totalNeed)
	if err != nil {
		return err
	}
	if txOutput != nil {
		txOutputs = append(txOutputs, txOutput)
	}
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

func (c *AccountSplitUtxoCommand) getBalanceHelper() (string, error) {
	as := &pb.AddressStatus{}
	as.Address = c.account
	var tokens []*pb.TokenDetail
	token := pb.TokenDetail{Bcname: c.cli.RootOptions.Name}
	tokens = append(tokens, &token)
	as.Bcs = tokens
	r, err := c.cli.XchainClient().GetBalance(context.Background(), as)
	if err != nil {
		return "0", err
	}
	return r.Bcs[0].Balance, nil
}

func (c *AccountSplitUtxoCommand) genSplitOutputs(toralNeed *big.Int) ([]*pb.TxOutput, error) {
	txOutputs := []*pb.TxOutput{}
	amount := big.NewInt(0)
	rest := toralNeed
	if big.NewInt(c.num).Cmp(rest) == 1 {
		return nil, errors.New("illegal splitutxo, splitutxo <= BALANCE required")
	}
	amount.Div(rest, big.NewInt(c.num))
	output := pb.TxOutput{}
	output.Amount = amount.Bytes()
	output.ToAddr = []byte(c.account)
	for i := int64(1); i < c.num && rest.Cmp(amount) == 1; i++ {
		tmpOutput := output
		txOutputs = append(txOutputs, &tmpOutput)
		rest.Sub(rest, amount)
	}
	output.Amount = rest.Bytes()
	txOutputs = append(txOutputs, &output)
	return txOutputs, nil
}

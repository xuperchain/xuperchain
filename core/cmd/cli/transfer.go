/*
 * Copyright (c) 2019, Baidu.com, Inc. All Rights Reserved.
 */

package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/utxo"
)

var (
	// ErrInvalidAmount error
	ErrInvalidAmount = errors.New("Invalid amount number")
	// ErrNegativeAmount error
	ErrNegativeAmount = errors.New("Amount in transaction can not be negative number")
	// ErrPutTx error
	ErrPutTx = errors.New("Put tx error")
	// ErrSelectUtxo error
	ErrSelectUtxo = errors.New("Select utxo error")
)

// TransferOptions transfer cmd options
type TransferOptions struct {
	BlockchainName string
	KeyPath        string
	CryptoType     string
	To             string
	Amount         string
	Fee            string
	Desc           []byte
	FrozenHeight   int64
	Version        int32
	// 支持账户转账
	From        string
	AccountPath string
}

// TransferCommand transfer cmd
type TransferCommand struct {
	cli *Cli
	cmd *cobra.Command

	to           string
	amount       string
	descfile     string
	fee          string
	frozenHeight int64
	version      int32
	// 支持账户转账
	from        string
	accountPath string
}

// NewTransferCommand new transfer cmd
func NewTransferCommand(cli *Cli) *cobra.Command {
	t := new(TransferCommand)
	t.cli = cli
	t.cmd = &cobra.Command{
		Use:   "transfer",
		Short: "Operate transfer trasaction, transfer tokens between accounts or aks",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return t.transfer(ctx)
		},
	}
	t.addFlags()
	return t.cmd
}

func (t *TransferCommand) addFlags() {
	t.cmd.Flags().StringVar(&t.to, "to", "", "common transfer transaction to whom")
	t.cmd.Flags().StringVar(&t.amount, "amount", "0", "transfer tokens")
	t.cmd.Flags().StringVar(&t.descfile, "desc", "", "desc file of tx, eg. contract or tdpos consensus")
	t.cmd.Flags().StringVar(&t.fee, "fee", "0", "fee of one tx")
	t.cmd.Flags().Int64Var(&t.frozenHeight, "frozen", 0, "frozen height of one tx")
	t.cmd.Flags().Int32Var(&t.version, "txversion", utxo.TxVersion, "tx version")
	t.cmd.Flags().StringVar(&t.from, "from", "", "account name")
	t.cmd.Flags().StringVar(&t.accountPath, "accountPath", "", "key path of account")
}

func readKeys(file string) (string, error) {
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	buf = bytes.TrimSpace(buf)
	return string(buf), nil
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

type invokeRequestWraper struct {
	pb.InvokeRequest
	// 取巧的手段来shadow pb.InvokeRequest里面的Args字段
	Args map[string]string `json:"args,omitempty"`
}

func newFeeAccount(fee string) *pb.TxDataAccount {
	return &pb.TxDataAccount{
		Address: utxo.FeePlaceholder,
		Amount:  fee,
	}
}

func (t *TransferCommand) getDesc() ([]byte, error) {
	if t.descfile == "" {
		return []byte("transfer from console"), nil
	}
	return ioutil.ReadFile(t.descfile)
}

func (t *TransferCommand) transfer(ctx context.Context) error {
	desc, err := t.getDesc()
	if err != nil {
		return err
	}
	opt := TransferOptions{
		BlockchainName: t.cli.RootOptions.Name,
		KeyPath:        t.cli.RootOptions.Keys,
		CryptoType:     t.cli.RootOptions.CryptoType,
		To:             t.to,
		Amount:         t.amount,
		Fee:            t.fee,
		Desc:           desc,
		FrozenHeight:   t.frozenHeight,
		Version:        t.version,
		From:           t.from,
		AccountPath:    t.accountPath,
	}

	txid, err := t.cli.Transfer(ctx, &opt)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", txid)
	return nil
}

func init() {
	AddCommand(NewTransferCommand)
}

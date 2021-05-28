/*
 * Copyright (c) 2021. Baidu Inc. All Rights Reserved.
 */

package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/xuperchain/xupercore/bcs/ledger/xledger/state/utxo"
)

// GovernTransferCommand transfer govern token struct
type GovernTransferCommand struct {
	cli *Cli
	cmd *cobra.Command

	to     string
	amount string
	fee    string
}

// NewGovernTransferCommand new transfer govern token cmd
func NewGovernTransferCommand(cli *Cli) *cobra.Command {
	t := new(GovernTransferCommand)
	t.cli = cli
	t.cmd = &cobra.Command{
		Use:   "transfer",
		Short: "Transfer govern token.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return t.transferGovernToken(ctx)
		},
	}
	t.addFlags()

	return t.cmd
}

func (c *GovernTransferCommand) addFlags() {
	c.cmd.Flags().StringVar(&c.to, "to", "", "govern token receiver.")
	c.cmd.Flags().StringVar(&c.amount, "amount", "0", "govern token amount.")
	c.cmd.Flags().StringVar(&c.fee, "fee", "0", "The fee to transfer govern token.")
}

func (c *GovernTransferCommand) transferGovernToken(ctx context.Context) error {
	ct := &CommTrans{
		Amount:       "0",
		Fee:          c.fee,
		FrozenHeight: 0,
		Version:      utxo.TxVersion,

		MethodName: "Transfer",
		Args:       make(map[string][]byte),

		IsQuick: false,

		ChainName:    c.cli.RootOptions.Name,
		Keys:         c.cli.RootOptions.Keys,
		XchainClient: c.cli.XchainClient(),
		CryptoType:   c.cli.RootOptions.Crypto,
		RootOptions:  c.cli.RootOptions,
	}

	var err error
	ct.To, err = readAddress(ct.Keys)
	if err != nil {
		return err
	}

	if c.to != "" {
		ct.ModuleName = "xkernel"
		ct.ContractName = "$govern_token"
		ct.Args["to"] = []byte(c.to)
		ct.Args["amount"] = []byte(c.amount)
	}

	err = ct.Transfer(ctx)
	if err != nil {
		return err
	}

	return nil
}

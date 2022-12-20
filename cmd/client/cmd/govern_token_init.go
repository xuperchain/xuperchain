/*
 * Copyright (c) 2021. Baidu Inc. All Rights Reserved.
 */

package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/xuperchain/xupercore/bcs/ledger/xledger/state/utxo"
)

// GovernInitCommand transfer govern token struct
type GovernInitCommand struct {
	cli *Cli
	cmd *cobra.Command

	fee string
}

// NewGovernInitCommand new transfer govern token cmd
func NewGovernInitCommand(cli *Cli) *cobra.Command {
	t := new(GovernInitCommand)
	t.cli = cli
	t.cmd = &cobra.Command{
		Use:   "init",
		Short: "Init govern token.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return t.initGovernToken(ctx)
		},
	}
	t.addFlags()

	return t.cmd
}

func (c *GovernInitCommand) addFlags() {
	c.cmd.Flags().StringVar(&c.fee, "fee", "0", "The fee to initialize govern token.")
}

func (c *GovernInitCommand) initGovernToken(ctx context.Context) error {
	ct := &CommTrans{
		Amount:       "0",
		Fee:          c.fee,
		FrozenHeight: 0,
		Version:      utxo.TxVersion,

		MethodName: "Init",
		Args:       make(map[string][]byte),

		IsQuick: false,

		ChainName:    c.cli.RootOptions.Name,
		Keys:         c.cli.RootOptions.Keys,
		XchainClient: c.cli.XchainClient(),
		CryptoType:   c.cli.RootOptions.Crypto,
	}

	var err error
	ct.To, err = readAddress(ct.Keys)
	if err != nil {
		return err
	}

	ct.ModuleName = "xkernel"
	ct.ContractName = "$govern_token"

	err = ct.Transfer(ctx)
	if err != nil {
		return err
	}

	return nil
}

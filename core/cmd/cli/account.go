/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package main

import "github.com/spf13/cobra"

// AccountCommand account cmd entrance
type AccountCommand struct {
	cli *Cli
	cmd *cobra.Command
}

// NewAccountCommand new account cmd
func NewAccountCommand(cli *Cli) *cobra.Command {
	c := new(AccountCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "account",
		Short: "Operate an account or address: balance|new|newkeys|split|merge|list-utxo.",
	}
	c.cmd.AddCommand(NewAccountBalanceCommand(cli))
	c.cmd.AddCommand(NewAccountNewkeysCommand(cli))
	c.cmd.AddCommand(NewAccountNewCommand(cli))
	c.cmd.AddCommand(NewAccountContractsCommand(cli))
	c.cmd.AddCommand(NewAccountQueryCommand(cli))
	return c.cmd
}

func init() {
	AddCommand(NewAccountCommand)
}

/*
 * Copyright (c) 2021. Baidu Inc. All Rights Reserved.
 */

package cmd

import (
	"github.com/spf13/cobra"
)

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
		Short: "Operate an account or address: balance|new|newkeys|contracts|restore|decrypt.",
	}
	c.cmd.AddCommand(NewAccountBalanceCommand(cli))
	c.cmd.AddCommand(NewAccountNewkeysCommand(cli))
	c.cmd.AddCommand(NewAccountNewCommand(cli))
	c.cmd.AddCommand(NewAccountContractsCommand(cli))
	c.cmd.AddCommand(NewAccountQueryCommand(cli))
	c.cmd.AddCommand(NewAccountRestoreCommand(cli))
	c.cmd.AddCommand(NewAccountDecryptCommand(cli))
	return c.cmd
}

func init() {
	AddCommand(NewAccountCommand)
}

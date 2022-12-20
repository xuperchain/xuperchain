/*
 * Copyright (c) 2021. Baidu Inc. All Rights Reserved.
 */

package cmd

import (
	"github.com/spf13/cobra"
)

// NewAccountCommand new account cmd
func NewAccountCommand(cli *Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "Operate an account or address: balance|new|newkeys|contracts|restore|decrypt.",
	}
	cmd.AddCommand(NewAccountBalanceCommand(cli))
	cmd.AddCommand(NewAccountNewkeysCommand(cli))
	cmd.AddCommand(NewAccountNewCommand(cli))
	cmd.AddCommand(NewAccountContractsCommand(cli))
	cmd.AddCommand(NewAccountQueryCommand(cli))
	cmd.AddCommand(NewAccountRestoreCommand(cli))
	cmd.AddCommand(NewAccountDecryptCommand(cli))
	return cmd
}

func init() {
	AddCommand(NewAccountCommand)
}

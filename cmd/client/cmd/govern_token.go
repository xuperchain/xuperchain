/*
 * Copyright (c) 2021. Baidu Inc. All Rights Reserved.
 */

package cmd

import (
	"github.com/spf13/cobra"
)

// NewGovernTokenCommand new govern token cmd
func NewGovernTokenCommand(cli *Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "governToken",
		Short: "governToken: init|transfer|query.",
	}
	cmd.AddCommand(NewGovernInitCommand(cli))
	cmd.AddCommand(NewGovernTransferCommand(cli))
	cmd.AddCommand(NewGovernTokenQueryCommand(cli))
	return cmd
}

func init() {
	AddCommand(NewGovernTokenCommand)
}

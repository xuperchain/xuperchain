/*
 * Copyright (c) 2021, Baidu.com, Inc. All Rights Reserved.
 */

package cmd

import (
	"github.com/spf13/cobra"
)

// NewTxCommand new tx cmd
func NewTxCommand(cli *Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tx",
		Short: "Operate tx command, query",
	}
	cmd.AddCommand(NewTxQueryCommand(cli))
	return cmd
}

func init() {
	AddCommand(NewTxCommand)
}

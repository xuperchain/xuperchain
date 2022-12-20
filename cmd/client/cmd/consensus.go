/*
 * Copyright (c) 2021. Baidu Inc. All Rights Reserved.
 */

package cmd

import (
	"github.com/spf13/cobra"
)

// NewConsensusCommand new consensus cmd
func NewConsensusCommand(cli *Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "consensus",
		Short: "Consensus module: status|invoke.",
	}
	cmd.AddCommand(NewConsensusInvokeCommand(cli))
	cmd.AddCommand(NewConsensusStatusCommand(cli))
	return cmd
}

func init() {
	AddCommand(NewConsensusCommand)
}

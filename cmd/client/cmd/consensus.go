/*
 * Copyright (c) 2021. Baidu Inc. All Rights Reserved.
 */

package cmd

import (
	"github.com/spf13/cobra"
)

// AccountCommand account cmd entrance
type ConsensusCommand struct {
	cli *Cli
	cmd *cobra.Command
}

// NewAccountCommand new account cmd
func NewConsensusCommand(cli *Cli) *cobra.Command {
	c := new(AccountCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "consensus",
		Short: "Consensus module: status|invoke.",
	}
	c.cmd.AddCommand(NewConsensusInvokeCommand(cli))
	c.cmd.AddCommand(NewConsensusStatusCommand(cli))
	return c.cmd
}

func init() {
	AddCommand(NewConsensusCommand)
}

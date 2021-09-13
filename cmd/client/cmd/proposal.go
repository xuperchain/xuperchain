/*
 * Copyright (c) 2021. Baidu Inc. All Rights Reserved.
 */

package cmd

import (
	"github.com/spf13/cobra"
)

// ProposalCommand proposal cmd entrance
type ProposalCommand struct {
	cli *Cli
	cmd *cobra.Command
}

// NewProposalCommand new proposal cmd
func NewProposalCommand(cli *Cli) *cobra.Command {
	c := new(ProposalCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "proposal",
		Short: "proposal: propose|vote|thaw|query.",
	}
	c.cmd.AddCommand(NewProposalProposeCommand(cli))
	c.cmd.AddCommand(NewProposalQueryCommand(cli))
	c.cmd.AddCommand(NewProposalVoteCommand(cli))
	c.cmd.AddCommand(NewProposalThawCommand(cli))
	return c.cmd
}

func init() {
	AddCommand(NewProposalCommand)
}

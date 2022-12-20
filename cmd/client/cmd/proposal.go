/*
 * Copyright (c) 2021. Baidu Inc. All Rights Reserved.
 */

package cmd

import (
	"github.com/spf13/cobra"
)

// NewProposalCommand new proposal cmd
func NewProposalCommand(cli *Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proposal",
		Short: "proposal: propose|vote|thaw|query.",
	}
	cmd.AddCommand(NewProposalProposeCommand(cli))
	cmd.AddCommand(NewProposalQueryCommand(cli))
	cmd.AddCommand(NewProposalVoteCommand(cli))
	cmd.AddCommand(NewProposalThawCommand(cli))
	return cmd
}

func init() {
	AddCommand(NewProposalCommand)
}

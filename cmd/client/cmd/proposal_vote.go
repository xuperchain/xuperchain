/*
 * Copyright (c) 2021. Baidu Inc. All Rights Reserved.
 */

package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/xuperchain/xupercore/bcs/ledger/xledger/state/utxo"
)

// ProposalVoteCommand vote a proposal struct
type ProposalVoteCommand struct {
	cli *Cli
	cmd *cobra.Command

	proposalID string
	amount     string
	fee        string
}

// NewProposalVoteCommand vote a proposal cmd
func NewProposalVoteCommand(cli *Cli) *cobra.Command {
	t := new(ProposalVoteCommand)
	t.cli = cli
	t.cmd = &cobra.Command{
		Use:   "vote",
		Short: "Vote a proposal.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return t.voteProposal(ctx)
		},
	}
	t.addFlags()

	return t.cmd
}

func (c *ProposalVoteCommand) addFlags() {
	c.cmd.Flags().StringVarP(&c.proposalID, "pid", "p", "", "proposal id.")
	c.cmd.Flags().StringVar(&c.amount, "amount", "", "amount.")
	c.cmd.Flags().StringVar(&c.fee, "fee", "0", "The fee to vote a proposal.")
}

func (c *ProposalVoteCommand) voteProposal(ctx context.Context) error {
	ct := &CommTrans{
		Amount:       "0",
		Fee:          c.fee,
		FrozenHeight: 0,
		Version:      utxo.TxVersion,

		MethodName: "Vote",
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

	if c.proposalID == "" || c.amount == "" {
		return fmt.Errorf("proposal id or amount is nil")
	}

	ct.ModuleName = "xkernel"
	ct.ContractName = "$proposal"
	ct.Args["proposal_id"] = []byte(c.proposalID)
	ct.Args["amount"] = []byte(c.amount)

	err = ct.Transfer(ctx)
	if err != nil {
		return err
	}

	return nil
}

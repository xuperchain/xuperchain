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

// ProposalThawCommand thaw a proposal struct
type ProposalThawCommand struct {
	cli *Cli
	cmd *cobra.Command

	proposalID string
	amount     string
	fee        string
}

// NewProposalThawCommand vote a proposal cmd
func NewProposalThawCommand(cli *Cli) *cobra.Command {
	t := new(ProposalThawCommand)
	t.cli = cli
	t.cmd = &cobra.Command{
		Use:   "thaw",
		Short: "Thaw a proposal.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return t.thawProposal(ctx)
		},
	}
	t.addFlags()

	return t.cmd
}

func (c *ProposalThawCommand) addFlags() {
	c.cmd.Flags().StringVarP(&c.proposalID, "pid", "p", "", "proposal id.")
	c.cmd.Flags().StringVar(&c.fee, "fee", "0", "The fee to thaw a proposal.")
}

func (c *ProposalThawCommand) thawProposal(ctx context.Context) error {
	ct := &CommTrans{
		Amount:       "0",
		Fee:          c.fee,
		FrozenHeight: 0,
		Version:      utxo.TxVersion,

		MethodName: "Thaw",
		Args:       make(map[string][]byte),

		IsQuick: false,

		ChainName:    c.cli.RootOptions.Name,
		Keys:         c.cli.RootOptions.Keys,
		XchainClient: c.cli.XchainClient(),
		CryptoType:   c.cli.RootOptions.Crypto,
		RootOptions:  c.cli.RootOptions,
	}

	var err error
	ct.To, err = readAddress(ct.Keys)
	if err != nil {
		return err
	}

	if c.proposalID == "" {
		return fmt.Errorf("proposal id or amount is nil")
	}

	ct.ModuleName = "xkernel"
	ct.ContractName = "$proposal"
	ct.Args["proposal_id"] = []byte(c.proposalID)

	err = ct.Transfer(ctx)
	if err != nil {
		return err
	}

	return nil
}

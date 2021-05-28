/*
 * Copyright (c) 2021. Baidu Inc. All Rights Reserved.
 */

package cmd

import (
	"context"
	"io/ioutil"

	"github.com/spf13/cobra"

	"github.com/xuperchain/xupercore/bcs/ledger/xledger/state/utxo"
)

// ProposalProposeCommand proposal a proposal struct
type ProposalProposeCommand struct {
	cli *Cli
	cmd *cobra.Command

	proposal string
	fee      string
}

// NewProposalProposeCommand propose a proposal cmd
func NewProposalProposeCommand(cli *Cli) *cobra.Command {
	t := new(ProposalProposeCommand)
	t.cli = cli
	t.cmd = &cobra.Command{
		Use:   "propose",
		Short: "Propose a proposal.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return t.proposeProposal(ctx)
		},
	}
	t.addFlags()

	return t.cmd
}

func (c *ProposalProposeCommand) addFlags() {
	c.cmd.Flags().StringVarP(&c.proposal, "proposal", "p", "", "proposal.")
	c.cmd.Flags().StringVar(&c.fee, "fee", "0", "The fee to propose a proposal.")
}

func (c *ProposalProposeCommand) proposeProposal(ctx context.Context) error {
	ct := &CommTrans{
		Amount:       "0",
		Fee:          c.fee,
		FrozenHeight: 0,
		Version:      utxo.TxVersion,

		MethodName: "Propose",
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

	proposal, err := c.getProposal()
	if err != nil {
		return err
	}

	ct.ModuleName = "xkernel"
	ct.ContractName = "$proposal"
	ct.Args["proposal"] = proposal

	err = ct.Transfer(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (c *ProposalProposeCommand) getProposal() ([]byte, error) {
	if c.proposal == "" {
		return []byte("no proposal"), nil
	}
	return ioutil.ReadFile(c.proposal)
}

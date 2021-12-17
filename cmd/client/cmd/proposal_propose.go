/*
 * Copyright (c) 2021. Baidu Inc. All Rights Reserved.
 */

package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/big"

	"github.com/spf13/cobra"

	"github.com/xuperchain/xuperchain/service/pb"
	"github.com/xuperchain/xupercore/bcs/ledger/xledger/state/utxo"
)

// ProposalProposeCommand proposal a proposal struct
type ProposalProposeCommand struct {
	cli *Cli
	cmd *cobra.Command

	proposal string
	fee      string
}

type proposalArgs struct {
	StopVotingHeight string `json:"stop_vote_height"`
}
type proposalData struct {
	Args proposalArgs `json:"args"`
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

	if err := c.validateProposal(proposal); err != nil {
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

func (c *ProposalProposeCommand) validateProposal(data []byte) error {
	pData := &proposalData{}
	if err := json.Unmarshal(data, pData); err != nil {
		return err
	}
	in := &pb.CommonIn{}

	status, err := c.cli.XchainClient().GetSystemStatus(context.Background(), in)
	if err != nil {
		return err
	}

	stopVotingHeight, ok := big.NewInt(0).SetString(pData.Args.StopVotingHeight, 10)
	if !ok {
		return errors.New("invalid stop_voting_height")
	}

	if big.NewInt(status.SystemsStatus.BcsStatus[0].Meta.TrunkHeight).Cmp(stopVotingHeight) >= 0 {
		return errors.New("stop voting height must be larger than current trunk height")
	}
	return nil
}

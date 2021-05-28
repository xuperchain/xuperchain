/*
 * Copyright (c) 2021. Baidu Inc. All Rights Reserved.
 */

package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/xuperchain/xupercore/bcs/ledger/xledger/state/utxo"
)

// VoteCommand vote cmd
type VoteCommand struct {
	cli *Cli
	cmd *cobra.Command

	frozenHeight int64
	amount       string
}

// NewVoteCommand new vote
func NewVoteCommand(cli *Cli) *cobra.Command {
	c := new(VoteCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "vote [options] txid",
		Short: "Operate vote command",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return c.vote(ctx, args[0])
		},
	}
	c.addFlags()
	return c.cmd
}

func (c *VoteCommand) addFlags() {
	c.cmd.Flags().Int64Var(&c.frozenHeight, "frozen", 0, "frozen height for your used tokens")
	c.cmd.Flags().StringVar(&c.amount, "amount", "0", "amount of tokens")
}

func (c *VoteCommand) vote(ctx context.Context, txid string) error {
	contractDesc := ContractDesc{
		Module: "proposal",
		Method: "Vote",
		Args: map[string]interface{}{
			"txid": txid,
		},
	}
	myaddress, err := readAddress(c.cli.RootOptions.Keys)
	if err != nil {
		return err
	}
	desc, _ := json.Marshal(contractDesc)
	opt := TransferOptions{
		BlockchainName: c.cli.RootOptions.Name,
		KeyPath:        c.cli.RootOptions.Keys,
		CryptoType:     c.cli.RootOptions.Crypto,
		To:             myaddress,
		Amount:         c.amount,
		Desc:           desc,
		FrozenHeight:   c.frozenHeight,
		Version:        utxo.TxVersion,
	}
	newtxid, err := c.cli.Transfer(ctx, &opt)
	if err != nil {
		return err
	}
	fmt.Println(newtxid)
	return nil
}

func init() {
	AddCommand(NewVoteCommand)
}

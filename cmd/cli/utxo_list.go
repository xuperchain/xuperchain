/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/xuperchain/xuperunion/pb"
)

// ListUtxoCommand necessary parmeters for query utxo records
type ListUtxoCommand struct {
	cli         *Cli
	cmd         *cobra.Command
	addr        string
	utxoItemNum int64
}

// NewListUtxoCommand an entry to query utxo records
func NewListUtxoCommand(cli *Cli) *cobra.Command {
	c := new(ListUtxoCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "list",
		Short: "Get utxo records info of an user.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return c.queryUtxoRecords(ctx)
		},
	}
	c.addFlags()
	return c.cmd
}

func (c *ListUtxoCommand) addFlags() {
	c.cmd.Flags().StringVarP(&c.addr, "address", "A", "", "address")
	c.cmd.Flags().Int64VarP(&c.utxoItemNum, "num", "N", 1, "utxo items to be displayed")
}

func (c *ListUtxoCommand) queryUtxoRecords(ctx context.Context) error {
	client := c.cli.XchainClient()
	if c.addr == "" {
		c.addr, _ = readAddress(c.cli.RootOptions.Keys)
	}
	request := &pb.UtxoRecordDetail{
		Bcname:       c.cli.RootOptions.Name,
		AccountName:  c.addr,
		DisplayCount: c.utxoItemNum,
	}
	response, err := client.QueryUtxoRecord(ctx, request)
	if err != nil {
		return err
	}
	output, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))

	return nil
}

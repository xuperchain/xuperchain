/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/xuperchain/xuperchain/core/pb"
)

// QueryVotedRecordsCommand query votedrecords cmd
type QueryVotedRecordsCommand struct {
	cli  *Cli
	cmd  *cobra.Command
	addr string
}

// NewQueryVotedRecordsCommand new query votedrecords cmd
func NewQueryVotedRecordsCommand(cli *Cli) *cobra.Command {
	c := new(QueryVotedRecordsCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "query-voted-records",
		Short: "Get all records who voted for an user",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return c.queryVotedRecords(ctx)
		},
	}
	c.addFlags()
	return c.cmd
}

func (c *QueryVotedRecordsCommand) addFlags() {
	c.cmd.Flags().StringVarP(&c.addr, "addr", "a", "", "user address")
}

func (c *QueryVotedRecordsCommand) queryVotedRecords(ctx context.Context) error {
	client := c.cli.XchainClient()
	request := &pb.DposVotedRecordsRequest{
		Bcname:  c.cli.RootOptions.Name,
		Address: c.addr,
	}
	response, err := client.DposVotedRecords(ctx, request)
	if err != nil {
		fmt.Println(err)
		return err
	}

	output, err := json.MarshalIndent(response.VotedTxidRecords, "", "  ")
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(string(output))
	return nil
}

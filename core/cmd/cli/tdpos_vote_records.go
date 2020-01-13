/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/xuperchain/xuperchain/core/pb"

	"github.com/spf13/cobra"
)

// QueryVoteRecordsCommand query vote records
type QueryVoteRecordsCommand struct {
	cli  *Cli
	cmd  *cobra.Command
	addr string
}

// NewQueryVoteRecordsCommand new query vote records
func NewQueryVoteRecordsCommand(cli *Cli) *cobra.Command {
	c := new(QueryVoteRecordsCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "query-vote-records",
		Short: "Get all vote records voted by an user.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return c.queryVoteRecords(ctx)
		},
	}
	c.addFlags()
	return c.cmd
}

func (c *QueryVoteRecordsCommand) addFlags() {
	c.cmd.Flags().StringVarP(&c.addr, "addr", "a", "", "user address")
}

func (c *QueryVoteRecordsCommand) queryVoteRecords(ctx context.Context) error {
	client := c.cli.XchainClient()
	request := &pb.DposVoteRecordsRequest{
		Bcname:  c.cli.RootOptions.Name,
		Address: c.addr,
	}
	response, err := client.DposVoteRecords(ctx, request)
	if err != nil {
		fmt.Println(err)
		return err
	}

	output, err := json.MarshalIndent(response.VoteTxidRecords, "", "  ")
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(string(output))
	return nil
}

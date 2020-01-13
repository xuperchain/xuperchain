/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package main

import (
	"context"
	"fmt"
	"github.com/xuperchain/xuperchain/core/pb"

	"github.com/spf13/cobra"
)

// QueryNomineeRecordsCommand query nominee records
type QueryNomineeRecordsCommand struct {
	cli  *Cli
	cmd  *cobra.Command
	addr string
}

// NewQueryNomineeRecordsCommand new query nominee records
func NewQueryNomineeRecordsCommand(cli *Cli) *cobra.Command {
	c := new(QueryNomineeRecordsCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "query-nominee-record",
		Short: "Get records who nominated user as a candidate",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return c.queryNomineeRecord(ctx)
		},
	}
	c.addFlags()
	return c.cmd
}

func (c *QueryNomineeRecordsCommand) addFlags() {
	c.cmd.Flags().StringVarP(&c.addr, "addr", "a", "", "user address")
}

func (c *QueryNomineeRecordsCommand) queryNomineeRecord(ctx context.Context) error {
	client := c.cli.XchainClient()
	request := &pb.DposNomineeRecordsRequest{
		Bcname:  c.cli.RootOptions.Name,
		Address: c.addr,
	}
	response, err := client.DposNomineeRecords(ctx, request)
	if err != nil {
		return err
	}

	fmt.Printf("nominee:%s\n", c.addr)
	fmt.Printf("txid:%s\n", response.Txid)
	return nil
}

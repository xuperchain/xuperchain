// Copyright (c) 2019. Baidu Inc. All Rights Reserved.
package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xuperchain/xuperchain/core/pb"
)

// QueryStatusCommand query vote records
type QueryStatusCommand struct {
	cli *Cli
	cmd *cobra.Command
}

// NewQueryStatusCommand new query vote records
func NewQueryStatusCommand(cli *Cli) *cobra.Command {
	c := new(QueryStatusCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "status",
		Short: "Get status of tdpos consensus.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return c.queryConsensusStatus(ctx)
		},
	}
	return c.cmd
}

func (c *QueryStatusCommand) queryConsensusStatus(ctx context.Context) error {
	cli := c.cli.XchainClient()
	request := &pb.DposStatusRequest{
		Bcname: c.cli.RootOptions.Name,
	}
	response, err := cli.DposStatus(ctx, request)
	if err != nil {
		return err
	}
	output, err := json.MarshalIndent(response.GetStatus(), "", "  ")
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(string(output))
	return nil
}

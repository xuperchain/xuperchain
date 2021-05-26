/*
 * Copyright (c) 2021. Baidu Inc. All Rights Reserved.
 */

package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/xuperchain/xuperchain/service/xchainpb/pb"
)

// QueryCandidatesCommand query candidates cmd
type QueryCandidatesCommand struct {
	cli *Cli
	cmd *cobra.Command
}

// NewQueryCandidatesCommand new query candidates cmd
func NewQueryCandidatesCommand(cli *Cli) *cobra.Command {
	c := new(QueryCandidatesCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "query-candidates options",
		Short: "Get all candidates for tdpos consensus.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return c.queryCandidates(ctx)
		},
	}
	return c.cmd
}

func (c *QueryCandidatesCommand) queryCandidates(ctx context.Context) error {
	client := c.cli.XchainClient()
	request := &pb.DposCandidatesRequest{
		Bcname: c.cli.RootOptions.Name,
	}
	res, err := client.DposCandidates(ctx, request)
	if err != nil {
		return err
	}
	output, err := json.MarshalIndent(res.CandidatesInfo, "", "  ")
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(string(output))
	return nil
}

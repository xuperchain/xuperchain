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

// QueryCheckResultCommand query res
type QueryCheckResultCommand struct {
	cli  *Cli
	cmd  *cobra.Command
	term int64
}

// NewQueryCheckResultCommand new query res
func NewQueryCheckResultCommand(cli *Cli) *cobra.Command {
	c := new(QueryCheckResultCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "query-checkResult",
		Short: "QueryCheckResult get check results of a specific term.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return c.queryCheckResult(ctx)
		},
	}
	c.addFlags()
	return c.cmd
}

func (c *QueryCheckResultCommand) addFlags() {
	c.cmd.Flags().Int64VarP(&c.term, "term", "t", 1, "term of checkresult")
}

func (c *QueryCheckResultCommand) queryCheckResult(ctx context.Context) error {
	client := c.cli.XchainClient()
	request := &pb.DposCheckResultsRequest{
		Bcname: c.cli.RootOptions.Name,
		Term:   c.term,
	}
	response, err := client.DposCheckResults(ctx, request)
	if err != nil {
		return err
	}
	fmt.Printf("term:%d\n", c.term)
	fmt.Printf("checkResult:%v\n", response.CheckResult)
	return nil
}

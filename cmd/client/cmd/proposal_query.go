/*
 * Copyright (c) 2021. Baidu Inc. All Rights Reserved.
 */

package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xuperchain/xupercore/kernel/contract/proposal/utils"
)

// ProposalQueryCommand proposal query cmd
type ProposalQueryCommand struct {
	cli *Cli
	cmd *cobra.Command

	module     string
	args       string
	methodName string
	isMulti    bool
	verbose    bool
	multiAddrs string
	proposalID string
}

// NewProposalQueryCommand new proposal query cmd
func NewProposalQueryCommand(cli *Cli) *cobra.Command {
	c := new(ProposalQueryCommand)
	c.cli = cli
	c.module = "xkernel"
	c.cmd = &cobra.Command{
		Use:     "query",
		Short:   "Query a proposal",
		Example: c.example(),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return c.query(ctx)
		},
	}
	c.addFlags()
	return c.cmd
}

func (c *ProposalQueryCommand) addFlags() {
	c.cmd.Flags().StringVarP(&c.proposalID, "pid", "p", "1", "proposal id.")
}

func (c *ProposalQueryCommand) example() string {
	return `
xchain-cli proposal query -p "your proposal id"
`
}

func (c *ProposalQueryCommand) query(ctx context.Context) error {
	ct := &CommTrans{
		ModuleName:   "xkernel",
		ContractName: utils.ProposalKernelContract,
		MethodName:   "Query",
		Args:         make(map[string][]byte),
		IsQuick:      c.isMulti,
		Keys:         c.cli.RootOptions.Keys,
		MultiAddrs:   c.multiAddrs,

		ChainName:    c.cli.RootOptions.Name,
		XchainClient: c.cli.XchainClient(),
	}

	if c.proposalID == "" {
		return fmt.Errorf("no proposal id found")
	}
	ct.Args["proposal_id"] = []byte(c.proposalID)

	response, _, err := ct.GenPreExeRes(ctx)
	if c.verbose {
		for _, req := range response.GetResponse().GetRequests() {
			limits := req.GetResourceLimits()
			for _, limit := range limits {
				fmt.Println(limit.Type.String(), ": ", limit.Limit)
			}
		}
	}
	return err
}

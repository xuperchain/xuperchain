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

// GovernTokenQueryCommand govern token query cmd
type GovernTokenQueryCommand struct {
	cli *Cli
	cmd *cobra.Command

	module     string
	args       string
	methodName string
	isMulti    bool
	verbose    bool
	multiAddrs string
	account    string
}

// NewContractQueryCommand new wasm/native/evm query cmd
func NewGovernTokenQueryCommand(cli *Cli) *cobra.Command {
	c := new(GovernTokenQueryCommand)
	c.cli = cli
	c.module = "xkernel"
	c.cmd = &cobra.Command{
		Use:     "query",
		Short:   "Query account's govern token balance",
		Example: c.example(),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return c.query(ctx)
		},
	}
	c.addFlags()
	return c.cmd
}

func (c *GovernTokenQueryCommand) addFlags() {
	c.cmd.Flags().StringVarP(&c.account, "account", "a", "", "govern token account.")
}

func (c *GovernTokenQueryCommand) example() string {
	return `
xchain-cli governToken query -a "your account"
`
}

func (c *GovernTokenQueryCommand) query(ctx context.Context) error {
	ct := &CommTrans{
		ModuleName:   "xkernel",
		ContractName: utils.GovernTokenKernelContract,
		MethodName:   "Query",
		Args:         make(map[string][]byte),
		IsQuick:      c.isMulti,
		Keys:         c.cli.RootOptions.Keys,
		MultiAddrs:   c.multiAddrs,

		ChainName:    c.cli.RootOptions.Name,
		XchainClient: c.cli.XchainClient(),
	}

	if c.account == "" {
		return fmt.Errorf("no account found")
	}
	ct.Args["account"] = []byte(c.account)

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

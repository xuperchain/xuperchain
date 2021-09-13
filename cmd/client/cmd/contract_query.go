/*
 * Copyright (c) 2021. Baidu Inc. All Rights Reserved.
 */

package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/xuperchain/xupercore/kernel/contract/bridge"
)

// ContractQueryCommand wasm/native/evm query cmd
type ContractQueryCommand struct {
	cli *Cli
	cmd *cobra.Command

	module     string
	args       string
	methodName string
	isMulti    bool
	verbose    bool
	multiAddrs string
	abiFile    string
}

// NewContractQueryCommand new wasm/native/evm query cmd
func NewContractQueryCommand(cli *Cli, module string) *cobra.Command {
	c := new(ContractQueryCommand)
	c.cli = cli
	c.module = module
	c.cmd = &cobra.Command{
		Use:     "query [options] code",
		Short:   "query info from contract code by customizing contract method",
		Example: c.example(),
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return c.query(ctx, args[0])
		},
	}
	c.addFlags()
	return c.cmd
}

func (c *ContractQueryCommand) addFlags() {
	c.cmd.Flags().StringVarP(&c.args, "args", "a", "{}", "query method args")
	c.cmd.Flags().StringVarP(&c.methodName, "method", "", "get", "contract method name")
	c.cmd.Flags().BoolVarP(&c.isMulti, "isMulti", "m", false, "multisig scene")
	c.cmd.Flags().BoolVarP(&c.verbose, "verbose", "v", false, "show query result verbosely")
	c.cmd.Flags().StringVarP(&c.multiAddrs, "multiAddrs", "A", "data/acl/addrs", "multiAddrs if multisig scene")
	if c.module == string(bridge.TypeEvm) {
		c.cmd.Flags().StringVarP(&c.abiFile, "abi", "", "", "the abi file of contract")
	}
}

func (c *ContractQueryCommand) example() string {
	return `
xchain wasm|native|evm query $codeaddr -a '{"Your contract parameters in json format"}' --method get
`
}

func (c *ContractQueryCommand) query(ctx context.Context, codeName string) error {
	ct := &CommTrans{
		ModuleName:   c.module,
		ContractName: codeName,
		MethodName:   c.methodName,
		Args:         make(map[string][]byte),
		IsQuick:      c.isMulti,
		Keys:         c.cli.RootOptions.Keys,
		MultiAddrs:   c.multiAddrs,

		ChainName:    c.cli.RootOptions.Name,
		XchainClient: c.cli.XchainClient(),
	}

	// generate preExe params
	args := make(map[string]interface{})
	err := json.Unmarshal([]byte(c.args), &args)
	if err != nil {
		return err
	}
	if c.module == string(bridge.TypeEvm) {
		if ct.Args, err = convertToXuper3EvmArgs(args); err != nil {
			return err
		}
	} else {
		if ct.Args, err = convertToXuper3Args(args); err != nil {
			return err
		}
	}

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

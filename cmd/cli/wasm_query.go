/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package main

import (
	"context"
	"encoding/json"

	"github.com/spf13/cobra"
)

// WasmQueryCommand wasm query cmd
type WasmQueryCommand struct {
	cli *Cli
	cmd *cobra.Command

	args       string
	methodName string
}

// NewWasmQueryCommand new wasm query cmd
func NewWasmQueryCommand(cli *Cli) *cobra.Command {
	c := new(WasmQueryCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:     "query [options] code",
		Short:   "query info from wasm code by customizing contract method",
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

func (c *WasmQueryCommand) addFlags() {
	c.cmd.Flags().StringVarP(&c.args, "args", "a", "{}", "query method args")
	c.cmd.Flags().StringVarP(&c.methodName, "method", "", "get", "contract method name")
}

func (c *WasmQueryCommand) example() string {
	return `
xchain wasm query $codeaddr -a '{"Your contract parameters in json format"}' --method get
`
}

func (c *WasmQueryCommand) query(ctx context.Context, codeName string) error {
	ct := &CommTrans{
		ModuleName:   "wasm",
		ContractName: codeName,
		MethodName:   c.methodName,
		Args:         make(map[string][]byte),
		IsMulti:      false,
		Keys:         c.cli.RootOptions.Keys,

		ChainName:    c.cli.RootOptions.Name,
		XchainClient: c.cli.XchainClient(),
	}

	args := make(map[string]interface{})
	if c.args != "" {
		json.Unmarshal([]byte(c.args), &args)
	}
	var err error
	ct.Args, err = convertToXuper3Args(args)
	if err != nil {
		return err
	}

	_, _, err = ct.GenPreExeRes(ctx)
	//fmt.Println(preExeRPCRes.GetResponse().GetResponse())
	return err
}

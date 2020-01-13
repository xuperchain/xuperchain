/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package main

import (
	"context"
	"encoding/json"

	"github.com/spf13/cobra"
)

// NativeQueryCommand native query cmd
type NativeQueryCommand struct {
	cli *Cli
	cmd *cobra.Command

	args       string
	methodName string
}

// NewNativeQueryCommand new native query cmd
func NewNativeQueryCommand(cli *Cli) *cobra.Command {
	c := new(NativeQueryCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:     "query",
		Short:   "[Deprecated] Query storage of native contract.",
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

func (c *NativeQueryCommand) addFlags() {
	c.cmd.Flags().StringVarP(&c.args, "args", "a", "{}", "query args")
	c.cmd.Flags().StringVarP(&c.methodName, "method", "", "query", "method name")
}

func (c *NativeQueryCommand) example() string {
	return `
xchain native query $code --method query or get or others --args {"keys":["a", "b"]}
`
}

func (c *NativeQueryCommand) query(ctx context.Context, codeName string) error {
	ct := &CommTrans{
		ModuleName:   "native",
		ContractName: codeName,
		MethodName:   c.methodName,
		Args:         make(map[string][]byte),
		IsQuick:      false,
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
	return err

	//fmt.Println(string(preExeRPCRes.GetResponse().GetResponse()))
}

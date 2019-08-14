/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package main

import (
	"context"
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/xuperchain/xuperunion/utxo"
)

// WasmInvokeCommand wasm invoke cmd
type WasmInvokeCommand struct {
	cli *Cli
	cmd *cobra.Command

	args       string
	account    string
	fee        string
	isMulti    bool
	multiAddrs string
	output     string
	methodName string
}

// NewWasmInvokeCommand new wasm invoke cmd
func NewWasmInvokeCommand(cli *Cli) *cobra.Command {
	c := new(WasmInvokeCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:     "invoke [options] code",
		Short:   "invoke from wasm code by customizing contract method",
		Example: c.example(),
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return c.invoke(ctx, args[0])
		},
	}
	c.addFlags()
	return c.cmd
}

func (c *WasmInvokeCommand) addFlags() {
	c.cmd.Flags().StringVarP(&c.args, "args", "a", "{}", "contract method args")
	c.cmd.Flags().StringVarP(&c.account, "account", "", "", "account name")
	c.cmd.Flags().StringVar(&c.fee, "fee", "", "fee of one tx")
	c.cmd.Flags().BoolVarP(&c.isMulti, "isMulti", "m", false, "multisig scene")
	c.cmd.Flags().StringVarP(&c.multiAddrs, "multiAddrs", "A", "data/acl/addrs", "multiAddrs if multisig scene")
	c.cmd.Flags().StringVarP(&c.output, "output", "o", "./tx.out", "tx draw data")
	c.cmd.Flags().StringVarP(&c.methodName, "method", "", "invoke", "contract method name")

}

func (c *WasmInvokeCommand) example() string {
	return `
xchain wasm invoke $codeaddr --method invoke -a '{"Your method args in json format"}'
`
}

func (c *WasmInvokeCommand) invoke(ctx context.Context, codeName string) error {
	ct := &CommTrans{
		Fee:          c.fee,
		FrozenHeight: 0,
		Version:      utxo.TxVersion,
		From:         c.account,
		ModuleName:   "wasm",
		ContractName: codeName,
		MethodName:   c.methodName,
		Args:         make(map[string][]byte),
		MultiAddrs:   c.multiAddrs,
		IsQuick:      c.isMulti,
		Output:       c.output,
		ChainName:    c.cli.RootOptions.Name,
		Keys:         c.cli.RootOptions.Keys,
		XchainClient: c.cli.XchainClient(),
		CryptoType:   c.cli.RootOptions.CryptoType,
	}

	// generate preExe params
	args := make(map[string]interface{})
	err := json.Unmarshal([]byte(c.args), &args)
	if err != nil {
		return err
	}
	ct.Args, err = convertToXuper3Args(args)
	if err != nil {
		return err
	}

	if c.isMulti {
		err = ct.GenerateMultisigGenRawTx(ctx)
	} else {
		err = ct.Transfer(ctx)
	}

	return err
}

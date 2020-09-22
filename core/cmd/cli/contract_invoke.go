/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/xuperchain/xuperchain/core/utxo"
)

// ContractInvokeCommand wasm invoke cmd
type ContractInvokeCommand struct {
	cli *Cli
	cmd *cobra.Command

	module     string
	args       string
	account    string
	fee        string
	isMulti    bool
	multiAddrs string
	output     string
	methodName string
	amount     string
	debug      bool
}

// NewContractInvokeCommand new wasm invoke cmd
func NewContractInvokeCommand(cli *Cli, module string) *cobra.Command {
	c := new(ContractInvokeCommand)
	c.cli = cli
	c.module = module
	c.cmd = &cobra.Command{
		Use:     "invoke [options] code",
		Short:   "invoke from contract code by customizing contract method",
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

func (c *ContractInvokeCommand) addFlags() {
	c.cmd.Flags().StringVarP(&c.args, "args", "a", "{}", "contract method args")
	c.cmd.Flags().StringVarP(&c.account, "account", "", "", "account name")
	c.cmd.Flags().StringVar(&c.fee, "fee", "", "fee of one tx")
	c.cmd.Flags().BoolVarP(&c.isMulti, "isMulti", "m", false, "multisig scene")
	c.cmd.Flags().StringVarP(&c.multiAddrs, "multiAddrs", "A", "data/acl/addrs", "multiAddrs if multisig scene")
	c.cmd.Flags().StringVarP(&c.output, "output", "o", "./tx.out", "tx draw data")
	c.cmd.Flags().StringVarP(&c.methodName, "method", "", "invoke", "contract method name")
	c.cmd.Flags().StringVarP(&c.amount, "amount", "", "", "the amount transfer to contract")
	c.cmd.Flags().BoolVarP(&c.debug, "debug", "", false, "debug print tx instead of posting")

}

func (c *ContractInvokeCommand) example() string {
	return `
xchain wasm|native invoke $codeaddr --method invoke -a '{"Your method args in json format"}'
`
}

func (c *ContractInvokeCommand) invoke(ctx context.Context, codeName string) error {
	ct := &CommTrans{
		Fee:          c.fee,
		FrozenHeight: 0,
		Version:      utxo.TxVersion,
		From:         c.account,
		ModuleName:   c.module,
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
		DebugTx:      c.debug,
		CliConf:      c.cli.RootOptions.CliConf,
	}
	// transfer to contract
	if c.amount != "" {
		ct.To = ct.ContractName
		ct.Amount = c.amount
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

func convertToXuper3Args(args map[string]interface{}) (map[string][]byte, error) {
	argmap := make(map[string][]byte)
	for k, v := range args {
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("bad key %s, expect string value, got %v", k, v)
		}
		argmap[k] = []byte(s)
	}
	return argmap, nil
}

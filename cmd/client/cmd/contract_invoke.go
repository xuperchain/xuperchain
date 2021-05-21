/*
 * Copyright (c) 2021. Baidu Inc. All Rights Reserved.
 */

package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/xuperchain/xupercore/bcs/ledger/xledger/state/utxo"
	"github.com/xuperchain/xupercore/kernel/contract/bridge"
)

// ContractInvokeCommand wasm/native/evm invoke cmd
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
	abiFile    string
}

// NewContractInvokeCommand new wasm/native/evm invoke cmd
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
	if c.module == string(bridge.TypeEvm) {
		c.cmd.Flags().StringVarP(&c.abiFile, "abi", "", "", "the abi file of contract")
	}
}

func (c *ContractInvokeCommand) example() string {
	return `
xchain wasm|native|evm invoke $codeaddr --method invoke -a '{"Your method args in json format"}'
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
		CryptoType:   c.cli.RootOptions.Crypto,
		DebugTx:      c.debug,
		CliConf:      c.cli.CliConf,
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
	if c.module == string(bridge.TypeEvm) {
		if ct.Args, err = convertToXuper3EvmArgs(args); err != nil {
			return err
		}
	} else {
		if ct.Args, err = convertToXuper3Args(args); err != nil {
			return err
		}
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

// evm contract args to xuper3 args.
func convertToXuper3EvmArgs(args map[string]interface{}) (map[string][]byte, error) {
	input, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}

	// 此处与 server 端结构相同，如果 jsonEncoded 字段修改，server 端也要修改（core/contract/evm/creator.go）。
	ret := map[string][]byte{
		"input":       input,
		"jsonEncoded": []byte("true"),
	}
	return ret, nil
}

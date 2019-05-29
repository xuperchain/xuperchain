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

// MultisigGenCommand multisig generate struct
type MultisigGenCommand struct {
	cli *Cli
	cmd *cobra.Command

	to           string
	amount       string
	descfile     string
	fee          string
	frozenHeight int64
	version      int32
	output       string
	multiAddrs   string
	from         string
	// contract params
	moduleName   string
	contractName string
	methodName   string
	args         string
}

// NewMultisigGenCommand multisig gen init method
func NewMultisigGenCommand(cli *Cli) *cobra.Command {
	c := new(MultisigGenCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "gen",
		Short: "Generate a raw transaction.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return c.generateTx(ctx)
		},
	}
	c.addFlags()
	return c.cmd
}

func (c *MultisigGenCommand) addFlags() {
	c.cmd.Flags().StringVar(&c.to, "to", "", "Target account/address of transfer.")
	c.cmd.Flags().StringVar(&c.amount, "amount", "0", "Token amount to be transferred.")
	c.cmd.Flags().StringVar(&c.descfile, "desc", "", "Desc file with the format of json for contract.")
	c.cmd.Flags().StringVar(&c.fee, "fee", "", "Fee to run a transaction.")
	c.cmd.Flags().Int64Var(&c.frozenHeight, "frozen", 0, "Frozen height of a transaction.")
	c.cmd.Flags().Int32Var(&c.version, "txversion", utxo.TxVersion, "Tx version.")
	c.cmd.Flags().StringVarP(&c.output, "output", "o", "./tx.out", "Serialized transaction data file.")
	c.cmd.Flags().StringVarP(&c.multiAddrs, "multiAddrs", "A", "data/acl/addrs", "MultiAddrs to fill required accounts/addresses.")
	c.cmd.Flags().StringVar(&c.from, "from", "", "Initiator of an transaction.")
	c.cmd.Flags().StringVar(&c.moduleName, "module", "", "Contract type: xkernel or wasm or native, native is deprecated.")
	c.cmd.Flags().StringVar(&c.contractName, "contract", "", "Contract name to be called.")
	c.cmd.Flags().StringVar(&c.methodName, "method", "", "Contract method to be called, It has been implemented in the target contract.")
	c.cmd.Flags().StringVar(&c.args, "args", "", "Contract parameters with json format for target contract method.")
}

// gen 命令入口
func (c *MultisigGenCommand) generateTx(ctx context.Context) error {
	ct := &CommTrans{
		To:           c.to,
		Amount:       c.amount,
		Descfile:     c.descfile,
		Fee:          c.fee,
		FrozenHeight: c.frozenHeight,
		Version:      c.version,
		From:         c.from,
		ModuleName:   c.moduleName,
		ContractName: c.contractName,
		MethodName:   c.methodName,
		Args:         make(map[string][]byte),
		MultiAddrs:   c.multiAddrs,
		Output:       c.output,
		IsPrint:      true,

		ChainName:    c.cli.RootOptions.Name,
		Keys:         c.cli.RootOptions.Keys,
		XchainClient: c.cli.XchainClient(),
	}

	if c.args != "" {
		err := json.Unmarshal([]byte(c.args), &ct.Args)
		if err != nil {
			return err
		}
	}

	return ct.GenerateMultisigGenRawTx(ctx)
}

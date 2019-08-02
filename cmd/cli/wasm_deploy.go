/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package main

import (
	"context"
	"encoding/json"
	"io/ioutil"

	"github.com/golang/protobuf/proto"
	"github.com/spf13/cobra"

	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/utxo"
)

// WasmDeployCommand wasm deploy cmd
type WasmDeployCommand struct {
	cli *Cli
	cmd *cobra.Command

	account      string
	contractName string
	args         string
	runtime      string
	fee          string
	isMulti      bool
	multiAddrs   string
	output       string
}

// NewWasmDeployCommand new wasm deploy cmd
func NewWasmDeployCommand(cli *Cli) *cobra.Command {
	c := new(WasmDeployCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "deploy [options] code path",
		Short: "deploy wasm code",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return c.deploy(ctx, args[0])
		},
	}
	c.addFlags()
	return c.cmd
}

func (c *WasmDeployCommand) addFlags() {
	c.cmd.Flags().StringVarP(&c.args, "arg", "a", "{}", "init arguments according your contract")
	c.cmd.Flags().StringVarP(&c.contractName, "cname", "n", "", "contract name")
	c.cmd.Flags().StringVarP(&c.account, "account", "", "", "account name")
	c.cmd.Flags().StringVarP(&c.runtime, "runtime", "", "c", "if contract code use go lang, then go or if use c lang, then c")
	c.cmd.Flags().StringVar(&c.fee, "fee", "", "fee of one tx")
	c.cmd.Flags().BoolVarP(&c.isMulti, "isMulti", "m", false, "multisig scene")
	c.cmd.Flags().StringVarP(&c.multiAddrs, "multiAddrs", "A", "data/acl/addrs", "multiAddrs if multisig scene")
	c.cmd.Flags().StringVarP(&c.output, "output", "o", "./tx.out", "tx draw data")
}

func (c *WasmDeployCommand) deploy(ctx context.Context, codepath string) error {
	ct := &CommTrans{
		Amount:       "0",
		Fee:          c.fee,
		FrozenHeight: 0,
		Version:      utxo.TxVersion,
		ModuleName:   "xkernel",
		ContractName: c.contractName,
		MethodName:   "Deploy",
		Args:         make(map[string][]byte),
		MultiAddrs:   c.multiAddrs,
		From:         c.account,
		Output:       c.output,
		IsQuick:      c.isMulti,
		ChainName:    c.cli.RootOptions.Name,
		Keys:         c.cli.RootOptions.Keys,
		XchainClient: c.cli.XchainClient(),
		CryptoType:   c.cli.RootOptions.CryptoType,
	}

	var err error
	ct.To, err = readAddress(ct.Keys)
	if err != nil {
		return err
	}

	// generate preExe params
	args := make(map[string]interface{})
	err = json.Unmarshal([]byte(c.args), &args)
	if err != nil {
		return err
	}
	x3args, err := convertToXuper3Args(args)
	if err != nil {
		return err
	}
	initArgs, _ := json.Marshal(x3args)
	codebuf, err := ioutil.ReadFile(codepath)
	if err != nil {
		return err
	}
	descbuf := c.prepareCodeDesc()
	ct.Args = map[string][]byte{
		"account_name":  []byte(c.account),
		"contract_name": []byte(c.contractName),
		"contract_code": codebuf,
		"contract_desc": descbuf,
		"init_args":     initArgs,
	}

	if c.isMulti {
		err = ct.GenerateMultisigGenRawTx(ctx)
	} else {
		err = ct.Transfer(ctx)
	}

	return err
}

func (c *WasmDeployCommand) prepareCodeDesc() []byte {
	desc := &pb.WasmCodeDesc{
		Runtime: c.runtime,
	}
	buf, _ := proto.Marshal(desc)
	return buf
}

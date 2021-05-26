/*
 * Copyright (c) 2021. Baidu Inc. All Rights Reserved.
 */

package cmd

import (
	"context"
	"encoding/json"
	"io/ioutil"

	"github.com/golang/protobuf/proto"
	"github.com/spf13/cobra"

	"github.com/xuperchain/xuperchain/service/xchainpb/pb"
	"github.com/xuperchain/xupercore/bcs/ledger/xledger/state/utxo"
	"github.com/xuperchain/xupercore/kernel/contract/bridge"
)

// ContractDeployCommand wasm/native/evm deploy cmd
type ContractDeployCommand struct {
	cli *Cli
	cmd *cobra.Command

	module       string
	account      string
	contractName string
	args         string
	runtime      string
	fee          string
	isMulti      bool
	multiAddrs   string
	output       string
	abiFile      string
}

// NewContractDeployCommand new wasm/native/evm deploy cmd
func NewContractDeployCommand(cli *Cli, module string) *cobra.Command {
	c := new(ContractDeployCommand)
	c.cli = cli
	c.module = module
	c.cmd = &cobra.Command{
		Use:   "deploy [options] code path",
		Short: "deploy contract code",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return c.deploy(ctx, args[0])
		},
	}
	c.addFlags()
	return c.cmd
}

func (c *ContractDeployCommand) addFlags() {
	c.cmd.Flags().StringVarP(&c.args, "arg", "a", "{}", "init arguments according your contract")
	c.cmd.Flags().StringVarP(&c.contractName, "cname", "n", "", "contract name")
	c.cmd.Flags().StringVarP(&c.account, "account", "", "", "account name")
	c.cmd.Flags().StringVarP(&c.runtime, "runtime", "", "c", "if contract code use go lang, then go or if use c lang, then c")
	c.cmd.Flags().StringVar(&c.fee, "fee", "", "fee of one tx")
	c.cmd.Flags().BoolVarP(&c.isMulti, "isMulti", "m", false, "multisig scene")
	c.cmd.Flags().StringVarP(&c.multiAddrs, "multiAddrs", "A", "data/acl/addrs", "multiAddrs if multisig scene")
	c.cmd.Flags().StringVarP(&c.output, "output", "o", "./tx.out", "tx draw data")
	if c.module == string(bridge.TypeEvm) {
		c.cmd.Flags().StringVarP(&c.abiFile, "abi", "", "", "the abi file of contract")
	}
}

func (c *ContractDeployCommand) deploy(ctx context.Context, codepath string) error {
	ct := &CommTrans{
		Amount:       "0",
		Fee:          c.fee,
		FrozenHeight: 0,
		Version:      utxo.TxVersion,
		ModuleName:   "xkernel",
		ContractName: "$contract",
		MethodName:   "deployContract",
		Args:         make(map[string][]byte),
		MultiAddrs:   c.multiAddrs,
		From:         c.account,
		Output:       c.output,
		IsQuick:      c.isMulti,
		ChainName:    c.cli.RootOptions.Name,
		Keys:         c.cli.RootOptions.Keys,
		XchainClient: c.cli.XchainClient(),
		CryptoType:   c.cli.RootOptions.Crypto,
		RootOptions:  c.cli.RootOptions,
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

	var codeBuf, abiCode []byte
	var x3args map[string][]byte

	if c.module == string(bridge.TypeEvm) {
		if abiCode, err = ioutil.ReadFile(c.abiFile); err != nil {
			return err
		}

		if x3args, err = convertToXuper3EvmArgs(args); err != nil {
			return err
		}
	} else {
		if x3args, err = convertToXuper3Args(args); err != nil {
			return err
		}
	}

	codeBuf, err = ioutil.ReadFile(codepath)
	if err != nil {
		return err
	}

	initArgs, _ := json.Marshal(x3args)

	descBuf := c.prepareCodeDesc()

	ct.Args = map[string][]byte{
		"account_name":  []byte(c.account),
		"contract_name": []byte(c.contractName),
		"contract_code": codeBuf,
		"contract_desc": descBuf,
		"init_args":     initArgs,
		"contract_abi":  abiCode,
	}

	if c.isMulti {
		err = ct.GenerateMultisigGenRawTx(ctx)
	} else {
		err = ct.Transfer(ctx)
	}

	return err
}

func (c *ContractDeployCommand) prepareCodeDesc() []byte {
	desc := &pb.WasmCodeDesc{
		Runtime:      c.runtime,
		ContractType: c.module,
	}
	buf, _ := proto.Marshal(desc)
	return buf
}

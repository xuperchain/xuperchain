/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/burrow/deploy/compile"
	"github.com/spf13/cobra"

	"github.com/xuperchain/xuperchain/core/common/log"
	"github.com/xuperchain/xuperchain/core/contract/bridge"
	"github.com/xuperchain/xuperchain/core/pb"
	"github.com/xuperchain/xuperchain/core/utxo"
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
	c.cmd.Flags().StringVarP(&c.abiFile, "abi", "", "", "the abi file of contract")
}

func (c *ContractDeployCommand) deploy(ctx context.Context, codepath string) error {
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

	var codeBuf, abiCode []byte
	var evmCode string
	if c.module == string(bridge.TypeEvm) {
		if c.abiFile != "" {
			evmCode, abiCode, err = readEVMCodeAndAbi(c.abiFile)
			if err != nil {
				return err
			}
		} else {
			evmCode, abiCode, err = compileSolidityForEVM(codepath)
			if err != nil {
				return err
			}
			codeBuf, err = hex.DecodeString(evmCode)
			if err != nil {
				return err
			}
		}

	} else {
		codeBuf, err = ioutil.ReadFile(codepath)
		if err != nil {
			return err
		}
	}

	// generate preExe params
	args := make(map[string]interface{})
	err = json.Unmarshal([]byte(c.args), &args)
	if err != nil {
		return err
	}

	var x3args map[string][]byte
	if c.module == string(bridge.TypeEvm) && c.args != "" {
		x3args, ct.AbiCode, err = convertToEvmArgsWithAbiData(abiCode, "", args)
		if err != nil {
			return err
		}
		callData := hex.EncodeToString(x3args["input"])
		evmCode = evmCode + callData
		codeBuf, err = hex.DecodeString(evmCode)
		if err != nil {
			return err
		}
	} else {
		x3args, err = convertToXuper3Args(args)
		if err != nil {
			return err
		}
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

func readEVMCodeAndAbi(abiFilePath string) (string, []byte, error) {
	if _, err := os.Stat(abiFilePath); err != nil {
		fmt.Printf("abifile not found,%s\n", abiFilePath)
		return "", nil, fmt.Errorf("Abi doesn't exist for =>\t%s", abiFilePath)
	}
	sol, err := compile.LoadSolidityContract(abiFilePath)
	if err != nil {
		return "", nil, err
	}

	return sol.Evm.Bytecode.Object, sol.Abi, nil
}

func compileSolidityForEVM(solidityFile string) (string, []byte, error) {
	filePath, fileName := filepath.Split(solidityFile)
	contractCode, abiCode, err := compileSolidity(filePath, fileName, false)
	if err != nil {
		return "", nil, err
	}
	return contractCode, []byte(abiCode), nil
}

func compileSolidity(filePath string, fileName string, optimize bool) (string, string, error) {
	fileSuffix := path.Ext(fileName)
	fileNameOnly := strings.TrimSuffix(fileName, fileSuffix)

	logger, _ := log.NewLoggerForEVM()

	// compile contracts
	resp, err := compile.EVM(fileName, optimize, filePath, nil, logger)
	if err != nil {
		return "", "", err
	}
	var solidityContract compile.SolidityContract
	if len(resp.Objects) == 1 {
		// only one contract
		// check that the file name is the same as the contract name
		if resp.Objects[0].Objectname != fileNameOnly {
			return "", "", fmt.Errorf("No contract found\n")
		}
		solidityContract = resp.Objects[0].Contract
	} else {
		// find the contract to be deployed from several contracts
		// please check that the file name is the same as the contract name
		var i int
		for i = range resp.Objects {
			if resp.Objects[i].Objectname == fileNameOnly {
				break
			}

			// find the last contract object, but no contract fount yet
			if i == len(resp.Objects)-1 {
				return "", "", fmt.Errorf("No contract found\n")
			}
		}
		solidityContract = resp.Objects[i].Contract
	}

	// save the contract's binary
	err = solidityContract.Save(filePath, fmt.Sprintf("%s.bin", fileNameOnly))
	if err != nil {
		fmt.Printf("Error: Solidity save error: %s \n", fileNameOnly)
		return "", "", err
	}

	// construct the byteCode
	if solidityContract.Evm.Bytecode.Object == "" {
		return "", "", fmt.Errorf("Solidity compile result is nil\n")
	}
	return solidityContract.Evm.Bytecode.Object, string(solidityContract.Abi), nil
}

/*
 * Copyright (c) 2021. Baidu Inc. All Rights Reserved.
 */

package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/hyperledger/burrow/crypto"
	"github.com/xuperchain/xupercore/bcs/contract/evm"
)

// EVMAddrTransCommand address translation between EVM and xchain
type EVMAddrTransCommand struct {
	cli *Cli
	cmd *cobra.Command

	transType string
	from      string
}

// NewEVMAddrTransCommand new trans cmd
func NewEVMAddrTransCommand(cli *Cli) *cobra.Command {
	c := new(EVMAddrTransCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "addr-trans -t [x2e/e2x] -f from_address",
		Short: "Address translation between EVM and xchain",
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return c.trans(ctx)
		},
	}
	c.addFlags()
	return c.cmd
}

func (c *EVMAddrTransCommand) addFlags() {
	c.cmd.Flags().StringVarP(&c.transType, "type", "t", "e2x", "the kind of address translation between EVM and xchain")
	c.cmd.Flags().StringVarP(&c.from, "from", "f", "", "from address")
}

func (c *EVMAddrTransCommand) trans(ctx context.Context) error {
	if c.transType != "x2e" && c.transType != "e2x" {
		return fmt.Errorf("wrong transType, must be x2e or e2x")
	}

	var addr, addrType string
	var err error
	switch c.transType {
	case "x2e":
		addr, addrType, err = evm.DetermineXchainAddress(c.from)
		if err != nil {
			return err
		}

	case "e2x":
		evmAddr, err := crypto.AddressFromHexString(c.from)
		if err != nil {
			return err
		}
		addr, addrType, err = evm.DetermineEVMAddress(evmAddr)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("wrong transType, must be x2e or e2x")
	}

	fmt.Printf("result, %s\t%s\n", addr, addrType)

	return nil
}

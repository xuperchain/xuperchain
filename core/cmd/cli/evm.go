/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package main

import "github.com/spf13/cobra"

// EvmCommand evm cmd struct
type EvmCommand struct {
}

// NewEvmCommand new evm cmd
func NewEvmCommand(cli *Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "evm",
		Short: "Operate a native contract: deploy|upgrade|invoke|query",
	}
	cmd.AddCommand(NewContractDeployCommand(cli, "evm"))
	cmd.AddCommand(NewContractInvokeCommand(cli, "evm"))
	cmd.AddCommand(NewContractQueryCommand(cli, "evm"))
	cmd.AddCommand(NewContractUpgradeCommand(cli))
	cmd.AddCommand(NewEVMAddrTransCommand(cli))
	return cmd
}

func init() {
	AddCommand(NewEvmCommand)
}

/*
 * Copyright (c) 2021. Baidu Inc. All Rights Reserved.
 */

package cmd

import "github.com/spf13/cobra"

// NewEvmCommand new evm cmd
func NewEvmCommand(cli *Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "evm",
		Short: "Operate an evm contract: deploy|upgrade|invoke|query",
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

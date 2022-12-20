/*
 * Copyright (c) 2021. Baidu Inc. All Rights Reserved.
 */

package cmd

import "github.com/spf13/cobra"

// NewNativeCommand new native cmd
func NewNativeCommand(cli *Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "native",
		Short: "Operate a native contract: deploy|upgrade|invoke|query",
	}
	cmd.AddCommand(NewContractDeployCommand(cli, "native"))
	cmd.AddCommand(NewContractInvokeCommand(cli, "native"))
	cmd.AddCommand(NewContractQueryCommand(cli, "native"))
	cmd.AddCommand(NewContractUpgradeCommand(cli))
	return cmd
}

func init() {
	AddCommand(NewNativeCommand)
}

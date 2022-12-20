/*
 * Copyright (c) 2021. Baidu Inc. All Rights Reserved.
 */

package cmd

import "github.com/spf13/cobra"

// NewWasmCommand new wasm cmd
func NewWasmCommand(cli *Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wasm",
		Short: "Operate a command with wasm, deploy|invoke|query",
	}
	cmd.AddCommand(NewContractDeployCommand(cli, "wasm"))
	cmd.AddCommand(NewContractInvokeCommand(cli, "wasm"))
	cmd.AddCommand(NewContractQueryCommand(cli, "wasm"))
	cmd.AddCommand(NewContractUpgradeCommand(cli))
	return cmd
}

func init() {
	AddCommand(NewWasmCommand)
}

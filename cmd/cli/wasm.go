/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package main

import "github.com/spf13/cobra"

// WasmCommand wasm cmd
type WasmCommand struct {
}

// NewWasmCommand new wasm cmd
func NewWasmCommand(cli *Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wasm",
		Short: "Operate a command with wasm, deploy|invoke|query",
	}
	cmd.AddCommand(NewWasmDeployCommand(cli))
	cmd.AddCommand(NewWasmInvokeCommand(cli))
	cmd.AddCommand(NewWasmQueryCommand(cli))
	return cmd
}

func init() {
	AddCommand(NewWasmCommand)
}

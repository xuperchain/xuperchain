/*
 * Copyright (c) 2021. Baidu Inc. All Rights Reserved.
 */

package cmd

import "github.com/spf13/cobra"

// NewXKernelCommand new xkernel cmd
func NewXKernelCommand(cli *Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "xkernel",
		Short: "Operate a command with xkernel, invoke|query",
	}
	cmd.AddCommand(NewContractInvokeCommand(cli, "xkernel"))
	cmd.AddCommand(NewContractQueryCommand(cli, "xkernel"))
	return cmd
}

func init() {
	AddCommand(NewXKernelCommand)
}

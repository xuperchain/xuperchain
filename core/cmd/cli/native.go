/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package main

import "github.com/spf13/cobra"

// NativeCommand native cmd struct
type NativeCommand struct {
}

// NewNativeCommand new native cmd
func NewNativeCommand(cli *Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "native",
		Short: "[Deprecated] Operate a native contract: activate|deactivate|deploy|invoke|query|status.",
	}
	cmd.AddCommand(NewNativeActivateCommand(cli))
	cmd.AddCommand(NewNativeDeployCommand(cli))
	cmd.AddCommand(NewNativeStatusCommand(cli))
	cmd.AddCommand(NewNativeQueryCommand(cli))
	cmd.AddCommand(NewNativeDeactivateCommand(cli))
	cmd.AddCommand(NewNativeInvokeCommand(cli))
	return cmd
}

func init() {
	AddCommand(NewNativeCommand)
}

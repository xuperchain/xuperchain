/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package main

import (
	"context"

	"github.com/spf13/cobra"
)

// NewNativeDeactivateCommand new native deactivate cmd
func NewNativeDeactivateCommand(cli *Cli) *cobra.Command {
	c := new(NativeActivateCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "deactivate",
		Short: "[Deprecated] Deactivate a ready native contract, make the contract registered.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			pluginName := args[0]
			return c.run(ctx, pluginName, "deactivate")
		},
	}
	c.addFlags()
	return c.cmd
}

/*
 * Copyright (c) 2021. Baidu Inc. All Rights Reserved.
 */

package cmd

import (
	"github.com/spf13/cobra"
)

// NewNetURLCommand new neturl cmd
func NewNetURLCommand(cli *Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "netURL",
		Short: "Operate a netURL: gen|get|preview|convert.",
	}
	cmd.AddCommand(NewNetURLGenCommand(cli))
	cmd.AddCommand(NewNetURLGetCommand(cli))
	cmd.AddCommand(NewNetURLPreviewCommand(cli))
	cmd.AddCommand(NewNetURLConvertCommand(cli))
	return cmd
}

func init() {
	AddCommand(NewNetURLCommand)
}

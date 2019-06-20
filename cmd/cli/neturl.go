/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package main

import (
	"github.com/spf13/cobra"
)

// NetURLCommand neturl cmd
type NetURLCommand struct {
}

// NewNetURLCommand new neturl cmd
func NewNetURLCommand(cli *Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "netURL",
		Short: "Operate a netURL: gen|get|preview.",
	}
	cmd.AddCommand(NewNetURLGenCommand(cli))
	cmd.AddCommand(NewNetURLGetCommand(cli))
	cmd.AddCommand(NewNetURLPreviewCommand(cli))
	return cmd
}

func init() {
	AddCommand(NewNetURLCommand)
}

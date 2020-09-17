/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package main

import (
	"context"

	"github.com/spf13/cobra"

	p2p_base "github.com/xuperchain/xuperchain/core/p2p/base"
)

// NetURLConvertCommand neturl gen cmd
type NetURLConvertCommand struct {
	cli  *Cli
	cmd  *cobra.Command
	path string
}

// NewNetURLGenCommand new neturl gen cmd
func NewNetURLConvertCommand(cli *Cli) *cobra.Command {
	n := new(NetURLConvertCommand)
	n.cli = cli
	n.cmd = &cobra.Command{
		Use:   "convert [options]",
		Short: "Convert net private key to CA pem format",
		RunE: func(cmd *cobra.Command, args []string) error {
			return n.convertKey(context.TODO())
		},
	}
	n.addFlags()
	return n.cmd
}

func (n *NetURLConvertCommand) addFlags() {
	n.cmd.Flags().StringVar(&n.path, "path", "./data/netkeys", "path where net_private.key saved (default is ./data/netkeys)")
}

func (n *NetURLConvertCommand) convertKey(ctx context.Context) error {
	return p2p_base.GeneratePemKeyFromNetKey(n.path)
}

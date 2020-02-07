/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package main

import (
	"context"

	"github.com/spf13/cobra"

	p2p_base "github.com/xuperchain/xuperchain/core/p2p/base"
)

// NetURLGenCommand neturl gen cmd
type NetURLGenCommand struct {
	cli  *Cli
	cmd  *cobra.Command
	path string
}

// NewNetURLGenCommand new neturl gen cmd
func NewNetURLGenCommand(cli *Cli) *cobra.Command {
	n := new(NetURLGenCommand)
	n.cli = cli
	n.cmd = &cobra.Command{
		Use:   "gen [options]",
		Short: "Generate net url for p2p",
		RunE: func(cmd *cobra.Command, args []string) error {
			return n.genNetURL(context.TODO())
		},
	}
	n.addFlags()
	return n.cmd
}

func (n *NetURLGenCommand) addFlags() {
	n.cmd.Flags().StringVar(&n.path, "path", "./data/netkeys/", "path to save net url (default is ./data/netkeys/)")
}

func (n *NetURLGenCommand) genNetURL(ctx context.Context) error {
	return p2p_base.GenerateKeyPairWithPath(n.path)
}

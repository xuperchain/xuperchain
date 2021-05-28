/*
 * Copyright (c) 2021. Baidu Inc. All Rights Reserved.
 */

package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/xuperchain/xupercore/kernel/network/p2p"
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
	return p2p.GenerateKeyPairWithPath(n.path)
}

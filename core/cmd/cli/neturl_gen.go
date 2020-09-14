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
	cli    *Cli
	cmd    *cobra.Command
	path   string
	from   string
	format string
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
	n.cmd.Flags().StringVar(&n.from, "from", "", "gen from private key, net|pem, (net is libp2p format, pem is standard format)")
}

func (n *NetURLGenCommand) genNetURL(ctx context.Context) error {
	switch n.from {
	case "net":
		return p2p_base.GeneratePemKeyFromNetKey(n.path)
	case "pem":
		return p2p_base.GenerateNetKeyFromPemKey(n.path)
	default:
		return p2p_base.GenerateKeyPairWithPath(n.path)
	}
}

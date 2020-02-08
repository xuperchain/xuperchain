package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	p2p_base "github.com/xuperchain/xuperchain/core/p2p/base"
)

// NetURLPreviewCommand preview neturl using given params
type NetURLPreviewCommand struct {
	cli  *Cli
	cmd  *cobra.Command
	ip   string
	port string
	path string
}

// NewNetURLPreviewCommand new get neturl cmd
func NewNetURLPreviewCommand(cli *Cli) *cobra.Command {
	n := new(NetURLPreviewCommand)
	n.cli = cli
	n.cmd = &cobra.Command{
		Use:   "preview",
		Short: "preview net URL for p2p using given ip, port and key path",
		RunE: func(cmd *cobra.Command, args []string) error {
			return n.previewNetURL(context.TODO())
		},
	}
	n.addFlags()
	return n.cmd
}

func (n *NetURLPreviewCommand) addFlags() {
	n.cmd.Flags().StringVar(&n.ip, "ip", "127.0.0.1", "ip address of the p2p node (default is 127.0.0.1)")
	n.cmd.Flags().StringVar(&n.port, "port", "47101", "port of the p2p node (default is 47101)")
	n.cmd.Flags().StringVar(&n.path, "path", "./data/netkeys/", "path to save net url (default is ./data/netkeys/)")
}

func (n *NetURLPreviewCommand) previewNetURL(ctx context.Context) error {
	pid, err := p2p_base.GetPeerIDFromPath(n.path)
	if err != nil {
		fmt.Println("Parse net URL from key path failed, err=", err)
	}
	fmt.Printf("/ip4/%s/tcp/%s/p2p/%s\n", n.ip, n.port, pid)
	return nil
}

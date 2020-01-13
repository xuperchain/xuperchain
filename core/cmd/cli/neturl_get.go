/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/xuperchain/xuperchain/core/pb"
)

// NetURLGetCommand get neturl cmd
type NetURLGetCommand struct {
	cli *Cli
	cmd *cobra.Command
}

// NewNetURLGetCommand new get neturl cmd
func NewNetURLGetCommand(cli *Cli) *cobra.Command {
	n := new(NetURLGetCommand)
	n.cli = cli
	n.cmd = &cobra.Command{
		Use:   "get",
		Short: "Get net url for p2p",
		RunE: func(cmd *cobra.Command, args []string) error {
			return n.getNetURL(context.TODO())
		},
	}
	return n.cmd
}

func (n *NetURLGetCommand) getNetURL(ctx context.Context) error {
	client := n.cli.XchainClient()
	req := &pb.CommonIn{}
	res, err := client.GetNetURL(ctx, req)
	if err != nil {
		return err
	}
	output, err := json.MarshalIndent(res.RawUrl, "", "")
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(string(output))
	return nil
}

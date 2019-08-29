package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/xuperchain/xuperunion/global"
	"github.com/xuperchain/xuperunion/pb"
)

// XPoSXPowerQueryCommand structure for XPower query
type XPoSXPowerQueryCommand struct {
	cli *Cli
	cmd *cobra.Command
	// 查询address的xpower值
	address string
}

// NewXPoSXPowerQueryCommand new a cmd for xpower query
func NewXPoSXPowerQueryCommand(cli *Cli) *cobra.Command {
	c := new(XPoSXPowerQueryCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "xpower",
		Short: "query xpower of an address in the tip block",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return c.queryXPower(ctx)
		},
	}
	c.addFlags()

	return c.cmd
}

func (c *XPoSXPowerQueryCommand) addFlags() {
	c.cmd.Flags().StringVar(&c.address, "address", "", "address name")
}

func (c *XPoSXPowerQueryCommand) queryXPower(ctx context.Context) error {
	client := c.cli.XchainClient()
	xpowerRequest := &pb.XPowerRequest{
		Header: &pb.Header{
			Logid: global.Glogid(),
		},
		Bcname:  c.cli.RootOptions.Name,
		Address: c.address,
	}
	if len(c.address) == 0 {
		return errors.New("param error: address is required")
	}
	reply, err := client.QueryXPower(ctx, xpowerRequest)
	if err != nil {
		return err
	}

	if reply.Header.Error != pb.XChainErrorEnum_SUCCESS {
		return errors.New(reply.Header.Error.String())
	}
	if reply != nil {
		fmt.Println(reply.GetXpower())
	}

	return nil
}

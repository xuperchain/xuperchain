package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/xuperchain/xuperunion/global"
	"github.com/xuperchain/xuperunion/pb"
)

// XPoSSlotQueryCommand structure for slot query
type XPoSSlotQueryCommand struct {
	cli *Cli
	cmd *cobra.Command
}

// NewXPoSSlotQueryCommand new a cmd for slot query
func NewXPoSSlotQueryCommand(cli *Cli) *cobra.Command {
	c := new(XPoSSlotQueryCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "slot",
		Short: "query slot -> address info",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return c.querySlot(ctx)
		},
	}

	return c.cmd
}

func (c *XPoSSlotQueryCommand) querySlot(ctx context.Context) error {
	client := c.cli.XchainClient()
	slotRequest := &pb.Slot2AddressRequest{
		Header: &pb.Header{
			Logid: global.Glogid(),
		},
		Bcname: c.cli.RootOptions.Name,
	}

	reply, err := client.QuerySlot(ctx, slotRequest)
	if err != nil {
		return err
	}

	if reply.Header.Error != pb.XChainErrorEnum_SUCCESS {
		return errors.New(reply.Header.Error.String())
	}

	if reply != nil {
		slot2AddrArr := reply.GetSlot2Addr()
		for _, value := range slot2AddrArr {
			slot := value.GetSlot()
			address := value.GetAddr()
			fmt.Println("slotId:", "[", slot, "]", "address", "[", address, "]")
		}
	}

	return nil
}

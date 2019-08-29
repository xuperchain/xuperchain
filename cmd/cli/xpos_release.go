package main

import (
	"context"
	"errors"

	"github.com/spf13/cobra"

	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/utxo"
)

// XPoSRleaseSlotCommand structure for release a slot
type XPoSReleaseSlotCommand struct {
	cli *Cli
	cmd *cobra.Command

	descfile string
}

// NewXPoSReleaseSlotCommand new a cmd for release a slot
func NewXPoSReleaseSlotCommand(cli *Cli) *cobra.Command {
	c := new(XPoSReleaseSlotCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "release",
		Short: "release a slot",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return c.releaseSlot(ctx)
		},
	}
	c.addFlags()

	return c.cmd
}

func (c *XPoSReleaseSlotCommand) addFlags() {
	c.cmd.Flags().StringVar(&c.descfile, "desc", "", "The json config file for releasing a slot")
}

func (c *XPoSReleaseSlotCommand) releaseSlot(ctx context.Context) error {
	if len(c.descfile) == 0 {
		return errors.New("desc file required")
	}

	ct := &CommTrans{
		Amount:       "0",
		Fee:          "0",
		FrozenHeight: 0,
		Version:      utxo.TxVersion,

		MethodName: "ReleaseSlot",
		Args:       make(map[string][]byte),

		Descfile: c.descfile,
		IsQuick:  false,

		ChainName:       c.cli.RootOptions.Name,
		Keys:            c.cli.RootOptions.Keys,
		XchainClient:    c.cli.XchainClient(),
		CryptoType:      c.cli.RootOptions.CryptoType,
		TransactionType: pb.TransactionType_RELEASESLOT,
	}

	tranErr := ct.Transfer(ctx)
	if tranErr != nil {
		return tranErr
	}

	return nil
}

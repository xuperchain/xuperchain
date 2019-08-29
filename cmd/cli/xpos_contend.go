package main

import (
	"context"
	"errors"

	"github.com/spf13/cobra"

	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/utxo"
)

// XPoSContendSlotCommand structure for contend slot
type XPoSContendSlotCommand struct {
	cli *Cli
	cmd *cobra.Command

	// desc file for contend a slot
	descfile string
	fee      string
}

// NewXPoSContendSlotCommand structure for contend a slot
func NewXPoSContendSlotCommand(cli *Cli) *cobra.Command {
	c := new(XPoSContendSlotCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "contend",
		Short: "contend a slot",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return c.contendSlot(ctx)
		},
	}
	c.addFlags()

	return c.cmd
}

func (c *XPoSContendSlotCommand) addFlags() {
	c.cmd.Flags().StringVar(&c.descfile, "desc", "", "The json config file for contending a slot")
	c.cmd.Flags().StringVar(&c.fee, "fee", "0", "The fee to contend a slot")
}

func (c *XPoSContendSlotCommand) contendSlot(ctx context.Context) error {
	if len(c.descfile) == 0 {
		return errors.New("desc file required")
	}

	ct := &CommTrans{
		Amount:       "0",
		Fee:          c.fee,
		FrozenHeight: 0,
		Version:      utxo.TxVersion,

		MethodName: "ContendSlot",
		Args:       make(map[string][]byte),

		Descfile: c.descfile,
		IsQuick:  false,

		ChainName:       c.cli.RootOptions.Name,
		Keys:            c.cli.RootOptions.Keys,
		XchainClient:    c.cli.XchainClient(),
		CryptoType:      c.cli.RootOptions.CryptoType,
		TransactionType: pb.TransactionType_CONTENDSLOT,
	}

	tranErr := ct.Transfer(ctx)
	if tranErr != nil {
		return tranErr
	}

	return nil
}

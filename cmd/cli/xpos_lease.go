package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/xuperchain/xuperunion/utxo"
)

// XPoSLeaseCommand structure for lease
type XPoSLeaseCommand struct {
	cli *Cli
	cmd *cobra.Command
	// 发起租赁方
	from string
	// 租赁接收方
	to string
	// 租赁数量
	amount  string
	version int32
}

// NewXPoSLeaseCommand new a command for lease
func NewXPoSLeaseCommand(cli *Cli) *cobra.Command {
	c := new(XPoSLeaseCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "lease",
		Short: "lease utxo from A to B",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return c.lease(ctx)
		},
	}

	c.addFlags()
	return c.cmd
}

func (c *XPoSLeaseCommand) addFlags() {
	c.cmd.Flags().StringVar(&c.from, "from", "", "lease address")
	c.cmd.Flags().StringVar(&c.to, "to", "", "leased address")
	c.cmd.Flags().StringVar(&c.amount, "amount", "", "amount of leasing")
	c.cmd.Flags().Int32Var(&c.version, "txversion", utxo.TxVersion, "tx version")
}

func (c *XPoSLeaseCommand) lease(ctx context.Context) error {
	opt := TransferOptions{
		BlockchainName: c.cli.RootOptions.Name,
		KeyPath:        c.cli.RootOptions.Keys,
		CryptoType:     c.cli.RootOptions.CryptoType,
		To:             "lease#" + c.to + "_" + c.from,
		Amount:         c.amount,
		From:           c.from,
		Version:        c.version,
	}

	txid, tranErr := c.cli.Transfer(ctx, &opt)
	if tranErr != nil {
		return tranErr
	}
	fmt.Println("txid: ", txid)

	return nil
}

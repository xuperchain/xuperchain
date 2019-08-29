package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/utxo"
)

// XPoSCancelLeaseCommand structure for lease
type XPoSCancelLeaseCommand struct {
	cli *Cli
	cmd *cobra.Command
	// 取消租赁人
	from string
	// 取消租赁的钱的去处
	to string
	// 租赁生成的交易
	leaseTxid string
	// 租赁出去的数额
	amount  string
	version int32
}

// NewXPoSCancelLeaseCommand new a cmd for cancel lease
func NewXPoSCancelLeaseCommand(cli *Cli) *cobra.Command {
	c := new(XPoSCancelLeaseCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "cancel",
		Short: "cancel lease utxo from A to B",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return c.cancel(ctx)
		},
	}

	c.addFlags()
	return c.cmd
}

func (c *XPoSCancelLeaseCommand) addFlags() {
	c.cmd.Flags().StringVar(&c.from, "from", "", "leased address")
	c.cmd.Flags().StringVar(&c.to, "to", "", "lease address regularly")
	c.cmd.Flags().StringVar(&c.leaseTxid, "txid", "", "tx lease")
	c.cmd.Flags().StringVar(&c.amount, "amount", "0", "leased amount")
	c.cmd.Flags().Int32Var(&c.version, "txversion", utxo.TxVersion, "tx version")
}

func (c *XPoSCancelLeaseCommand) cancel(ctx context.Context) error {
	opt := TransferOptions{
		BlockchainName: c.cli.RootOptions.Name,
		KeyPath:        c.cli.RootOptions.Keys,
		CryptoType:     c.cli.RootOptions.CryptoType,
		To:             c.to,
		From:           c.from,
		LeaseTxid:      c.leaseTxid,
		Amount:         c.amount,
		Type:           pb.TransactionType_CANCELLEASE,
		Version:        c.version,
	}

	txid, tranErr := c.cli.Transfer(ctx, &opt)
	if tranErr != nil {
		return tranErr
	}
	fmt.Println("txid:", txid)

	return nil
}

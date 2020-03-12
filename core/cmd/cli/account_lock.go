package main

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/xuperchain/xuperchain/core/global"
	"github.com/xuperchain/xuperchain/core/pb"
)

type AccountLockCommand struct {
	cli *Cli
	cmd *cobra.Command
}

func NewAccountLockCommand(cli *Cli) *cobra.Command {
	c := new(AccountLockCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "lock",
		Short: "lock privateKey with keys and passcode.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return c.lock(ctx)
		},
	}
	return c.cmd
}

func (c *AccountLockCommand) lock(ctx context.Context) error {
	keypath := c.cli.RootOptions.Keys
	passcode := c.cli.RootOptions.Passcode
	data := &pb.AccountData{
		KeyPath:  keypath,
		PassCode: passcode,
		Header:   global.GHeader(),
	}
	_, err := c.cli.xclient.LockPrivateKey(ctx, data)
	if err == nil {
		fmt.Println("lock privateKey success")
	}
	return err
}

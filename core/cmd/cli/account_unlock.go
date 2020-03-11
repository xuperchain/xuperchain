package main

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/xuperchain/xuperchain/core/pb"
)

type AccountUnLockCommand struct {
	cli *Cli
	cmd *cobra.Command
}


func NewAccountUnLockCommand(cli *Cli) *cobra.Command {
	c := new(AccountUnLockCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "unlock",
		Short: "unlock privateKey with keys and passcode.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return c.unlock(ctx)
		},
	}
	return c.cmd
}

func (c *AccountUnLockCommand)unlock(ctx context.Context)error{
	keypath := c.cli.RootOptions.Keys
	passcode := c.cli.RootOptions.Passcode
	data := &pb.AccountData{
		KeyPath:  keypath,
		PassCode: passcode,
	}
	_, err := c.cli.xclient.UnLockPrivateKey(ctx, data)
	if err ==nil{
		fmt.Printf("unlock privateKey success,60s expired")
	}
	return err
}









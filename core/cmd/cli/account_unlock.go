package main

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/xuperchain/xuperchain/core/global"
	"github.com/xuperchain/xuperchain/core/pb"
	"strconv"
)

type AccountUnLockCommand struct {
	cli *Cli
	cmd *cobra.Command
	expiredTime int
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
	c.addFlags()
	return c.cmd
}

func (c *AccountUnLockCommand) addFlags(){
	c.cmd.Flags().IntVarP(&c.expiredTime,"expiredTime","e",60,"set time for unlock expired time, default 60 second,the unit is second")
}

func (c *AccountUnLockCommand)unlock(ctx context.Context)error{
	keypath := c.cli.RootOptions.Keys
	passcode := c.cli.RootOptions.Passcode
	expiredTime := c.expiredTime
	data := &pb.AccountData{
		KeyPath:  keypath,
		PassCode: passcode,
		ExpiredTime:int32(expiredTime),
		Header:     global.GHeader(),
	}
	_, err := c.cli.xclient.UnLockPrivateKey(ctx, data)
	if err ==nil{
		fmt.Println("unlock privateKey success,"+strconv.Itoa(expiredTime)+"s expired")
	}
	return err
}









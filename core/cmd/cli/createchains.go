/*
 * Copyright (c) 2019, Baidu.com, Inc. All Rights Reserved.
 */

package main

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"

	"github.com/xuperchain/xuperchain/core/common/config"
	"github.com/xuperchain/xuperchain/core/common/log"
	"github.com/xuperchain/xuperchain/core/contract/kernel"
)

// CreateChainCommand create chain cmd
type CreateChainCommand struct {
	cli          *Cli
	cmd          *cobra.Command
	ccRootConfig string
	ccOutput     string
}

// NewCreateChainVersion new create chain cmd
func NewCreateChainVersion(cli *Cli) *cobra.Command {
	c := new(CreateChainCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "createChain",
		Short: "Operate a blockchain: [OPTIONS].",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return c.createChain(ctx)
		},
	}
	c.addFlags()
	return c.cmd
}

func (c *CreateChainCommand) addFlags() {
	c.cmd.Flags().StringVar(&c.ccRootConfig, "rootconfig", "./data/config", "config for blockchain")
	c.cmd.Flags().StringVarP(&c.ccOutput, "output", "O", "./data/blockchain", "output path for blockchain")
}

func (c *CreateChainCommand) createChain(ctx context.Context) error {
	lcfg := config.LogConfig{
		Module:         "xchain",
		Filepath:       "logs",
		Filename:       "xchain",
		Fmt:            "logfmt",
		Console:        true,
		Level:          "debug",
		Async:          false,
		RotateInterval: 60,  // rotate every 60 minutes
		RotateBackups:  168, // keep old log files for 7 days
	}
	xlog, err := log.OpenLog(&lcfg)
	if err != nil {
		return err
	}
	k := kernel.Kernel{}
	k.Init(c.ccOutput, xlog, nil, "xuper")
	js := c.ccRootConfig + "/" + c.cli.RootOptions.Name + ".json"
	data, err := ioutil.ReadFile(js)
	if err != nil {
		fmt.Println("read file " + js + " error")
		return err
	}
	err = k.CreateBlockChain(c.cli.RootOptions.Name, data)
	if err != nil {
		fmt.Println("create block chain error " + c.cli.RootOptions.Name)
		return err
	}
	return nil
}

func init() {
	AddCommand(NewCreateChainVersion)
}

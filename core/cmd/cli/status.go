/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/xuperchain/xuperchain/core/global"
	"github.com/xuperchain/xuperchain/core/pb"
)

// StatusCommand status cmd
type StatusCommand struct {
	cli *Cli
	cmd *cobra.Command
}

// NewStatusCommand new status cmd
func NewStatusCommand(cli *Cli) *cobra.Command {
	s := new(StatusCommand)
	s.cli = cli
	s.cmd = &cobra.Command{
		Use:   "status",
		Short: "Operate a command to get status of current xchain server",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return s.printXchainStatus(ctx)
		},
	}
	return s.cmd
}

func (s *StatusCommand) printXchainStatus(ctx context.Context) error {
	client := s.cli.XchainClient()
	req := &pb.CommonIn{
		Header: global.GHeader(),
	}
	reply, err := client.GetSystemStatus(ctx, req)
	if err != nil {
		return err
	}
	if reply.Header.Error != pb.XChainErrorEnum_SUCCESS {
		return errors.New(reply.Header.Error.String())
	}
	status := FromSystemStatusPB(reply.GetSystemsStatus())
	output, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(output))
	return nil
}

func init() {
	AddCommand(NewStatusCommand)
}

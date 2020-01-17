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

// ContractStatDataQueryCommand contract statistic data query cmd
type ContractStatDataQueryCommand struct {
	cli *Cli
	cmd *cobra.Command
}

// NewContractStatDataQueryCommand new a command for ContractStatDataQueryCommand
func NewContractStatDataQueryCommand(cli *Cli) *cobra.Command {
	c := new(ContractStatDataQueryCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "query contract statistic data",
		Short: "query contract stat data based on bcname",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return c.queryContractStatData(ctx)
		},
	}
	c.addFlags()
	return c.cmd
}

func (c *ContractStatDataQueryCommand) addFlags() {
}

func (c *ContractStatDataQueryCommand) queryContractStatData(ctx context.Context) error {
	client := c.cli.XchainClient()
	request := &pb.ContractStatDataRequest{
		Header: &pb.Header{
			Logid: global.Glogid(),
		},
		Bcname: c.cli.RootOptions.Name,
	}
	reply, err := client.QueryContractStatData(ctx, request)
	if err != nil {
		return err
	}

	if reply.Header.Error != pb.XChainErrorEnum_SUCCESS {
		return errors.New(reply.Header.Error.String())
	}

	output, err := json.MarshalIndent(reply, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))

	return nil
}

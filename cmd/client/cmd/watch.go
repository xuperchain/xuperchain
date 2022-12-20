package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/golang/protobuf/proto"
	"github.com/spf13/cobra"

	"github.com/xuperchain/xuperchain/service/pb"
)

type watchCommand struct {
	cli *Cli
	cmd *cobra.Command

	filter      string
	oneLine     bool
	skipEmptyTx bool
}

func newWatchCommand(cli *Cli) *cobra.Command {
	c := new(watchCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "watch [options]",
		Short: "watch block event",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return c.watch(ctx)
		},
	}
	c.addFlags()
	return c.cmd
}

func (c *watchCommand) addFlags() {
	c.cmd.Flags().StringVarP(&c.filter, "filter", "f", "{}", "filter options")
	c.cmd.Flags().BoolVarP(&c.oneLine, "oneline", "", false, "whether print one event one line")
	c.cmd.Flags().BoolVarP(&c.skipEmptyTx, "skip-empty-tx", "", false, "whether print block with no tx matched")
}

func (c *watchCommand) watch(ctx context.Context) error {
	filter := &pb.BlockFilter{
		Bcname: c.cli.RootOptions.Name,
	}
	err := json.Unmarshal([]byte(c.filter), filter)
	if err != nil {
		return err
	}

	buf, _ := proto.Marshal(filter)
	request := &pb.SubscribeRequest{
		Type:   pb.SubscribeType_BLOCK,
		Filter: buf,
	}

	client := c.cli.EventClient()
	stream, err := client.Subscribe(ctx, request)
	if err != nil {
		return err
	}
	for {
		event, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		var block pb.FilteredBlock
		err = proto.Unmarshal(event.Payload, &block)
		if err != nil {
			return err
		}
		if len(block.GetTxs()) == 0 && c.skipEmptyTx {
			continue
		}
		c.printBlock(&block)
	}
}

func (c *watchCommand) printBlock(pbblock *pb.FilteredBlock) {
	block := FromFilteredBlockPB(pbblock)
	var buf []byte
	if c.oneLine {
		buf, _ = json.Marshal(block)
	} else {
		buf, _ = json.MarshalIndent(block, "", "  ")
	}
	fmt.Println(string(buf))
}

func init() {
	AddCommand(newWatchCommand)
}

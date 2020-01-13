/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/golang/protobuf/jsonpb"
	"github.com/spf13/cobra"

	"github.com/xuperchain/xuperchain/core/global"
	"github.com/xuperchain/xuperchain/core/pb"
)

// NativeStatusCommand native status cmd
type NativeStatusCommand struct {
	cli *Cli
	cmd *cobra.Command
}

// NewNativeStatusCommand new native status cmd
func NewNativeStatusCommand(cli *Cli) *cobra.Command {
	c := new(NativeStatusCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "status",
		Short: "[Deprecated] List status of a native contract.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return c.printStatus(ctx)
		},
	}
	return c.cmd
}

func (c *NativeStatusCommand) printStatus(ctx context.Context) error {
	request := &pb.NativeCodeStatusRequest{
		Header: global.GHeader(),
		Bcname: c.cli.RootOptions.Name,
	}
	m := make(map[string][]interface{})
	err := c.cli.RangeNodes(ctx, func(addr string, client pb.XchainClient, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s error:%s\n", addr, err)
			return nil
		}
		resp, err := client.NativeCodeStatus(ctx, request)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s error:%s\n", addr, err)
			return nil
		}
		if resp.Header.Error != pb.XChainErrorEnum_SUCCESS {
			fmt.Fprintf(os.Stderr, "%s error:%s\n", addr, resp.Header.Error)
			return nil
		}
		pbm := jsonpb.Marshaler{
			Indent:       "  ",
			EmitDefaults: true,
		}
		var list []interface{}
		for _, status := range resp.Status {
			buf := bytes.Buffer{}
			pbm.Marshal(&buf, status)
			var iface interface{}
			json.Unmarshal(buf.Bytes(), &iface)
			list = append(list, iface)
		}
		m[addr] = list
		return nil
	})
	if err != nil {
		return err
	}
	out, _ := json.MarshalIndent(m, "", "  ")
	fmt.Println(string(out))
	return nil
}

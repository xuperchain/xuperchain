/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/spf13/cobra"

	crypto_client "github.com/xuperchain/xuperchain/core/crypto/client"
	"github.com/xuperchain/xuperchain/core/crypto/hash"
	"github.com/xuperchain/xuperchain/core/global"
	"github.com/xuperchain/xuperchain/core/pb"
)

// NativeDeployCommand native deploy cmd
type NativeDeployCommand struct {
	cli *Cli
	cmd *cobra.Command

	name    string
	version string
	single  bool
}

// NewNativeDeployCommand new native deploy cmd
func NewNativeDeployCommand(cli *Cli) *cobra.Command {
	c := new(NativeDeployCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "deploy BINARY",
		Short: "[Deprecated] Deploy a native contract.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return c.deploy(ctx, args[0])
		},
	}
	c.addFlags()
	return c.cmd
}

func (c *NativeDeployCommand) addFlags() {
	c.cmd.Flags().StringVarP(&c.name, "cname", "n", "", "native code name")
	c.cmd.MarkFlagRequired("cname")
	c.cmd.Flags().StringVarP(&c.version, "version", "v", "", "native code version")
	c.cmd.MarkFlagRequired("version")
	c.cmd.Flags().BoolVarP(&c.single, "single", "s", false, "only deploy on a single node")
}

func (c *NativeDeployCommand) signDesc(desc *pb.NativeCodeDesc, privkey []byte) ([]byte, error) {
	// create crypto client
	cryptoClient, cryptoErr := crypto_client.CreateCryptoClientFromJSONPrivateKey(privkey)
	if cryptoErr != nil {
		fmt.Println("fail to create crypto client, err=", cryptoErr)
		return nil, cryptoErr
	}

	key, err := cryptoClient.GetEcdsaPrivateKeyFromJSON(privkey)
	if err != nil {
		return nil, err
	}
	buf, _ := proto.Marshal(desc)
	digest := hash.DoubleSha256(buf)

	return cryptoClient.SignECDSA(key, digest)
}

func (c *NativeDeployCommand) deploy(ctx context.Context, codepath string) error {
	content, err := ioutil.ReadFile(codepath)
	if err != nil {
		return err
	}
	desc := &pb.NativeCodeDesc{
		Name:            c.name,
		Version:         c.version,
		Digest:          hash.DoubleSha256(content),
		XuperApiVersion: 3,
	}

	pubkey, err := readPublicKey(c.cli.RootOptions.Keys)
	if err != nil {
		return err
	}
	privkey, err := readPrivateKey(c.cli.RootOptions.Keys)
	if err != nil {
		return err
	}
	address, err := readAddress(c.cli.RootOptions.Keys)
	if err != nil {
		return err
	}
	sign, err := c.signDesc(desc, []byte(privkey))
	if err != nil {
		return err
	}

	request := &pb.DeployNativeCodeRequest{
		Header:  global.GHeader(),
		Bcname:  c.cli.RootOptions.Name,
		Desc:    desc,
		Code:    content,
		Address: address,
		Pubkey:  []byte(pubkey),
		Sign:    sign,
	}
	if c.single {
		resp, err := c.cli.XchainClient().DeployNativeCode(ctx, request)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s error:%s\n", c.cli.RootOptions.Host, err)
			return nil
		}
		if resp.Header.Error != pb.XChainErrorEnum_SUCCESS {
			fmt.Fprintf(os.Stderr, "%s error:%s\n", c.cli.RootOptions.Host, resp.Header.Error)
			return nil
		}
		fmt.Printf("%s ok\n", c.cli.RootOptions.Host)
		return nil
	}

	err = c.cli.RangeNodes(ctx, func(addr string, client pb.XchainClient, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s error:%s\n", addr, err)
			return nil
		}
		resp, err := client.DeployNativeCode(ctx, request)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s error:%s\n", addr, err)
			return nil
		}
		if resp.Header.Error != pb.XChainErrorEnum_SUCCESS {
			fmt.Fprintf(os.Stderr, "%s error:%s\n", addr, resp.Header.Error)
			return nil
		}
		fmt.Printf("%s ok\n", addr)
		return nil
	})
	return err
}

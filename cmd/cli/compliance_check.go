package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	//"encoding/hex"
	"encoding/base64"

	"github.com/golang/protobuf/proto"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/xuperchain/xuperunion/global"
	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/utxo/txhash"
)

type GetSignCommand struct {
	cli          *Cli
	cmd          *cobra.Command
	xcheckclient pb.XcheckClient

	tx   string
	host string
}

func NewGetSignCommand(cli *Cli) *cobra.Command {
	c := new(GetSignCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "get",
		Short: "get a sign from remote node.",
		Long:  `./xchain-cli multisig get --tx ./tx.out`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return c.get(ctx)
		},
	}
	c.addFlags()
	/*
		err := c.initXcheckClient()
		if err != nil {
			fmt.Errorf("connect to Xcheck service failed, error ", err)
		}*/
	return c.cmd
}

func (c *GetSignCommand) initXcheckClient() error {
	fmt.Println("host: ", c.host)
	fmt.Println("tx: ", c.tx)
	conn, err := grpc.Dial(c.host, grpc.WithInsecure(), grpc.WithMaxMsgSize(64<<20-1))
	if err != nil {
		return err
	}
	c.xcheckclient = pb.NewXcheckClient(conn)
	return nil
}

func (c *GetSignCommand) XcheckClient() pb.XcheckClient {
	c.initXcheckClient()
	return c.xcheckclient
}

func (c *GetSignCommand) addFlags() {
	c.cmd.Flags().StringVar(&c.tx, "tx", "./tx.out", "Serialized transaction data file")
	c.cmd.Flags().StringVar(&c.host, "host", "localhost:6718", "host to get signature from compliance check service")
}

func (c *GetSignCommand) get(ctx context.Context) error {
	data, err := ioutil.ReadFile(c.tx)
	if err != nil {
		return errors.New("Fail to open serialized transaction data file")
	}
	tx := &pb.Transaction{}
	err = proto.Unmarshal(data, tx)
	if err != nil {
		return errors.New("Fail to Unmarshal proto")
	}

	tx.Txid, err = txhash.MakeTransactionID(tx)

	txStatus := &pb.TxStatus{
		Bcname: c.cli.RootOptions.Name,
		Status: pb.TransactionStatus_UNCONFIRM,
		Tx:     tx,
		Header: &pb.Header{
			Logid: global.Glogid(),
		},
		Txid: tx.Txid,
	}
	// XcheckClient
	reply, err2 := c.XcheckClient().ComplianceCheck(ctx, txStatus)
	if err2 != nil {
		fmt.Println("check here new XcheckClient error", err2)
		return err2
	}
	if reply.Header.Error != pb.XChainErrorEnum_SUCCESS {
		return fmt.Errorf("Failed to post tx:%s, logid:%s", reply.Header.Error.String(), reply.Header.Logid)
	}
	sign := reply.GetSignature()
	fmt.Println(sign.GetPublicKey())
	fmt.Println(base64.StdEncoding.EncodeToString(sign.GetSign()))
	return nil
}

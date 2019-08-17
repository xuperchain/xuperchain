package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/golang/protobuf/proto"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/xuperchain/xuperunion/global"
	"github.com/xuperchain/xuperunion/pb"
)

type GetComplianceCheckSignCommand struct {
	cli          *Cli
	cmd          *cobra.Command
	xcheckclient pb.XcheckClient

	tx     string
	output string
}

func NewGetComplianceCheckSignCommand(cli *Cli) *cobra.Command {
	c := new(GetComplianceCheckSignCommand)
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
	return c.cmd
}

func (c *GetComplianceCheckSignCommand) initXcheckClient() error {
	conn, err := grpc.Dial(c.cli.RootOptions.Host, grpc.WithInsecure(), grpc.WithMaxMsgSize(64<<20-1))
	if err != nil {
		return err
	}
	c.xcheckclient = pb.NewXcheckClient(conn)
	return nil
}

func (c *GetComplianceCheckSignCommand) XcheckClient() pb.XcheckClient {
	c.initXcheckClient()
	return c.xcheckclient
}

func (c *GetComplianceCheckSignCommand) addFlags() {
	c.cmd.Flags().StringVar(&c.tx, "tx", "./tx.out", "Serialized transaction data file")
	c.cmd.Flags().StringVar(&c.output, "output", "./compliance_check_sign.out", "Generate signature file for a transaction.")
}

func (c *GetComplianceCheckSignCommand) get(ctx context.Context) error {
	data, err := ioutil.ReadFile(c.tx)
	if err != nil {
		return errors.New("Fail to open serialized transaction data file")
	}
	tx := &pb.Transaction{}
	err = proto.Unmarshal(data, tx)
	if err != nil {
		return errors.New("Fail to Unmarshal proto")
	}

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
		return fmt.Errorf("Failed to get sign for tx:%s, logid:%s", reply.Header.Error.String(), reply.Header.Logid)
	}
	signInfo := reply.GetSignature()
	signJSON, err3 := json.MarshalIndent(signInfo, "", "  ")
	if err3 != nil {
		return err3
	}
	fmt.Println(string(signJSON))
	err3 = ioutil.WriteFile(c.output, signJSON, 0755)
	if err3 != nil {
		return errors.New("WriteFile error")
	}
	return nil
}

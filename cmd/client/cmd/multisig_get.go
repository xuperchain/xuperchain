package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"

	"github.com/golang/protobuf/proto" //nolint:staticcheck
	"github.com/spf13/cobra"
	"github.com/xuperchain/xupercore/bcs/ledger/xledger/state/utxo"
	"google.golang.org/grpc"

	"github.com/xuperchain/xuperchain/service/pb"
)

type GetComplianceCheckSignCommand struct {
	cli             *Cli
	cmd             *cobra.Command
	from            string
	xendorserclient pb.XendorserClient
	tx              string
	version         int32
	output          string
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

func (c *GetComplianceCheckSignCommand) initXEndorserClient() error {
	//nolint:staticcheck
	conn, err := grpc.Dial(c.cli.RootOptions.Host, grpc.WithInsecure(), grpc.WithMaxMsgSize(64<<20-1))
	if err != nil {
		return err
	}

	c.xendorserclient = pb.NewXendorserClient(conn)
	return nil
}

func (c *GetComplianceCheckSignCommand) XEndorserClient() pb.XendorserClient {
	_ = c.initXEndorserClient()
	return c.xendorserclient
}

func (c *GetComplianceCheckSignCommand) addFlags() {
	c.cmd.Flags().StringVar(&c.tx, "tx", "./tx.out", "Serialized transaction data file")
	c.cmd.Flags().StringVar(&c.output, "output", "./compliance_check_sign.out", "Generate signature file for a transaction.")
	c.cmd.Flags().Int32Var(&c.version, "txversion", utxo.TxVersion, "Tx version.")
	c.cmd.Flags().StringVar(&c.from, "from", "", "Initiator of an transaction.")
}

func (c *GetComplianceCheckSignCommand) get(ctx context.Context) error {
	data, err := os.ReadFile(c.tx)
	if err != nil {
		return errors.New("Fail to open serialized transaction data file")
	}
	tx := &pb.Transaction{}
	err = proto.Unmarshal(data, tx)
	if err != nil {
		return errors.New("Fail to Unmarshal proto")
	}

	ct := &CommTrans{
		FrozenHeight: 0,
		Version:      c.version,
		From:         c.from,
		Output:       c.output,
		ChainName:    c.cli.RootOptions.Name,
		Keys:         c.cli.RootOptions.Keys,
		XchainClient: c.cli.XchainClient(),
		CryptoType:   c.cli.RootOptions.Crypto,
		RootOptions:  c.cli.RootOptions,
	}

	var fromAddr string
	if c.from != "" {
		fromAddr = c.from
	} else {
		fromAddr, err = readAddress(ct.Keys)
		if err != nil {
			return err
		}
	}

	totalNeed := big.NewInt(0).SetInt64(int64(c.cli.RootOptions.ComplianceCheck.ComplianceCheckEndorseServiceFee))
	utxoInput := &pb.UtxoInput{
		Bcname:    ct.ChainName,
		Address:   fromAddr,
		TotalNeed: totalNeed.String(),
		NeedLock:  false,
	}
	//选取黄反检查需要的utxo
	utxoOutput, err := c.cli.XchainClient().SelectUTXO(ctx, utxoInput)
	if err != nil {
		fmt.Println("select utxo error", err)
		return err
	}

	//组装小费tx
	feeTx, err := ct.GenComplianceCheckTx(utxoOutput)
	if err != nil {
		fmt.Println("gen compliance check tx error", err)
		return err
	}

	txStatus := &pb.TxStatus{
		Bcname: c.cli.RootOptions.Name,
		Tx:     tx,
	}
	requestData, err := json.Marshal(txStatus)
	if err != nil {
		fmt.Printf("json encode txStatus failed: %v", err)
		return err
	}
	endorserRequest := &pb.EndorserRequest{
		RequestName: "ComplianceCheck",
		BcName:      c.cli.RootOptions.Name,
		Fee:         feeTx,
		RequestData: requestData,
	}
	// XEndorserClient
	reply, err := c.XEndorserClient().EndorserCall(ctx, endorserRequest)
	if err != nil {
		fmt.Println("check here new XendorserClient error", err)
		return err
	}
	if reply.Header.Error != pb.XChainErrorEnum_SUCCESS {
		return fmt.Errorf("Failed to get sign for tx:%s, logid:%s", reply.Header.Error.String(), reply.Header.Logid)
	}
	signInfo := reply.GetEndorserSign()
	signJSON, err3 := json.MarshalIndent(signInfo, "", "  ")
	if err3 != nil {
		return err3
	}
	fmt.Println(string(signJSON))
	err3 = os.WriteFile(c.output, signJSON, 0755)
	if err3 != nil {
		return errors.New("WriteFile error")
	}
	return nil
}

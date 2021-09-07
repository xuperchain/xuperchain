package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"

	"github.com/xuperchain/xuperchain/service/pb"
	"github.com/xuperchain/xupercore/bcs/ledger/xledger/state/utxo"
	"github.com/xuperchain/xupercore/lib/utils"
)

// 本文件封装了和共识模块有关的client调用接口, 具体格式为:
// xchain-cli consensus invoke 当前共识kernel调用
//   --type 标识共识名称，需符合当前共识状态
//   --method 标识共识方法，即调用的目标kernerl方法
//   --desc 标识输入参数，json格式
const (
	ModuleName = "xkernel"
)

type ConsensusInvokeCommand struct {
	cli *Cli
	cmd *cobra.Command

	module     string
	chain      string
	bucket     string
	method     string
	descfile   string
	account    string
	fee        string
	isMulti    bool
	multiAddrs string
	output     string
}

type invokeMethodFunc func(c *ConsensusInvokeCommand, ctx context.Context, ct *CommTrans) error

var invokeMap = map[string]invokeMethodFunc{
	"tdpos": (*ConsensusInvokeCommand).tdposInvoke,
	"xpos":  (*ConsensusInvokeCommand).tdposInvoke,
	"poa":   (*ConsensusInvokeCommand).xpoaInvoke,
	"xpoa":  (*ConsensusInvokeCommand).xpoaInvoke,
}

// NewConsensusCommand new consensus cmd
func NewConsensusInvokeCommand(cli *Cli) *cobra.Command {
	c := new(ConsensusInvokeCommand)
	c.cli = cli
	c.module = ModuleName
	c.cmd = &cobra.Command{
		Use:   "invoke",
		Short: "invoke consensus method",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return c.invoke(ctx)
		},
	}
	c.addFlags()
	return c.cmd
}

func (c *ConsensusInvokeCommand) addFlags() {
	c.cmd.Flags().StringVar(&c.descfile, "desc", "", "The json config file for consensus.")
	c.cmd.Flags().StringVarP(&c.bucket, "type", "t", "", "consensus bucket name")
	c.cmd.Flags().StringVarP(&c.method, "method", "", "", "kernel method name")
	c.cmd.Flags().StringVarP(&c.account, "account", "", "", "account name")
	c.cmd.Flags().StringVar(&c.fee, "fee", "", "fee of one tx")
	c.cmd.Flags().BoolVarP(&c.isMulti, "isMulti", "", false, "multisig scene")
	c.cmd.Flags().StringVarP(&c.multiAddrs, "multiAddrs", "A", "data/acl/addrs", "multiAddrs if multisig scene")
	c.cmd.Flags().StringVarP(&c.output, "output", "o", "./tx.out", "tx draw data")
}

func (c *ConsensusInvokeCommand) consensusRun(method invokeMethodFunc, ctx context.Context, ct *CommTrans) error {
	return method(c, ctx, ct)
}

func (c *ConsensusInvokeCommand) invoke(ctx context.Context) error {
	ct := &CommTrans{
		Version:      utxo.TxVersion,
		Amount:       "0",
		From:         c.account,
		ModuleName:   c.module,
		ContractName: "$" + c.bucket,
		MethodName:   c.method,
		Args:         make(map[string][]byte),
		MultiAddrs:   c.multiAddrs,
		IsQuick:      c.isMulti,
		ChainName:    c.cli.RootOptions.Name,
		Keys:         c.cli.RootOptions.Keys,
		XchainClient: c.cli.XchainClient(),
		CryptoType:   c.cli.RootOptions.Crypto,
		Fee:          c.fee,
		Output:       c.output,
	}
	if _, ok := invokeMap[c.bucket]; !ok {
		fmt.Printf("consensus type error.\n")
		return nil
	}
	// 所有共识命令统一输入一个当前Tipheight-3
	bcStatus := &pb.BCStatus{
		Header: &pb.Header{
			Logid: utils.GenLogId(),
		},
		Bcname: c.cli.RootOptions.Name,
	}
	status, err := c.cli.XchainClient().GetBlockChainStatus(ctx, bcStatus)
	if err != nil {
		return fmt.Errorf("get chain status error.\n")
	}
	ct.Args["height"] = []byte(fmt.Sprintf("%d", status.GetMeta().TrunkHeight-3))
	return c.consensusRun(invokeMap[c.bucket], ctx, ct)
}

func (c *ConsensusInvokeCommand) tdposInvoke(ctx context.Context, ct *CommTrans) error {
	if c.method == "getTdposInfos" {
		_, _, err := ct.GenPreExeRes(ctx)
		if err != nil {
			return err
		}
		return nil
	}
	// tdpos必须有input json数据
	if c.descfile == "" {
		return fmt.Errorf("tdpos needs desc file.\n")
	}
	args := make(map[string]interface{})
	desc, err := ioutil.ReadFile(c.descfile)
	if err != nil {
		return err
	}
	err = json.Unmarshal(desc, &args)
	if err != nil {
		return err
	}

	ct.Args, err = convertToXuper3Args(args)
	if err != nil {
		return err
	}
	ct.To, err = readAddress(ct.Keys)
	if err != nil {
		return err
	}
	if c.account == "" {
		initAk, _ := readAddress(c.cli.RootOptions.Keys)
		c.account = initAk
	}
	if c.isMulti { // 走多签
		return ct.GenerateMultisigGenRawTx(ctx)
	}
	return ct.Transfer(ctx)
}

func (c *ConsensusInvokeCommand) xpoaInvoke(ctx context.Context, ct *CommTrans) error {
	if c.method == "getValidates" {
		// TODO:: 读接口是否需要增加权限
		_, _, err := ct.GenPreExeRes(ctx)
		if err != nil {
			return err
		}
		return nil
	}
	if c.account == "" {
		return fmt.Errorf("xpoa needs acl account.\n")
	}
	ct.From = c.account
	// xpoa的account必须严格鉴权, 首先吊起acl访问
	client := c.cli.XchainClient()
	aclStatus := &pb.AclStatus{
		Header: &pb.Header{
			Logid: utils.GenLogId(),
		},
		Bcname:      c.cli.RootOptions.Name,
		AccountName: c.account,
	}
	reply, err := client.QueryACL(ctx, aclStatus)
	if err != nil {
		return err
	}
	args := map[string]interface{}{
		"height": string(ct.Args["height"]),
	}
	desc, err := ioutil.ReadFile(c.descfile)
	if err != nil {
		return err
	}
	err = json.Unmarshal(desc, &args)
	if err != nil {
		return err
	}
	// 此时填充acl信息
	acl := reply.GetAcl()
	if acl == nil {
		return fmt.Errorf("xpoa query acl error. pls check.\n")
	}
	aksB, err := json.Marshal(acl.AksWeight)
	if err != nil {
		return fmt.Errorf("xpoa query acl marshal error.\n")
	}
	args["aksWeight"] = string(aksB)
	if acl.Pm == nil {
		return fmt.Errorf("xpoa query acl error.\n")
	}
	args["acceptValue"] = fmt.Sprintf("%f", acl.Pm.AcceptValue)
	args["rule"] = fmt.Sprintf("%d", acl.Pm.Rule)
	ct.Args, err = convertToXuper3Args(args)
	if err != nil {
		return err
	}
	ct.To, err = readAddress(ct.Keys)
	if err != nil {
		return err
	}
	// xpoa必须走多签
	return ct.GenerateMultisigGenRawTx(ctx)
}

/*
 * Copyright (c) 2021. Baidu Inc. All Rights Reserved.
 */

package cmd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xuperchain/xupercore/lib/crypto/client"
	"github.com/xuperchain/xupercore/lib/crypto/client/base"
	"github.com/xuperchain/xupercore/lib/crypto/hash"
	"github.com/xuperchain/xupercore/lib/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/xuperchain/xuperchain/service/common"
	"github.com/xuperchain/xuperchain/service/pb"
)

// CommandFunc 代表了一个子命令，用于往Cli注册子命令
type CommandFunc func(c *Cli) *cobra.Command

var (
	// commands 用于收集所有的子命令，在启动的时候统一往Cli注册
	Commands []CommandFunc
)

// RootOptions 代表全局通用的flag，可以以嵌套结构体的方式组织flags.
type RootOptions struct {
	Host               string
	Name               string
	Keys               string
	Crypto             string
	Config             string
	TLS                TLSOptions            `yaml:"tls,omitempty"`
	EndorseServiceHost string                `yaml:"endorseServiceHost,omitempty"`
	ComplianceCheck    ComplianceCheckConfig `yaml:"complianceCheck,omitempty"`
	MinNewChainAmount  string                `yaml:"minNewChainAmount,omitempty"`
}

// Cli 是所有子命令执行的上下文.
type Cli struct {
	RootOptions RootOptions

	rootCmd *cobra.Command
	xclient pb.XchainClient

	eventClient pb.EventServiceClient
}

// NewCli new cli cmd
func NewCli() *Cli {
	rootCmd := &cobra.Command{
		Use:           "xchain-cli",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	return &Cli{
		rootCmd: rootCmd,
	}
}

func (c *Cli) SetVer(ver string) {
	c.rootCmd.Version = ver
}

func (c *Cli) initXchainClient() error {
	conn, err := grpc.Dial(c.RootOptions.Host, grpc.WithInsecure(), grpc.WithMaxMsgSize(64<<20-1))
	if err != nil {
		return err
	}
	c.xclient = pb.NewXchainClient(conn)
	c.eventClient = pb.NewEventServiceClient(conn)
	return nil
}

func (c *Cli) initFlags() error {
	// 参数设置优先级：1.命令行指定 2.配置文件指定 3.默认值
	// 加载配置文件
	var cfgFile string
	rootFlag := c.rootCmd.PersistentFlags()
	rootFlag.StringVarP(&cfgFile, "conf", "C", "./conf/xchain-cli.yaml", "client config file")
	c.RootOptions = NewRootOptions()

	// 设置命令行参数和默认值
	rootFlag.StringP("host", "H", c.RootOptions.Host, "server node ip:port")
	rootFlag.String("name", c.RootOptions.Name, "block chain name")
	rootFlag.String("keys", c.RootOptions.Keys, "directory of keys")
	rootFlag.String("crypto", c.RootOptions.Crypto, "crypto type")
	viper.BindPFlags(rootFlag)
	err := c.RootOptions.LoadConfig(cfgFile)
	if err != nil {
		fmt.Printf("load client config failed.config:%s err:%v\n", cfgFile, err)
		os.Exit(-1)
	}

	cobra.OnInitialize(func() {
		viper.Unmarshal(&c.RootOptions)
		err = c.initXchainClient()
		if err != nil {
			fmt.Printf("init xchain client failed.err:%v\n", err)
			os.Exit(-1)
		}
	})

	return nil
}

// Init cmd init entrance
func (c *Cli) Init() error {
	err := c.initFlags()
	if err != nil {
		return err
	}
	return nil
}

// AddCommands add sub commands
func (c *Cli) AddCommands(cmds []CommandFunc) {
	for _, cmd := range cmds {
		c.rootCmd.AddCommand(cmd(c))
	}
}

// Execute exe cmd
func (c *Cli) Execute() {
	err := c.rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

// XchainClient get xchain client
func (c *Cli) XchainClient() pb.XchainClient {
	return c.xclient
}

// EventClient get EventService client
func (c *Cli) EventClient() pb.EventServiceClient {
	return c.eventClient
}

// GetNodes get all nodes
func (c *Cli) GetNodes(ctx context.Context) ([]string, error) {
	req := &pb.CommonIn{
		Header: &pb.Header{
			Logid: utils.GenLogId(),
		},
	}
	reply, err := c.xclient.GetSystemStatus(ctx, req)
	if err != nil {
		return nil, err
	}
	if reply.Header.Error != pb.XChainErrorEnum_SUCCESS {
		return nil, errors.New(reply.Header.Error.String())
	}
	nodes := reply.GetSystemsStatus().GetPeerUrls()
	return nodes, nil
}

func genCreds(certPath, serverName string) (credentials.TransportCredentials, error) {
	bs, err := ioutil.ReadFile(certPath + "/cert.crt")

	if err != nil {
		return nil, err
	}
	certPool := x509.NewCertPool()
	ok := certPool.AppendCertsFromPEM(bs)
	if !ok {
		return nil, err
	}

	certificate, err := tls.LoadX509KeyPair(certPath+"/key.pem", certPath+"/private.key")
	if err != nil {
		return nil, err
	}
	creds := credentials.NewTLS(
		&tls.Config{
			ServerName:   serverName,
			Certificates: []tls.Certificate{certificate},
			RootCAs:      certPool,
			ClientCAs:    certPool,
			ClientAuth:   tls.RequireAndVerifyClientCert,
		})
	return creds, nil
}

// RangeNodes exe func in all nodes
func (c *Cli) RangeNodes(ctx context.Context, f func(addr string, client pb.XchainClient, err error) error) error {
	nodes, err := c.GetNodes(ctx)
	if err != nil {
		return err
	}
	options := []grpc.DialOption{grpc.WithMaxMsgSize(64<<20 - 1)}
	if c.RootOptions.TLS.Enable {
		cred, err := genCreds(c.RootOptions.TLS.Cert, c.RootOptions.TLS.Server)
		if err != nil {
			return err
		}
		options = append(options, grpc.WithTransportCredentials(cred))
	} else {
		options = append(options, grpc.WithInsecure())
	}

	for _, addr := range nodes {
		conn, err := grpc.Dial(addr, options...)
		if err != nil {
			err = f(addr, nil, err)
			if err != nil {
				return err
			}
		}
		client := pb.NewXchainClient(conn)
		err = f(addr, client, err)
		if err != nil {
			return err
		}
	}

	optionsRPC := []grpc.DialOption{grpc.WithMaxMsgSize(64<<20 - 1), grpc.WithInsecure()}
	conn, err := grpc.Dial(c.RootOptions.Host, optionsRPC...)
	if err != nil {
		return err
	}
	client := pb.NewXchainClient(conn)
	err = f(c.RootOptions.Host, client, err)
	if err != nil {
		return err
	}
	return nil
}

// Transfer transfer cli entrance
func (c *Cli) Transfer(ctx context.Context, opt *TransferOptions) (string, error) {

	from := newAK(opt.KeyPath)
	fromInfo, err := from.Info()
	if err != nil {
		return "", err
	}

	// create crypto client
	crypto, cryptoErr := client.CreateCryptoClient(opt.CryptoType)
	if cryptoErr != nil {
		fmt.Println("fail to create crypto client, err=", cryptoErr)
		return "", cryptoErr
	}

	return c.transfer(ctx, c.xclient, opt, fromInfo, crypto)
}

func (c *Cli) transfer(ctx context.Context, client pb.XchainClient, opt *TransferOptions, initiator AKInfo,
	cryptoClient base.CryptoClient) (string, error) {

	if opt.From == "" {
		opt.From = initiator.address
	}
	return c.transferSupportAccount(ctx, client, opt, initiator, cryptoClient)
}

func (c *Cli) transferSupportAccount(ctx context.Context, client pb.XchainClient, opt *TransferOptions,
	initiator AKInfo, cryptoClient base.CryptoClient) (string, error) {

	// 组装交易
	txStatus, err := assembleTxSupportAccount(ctx, client, opt, initiator, cryptoClient)
	if err != nil {
		return "", err
	}

	// 签名和生成 Tx ID
	signInfo, err := initiator.KeyPair.SignTx(txStatus.Tx, cryptoClient)
	if err != nil {
		return "", err
	}

	txStatus.Tx.InitiatorSigns = append(txStatus.Tx.InitiatorSigns, signInfo)
	txStatus.Tx.AuthRequireSigns, err = signTx(opt, cryptoClient, txStatus.Tx, initiator.KeyPair)
	if err != nil {
		return "", fmt.Errorf("Failed to signTx %s", err)
	}
	txStatus.Tx.Txid, err = common.MakeTxId(txStatus.Tx)
	if err != nil {
		return "", fmt.Errorf("Failed to gen txid %s", err)
	}
	txStatus.Txid = txStatus.Tx.Txid

	if opt.Debug {
		ttx := FromPBTx(txStatus.Tx)
		out, _ := json.MarshalIndent(ttx, "", "  ")
		fmt.Println(string(out))
		return hex.EncodeToString(txStatus.GetTxid()), nil
	}

	// 提交
	reply, err := client.PostTx(ctx, txStatus)
	if err != nil {
		return "", fmt.Errorf("transferSupportAccount post tx err %s", err)
	}
	if reply.Header.Error != pb.XChainErrorEnum_SUCCESS {
		return "", fmt.Errorf("Failed to post tx: %s", reply.Header.String())
	}
	return hex.EncodeToString(txStatus.GetTxid()), nil
}

func assembleTxSupportAccount(ctx context.Context, client pb.XchainClient, opt *TransferOptions, initiator AKInfo,
	cryptoClient base.CryptoClient) (*pb.TxStatus, error) {

	bigZero := big.NewInt(0)
	totalNeed := big.NewInt(0)
	tx := &pb.Transaction{
		Version:   opt.Version,
		Coinbase:  false,
		Desc:      opt.Desc,
		Nonce:     utils.GenNonce(),
		Timestamp: time.Now().UnixNano(),
		Initiator: initiator.address,
	}
	account := &pb.TxDataAccount{
		Address:      opt.To,
		Amount:       opt.Amount,
		FrozenHeight: opt.FrozenHeight,
	}
	accounts := []*pb.TxDataAccount{account}
	if opt.Fee != "" && opt.Fee != "0" {
		accounts = append(accounts, newFeeAccount(opt.Fee))
	}
	// 组装output
	for _, acc := range accounts {
		amount, ok := big.NewInt(0).SetString(acc.Amount, 10)
		if !ok {
			return nil, ErrInvalidAmount
		}
		if amount.Cmp(bigZero) < 0 {
			return nil, ErrNegativeAmount
		}
		totalNeed.Add(totalNeed, amount)
		txOutput := &pb.TxOutput{}
		txOutput.ToAddr = []byte(acc.Address)
		txOutput.Amount = amount.Bytes()
		txOutput.FrozenHeight = acc.FrozenHeight
		tx.TxOutputs = append(tx.TxOutputs, txOutput)
	}
	// 组装input 和 剩余output
	txInputs, deltaTxOutput, err := assembleTxInputsSupportAccount(ctx, client, opt, totalNeed, initiator, cryptoClient)
	if err != nil {
		return nil, err
	}
	tx.TxInputs = txInputs
	if deltaTxOutput != nil {
		tx.TxOutputs = append(tx.TxOutputs, deltaTxOutput)
	}
	// 设置auth require
	tx.AuthRequire, err = genAuthRequirement(opt.From, opt.AccountPath)
	if err != nil {
		return nil, err
	}

	preExeRPCReq := &pb.InvokeRPCRequest{
		Bcname:   opt.BlockchainName,
		Requests: []*pb.InvokeRequest{},
		Header: &pb.Header{
			Logid: utils.GenLogId(),
		},
		Initiator:   initiator.address,
		AuthRequire: tx.AuthRequire,
	}

	preExeRes, err := client.PreExec(ctx, preExeRPCReq)
	if err != nil {
		return nil, err
	}

	tx.ContractRequests = preExeRes.GetResponse().GetRequests()
	tx.TxInputsExt = preExeRes.GetResponse().GetInputs()
	tx.TxOutputsExt = preExeRes.GetResponse().GetOutputs()

	txStatus := &pb.TxStatus{
		Bcname: opt.BlockchainName,
		Status: pb.TransactionStatus_UNCONFIRM,
		Tx:     tx,
	}
	txStatus.Header = &pb.Header{
		Logid: utils.GenLogId(),
	}
	return txStatus, nil
}

// genAuthRequirement generates auth requirement for AK or Account
func genAuthRequirement(from, path string) (authRequirements []string, err error) {

	// generate auth requirement for AK: require self
	if path == "" {
		authRequirements = []string{from}
		return
	}

	// generate auth requirements for account: require all AK in account path
	// TODO: is that necessary for all AK
	akList, err := listAKs(path)
	if err != nil {
		return nil, err
	}
	for _, ak := range akList {
		authRequirement, err := ak.AuthRequirementFrom(from)
		if err != nil {
			return nil, fmt.Errorf("requires auth error: %s", err)
		}
		authRequirements = append(authRequirements, authRequirement)
	}
	return authRequirements, nil
}

// signTx generates Tx signatures for AK or Account
func signTx(opt *TransferOptions, crypto base.CryptoClient, tx *pb.Transaction,
	initiator KeyPair) ([]*pb.SignatureInfo, error) {

	// generate for AK
	if opt.AccountPath == "" {
		return signTxForAK(tx, initiator, crypto)
	}

	// generate for account
	return signTxForAccount(tx, opt.AccountPath, crypto)
}

// signTxForAK generates transaction signature for AK with its key pair
func signTxForAK(tx *pb.Transaction, keyPair KeyPair, crypto base.CryptoClient) ([]*pb.SignatureInfo, error) {
	sign, err := keyPair.SignTx(tx, crypto)
	if err != nil {
		return nil, err
	}
	return []*pb.SignatureInfo{sign}, nil
}

// signTxForAccount generates transaction signatures for account: signed by each AK in account path
// Params:
// 	tx: transaction
//	path: root path for accounts
// 	crypto: crypto client
// TODO: is that necessary for all AK
func signTxForAccount(tx *pb.Transaction, path string, crypto base.CryptoClient) ([]*pb.SignatureInfo, error) {
	signs := []*pb.SignatureInfo{}

	akList, err := listAKs(path)
	if err != nil {
		return nil, err
	}
	for _, ak := range akList {
		sign, err := ak.SignTx(tx, crypto)
		if err != nil {
			return nil, fmt.Errorf("sign error: %s", err)
		}
		signs = append(signs, sign)
	}
	return signs, nil
}

func assembleTxInputsSupportAccount(ctx context.Context, client pb.XchainClient, opt *TransferOptions,
	totalNeed *big.Int, initiator AKInfo, cryptoClient base.CryptoClient) ([]*pb.TxInput, *pb.TxOutput, error) {

	ui := &pb.UtxoInput{
		Bcname:    opt.BlockchainName,
		Address:   opt.From,
		TotalNeed: totalNeed.String(),
		NeedLock:  true,
		Publickey: initiator.publicKey,
	}

	sign, err := computeSelectUtxoSign(opt.BlockchainName, initiator.address, totalNeed.String(), initiator.secretKey,
		strconv.FormatBool(true), cryptoClient)
	if err != nil {
		return nil, nil, err
	}
	ui.UserSign = sign
	utxoRes, selectErr := client.SelectUTXO(ctx, ui)
	if selectErr != nil || utxoRes.Header.Error != pb.XChainErrorEnum_SUCCESS {
		return nil, nil, ErrSelectUtxo
	}
	var txTxInputs []*pb.TxInput
	var txOutput *pb.TxOutput
	for _, utxo := range utxoRes.UtxoList {
		txInput := new(pb.TxInput)
		txInput.RefTxid = utxo.RefTxid
		txInput.RefOffset = utxo.RefOffset
		txInput.FromAddr = utxo.ToAddr
		txInput.Amount = utxo.Amount
		txTxInputs = append(txTxInputs, txInput)
	}
	utxoTotal, ok := big.NewInt(0).SetString(utxoRes.TotalSelected, 10)
	if !ok {
		return nil, nil, ErrSelectUtxo
	}
	// 多出来的utxo需要再转给自己
	if utxoTotal.Cmp(totalNeed) > 0 {
		delta := utxoTotal.Sub(utxoTotal, totalNeed)
		txOutput = &pb.TxOutput{
			ToAddr: []byte(opt.From), // 收款人就是汇款人自己
			Amount: delta.Bytes(),
		}
	}
	return txTxInputs, txOutput, nil
}

func computeSelectUtxoSign(bcName, account, need, initScrKey, isLock string, cryptoClient base.CryptoClient) ([]byte, error) {
	privateKey, err := cryptoClient.GetEcdsaPrivateKeyFromJsonStr(initScrKey)
	if err != nil {
		return nil, err
	}

	hashStr := bcName + account + need + isLock
	doubleHash := hash.DoubleSha256([]byte(hashStr))
	signResult, err := cryptoClient.SignECDSA(privateKey, doubleHash)
	if err != nil {
		return nil, err
	}
	return signResult, nil
}

// AddCommand add sub cmd
func AddCommand(cmd CommandFunc) {
	Commands = append(Commands, cmd)
}

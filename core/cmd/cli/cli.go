/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	crypto_client "github.com/xuperchain/xuperunion/crypto/client"
	crypto_base "github.com/xuperchain/xuperunion/crypto/client/base"
	"github.com/xuperchain/xuperunion/global"
	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/utxo/txhash"
)

// CommandFunc 代表了一个子命令，用于往Cli注册子命令
type CommandFunc func(c *Cli) *cobra.Command

var (
	// commands 用于收集所有的子命令，在启动的时候统一往Cli注册
	commands []CommandFunc
)

// RootOptions 代表全局通用的flag，可以以嵌套结构体的方式组织flags.
type RootOptions struct {
	Host       string
	Name       string
	Keys       string
	TLS        TLSOptions
	CryptoType string
	Xuper3     bool
}

// TLSOptions TLS part
type TLSOptions struct {
	Cert   string
	Server string
	Enable bool
}

// Cli 是所有子命令执行的上下文.
type Cli struct {
	RootOptions RootOptions

	rootCmd *cobra.Command
	xclient pb.XchainClient
}

// NewCli new cli cmd
func NewCli() *Cli {
	rootCmd := &cobra.Command{
		Use:           "xchain-cli",
		SilenceErrors: true,
		SilenceUsage:  true,
		Version:       xchainVersion(),
	}
	return &Cli{
		rootCmd: rootCmd,
	}
}

func xchainVersion() string {
	return fmt.Sprintf("%s-%s %s", buildVersion, commitHash, buildDate)
}

func (c *Cli) initXchainClient() error {
	conn, err := grpc.Dial(c.RootOptions.Host, grpc.WithInsecure(), grpc.WithMaxMsgSize(64<<20-1))
	if err != nil {
		return err
	}
	c.xclient = pb.NewXchainClient(conn)
	return nil
}

func (c *Cli) initFlags() error {
	var cfgFile string
	rootFlags := c.rootCmd.PersistentFlags()
	rootFlags.StringVar(&cfgFile, "config", "", "config file (default is ./xchain.yaml)")
	rootFlags.StringP("host", "H", "127.0.0.1:37101", "server node ip:port")
	rootFlags.String("name", "xuper", "block chain name")
	rootFlags.String("keys", "data/keys", "directory of keys")
	rootFlags.String("cryptotype", crypto_client.CryptoTypeDefault, "crypto type, eg. default")
	viper.BindPFlags(rootFlags)

	cobra.OnInitialize(func() {
		if cfgFile != "" {
			viper.SetConfigFile(cfgFile)
		}
		viper.SetConfigName("xchain")
		viper.AddConfigPath(".")
		viper.AddConfigPath("./conf")
		viper.AddConfigPath(os.Getenv("HOME"))
		viper.ReadInConfig()
		// viper按照如下顺序查找一个flag key:
		// - pflag里面的被命令行显式设置的key
		// - 环境变量显式设置的
		// - 配置文件显式设置的
		// - KV存储的
		// - 通过viper设置的default flag
		// - 如果前面都没有变化，最后使用pflag的默认值
		// 所以在Unmarshal的时候命令行里面显式设置的flag会覆盖配置文件里面的flag
		// 如果配置文件没有这个flag，会用pflag的默认值
		//
		// 如果想使用嵌套struct的flag，则在设置pflag的flag name的时候需要使用如下的方式
		// rootFlags.String("topic.key", "", "")
		viper.Unmarshal(&c.RootOptions)

		err := c.initXchainClient()
		if err != nil {
			fmt.Printf("init xchain client:%s\n", err)
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

// GetNodes get all nodes
func (c *Cli) GetNodes(ctx context.Context) ([]string, error) {
	req := &pb.CommonIn{
		Header: global.GHeader(),
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
	fromAddr, err := readAddress(opt.KeyPath)
	if err != nil {
		return "", err
	}
	fromPubkey, err := readPublicKey(opt.KeyPath)
	if err != nil {
		return "", err
	}

	fromScrkey, err := readPrivateKey(opt.KeyPath)
	if err != nil {
		return "", err
	}

	// create crypto client
	cryptoClient, cryptoErr := crypto_client.CreateCryptoClient(opt.CryptoType)
	if cryptoErr != nil {
		fmt.Println("fail to create crypto client, err=", cryptoErr)
		return "", cryptoErr
	}

	return c.transfer(ctx, c.xclient, opt, fromAddr, fromPubkey, fromScrkey, cryptoClient)
}

func (c *Cli) transfer(ctx context.Context, client pb.XchainClient, opt *TransferOptions, fromAddr,
	fromPubkey, fromScrkey string, cryptoClient crypto_base.CryptoClient) (string, error) {
	if opt.From == "" {
		opt.From = fromAddr
	}
	return c.tansferSupportAccount(ctx, client, opt, fromAddr, fromPubkey, fromScrkey, cryptoClient)
}

func (c *Cli) tansferSupportAccount(ctx context.Context, client pb.XchainClient, opt *TransferOptions,
	initAddr, initPubkey, initScrkey string, cryptoClient crypto_base.CryptoClient) (string, error) {
	// 组装交易
	txStatus, err := assembleTxSupportAccount(ctx, client, opt, initAddr)
	if err != nil {
		return "", err
	}

	// 签名和生成txid
	signTx, err := txhash.ProcessSignTx(cryptoClient, txStatus.Tx, []byte(initScrkey))
	if err != nil {
		return "", err
	}
	signInfo := &pb.SignatureInfo{
		PublicKey: initPubkey,
		Sign:      signTx,
	}
	txStatus.Tx.InitiatorSigns = append(txStatus.Tx.InitiatorSigns, signInfo)
	txStatus.Tx.AuthRequireSigns, err = genAuthRequireSigns(opt, cryptoClient, txStatus.Tx, initScrkey, initPubkey)
	if err != nil {
		return "", fmt.Errorf("Failed to genAuthRequireSigns %s", err)
	}
	txStatus.Tx.Txid, err = txhash.MakeTransactionID(txStatus.Tx)
	if err != nil {
		return "", fmt.Errorf("Failed to gen txid %s", err)
	}
	txStatus.Txid = txStatus.Tx.Txid

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

func assembleTxSupportAccount(ctx context.Context, client pb.XchainClient, opt *TransferOptions, initAddr string) (*pb.TxStatus, error) {
	bigZero := big.NewInt(0)
	totalNeed := big.NewInt(0)
	tx := &pb.Transaction{
		Version:   opt.Version,
		Coinbase:  false,
		Desc:      opt.Desc,
		Nonce:     global.GenNonce(),
		Timestamp: time.Now().UnixNano(),
		Initiator: initAddr,
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
	txInputs, deltaTxOutput, err := assembleTxInputsSupportAccount(ctx, client, opt, totalNeed)
	if err != nil {
		return nil, err
	}
	tx.TxInputs = txInputs
	if deltaTxOutput != nil {
		tx.TxOutputs = append(tx.TxOutputs, deltaTxOutput)
	}
	// 设置auth require
	tx.AuthRequire, err = genAuthRequire(opt.From, opt.AccountPath)
	if err != nil {
		return nil, err
	}

	preExeRPCReq := &pb.InvokeRPCRequest{
		Bcname:      opt.BlockchainName,
		Requests:    []*pb.InvokeRequest{},
		Header:      global.GHeader(),
		Initiator:   initAddr,
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
		Logid: global.Glogid(),
	}
	return txStatus, nil
}

func genAuthRequire(from, path string) ([]string, error) {
	authRequire := []string{}
	if path == "" {
		authRequire = append(authRequire, from)
		return authRequire, nil
	}
	dir, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, fi := range dir {
		if fi.IsDir() {
			addr, err := readAddress(path + "/" + fi.Name())
			if err != nil {
				return nil, err
			}
			authRequire = append(authRequire, from+"/"+addr)
		}
	}
	return authRequire, nil
}

func genAuthRequireSigns(opt *TransferOptions, cryptoClient crypto_base.CryptoClient, tx *pb.Transaction, initScrkey, initPubkey string) ([]*pb.SignatureInfo, error) {
	authRequireSigns := []*pb.SignatureInfo{}
	if opt.AccountPath == "" {
		signTx, err := txhash.ProcessSignTx(cryptoClient, tx, []byte(initScrkey))
		if err != nil {
			return nil, err
		}
		signInfo := &pb.SignatureInfo{
			PublicKey: initPubkey,
			Sign:      signTx,
		}
		authRequireSigns = append(authRequireSigns, signInfo)
		return authRequireSigns, nil
	}

	dir, err := ioutil.ReadDir(opt.AccountPath)
	if err != nil {
		return nil, err
	}
	for _, fi := range dir {
		if fi.IsDir() {
			sk, err := readPrivateKey(opt.AccountPath + "/" + fi.Name())
			if err != nil {
				return nil, err
			}
			pk, err := readPublicKey(opt.AccountPath + "/" + fi.Name())
			if err != nil {
				return nil, err
			}
			signTx, err := txhash.ProcessSignTx(cryptoClient, tx, []byte(sk))
			if err != nil {
				return nil, err
			}
			signInfo := &pb.SignatureInfo{
				PublicKey: pk,
				Sign:      signTx,
			}
			authRequireSigns = append(authRequireSigns, signInfo)
		}
	}
	return authRequireSigns, nil
}

func assembleTxInputsSupportAccount(ctx context.Context, client pb.XchainClient, opt *TransferOptions, totalNeed *big.Int) ([]*pb.TxInput, *pb.TxOutput, error) {
	ui := &pb.UtxoInput{
		Bcname:    opt.BlockchainName,
		Address:   opt.From,
		TotalNeed: totalNeed.String(),
		NeedLock:  true,
	}
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

// AddCommand add sub cmd
func AddCommand(cmd CommandFunc) {
	commands = append(commands, cmd)
}

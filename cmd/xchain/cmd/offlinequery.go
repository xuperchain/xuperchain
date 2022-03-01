package cmd

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"unsafe"

	cmdcli "github.com/xuperchain/xuperchain/cmd/client/cmd"
	"github.com/xuperchain/xuperchain/service/pb"
	"github.com/xuperchain/xupercore/bcs/ledger/xledger/ledger"
	"github.com/xuperchain/xupercore/bcs/ledger/xledger/state"
	sctx "github.com/xuperchain/xupercore/bcs/ledger/xledger/state/context"
	"github.com/xuperchain/xupercore/kernel/common/xconfig"
	"github.com/xuperchain/xupercore/lib/crypto/client"
	"github.com/xuperchain/xupercore/lib/logs"
	"github.com/xuperchain/xupercore/lib/utils"
	xutils "github.com/xuperchain/xupercore/lib/utils"

	"github.com/spf13/cobra"
)

type OfflineQueryCommand struct {
	BaseCmd
	ChainName  string
	EnvConf    string
	CryptoType string
}

func GetOfflineQueryCommand() *OfflineQueryCommand {
	c := new(OfflineQueryCommand)
	c.Cmd = &cobra.Command{
		Use:   "offlineQuery",
		Short: "offline query ledger (Please stop node before offline query!)",
	}

	c.Cmd.PersistentFlags().StringVarP(&c.ChainName,
		"name", "n", "xuper", "block chain name")
	c.Cmd.PersistentFlags().StringVarP(&c.EnvConf,
		"env_conf", "e", "./conf/env.yaml", "env config file path")
	c.Cmd.PersistentFlags().StringVarP(&c.CryptoType,
		"crypto", "c", "default", "crypto type")

	c.Cmd.AddCommand(NewOfflineQueryBlockCommand(c).GetCmd())
	c.Cmd.AddCommand(NewOfflineQueryTxCommand(c).GetCmd())
	c.Cmd.AddCommand(NewOfflineQueryKVStoreCommand(c).GetCmd())

	return c
}

func (oq *OfflineQueryCommand) createLedgerAndStateHandle() (*ledger.Ledger, *state.State, error) {
	econf, err := createEnvConfig(oq.EnvConf)
	if err != nil {
		return nil, nil, err
	}

	return createHandle(econf, oq.ChainName, oq.CryptoType)
}

type OfflineQueryBlockCommand struct {
	BaseCmd
	root     *OfflineQueryCommand
	ByHeight bool
}

func NewOfflineQueryBlockCommand(root *OfflineQueryCommand) *OfflineQueryBlockCommand {
	c := new(OfflineQueryBlockCommand)
	c.root = root
	c.Cmd = &cobra.Command{
		Use:   "block",
		Short: "offline query block info (Please stop node before offline query!)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("expect blockId")
			}
			ctx := context.Background()
			if c.ByHeight {
				height, err := strconv.Atoi(args[0])
				if err != nil {
					return err
				}
				return c.QueryByBlockHeight(ctx, int64(height))
			}
			return c.QueryByBlockId(ctx, args[0])
		},
	}
	c.Cmd.Flags().BoolVarP(&c.ByHeight, "byHeight", "N", false, "Get block by height.")
	return c
}

func (oqb *OfflineQueryBlockCommand) QueryByBlockId(ctx context.Context, blockId string) error {
	fmt.Println("query block by blockId", blockId)
	ledgerHandle, stateHandle, err := oqb.root.createLedgerAndStateHandle()
	if err != nil {
		return err
	}
	defer ledgerHandle.Close()
	defer stateHandle.Close()

	targetBlockId, err := hex.DecodeString(blockId)
	if err != nil {
		return err
	}
	targetBlock, err := ledgerHandle.QueryBlock(targetBlockId)
	if err != nil {
		fmt.Printf("query block by blockId %s failed.", blockId)
		return err
	}
	// attention：这里利用指针强转，必须保持前后结构内容一致
	output, err := json.MarshalIndent(cmdcli.FromInternalBlockPB((*pb.InternalBlock)(unsafe.Pointer(targetBlock))), "", "  ")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(output))
	return nil
}

func (oqb *OfflineQueryBlockCommand) QueryByBlockHeight(ctx context.Context, height int64) error {
	fmt.Println("query block by height", height)
	ledgerHandle, stateHandle, err := oqb.root.createLedgerAndStateHandle()
	if err != nil {
		return err
	}
	defer ledgerHandle.Close()
	defer stateHandle.Close()

	targetBlock, err := ledgerHandle.QueryBlockByHeight(height)
	if err != nil {
		fmt.Printf("query block by height %d failed.", height)
		return err
	}
	// attention：这里利用指针强转，必须保持前后结构内容一致
	output, err := json.MarshalIndent(cmdcli.FromInternalBlockPB((*pb.InternalBlock)(unsafe.Pointer(targetBlock))), "", "  ")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(output))
	return nil
}

type OfflineQueryTxCommand struct {
	BaseCmd
	root *OfflineQueryCommand
	TxId string
}

func NewOfflineQueryTxCommand(root *OfflineQueryCommand) *OfflineQueryTxCommand {
	c := new(OfflineQueryTxCommand)
	c.root = root
	c.Cmd = &cobra.Command{
		Use:   "tx",
		Short: "offline query tx info (Please stop node before offline query!)",
	}
	c.Cmd.AddCommand(&cobra.Command{
		Use:   "query",
		Short: "query transaction by TxId",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("expect TxId")
			}
			ctx := context.TODO()
			return c.QueryByTxId(ctx, args[0])
		},
	})
	return c
}

func (oqt *OfflineQueryTxCommand) QueryByTxId(ctx context.Context, txid string) error {
	fmt.Println("query tx by txid", txid)
	ledgerHandle, stateHandle, err := oqt.root.createLedgerAndStateHandle()
	if err != nil {
		return err
	}
	defer ledgerHandle.Close()
	defer stateHandle.Close()

	targetTxId, err := hex.DecodeString(txid)
	if err != nil {
		fmt.Println(err)
		return err
	}

	targetTxInfo, has, err := stateHandle.QueryTx(targetTxId)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if !has {
		fmt.Printf("txid:%s not found", txid)
		return nil
	}

	// attention：这里利用指针强转，必须保持前后结构内容一致
	output, err := json.MarshalIndent(cmdcli.FromPBTx((*pb.Transaction)(unsafe.Pointer(targetTxInfo))), "", "  ")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(output))
	return nil
}

type OfflineQueryKVStoreCommand struct {
	BaseCmd
	root   *OfflineQueryCommand
	Bucket string
	Key    string
	Height int64

	DecodeType string
}

func NewOfflineQueryKVStoreCommand(root *OfflineQueryCommand) *OfflineQueryKVStoreCommand {
	c := new(OfflineQueryKVStoreCommand)
	c.root = root
	c.Cmd = &cobra.Command{
		Use:   "get",
		Short: "offline query KVStore info (Please stop node before offline query!)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if c.Bucket == "" {
				return errors.New("expect bucket")
			}
			if c.Key == "" {
				return errors.New("expect key")
			}
			ctx := context.Background()
			return c.Get(ctx, c.Bucket, c.Key, c.Height)
		},
	}
	c.Cmd.Flags().StringVarP(&c.Bucket, "bucket", "b", "", "bucket space in kvstore")
	c.Cmd.Flags().StringVarP(&c.Key, "key", "k", "", "key in kvstore")
	c.Cmd.Flags().Int64VarP(&c.Height, "height", "N", 0, "snapshoot query by height. The default value is 0, query the latest value.")
	c.Cmd.Flags().StringVarP(&c.DecodeType, "decode", "d", "native", "val decode type. [native|json|hex]")
	return c
}

func (oqkv *OfflineQueryKVStoreCommand) Get(ctx context.Context, bucket string, key string, height int64) error {
	fmt.Printf("query kvstore info bucket=%s key=%s height=%d\n", bucket, key, height)

	ledgerHandle, stateHandle, err := oqkv.root.createLedgerAndStateHandle()
	if err != nil {
		return err
	}
	defer ledgerHandle.Close()
	defer stateHandle.Close()

	var val []byte
	if height > 0 {
		// 在指定高度进行快照读取操作
		block, errQuery := ledgerHandle.QueryBlockHeaderByHeight(height)
		if errQuery != nil {
			fmt.Println(errQuery)
			return errQuery
		}
		readerSnapshot, errCreate := stateHandle.CreateXMSnapshotReader(block.Blockid)
		if errCreate != nil {
			fmt.Println(errCreate)
			return errCreate
		}

		val, err = readerSnapshot.Get(bucket, []byte(key))
		if err != nil {
			fmt.Println(err)
			return err
		}
	} else {
		// 读取最新值
		reader := stateHandle.CreateXMReader()
		versionedData, errGet := reader.Get(bucket, []byte(key))
		if errGet != nil {
			fmt.Println(errGet)
			return errGet
		}
		val = versionedData.GetPureData().GetValue()
	}

	if len(val) == 0 {
		fmt.Println("val is nil")
		return nil
	}

	// 结果值转换
	switch oqkv.DecodeType {
	case "json":
		jsonVal := map[string]interface{}{}
		err = json.Unmarshal(val, &jsonVal)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(jsonVal)
	case "hex":
		fmt.Println(hex.EncodeToString(val))
	case "native":
		fmt.Println(string(val))
	default:
		fmt.Println(string(val))
	}
	return nil
}

func createEnvConfig(path string) (*xconfig.EnvConf, error) {
	if !xutils.FileIsExist(path) {
		fmt.Printf("config file not exist.env_conf:%s\n", path)
		return nil, fmt.Errorf("config file not exist")
	}

	econf, err := xconfig.LoadEnvConf(path)
	if err != nil {
		fmt.Printf("load env config failed.env_conf:%s err:%v\n", path, err)
		return nil, fmt.Errorf("load env config failed")
	}
	return econf, nil
}

func createHandle(econf *xconfig.EnvConf, chainName string, cryptoType string) (*ledger.Ledger, *state.State, error) {
	logs.InitLog(econf.GenConfFilePath(econf.LogConf), econf.GenDirAbsPath(econf.LogDir))
	lctx, err := ledger.NewLedgerCtx(econf, chainName)
	if err != nil {
		return nil, nil, err
	}

	ledgerPath := lctx.EnvCfg.GenDataAbsPath(lctx.EnvCfg.ChainDir)
	ledgerPath = filepath.Join(ledgerPath, lctx.BCName)
	if !utils.PathExists(ledgerPath) {
		return nil, nil, errors.New("invalid name:" + lctx.BCName)
	}

	ledgerHandle, err := ledger.OpenLedger(lctx)
	if err != nil {
		return nil, nil, err
	}
	crypt, err := client.CreateCryptoClient(cryptoType)
	if err != nil {
		return nil, nil, err
	}
	ctx, err := sctx.NewStateCtx(econf, chainName, ledgerHandle, crypt)
	if err != nil {
		return nil, nil, err
	}
	stateHandle, err := state.NewState(ctx)
	if err != nil {
		return nil, nil, err
	}

	return ledgerHandle, stateHandle, nil
}

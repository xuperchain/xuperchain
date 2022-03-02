package cmd

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/xuperchain/xupercore/lib/utils"
)

func getOfflineQueryCommand() *OfflineQueryCommand {
	c := GetOfflineQueryCommand()
	curPath := utils.GetCurFileDir()
	c.ChainName = "xuper"
	c.EnvConf = filepath.Join(curPath, "./mockLedger/conf/env.yaml")
	c.CryptoType = "default"
	c.RootPath = filepath.Join(curPath, "./mockLedger")
	return c
}

func TestNewOfflineQueryStatusCommand(t *testing.T) {
	cmd := NewOfflineQueryStatusCommand(getOfflineQueryCommand())
	if err := cmd.QueryStatus(context.Background()); err != nil {
		t.Error(err)
	}
}

func TestNewOfflineQueryBlockCommand(t *testing.T) {
	cmd := NewOfflineQueryBlockCommand(getOfflineQueryCommand())

	if err := cmd.QueryByBlockId(context.Background(), "8f5998b6340103f3635f9a2f3bb8c9882ec9e1d1b4817eb93ac8871726bed08b"); err != nil {
		t.Error(err)
	}

	if err := cmd.QueryByBlockHeight(context.Background(), 11); err != nil {
		t.Error(err)
	}
}

func TestNewOfflineQueryTxCommand(t *testing.T) {
	cmd := NewOfflineQueryTxCommand(getOfflineQueryCommand())

	if err := cmd.QueryByTxId(context.Background(), "6afe271528b2a283130d7b814c1bfe2df986c3695247a491170059a6e4495127"); err != nil {
		t.Error(err)
	}
}

func TestNewOfflineQueryKVStoreCommand(t *testing.T) {
	cmd := NewOfflineQueryKVStoreCommand(getOfflineQueryCommand())

	// 查询最新值
	if err := cmd.Get(context.Background(), "governToken", []byte("balanceOf_TeyyPLpp9L7QAcxHangtcHTu7HUZ6iydY"), -1); err != nil {
		t.Error(err)
	}

	// 按照区块高度10进行快照查询，创建治理代币交易在高度8的区块中，这里可以读取成功
	if err := cmd.Get(context.Background(), "governToken", []byte("balanceOf_TeyyPLpp9L7QAcxHangtcHTu7HUZ6iydY"), 10); err != nil {
		t.Error(err)
	}

	// 按照区块高度7进行快照查询，创建治理代币交易在高度8的区块中，这里读取结果为nil
	if err := cmd.Get(context.Background(), "governToken", []byte("balanceOf_TeyyPLpp9L7QAcxHangtcHTu7HUZ6iydY"), 7); err != nil {
		t.Error(err)
	}
}

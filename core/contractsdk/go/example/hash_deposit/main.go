package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/xuperchain/xuperchain/core/contractsdk/go/code"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/driver"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/utils"
)

const (
	UserBucket = "USER_"
	HashBucket = "HASH_"
)

type hashDeposit struct {
}
type fileInfo struct {
	UserID   string `json:"user_id" validte:"required"`
	HashID   string `json:"hash_id" validte:"required"`
	FileName string `json:"file_name" validte:"required"`
}

func (hd *hashDeposit) Initialize(ctx code.Context) code.Response {
	return code.OK([]byte("ok"))
}

func (hd *hashDeposit) StoreFileInfo(ctx code.Context) code.Response {
	args := fileInfo{}
	err := utils.Unmarshal(ctx.Args(), &args)
	if err != nil {
		return code.Error(err)
	}

	userKey := UserBucket + args.UserID + "/" + args.HashID
	hashKey := HashBucket + args.HashID

	value, _ := json.Marshal(args)

	if _, err = ctx.GetObject([]byte(hashKey)); err == nil {
		return code.Error(fmt.Errorf("hash id %s already exists\n", args.HashID))
	}
	if err := ctx.PutObject([]byte(userKey), []byte(value)); err != nil {
		return code.Error(err)
	}

	if err := ctx.PutObject([]byte(hashKey), []byte(value)); err != nil {
		return code.Error(err)
	}
	return code.OK([]byte(value))
}

func (hd *hashDeposit) QueryUserList(ctx code.Context) code.Response {
	prefix := UserBucket
	iter := ctx.NewIterator(code.PrefixRange([]byte(prefix)))
	defer iter.Close()
	users := []string{}
	for iter.Next() {
		userKey := string(iter.Key()[len(UserBucket):])
		users = append(users, strings.Split(userKey, "/")[0])
	}
	if err := iter.Error(); err != nil {
		return code.Error(err)
	}
	// TODO
	return code.JSON(users)
}

func (hd *hashDeposit) QueryFileInfoByUser(ctx code.Context) code.Response {
	args := struct {
		UserID string `json:"user_id" validte:"required"`
	}{}
	if err := utils.Unmarshal(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}

	prefix := UserBucket + args.UserID + "/"
	iter := ctx.NewIterator(code.PrefixRange([]byte(prefix)))
	defer iter.Close()

	fileInfos := []fileInfo{}
	for iter.Next() {
		info := fileInfo{}
		if err := json.Unmarshal(iter.Value(), &info); err != nil {
			return code.Error(err)
		}
		fileInfos = append(fileInfos, info)
	}
	if err := iter.Error(); err != nil {
		return code.Error(err)
	}
	return code.JSON(fileInfos)
}

func (hd *hashDeposit) QueryFileInfoByHash(ctx code.Context) code.Response {
	args := struct {
		HashID string `json:"hash_id" validte:"required"`
	}{}
	if err := utils.Unmarshal(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}
	key := HashBucket + args.HashID
	value, err := ctx.GetObject([]byte(key))
	if err != nil {
		return code.Error(err)
	}
	return code.OK(value)
}

func main() {
	driver.Serve(new(hashDeposit))
}

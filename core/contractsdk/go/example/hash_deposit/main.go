package main

import (
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

func (hd *hashDeposit) Initialize(ctx code.Context) code.Response {
	return code.OK([]byte("ok~"))
}

func (hd *hashDeposit) StoreFileInfo(ctx code.Context) code.Response {
	args := struct {
		UsedID   string `json:"user_id" required:"true"`
		HashID   string `json:"hash_id" required:"true"`
		FileName string `json:"file_name" required:"true"`
	}{}
	err := utils.Validate(ctx.Args(), &args)
	if err != nil {
		return code.Error(err)
	}

	userKey := UserBucket + args.UsedID + "/" + args.HashID
	hashKey := HashBucket + args.HashID
	value := args.UsedID + "\t" + args.HashID + "\t" + args.FileName

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
	key := UserBucket
	iter := ctx.NewIterator([]byte(key), []byte(key+"~"))
	defer iter.Close()
	builder := strings.Builder{}
	for iter.Next() {
		userKey := string(iter.Key()[len(UserBucket):])
		builder.WriteString(strings.Split(userKey, "/")[0])
		builder.WriteString("\n")
	}
	if err := iter.Error(); err != nil {
		return code.Error(err)
	}
	return code.OK([]byte(builder.String()))
}

func (hd *hashDeposit) QueryFileInfoByUser(ctx code.Context) code.Response {
	args := struct {
		UserID string `json:"user_id" required:"true"`
	}{}
	if err := utils.Validate(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}

	builder := strings.Builder{}

	start := UserBucket + args.UserID
	end := start + "~"
	iter := ctx.NewIterator([]byte(start), []byte(end))

	defer iter.Close()
	for iter.Next() {
		builder.Write(iter.Value())
		builder.WriteString("\n")
	}
	if err := iter.Error(); err != nil {
		return code.Error(err)
	}
	return code.OK([]byte(builder.String()))
}

func (hd *hashDeposit) QueryFileInfoByHash(ctx code.Context) code.Response {
	args := struct {
		HashID string `json:"hash_id" required:"true"`
	}{}
	if err := utils.Validate(ctx.Args(), &args); err != nil {
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

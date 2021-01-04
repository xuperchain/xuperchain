package main

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/xuperchain/xuperchain/core/contractsdk/go/code"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/driver"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/utils"
)

const (
	GOODS          = "GOODS_"
	GOODSRECORD    = "GOODSSRECORD_"
	GOODSRECORDTOP = "GOODSSRECORDTOP_"
	CREATE         = "CREATE"
	ADMIN          = "ADMIN"
)

type sourceTrace struct {
}

func (st *sourceTrace) Initialize(ctx code.Context) code.Response {
	args := struct {
		Admin string `json:"admin" required:"true"`
	}{}
	if err := utils.Validate(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}
	if err := ctx.PutObject([]byte(ADMIN), []byte(args.Admin)); err != nil {
		return code.Error(err)
	}

	return code.OK([]byte("ok"))
}

func (st *sourceTrace) CreateGoods(ctx code.Context) code.Response {
	caller := ctx.Initiator()
	if caller == "" {
		return code.Error(utils.ErrMissingCaller)
	}

	admin, err := ctx.GetObject([]byte(ADMIN))
	if err != nil {
		return code.Error(err)
	}

	if string(admin) != caller {
		return code.Error(utils.ErrPermissionDenied)
	}

	args := struct {
		Id   string `json:"id" required:"true''"`
		Desc string `json:"desc" required:"desc"`
	}{}
	if err := utils.Validate(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}
	goodsKey := GOODS + args.Id

	if _, err := ctx.GetObject([]byte(goodsKey)); err == nil {
		return code.Error(fmt.Errorf("goods %s already exists", args.Id))
	}

	if err := ctx.PutObject([]byte(goodsKey), []byte(args.Desc)); err != nil {
		return code.Error(err)
	}

	goodsRecordsKey := GOODSRECORD + args.Id + "_0"
	goodsRecordsTopKey := GOODSRECORDTOP + args.Id
	if err := ctx.PutObject([]byte(goodsRecordsKey), []byte(CREATE)); err != nil {
		return code.Error(err)
	}
	if err := ctx.PutObject([]byte(goodsRecordsTopKey), []byte("0")); err != nil {
		return code.Error(err)
	}
	return code.OK([]byte(args.Id))

}

func (st *sourceTrace) UpdateGoods(ctx code.Context) code.Response {
	caller := ctx.Initiator()
	if caller == "" {
		return code.Error(utils.ErrMissingCaller)
	}

	admin, err := ctx.GetObject([]byte(ADMIN))
	if err != nil {
		return code.Error(err)
	}

	if string(admin) != caller {
		return code.Error(utils.ErrPermissionDenied)
	}
	args := struct {
		Id     string `json:"id" required:"true"`
		Reason string `json:"reason" required:"true"`
	}{}

	if err := utils.Validate(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}

	topRecordByte, err := ctx.GetObject([]byte(GOODSRECORDTOP + args.Id))
	if err != nil {
		return code.Error(err)
	}
	topRecord, _ := big.NewInt(0).SetString(string(topRecordByte), 10)
	topRecord = topRecord.Add(topRecord, big.NewInt(1))

	if err := ctx.PutObject([]byte(GOODSRECORD+args.Id+"_"+topRecord.String()), []byte(args.Reason)); err != nil {
		return code.Error(err)
	}
	if err := ctx.PutObject([]byte(GOODSRECORDTOP+args.Id), []byte(topRecord.String())); err != nil {
		return code.Error(err)
	}
	return code.OK([]byte(topRecord.String()))
}

func (st *sourceTrace) QueryRecords(ctx code.Context) code.Response {
	args := struct {
		Id string `json:"id" required:"true"`
	}{}
	if err := utils.Validate(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}

	goodsRecordsKey := GOODSRECORD + args.Id + "_"
	start := goodsRecordsKey
	end := start + "~"
	iter := ctx.NewIterator([]byte(start), []byte(end))
	defer iter.Close()

	buf := strings.Builder{}
	for iter.Next() {
		goodsRecord := string(iter.Key())[len(goodsRecordsKey):]
		reason := iter.Value()

		buf.WriteString("updateRecord=")
		buf.WriteString(goodsRecord)
		buf.WriteString(",reason=")
		buf.Write(reason)
		buf.WriteString(("\n"))
	}
	if err := iter.Error(); err != nil {
		return code.Error(err)
	}
	return code.OK([]byte(buf.String()))
}

func main() {
	driver.Serve(new(sourceTrace))
}

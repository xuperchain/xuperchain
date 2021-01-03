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
	ADMIN          = "admin"
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
		return code.Error(fmt.Errorf("goods type %s already exists", goodsKey[len(GOODS):]))
	}

	if err := ctx.PutObject([]byte(goodsKey), []byte(args.Desc)); err != nil {
		return code.Error(err)
	}

	goodsRecordsKey := GOODSRECORD + args.Id + "_0"
	goodsRecordsTopKey := GOODSRECORDTOP + args.Id
	if err := ctx.PutObject([]byte(goodsRecordsKey), []byte(CREATE)); err != nil {
		return code.Error(err)
	}
	value := []byte("0")
	if err := ctx.PutObject([]byte(goodsRecordsTopKey), value); err != nil {
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
	topRecord, _ := big.NewFloat(0).SetString(string(topRecordByte))
	topRecord = topRecord.Add(topRecord, big.NewFloat(1))

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
	value, err := ctx.GetObject([]byte(GOODS + args.Id))
	_ = value // TODO @fengjin
	if err != nil {
		return code.Error(err)
	}
	goodsRecordsKey := GOODSRECORD + args.Id + "_"
	start := goodsRecordsKey
	end := start + "~"
	iter := ctx.NewIterator([]byte(start), []byte(end))
	defer iter.Close()
	buf := strings.Builder{}
	for iter.Next() {
		goodsRecord := string(iter.Key())[len(GOODSRECORD):] //TODO @fengjin 确认下标没问题
		pos := strings.Index(goodsRecord, "_")
		goodsId := goodsRecord[:pos]
		updateRecord := goodsRecord[pos+1:]
		reason := iter.Value()
		buf.WriteString("goodsId=")
		buf.WriteString(goodsId)
		buf.WriteString(",updateRecord=")
		buf.WriteString(updateRecord)
		buf.WriteString(",reason=")
		buf.Write(reason)
	}
	if err := iter.Error(); err != nil {
		return code.Error(err)
	}
	return code.OK([]byte(buf.String()))
}

func main() {
	driver.Serve(new(sourceTrace))
}

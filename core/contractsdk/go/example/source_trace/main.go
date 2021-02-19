package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/xuperchain/xuperchain/core/contractsdk/go/code"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/driver"
)

const (
	GOODS          = "GOODS_"
	GOODSRECORD    = "GOODSSRECORD_"
	GOODSRECORDTOP = "GOODSSRECORDTOP_"
	CREATE         = "CREATE"
	ADMIN          = "ADMIN"
)

type updateRecord struct {
	UpdateReccord string `json:"update_record'`
	Reason        string `json:"reason"`
}

type sourceTrace struct {
}

func (st *sourceTrace) Initialize(ctx code.Context) code.Response {
	args := struct {
		Admin string `json:"admin" validte:"required"`
	}{}
	if err := code.Unmarshal(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}
	if err := ctx.PutObject([]byte(ADMIN), []byte(args.Admin)); err != nil {
		return code.Error(err)
	}

	return code.OK([]byte("ok"))
}

func (st *sourceTrace) CreateGoods(ctx code.Context) code.Response {
	initiator := ctx.Initiator()
	if initiator == "" {
		return code.Error(code.ErrMissingInitiator)
	}

	admin, err := ctx.GetObject([]byte(ADMIN))
	if err != nil {
		return code.Error(err)
	}

	if string(admin) != initiator {
		return code.Error(code.ErrPermissionDenied)
	}

	args := struct {
		Id   string `json:"id" required:"true''"`
		Desc string `json:"desc" required:"desc"`
	}{}
	if err := code.Unmarshal(ctx.Args(), &args); err != nil {
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
	initiator := ctx.Initiator()
	if initiator == "" {
		return code.Error(code.ErrMissingInitiator)
	}

	admin, err := ctx.GetObject([]byte(ADMIN))
	if err != nil {
		return code.Error(err)
	}

	if string(admin) != initiator {
		return code.Error(code.ErrPermissionDenied)
	}
	args := struct {
		Id     string `json:"id" validte:"required"`
		Reason string `json:"reason" validte:"required"`
	}{}

	if err := code.Unmarshal(ctx.Args(), &args); err != nil {
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
		Id string `json:"id" validte:"required"`
	}{}
	if err := code.Unmarshal(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}
	goodsKey := GOODS + args.Id

	if _, err := ctx.GetObject([]byte(goodsKey)); err != nil {
		return code.Error(errors.New("goods not found"))
	}

	goodsRecordsKey := GOODSRECORD + args.Id + "_"
	iter := ctx.NewIterator(code.PrefixRange([]byte(goodsRecordsKey)))
	defer iter.Close()
	records := []updateRecord{}

	for iter.Next() {
		goodsRecord := string(iter.Key())[len(goodsRecordsKey):]
		reason := iter.Value()
		records = append(records, updateRecord{
			goodsRecord,
			string(reason),
		})
	}
	if err := iter.Error(); err != nil {
		return code.Error(err)
	}
	recordsByte, _ := json.Marshal(records)

	return code.OK(recordsByte)
}

func main() {
	driver.Serve(new(sourceTrace))
}

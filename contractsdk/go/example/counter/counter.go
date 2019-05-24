package main

import (
	"strconv"

	"github.com/xuperchain/xuperunion/contractsdk/go/code"
	"github.com/xuperchain/xuperunion/contractsdk/go/driver"
)

type counter struct{}

func (c *counter) Initialize(ctx code.Context) code.Response {
	creator, ok := ctx.Args()["creator"].(string)
	if !ok {
		return code.Errors("missing creator")
	}
	err := ctx.PutObject([]byte("creator"), []byte(creator))
	if err != nil {
		return code.Error(err)
	}
	return code.OK(nil)
}

func (c *counter) Increase(ctx code.Context) code.Response {
	key, ok := ctx.Args()["key"].(string)
	if !ok {
		return code.Errors("missing key")
	}
	value, err := ctx.GetObject([]byte(key))
	cnt := 0
	if err == nil {
		cnt, _ = strconv.Atoi(string(value))
	}

	cntstr := strconv.Itoa(cnt + 1)

	err = ctx.PutObject([]byte(key), []byte(cntstr))
	if err != nil {
		return code.Error(err)
	}
	return code.Response{
		Status:  200,
		Message: cntstr,
	}
}

func (c *counter) Get(ctx code.Context) code.Response {
	key, ok := ctx.Args()["key"].(string)
	if !ok {
		return code.Errors("missing key")
	}
	value, err := ctx.GetObject([]byte(key))
	if err != nil {
		return code.Error(err)
	}
	return code.Response{
		Status:  200,
		Message: string(value),
	}
}

func main() {
	driver.Serve(new(counter))
}

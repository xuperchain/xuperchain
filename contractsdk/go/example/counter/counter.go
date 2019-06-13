package main

import (
	"strconv"

	"github.com/xuperchain/xuperunion/contractsdk/go/code"
	"github.com/xuperchain/xuperunion/contractsdk/go/driver"
)

type counter struct{}

func (c *counter) Initialize(ctx code.Context) code.Response {
	creator, ok := ctx.Args()["creator"]
	if !ok {
		return code.Errors("missing creator")
	}
	err := ctx.PutObject([]byte("creator"), creator)
	if err != nil {
		return code.Error(err)
	}
	return code.OK(nil)
}

func (c *counter) Increase(ctx code.Context) code.Response {
	key, ok := ctx.Args()["key"]
	if !ok {
		return code.Errors("missing key")
	}
	value, err := ctx.GetObject(key)
	cnt := 0
	if err == nil {
		cnt, _ = strconv.Atoi(string(value))
	}

	cntstr := strconv.Itoa(cnt + 1)

	err = ctx.PutObject(key, []byte(cntstr))
	if err != nil {
		return code.Error(err)
	}
	return code.OK([]byte(cntstr))
}

func (c *counter) Get(ctx code.Context) code.Response {
	key, ok := ctx.Args()["key"]
	if !ok {
		return code.Errors("missing key")
	}
	value, err := ctx.GetObject(key)
	if err != nil {
		return code.Error(err)
	}
	return code.OK(value)
}

func main() {
	driver.Serve(new(counter))
}

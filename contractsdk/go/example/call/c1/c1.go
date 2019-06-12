package main

import (
	"math/big"
	"strconv"

	"github.com/xuperchain/xuperunion/contractsdk/go/code"
	"github.com/xuperchain/xuperunion/contractsdk/go/driver"
)

type c1 struct{}

func (c *c1) Initialize(ctx code.Context) code.Response {
	return code.OK(nil)
}

func (c *c1) Invoke(ctx code.Context) code.Response {
	// 获取cnt变量
	var cnt int
	cntstr, _ := ctx.GetObject([]byte("cnt"))
	if cntstr != nil {
		cnt, _ = strconv.Atoi(string(cntstr))
	}

	// 发起转账
	args := ctx.Args()
	toaddr := string(args["to"])
	amount := big.NewInt(1)
	err := ctx.Transfer(toaddr, amount)
	if err != nil {
		return code.Error(err)
	}

	// 发起跨合约调用
	callArgs := map[string][]byte{
		"to": []byte(toaddr),
	}
	resp, err := ctx.Call("wasm", "c2", "invoke", callArgs)
	if err != nil {
		return code.Error(err)
	}
	if code.IsStatusError(resp.Status) {
		return *resp
	}

	// 根据合约调用结果记录到call变量里面并持久化
	err = ctx.PutObject([]byte("call"), resp.Body)
	if err != nil {
		return code.Error(err)
	}

	// 对cnt变量加1并持久化
	cnt = cnt + 1
	err = ctx.PutObject([]byte("cnt"), []byte(strconv.Itoa(cnt)))
	if err != nil {
		return code.Error(err)
	}

	cntstr = []byte(strconv.Itoa(cnt))
	return code.Response{
		Status:  200,
		Message: string(cntstr) + ":" + string(resp.Body),
		Body:    cntstr,
	}
}

func main() {
	driver.Serve(new(c1))
}

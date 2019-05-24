package main

import (
	"math/big"
	"strconv"

	"github.com/xuperchain/xuperunion/contractsdk/go/code"
	"github.com/xuperchain/xuperunion/contractsdk/go/driver"
)

type c2 struct{}

func (c *c2) Initialize(ctx code.Context) code.Response {
	return code.OK(nil)
}

func (c *c2) Invoke(ctx code.Context) code.Response {
	var cnt int
	cntstr, _ := ctx.GetObject([]byte("cnt"))
	if cntstr != nil {
		cnt, _ = strconv.Atoi(string(cntstr))
	}
	cnt += 1000

	args := ctx.Args()
	toaddr := args["to"].(string)
	amount := big.NewInt(1000)
	err := ctx.Transfer(toaddr, amount)
	if err != nil {
		return code.Error(err)
	}
	err = ctx.PutObject([]byte("cnt"), []byte(strconv.Itoa(cnt)))
	if err != nil {
		return code.Error(err)
	}

	cntstr = []byte(strconv.Itoa(cnt))
	return code.Response{
		Status:  200,
		Message: string(cntstr),
		Body:    cntstr,
	}
}

func (c *c2) Query(ctx code.Context) code.Response {
	keys, ok := ctx.Args()["keys"].([]interface{})
	if !ok {
		return code.Errors("argument keys not found")
	}
	result := make(map[string]string)
	for _, ikey := range keys {
		key := ikey.(string)
		value, _ := ctx.GetObject([]byte(key))
		result[key] = string(value)
	}
	return code.JSON(result)
}

func main() {
	driver.Serve(new(c2))
}

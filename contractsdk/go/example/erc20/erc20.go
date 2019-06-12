package main

import (
	"errors"
	"math/big"

	"github.com/xuperchain/xuperunion/contractsdk/go/code"
	"github.com/xuperchain/xuperunion/contractsdk/go/driver"
)

const (
	typeInt = "I"
	typeMap = "M"
)

type xdb struct {
	ctx   code.Context
	dirty map[string]*big.Int
}

func newXDB() *xdb {
	return &xdb{
		dirty: make(map[string]*big.Int),
	}
}

func (x *xdb) Int(name string) *big.Int {
	key := typeInt + name
	n, ok := x.dirty[key]
	if ok {
		return n
	}
	n = big.NewInt(0)
	x.dirty[key] = n
	return n
}

func (x *xdb) Map(name string) *xmap {
	return &xmap{
		db:   x,
		name: name,
	}
}

func (x *xdb) SetContext(ctx code.Context) {
	x.ctx = ctx
	for key, n := range x.dirty {
		value, err := ctx.GetObject([]byte(key))
		if err == nil {
			n.SetString(string(value), 10)
		}
	}
}

func (x *xdb) Commit() error {
	for key, n := range x.dirty {
		err := x.ctx.PutObject([]byte(key), []byte(n.String()))
		if err != nil {
			return err
		}
	}
	return nil
}

type xmap struct {
	db   *xdb
	name string
}

func (x *xmap) Get(mkey string) *big.Int {
	key := typeMap + x.name + "\x00" + mkey
	n, ok := x.db.dirty[key]
	if ok {
		return n
	}
	value, err := x.db.ctx.GetObject([]byte(key))
	n = big.NewInt(0)
	if err == nil {
		_, ok := n.SetString(string(value), 10)
		if !ok {
			panic(string(value))
		}
	}
	x.db.dirty[key] = n
	return n
}

type erc20 struct {
	xdb         *xdb
	totalSupply *big.Int
	balanceOf   *xmap
	allowance   *xmap
}

func newERC20() *erc20 {
	xdb := newXDB()
	return &erc20{
		xdb:         xdb,
		totalSupply: xdb.Int("totalSupply"),
		balanceOf:   xdb.Map("balanceOf"),
		allowance:   xdb.Map("allowance"),
	}
}

func (e *erc20) Initialize(ctx code.Context) code.Response {
	initSupplyStr := string(ctx.Args()["initSupply"])
	if initSupplyStr == "" {
		return code.Errors("missing initSupply")
	}
	caller := ctx.Caller()
	initSupply, ok := big.NewInt(0).SetString(initSupplyStr, 10)
	if !ok {
		return code.Errors("bad initSupply number")
	}

	if initSupply.Cmp(big.NewInt(0)) <= 0 {
		return code.Errors("amount must bigger than 0")
	}

	e.xdb.SetContext(ctx)
	e.totalSupply.Set(initSupply)
	e.balanceOf.Get(caller).Set(initSupply)

	err := e.xdb.Commit()
	if err != nil {
		return code.Error(err)
	}
	return code.OK(nil)
}

func (e *erc20) transfer(from, to string, amount *big.Int) error {
	if amount.Cmp(big.NewInt(0)) <= 0 {
		return errors.New("amount must bigger than 0")
	}
	fromAmount := e.balanceOf.Get(from)
	toAmount := e.balanceOf.Get(to)
	if fromAmount.Cmp(amount) < 0 {
		return errors.New("balance of from less than amount")
	}
	fromAmount.Sub(fromAmount, amount)
	toAmount.Add(toAmount, amount)
	return nil
}

func (e *erc20) Transfer(ctx code.Context) code.Response {
	from := ctx.Caller()
	to := string(ctx.Args()["to"])
	if to == "" {
		return code.Errors("missing to argument")
	}
	amountstr := string(ctx.Args()["amount"])
	if amountstr == "" {
		return code.Errors("missing amount argument")
	}

	amount, ok := big.NewInt(0).SetString(amountstr, 10)
	if !ok {
		return code.Errors("bad amount number")
	}

	err := e.transfer(from, to, amount)
	if err != nil {
		return code.Error(err)
	}
	return code.OK(nil)
}

// TransferFrom(from string, to string, amount string)
func (e *erc20) TransferFrom(ctx code.Context) code.Response {
	caller := ctx.Caller()
	from := string(ctx.Args()["from"])
	if from == "" {
		return code.Errors("missing from argument")
	}

	to := string(ctx.Args()["to"])
	if to == "" {
		return code.Errors("missing to argument")
	}
	amountstr := string(ctx.Args()["amount"])
	if amountstr == "" {
		return code.Errors("missing amount argument")
	}

	amount, ok := big.NewInt(0).SetString(amountstr, 10)
	if !ok {
		return code.Errors("bad amount number")
	}

	allowance := e.allowance.Get(from + "_" + caller)
	if allowance.Cmp(amount) < 0 {
		return code.Errors("allowance less than amount")
	}

	err := e.transfer(from, to, amount)
	if err != nil {
		return code.Error(err)
	}
	allowance.Sub(allowance, amount)
	return code.OK(nil)
}

// Approve(spender string, amount string)
func (e *erc20) Approve(ctx code.Context) code.Response {
	caller := ctx.Caller()
	spender := string(ctx.Args()["spender"])
	if spender == "" {
		return code.Errors("missing spender argument")
	}
	amountstr := string(ctx.Args()["amount"])
	if amountstr == "" {
		return code.Errors("missing amount argument")
	}

	amount := e.allowance.Get(caller + "_" + spender)
	_, ok := amount.SetString(amountstr, 10)
	if !ok {
		return code.Errors("bad amount number")
	}

	return code.OK(nil)
}

// Allowance(owner string, spender string)
func (e *erc20) Allowance(ctx code.Context) code.Response {
	spender := string(ctx.Args()["spender"])
	if spender == "" {
		return code.Errors("missing spender argument")
	}
	owner := string(ctx.Args()["owner"])
	if owner == "" {
		return code.Errors("missing owner argument")
	}

	amount := e.allowance.Get(owner + "_" + spender)
	return code.OK([]byte(amount.String()))
}

func (e *erc20) Invoke(ctx code.Context) code.Response {
	var resp code.Response
	e.xdb.SetContext(ctx)
	action := string(ctx.Args()["action"])
	if action == "" {
		return code.Errors("missing action")
	}
	switch action {
	case "transfer":
		resp = e.Transfer(ctx)
	case "transferFrom":
		resp = e.TransferFrom(ctx)
	case "approve":
		resp = e.Approve(ctx)
	default:
		resp = code.Errors("bad action " + action)
	}
	if code.IsStatusError(resp.Status) {
		return resp
	}
	err := e.xdb.Commit()
	if err != nil {
		return code.Error(err)
	}
	return resp
}

func (e *erc20) Balance(ctx code.Context) code.Response {
	address := string(ctx.Args()["address"])
	if address == "" {
		return code.Errors("missing address argument")
	}
	amount := e.balanceOf.Get(address)
	return code.OK([]byte(amount.String()))
}

func (e *erc20) TotalSupply(ctx code.Context) code.Response {
	return code.OK([]byte(e.totalSupply.String()))
}

func (e *erc20) Query(ctx code.Context) code.Response {
	e.xdb.SetContext(ctx)
	action := string(ctx.Args()["action"])
	if action == "" {
		return code.Errors("missing action")
	}
	switch action {
	case "balanceOf":
		return e.Balance(ctx)
	case "totalSupply":
		return e.TotalSupply(ctx)
	case "allowance":
		return e.Allowance(ctx)
	default:
		return code.Errors("bad action " + action)
	}
}

func main() {
	driver.Serve(newERC20())
}

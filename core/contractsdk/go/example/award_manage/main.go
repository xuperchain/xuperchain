package main

import (
	"errors"
	"math/big"

	"github.com/xuperchain/xuperchain/core/contractsdk/go/code"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/driver"
)

const (
	BALANCEPRE   = "balanceOf/"
	ALLOWANCEPRE = "allowanceOf/"
	MASTERPRE    = "admin"
	TOTAL_SUPPLY = "TotalSupply"
)

type awardManage struct{}

func (am *awardManage) Initialize(ctx code.Context) code.Response {
	args := struct {
		TotalSupply *big.Int `json:"total_supply" validate:"gt=0"`
	}{}
	if err := code.Unmarshal(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}

	initiator := ctx.Initiator()
	if len(initiator) == 0 {
		return code.Error(errors.New("missing initiator"))
	}
	err := ctx.PutObject([]byte(BALANCEPRE+initiator), []byte(args.TotalSupply.String()))
	if err != nil {
		return code.Error(err)
	}

	if err := ctx.PutObject([]byte(TOTAL_SUPPLY), []byte(args.TotalSupply.String())); err != nil {
		return code.Error(err)
	}

	if err := ctx.PutObject([]byte(MASTERPRE), []byte(initiator)); err != nil {
		return code.Error(err)
	}

	return code.OK([]byte(args.TotalSupply.String()))
}

func (am *awardManage) AddAward(ctx code.Context) code.Response {
	initiator := ctx.Initiator()
	if initiator == "" {
		return code.Error(code.ErrMissingInitiator)
	}
	masterBytes, err := ctx.GetObject([]byte(MASTERPRE))
	if err != nil {
		return code.Error(err)
	}
	master := string(masterBytes)

	if master != initiator {
		return code.Error(code.ErrPermissionDenied)
	}

	args := struct {
		Amount *big.Int `json:"amount" validate:"gt=0"`
	}{}

	if err := code.Unmarshal(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}

	totalSupplyByte, err := ctx.GetObject([]byte(TOTAL_SUPPLY))
	if err != nil {
		return code.Error(err)
	}

	totalSupply, _ := big.NewInt(0).SetString(string(totalSupplyByte), 10)
	totalSupply = big.NewInt(0).Add(totalSupply, args.Amount)

	if err := ctx.PutObject([]byte(TOTAL_SUPPLY), []byte(totalSupply.String())); err != nil {
		return code.Error(err)
	}
	return code.OK([]byte(totalSupply.String()))
}

func (am *awardManage) TotalSupply(ctx code.Context) code.Response {
	if value, err := ctx.GetObject([]byte(TOTAL_SUPPLY)); err != nil {
		return code.Error(err)
	} else {
		return code.OK(value)
	}
}

func (am *awardManage) Balance(ctx code.Context) code.Response {
	initiator := ctx.Initiator()
	if initiator == "" {
		return code.Error(code.ErrMissingInitiator)
	}
	value, err := ctx.GetObject([]byte(BALANCEPRE + initiator))
	if err != nil {
		return code.Error(err)
	}
	return code.OK(value)
}

func (am *awardManage) Allowance(ctx code.Context) code.Response {
	args := struct {
		From string `json:"from,excludes=/"`
		To   string `json:"to,excludes=/"`
	}{}
	if err := code.Unmarshal(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}

	value, err := ctx.GetObject([]byte(ALLOWANCEPRE + args.From + "/" + args.To))
	if err != nil {
		return code.Error(err)
	}
	return code.OK(value)
}

func (am *awardManage) Transfer(ctx code.Context) code.Response {
	initiator := ctx.Initiator()
	from := initiator
	if from == "" {
		return code.Error(code.ErrMissingInitiator)
	}
	args := struct {
		To    string   `json:"to" validate:"required,excludes=/"`
		Token *big.Int `json:"token" validate:"required"`
	}{}

	if err := code.Unmarshal(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}
	if from == args.To {
		return code.Error(errors.New("can not transfer to yourself"))
	}

	fromBalanceByte, err := ctx.GetObject([]byte(BALANCEPRE + from))
	if err != nil {
		return code.Error(err)
	}
	fromBalance, _ := big.NewInt(0).SetString(string(fromBalanceByte), 10)

	if fromBalance.Cmp(args.Token) < 0 {
		return code.Error(errors.New("balance not enough"))
	}

	toBalanceByte, err := ctx.GetObject([]byte(BALANCEPRE + args.To))
	if err != nil { // errors!=nil means account not found
		toBalanceByte = []byte("0")
	}
	toBalance, _ := big.NewInt(0).SetString(string(toBalanceByte), 10)

	if err := ctx.PutObject([]byte(BALANCEPRE+from), []byte(big.NewInt(0).Sub(fromBalance, args.Token).String())); err != nil {
		return code.Error(err)
	}
	if err := ctx.PutObject([]byte(BALANCEPRE+args.To), []byte(big.NewInt(0).Add(toBalance, args.Token).String())); err != nil {
		return code.Error(err)
	}

	return code.OK([]byte("ok"))
}

func (am *awardManage) TransferFrom(ctx code.Context) code.Response {
	args := struct {
		From  string   `json:"from" validate:"required,excludes=/"`
		Token *big.Int `json:"token" validate:"required"`
	}{}
	initiator := ctx.Initiator()
	if initiator == "" {
		return code.Error(code.ErrMissingInitiator)
	}
	if err := code.Unmarshal(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}

	fromBalanceByte, err := ctx.GetObject([]byte(BALANCEPRE + args.From))
	if err != nil {
		return code.Error(err)
	}

	fromBalance, _ := big.NewInt(0).SetString(string(fromBalanceByte), 10)
	if fromBalance.Cmp(args.Token) < 0 {
		return code.Error(code.ErrBalanceLow)
	}

	allowanceKey := ALLOWANCEPRE + args.From + "/" + initiator

	allowanceBalanceByte, err := ctx.GetObject([]byte(allowanceKey))
	if err != nil {
		return code.Error(err)
	}
	allowanceBalance, _ := big.NewInt(0).SetString(string(allowanceBalanceByte), 10)
	if allowanceBalance.Cmp(args.Token) < 0 {
		return code.Error(errors.New("allowance balance not enough"))
	}

	toBalance := big.NewInt(0)
	toBalanceByte, err := ctx.GetObject([]byte(BALANCEPRE + initiator))
	if err == nil {
		toBalance.SetString(string(toBalanceByte), 10)
	}

	fromBalance = fromBalance.Sub(fromBalance, args.Token)
	toBalance = toBalance.Add(toBalance, args.Token)
	allowanceBalance = allowanceBalance.Sub(allowanceBalance, args.Token)

	if err := ctx.PutObject([]byte(allowanceKey), []byte(allowanceBalance.String())); err != nil {
		return code.Error(err)
	}
	if err := ctx.PutObject([]byte(BALANCEPRE+args.From), []byte(fromBalance.String())); err != nil {
		return code.Error(err)
	}
	if err := ctx.PutObject([]byte(BALANCEPRE+initiator), []byte(toBalance.String())); err != nil {
		return code.Error(err)
	}
	return code.OK([]byte("ok"))
}

func (am *awardManage) Approve(ctx code.Context) code.Response {
	args := struct {
		To    string   `json:"to" validate:"required,excludes=/"`
		Token *big.Int `json:"token" validate:"required"`
	}{}
	from := ctx.Initiator()
	if len(from) == 0 {
		return code.Error(code.ErrMissingInitiator)
	}

	if err := code.Unmarshal(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}

	allowanceKey := []byte(ALLOWANCEPRE + from + "/" + args.To)

	allowanceByte := []byte("0")
	if value, err := ctx.GetObject(allowanceKey); err == nil {
		allowanceByte = value
	}

	allowanceBalance, _ := big.NewInt(0).SetString(string(allowanceByte), 10)
	allowanceBalance = allowanceBalance.Add(allowanceBalance, args.Token)
	if err := ctx.PutObject(allowanceKey, []byte(allowanceBalance.String())); err != nil {
		return code.Error(err)
	}
	return code.OK([]byte("ok"))
}

func main() {
	driver.Serve(new(awardManage))
}

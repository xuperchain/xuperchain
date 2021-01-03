package main

import (
	"errors"
	"math/big"

	"github.com/xuperchain/xuperchain/core/contractsdk/go/code"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/driver"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/utils"
)

const (
	BALANCEPRE   = "balanceOf_"
	ALLOWANCEPRE = "allowanceOf_"
	MASTERPRE    = "owner"
	TOTAL_SUPPLY = "TotalSupply"
)

type awardManage struct{}

func (am *awardManage) Initialize(ctx code.Context) code.Response {
	args := struct {
		TotalSupply *big.Int `json:"totalSupply" gt:"0"`
	}{}
	if err := utils.Validate(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}

	caller := ctx.Initiator()
	if len(caller) == 0 {
		return code.Error(errors.New("missing caller"))
	}
	err := ctx.PutObject([]byte(BALANCEPRE+caller), []byte(args.TotalSupply.String()))
	if err != nil {
		return code.Error(err)
	}

	if err := ctx.PutObject([]byte(TOTAL_SUPPLY), []byte(args.TotalSupply.String())); err != nil {
		return code.Error(err)
	}

	if err := ctx.PutObject([]byte(MASTERPRE), []byte(caller)); err != nil {
		return code.Error(err)
	}

	return code.OK([]byte(args.TotalSupply.String()))
}

func (am *awardManage) AddAward(ctx code.Context) code.Response {
	caller := ctx.Caller()
	if caller == "" {
		return code.Error(utils.ErrMissingCaller)
	}
	masterBytes, err := ctx.GetObject([]byte(MASTERPRE))
	if err != nil {
		return code.Error(err)
	}
	master := string(masterBytes)

	if master != caller {
		return code.Error(utils.ErrPermissionDenied)
	}

	args := struct {
		Amount *big.Float `json:"amount" gt:"0"`
	}{}

	if err := utils.Validate(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}

	totalSupplyByte, err := ctx.GetObject([]byte(TOTAL_SUPPLY))
	if err != nil {
		return code.Error(err)
	}

	totalSupply, _ := big.NewFloat(0).SetString(string(totalSupplyByte))
	totalSupply = big.NewFloat(0).Add(totalSupply, args.Amount)
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
	caller := ctx.Caller()
	if caller == "" {
		return code.Error(utils.ErrMissingCaller)
	}
	value, err := ctx.GetObject([]byte(BALANCEPRE + caller))
	if err != nil {
		return code.Error(err)
	}
	return code.OK(value)
}

func (am *awardManage) Allowance(ctx code.Context) code.Response {
	args := struct {
		From string `json:"from"`
		To   string `json:"to"`
	}{}
	if err := utils.Validate(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}

	value, err := ctx.GetObject([]byte(ALLOWANCEPRE + args.From + "_" + args.To))
	if err != nil {
		return code.Error(err)
	}
	return code.OK(value)

}

func (am *awardManage) Transfer(ctx code.Context) code.Response {
	from := ctx.Caller()
	if from == "" {
		return code.Error(utils.ErrMissingCaller)
	}
	args := struct {
		To    string     `json:"to" required:"true"`
		Token *big.Float `json:"token" required:"true"`
	}{}

	if err := utils.Validate(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}
	if from == args.To {
		return code.Error(errors.New("can not transfer to yourself"))
	}

	fromBalanceByte, err := ctx.GetObject([]byte(BALANCEPRE + from))
	if err != nil {
		return code.Error(err)
	}
	fromBalance, _ := big.NewFloat(0).SetString(string(fromBalanceByte))

	if fromBalance.Cmp(args.Token) < 0 {
		return code.Error(errors.New("balance not enough"))
	}

	toBalanceByte, err := ctx.GetObject([]byte(BALANCEPRE + args.To))
	if err != nil {
		toBalanceByte = []byte("0")
	}
	toBalance, _ := big.NewFloat(0).SetString(string(toBalanceByte))

	if err := ctx.PutObject([]byte(BALANCEPRE+from), []byte(big.NewFloat(0).Sub(fromBalance, args.Token).String())); err != nil {
		return code.Error(err)
	}
	if err := ctx.PutObject([]byte(BALANCEPRE+args.To), []byte(big.NewFloat(0).Add(toBalance, args.Token).String())); err != nil {
		return code.Error(err)
	}

	return code.OK([]byte("ok~"))

}

func (am *awardManage) TransferFrom(ctx code.Context) code.Response {
	args := struct {
		From  string     `json:"from" required:"true"`
		Token *big.Float `json:"token" required:"true"`
	}{}
	caller := ctx.Caller()
	if caller == "" { // TODO why,xchain's error?
		return code.Error(utils.ErrMissingCaller)
	}
	if err := utils.Validate(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}

	fromBalanceByte, err := ctx.GetObject([]byte(BALANCEPRE + args.From))
	if err != nil {
		return code.Error(err)
	}
	fromBalance, _ := big.NewFloat(0).SetString(string(string(fromBalanceByte)))
	if fromBalance.Cmp(args.Token) < 0 {
		return code.Error(utils.ErrBalanceLow)
	}
	allowanceKey := ALLOWANCEPRE + args.From + "_" + caller
	allowanceBalanceByte, err := ctx.GetObject([]byte(allowanceKey))
	if err != nil {
		return code.Error(err)
	}
	allowanceBalance, _ := big.NewFloat(0).SetString(string(allowanceBalanceByte))
	if allowanceBalance.Cmp(args.Token) < 0 {
		return code.Error(errors.New("allowance balance not enough"))
	}
	allowanceBalance = allowanceBalance.Sub(allowanceBalance, args.Token)
	if err := ctx.PutObject([]byte(allowanceKey), []byte(allowanceBalance.String())); err != nil {
		return code.Error(err)
	}
	return code.OK([]byte("ok~"))
}

func (am *awardManage) Approve(ctx code.Context) code.Response {
	args := struct {
		To    string     `json:"to" required:"true"`
		Token *big.Float `json:"token" required:"true"`
	}{}
	from := ctx.Caller()
	if len(from) == 0 {
		return code.Error(utils.ErrMissingCaller)
	}

	if err := utils.Validate(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}

	allowanceKey := []byte(ALLOWANCEPRE + from + "_" + args.To)

	allowanceByte := []byte("0")
	if value, err := ctx.GetObject(allowanceKey); err == nil {
		allowanceByte = value
	}

	allowanceBalance, _ := big.NewFloat(0).SetString(string(allowanceByte))
	allowanceBalance = allowanceBalance.Add(allowanceBalance, args.Token)
	if err := ctx.PutObject(allowanceKey, []byte(allowanceBalance.String())); err != nil {
		return code.Error(err)
	}
	return code.OK([]byte("ok~"))
}

func main() {
	driver.Serve(new(awardManage))
}

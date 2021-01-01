package main

import (
	"errors"
	"fmt"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/code"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/driver"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/utils"
	"math/big"
)

const (
	BALANCEPRE   = "balanceOf_"
	ALLOWANCEPRE = "allowanceOf_"
	MASTERPRE    = "owner"
)

type awardManage struct{}

func (am *awardManage) Initialize(ctx code.Context) code.Response {
	args := struct {
		TotalSupply *big.Int `json:"totalSupply",gt:"0"`
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

	if err := ctx.PutObject([]byte("TotalSupply"), []byte(args.TotalSupply.String())); err != nil {
		return code.Error(err)
	}

	if err := ctx.PutObject([]byte(MASTERPRE), []byte(caller)); err != nil {
		return code.Error(err)
	}

	return code.OK([]byte(args.TotalSupply.String()))

}

func (am *awardManage) AddAward(ctx code.Context) code.Response {
	caller := ctx.Caller()
	masterBytes, err := ctx.GetObject([]byte(MASTERPRE))
	if err != nil {
		return code.Error(err)
	}
	master := string(masterBytes)

	if master != caller {
		return code.Error(utils.ErrPermissionDenied)
	}
	args := struct {
		Amount *big.Float `json:"amount"`
	}{}
	if err := utils.Validate(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}

	totalSupplyByte, err := ctx.GetObject([]byte("TotalSupply"))
	if err != nil {
		return code.Error(err)
	}

	totalSupply, _ := big.NewFloat(0).SetString(string(totalSupplyByte))
	totalSupply = big.NewFloat(0).Add(totalSupply, args.Amount)
	if err := ctx.PutObject([]byte("TotalSupply"), []byte(totalSupply.String())); err != nil {
		return code.Error(err)
	}
	return code.OK([]byte(totalSupply.String()))
}

func (am *awardManage) TotalSupply(ctx code.Context) code.Response {
	if value, err := ctx.GetObject([]byte("TotalSupply")); err != nil {
		return code.Error(err)
	} else {
		return code.OK(value)
	}
}

func (am *awardManage) Balance(ctx code.Context) code.Response {
	args := struct {
		Caller string `json:"caller"`
	}{}
	if err := utils.Validate(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}
	value, err := ctx.GetObject([]byte(BALANCEPRE + args.Caller))
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
	args := struct {
		From  string     `json:"from",required:"true",eq:"",neq:"",lt:"",gt:""`
		To    string     `json:"to",required:"true"`
		Token *big.Float `json:"token",required:"true"`
	}{}

	if err := utils.Validate(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}
	if args.From == args.To {
		return code.Error(errors.New("can not transfer to yourself"))
	}

	fromBalanceByte, err := ctx.GetObject([]byte(BALANCEPRE + args.From))
	if err != nil {
		return code.Error(err)
	}
	fromBalance, succ := big.NewFloat(0).SetString(string(fromBalanceByte))
	if !succ {
		return code.Error(fmt.Errorf("parse %s as float error", string(fromBalanceByte)))
	}

	if fromBalance.Cmp(args.Token) < 0 {
		return code.Error(errors.New("balance not enough"))
	}

	toBalanceByte, err := ctx.GetObject([]byte(BALANCEPRE + args.To))
	if err != nil {
		return code.Error(err)
	}
	toBalance, _ := big.NewFloat(0).SetString(string(toBalanceByte))

	if err := ctx.PutObject([]byte(BALANCEPRE+args.From), []byte(big.NewFloat(0).Sub(fromBalance, args.Token).String())); err != nil {
		return code.Error(err)
	}
	if err := ctx.PutObject([]byte(BALANCEPRE+args.To), []byte(big.NewFloat(0).Add(toBalance, args.Token).String())); err != nil {
		return code.Error(err)
	}

	return code.OK([]byte("ok"))

}

func (am *awardManage) TransferFrom(ctx code.Context) code.Response {
	args := struct {
		From   string     `json:"from",required:"true"`
		To     string     `json:"to",required:"true"`
		Caller string     `json:"caller",required:"true"`
		Token  *big.Float `json:"token",required:"true"`
	}{}

	if err := utils.Validate(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}

	valueByte, err := ctx.GetObject([]byte(ALLOWANCEPRE + args.From + "_" + args.Caller))
	if err != nil {
		return code.Error(err)

	}
	value, _ := big.NewFloat(0).SetString(string(valueByte))
	if value.Cmp(args.Token) < 0 {
		return code.Error(utils.ErrBalanceLow)
	}

	fromBalanceByte, err := ctx.GetObject([]byte(BALANCEPRE + args.From))
	if err != nil {
		return code.Error(err)
	}
	fromBalance, _ := big.NewFloat(0).SetString(string(string(fromBalanceByte)))
	if fromBalance.Cmp(args.Token) < 0 {
		return code.Error(utils.ErrBalanceLow)
	}

	//toBalaneByte, err := ctx.GetObject([]byte(BALANCEPRE + args.To))
	//if err != nil {
	//	return code.Error(err)
	//}
	//toBalance, _ := big.NewFloat(0).SetString(string(toBalaneByte))
	//
	allowanceBalanceByte, err := ctx.GetObject([]byte(ALLOWANCEPRE + args.From + "_" + args.Caller))
	if err != nil {
		return code.Error(err)
	}
	allowanceBalance, _ := big.NewFloat(0).SetString(string(allowanceBalanceByte))
	allowanceBalance = allowanceBalance.Sub(allowanceBalance, args.Token)
	allowance_key := []byte("") // TODO
	if err := ctx.PutObject(allowance_key, []byte(allowanceBalance.String())); err != nil {
		return code.Error(err)
	}
	return code.OK([]byte("ok~"))
}

func (am *awardManage) Approve(ctx code.Context) code.Response {
	args := struct {
		From   string     `json:"from",required:"true"`
		To     string     `json:"to",required:"true"`
		Caller string     `json:"caller",required:"true"`
		Token  *big.Float `json:"token",required:"true"`
	}{}

	if err := utils.Validate(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}
	//value, err := ctx.GetObject([]byte(ALLOWANCEPRE + args.From + "_" + args.To))
	//if err != nil {
	//	return code.Error(err)
	//}

	fromBalanceByte, err := ctx.GetObject([]byte(BALANCEPRE + args.From))
	if err != nil {
		return code.Error(err)
	}

	fromBalance, _ := big.NewFloat(0).SetString(string(fromBalanceByte))
	if fromBalance.Cmp(args.Token) < 0 {
		return code.Error(utils.ErrBalanceLow)
	}

	//toBalanceByte, err := ctx.GetObject([]byte(BALANCEPRE + args.To))
	//if err != nil {
	//	return code.Error(err)
	//}
	////TODO @fengjin
	//toBalance, _ := big.NewFloat(0).SetString(string(toBalanceByte))

	allowanceKey := []byte(ALLOWANCEPRE + args.From + "_" + args.Caller)
	allowanceBalanceByte, err := ctx.GetObject(allowanceKey)
	if err != nil {
		return code.Error(err)
	}
	allowanceBalance, _ := big.NewFloat(0).SetString(string(allowanceBalanceByte))
	allowanceBalance = allowanceBalance.Sub(allowanceBalance, args.Token)
	if err := ctx.PutObject(allowanceKey, []byte(allowanceBalance.String())); err != nil {
		return code.Error(err)
	}
	return code.OK(nil)
}

func main() {
	driver.Serve(new(awardManage))
}

package main

import (
	"errors"
	"math/big"
	"math/rand"

	"github.com/xuperchain/xuperchain/core/contractsdk/go/code"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/driver"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/utils"
)

const (
	USERID  = "Userid"
	TICKTID = "Luckid"
	ADMIN   = "admin"
	RESULT  = "result"
	TICKETS = "tickets"
)

type luckDraw struct {
}

func (ld *luckDraw) Initialize(ctx code.Context) code.Response {
	args := struct {
		Admin string `json:"admin"`
	}{}
	if err := utils.Unmarshal(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}
	if err := ctx.PutObject([]byte(ADMIN), []byte(args.Admin)); err != nil {
		return code.Error(err)
	}
	err := ctx.PutObject([]byte(TICKETS), []byte("0"))
	if err != nil {
		return code.Error(err)
	}
	return code.OK(nil)
}

func (ld *luckDraw) isAdmin(ctx code.Context, initiator string) bool {
	admin, err := ctx.GetObject([]byte(ADMIN))
	if err != nil {
		return false
	}
	return string(admin) == initiator
}

func (ld *luckDraw) GetLuckId(ctx code.Context) code.Response {
	initiator := ctx.Initiator()
	if initiator == "" {
		return code.Error(utils.ErrMissingInitiator)
	}
	_, err := ctx.GetObject([]byte(RESULT))
	if err == nil {
		return code.Error(errors.New("the luck draw has finished"))
	}

	if userVal, err := ctx.GetObject([]byte(USERID + initiator)); err == nil {
		return code.OK(userVal)
	}

	lastIdByte, err := ctx.GetObject([]byte(TICKETS))
	if err != nil {
		return code.Error(err)
	}
	lastId, _ := big.NewInt(0).SetString(string(lastIdByte), 10)

	lastId = lastId.Add(lastId, big.NewInt(1))
	if err := ctx.PutObject([]byte(USERID+initiator), []byte(lastId.String())); err != nil {
		return code.Error(err)
	}
	if err := ctx.PutObject([]byte(TICKTID+lastId.String()), []byte(initiator)); err != nil {
		return code.Error(err)
	}
	if err := ctx.PutObject([]byte(TICKETS), []byte(lastId.String())); err != nil {
		return code.Error(err)
	}
	return code.OK([]byte(lastId.String()))
}

func (ld *luckDraw) StartLuckDraw(ctx code.Context) code.Response {
	initiator := ctx.Initiator()
	if initiator == "" {
		return code.Error(utils.ErrMissingInitiator)
	}
	if !ld.isAdmin(ctx, initiator) {
		return code.Error(utils.ErrPermissionDenied)
	}
	args := struct {
		Seed *big.Int `json:"seed" validte:"required"`
	}{}

	if err := utils.Unmarshal(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}

	lastIdByte, err := ctx.GetObject([]byte(TICKETS))
	if err != nil {
		return code.Error(err)
	}

	lastId, _ := big.NewInt(0).SetString(string(lastIdByte), 10)
	r := rand.NewSource(args.Seed.Int64())
	luckId := big.NewInt(r.Int63())
	luckId = luckId.Mod(luckId, lastId)
	luckId = luckId.Add(luckId, big.NewInt(1))

	luckUser, err := ctx.GetObject([]byte(TICKTID + luckId.String()))
	if err != nil {
		return code.Error(err)
	}

	if err := ctx.PutObject([]byte(RESULT), luckUser); err != nil {
		return code.Error(err)
	}
	return code.OK(luckUser)
}

func (ld *luckDraw) GetResult(ctx code.Context) code.Response {
	luckUser, err := ctx.GetObject([]byte(RESULT))
	if err != nil {
		return code.Error(err)
	}
	return code.OK(luckUser)
}

func main() {
	driver.Serve(new(luckDraw))
}

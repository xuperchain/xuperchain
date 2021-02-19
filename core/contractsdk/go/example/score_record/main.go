package main

import (
	"github.com/xuperchain/xuperchain/core/contractsdk/go/code"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/driver"
)

const (
	OWNER_KEY  = "owner"
	RECORD_KEY = "R_"
)

type scoreRecord struct {
}

func (sr *scoreRecord) Initialize(ctx code.Context) code.Response {
	args := struct {
		Owner string `json:"owner" validate:"required"`
	}{}

	if err := code.Unmarshal(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}
	if err := ctx.PutObject([]byte(OWNER_KEY), []byte(args.Owner)); err != nil {
		return code.Error(err)
	}
	return code.OK([]byte("ok"))
}

func (sc *scoreRecord) AddScore(ctx code.Context) code.Response {
	initiator := ctx.Initiator()
	if initiator == "" {
		return code.Error(code.ErrMissingInitiator)
	}
	owner, err := ctx.GetObject([]byte(OWNER_KEY))
	if err != nil {
		return code.Error(err)
	}
	if string(owner) != initiator {
		return code.Error(code.ErrPermissionDenied)
	}
	args := struct {
		UserId string `json:"user_id" validate:"required"`
		Data   string `json:"data" validate:"required"`
	}{}

	if err := code.Unmarshal(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}

	if err := ctx.PutObject([]byte(RECORD_KEY+args.UserId), []byte(args.Data)); err != nil {
		return code.Error(err)
	}
	return code.OK([]byte(args.UserId))
}

func (sr *scoreRecord) QueryScore(ctx code.Context) code.Response {
	args := struct {
		UserId string `json:"user_id" validate:"required"`
	}{}
	if err := code.Unmarshal(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}
	data, err := ctx.GetObject([]byte(RECORD_KEY + args.UserId))
	if err != nil {
		return code.Error(err)
	}

	return code.OK(data)
}

func (sr *scoreRecord) QueryOwner(ctx code.Context) code.Response {
	owner, err := ctx.GetObject([]byte(OWNER_KEY))
	if err != nil {
		return code.Error(err)
	}
	return code.OK(owner)
}

func main() {
	driver.Serve(new(scoreRecord))
}

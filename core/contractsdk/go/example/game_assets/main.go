package main

import (
	"fmt"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/code"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/driver"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/utils"
	"strings"
)

type gameAssets struct {
}

const (
	ADMIN      = "admin"
	ASSETTYPE  = "AssetType_"
	USERASSET  = "UserAsset_"
	ASSET2USER = "Asset2User_"
)

func (ga *gameAssets) Initialize(ctx code.Context) code.Response {
	args := struct {
		Admin string `json:"admin",required:"true"`
	}{}

	if err := utils.Validate(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}
	if err := ctx.PutObject([]byte(ADMIN), []byte(args.Admin)); err != nil {
		return code.Error(err)
	} else {
		return code.OK(nil)
	}

}

//TODO @fengjin
func (ga *gameAssets) isAdmin(ctx code.Context, caller string) bool {
	admin, err := ctx.GetObject([]byte(ADMIN))
	if err != nil { // return false if GetObject failed
		return false
	}
	return caller == string(admin)
}

func (ga *gameAssets) AddAssetType(ctx code.Context) code.Response {
	caller := ctx.Initiator()
	if caller == "" {
		return code.Error(utils.ErrMissingCaller)
	}
	if !ga.isAdmin(ctx, caller) {
		return code.Error(utils.ErrPermissionDenied)
	}
	args := struct {
		TypeID   string `json:"typeid",required:"true"`
		TypeDesc string `json:"typedesc",required:"true"`
	}{}
	if err := utils.Validate(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}
	assetKey := ASSETTYPE + args.TypeID
	if _, err := ctx.GetObject([]byte(assetKey)); err != nil {
		return code.Error(fmt.Errorf("asset type %s already exists", args.TypeID))
	}
	if err := ctx.PutObject([]byte(assetKey), []byte(args.TypeDesc)); err != nil {
		return code.Error(err)
	}
	return code.OK(nil)
}

func (ga *gameAssets) ListAssetType(ctx code.Context) code.Response {
	start := ASSETTYPE
	end := ASSETTYPE + "~"
	iter := ctx.NewIterator([]byte(start), []byte(end))
	result := []byte{}
	buf := strings.Builder{}
	for iter.Next() {
		buf.Write(iter.Key()[len([]byte(ASSETTYPE)):])
		buf.WriteString(":")
		buf.Write(iter.Value())
		buf.WriteString("\n")
	}
	return code.OK(result)
}

func (ga *gameAssets) getAssetByUser(ctx code.Context) code.Response {
	//这里也不想搞了
	caller := ctx.Initiator()
	if caller == "" {
		return code.Error(utils.ErrMissingCaller)
	}
	if !ga.isAdmin(ctx, caller) {
		return code.Error(utils.ErrPermissionDenied)
	}
	args := struct {
		UserID string `json:"userid",required:"false"`
	}{}
	if err := utils.Validate(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}

	//TODO @fengjin	空白slice 和nil以及slice的零值

	userId := caller
	if args.UserID != "" && len(args.UserID) > 0 {
		userId = args.UserID
	}
	userAssetKey := USERASSET + userId + "_"
	start := userAssetKey
	end := start + "~"
	iter := ctx.NewIterator([]byte(start), []byte(end))
	defer iter.Close()
	buf := strings.Builder{}
	var getObjectErr error
	for iter.Next() {
		if len(iter.Key()) > len(userAssetKey) {
			assetId := iter.Key()[len(userAssetKey):]
			typeId := iter.Value()
			assetTypeKey := ASSETTYPE + string(typeId)
			if assetDesc, err := ctx.GetObject([]byte(assetTypeKey)); err != nil {
				getObjectErr = err
				break
			} else {
				buf.WriteString("assetId=")
				buf.Write(assetId)
				buf.WriteString("typeId=")
				buf.Write(typeId)
				buf.WriteString("assetDesc=")
				buf.Write(assetDesc)
				buf.WriteString("\n")
			}
		}
	}
	if getObjectErr != nil { // TODO error 默认值
		return code.Error(getObjectErr)
	}
	if err := iter.Error(); err != nil {
		return code.Error(err)
	}
	return code.OK([]byte(buf.String()))
}

func (ga *gameAssets) NewAssetToUser(ctx code.Context) code.Response {
	caller := ctx.Initiator()
	if caller == "" {
		return code.Error(utils.ErrMissingCaller)
	}
	if !ga.isAdmin(ctx, caller) {
		return code.Error(utils.ErrPermissionDenied)
	}
	args := struct {
		UserId  string `json:"userid",required:"true"`
		TypeId  string `json:"typeid",required:"true"`
		AssetId string `json:"assetid",required:"true"`
	}{}
	if err := utils.Validate(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}
	assetKey := ASSET2USER + args.AssetId
	_, err := ctx.GetObject([]byte(assetKey))
	if err != nil {
		return code.Error(err)
	}
	userAssetKey := USERASSET + args.UserId + "_" + args.AssetId

	if err := ctx.PutObject([]byte(userAssetKey), []byte(args.TypeId)); err != nil {
		return code.Error(err)
	}
	if err := ctx.PutObject([]byte(assetKey), []byte(args.UserId)); err != nil {
		return code.Error(err)
	}
	return code.OK([]byte(args.AssetId))
}

func (ga *gameAssets) TradeAsset(ctx code.Context) code.Response {
	from := ctx.Initiator()
	if from == "" {
		return code.Error(utils.ErrMissingCaller)
	}
	args := struct {
		To      string `json:"to",required:"true"`
		AssetId string `json:"assetid",required:"true"`
	}{}
	if err := utils.Validate(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}
	userAssetKey := USERASSET + from + "_" + args.AssetId
	assetType, err := ctx.GetObject([]byte(userAssetKey))
	if err != nil {
		return code.Error(err)
	}
	if err := ctx.DeleteObject([]byte(userAssetKey)); err != nil {
		return code.Error(err)
	}

	assetKey := ASSET2USER + args.AssetId
	newuserAssetKey := USERASSET + args.To + "_" + args.AssetId
	if err := ctx.PutObject([]byte(newuserAssetKey), assetType); err != nil {
		return code.Error(err)
	}
	if err := ctx.PutObject([]byte(assetKey), []byte(args.To)); err != nil {
		return code.Error(err)
	}
	return code.OK([]byte(args.AssetId))
}

func main() {
	driver.Serve(new(gameAssets))
}

package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/xuperchain/xuperchain/core/contractsdk/go/code"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/driver"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/utils"
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
		TypeID   string `json:"type_id" required:"true"`
		TypeDesc string `json:"type_desc" required:"true"`
	}{}
	if err := utils.Validate(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}
	assetKey := ASSETTYPE + args.TypeID
	if _, err := ctx.GetObject([]byte(assetKey)); err == nil {
		return code.Error(fmt.Errorf("asset type %s already exists", args.TypeID))
	}
	if err := ctx.PutObject([]byte(assetKey), []byte(args.TypeDesc)); err != nil {
		return code.Error(err)
	}
	return code.OK([]byte(args.TypeID))
}

func (ga *gameAssets) ListAssetType(ctx code.Context) code.Response {
	start := ASSETTYPE
	end := ASSETTYPE + "~"
	iter := ctx.NewIterator([]byte(start), []byte(end))
	buf := strings.Builder{}
	for iter.Next() {
		buf.Write(iter.Key()[len([]byte(ASSETTYPE)):])
		buf.WriteString(":")
		buf.Write(iter.Value())
		buf.WriteString("\n")
	}
	return code.OK([]byte(buf.String()))
}

func (ga *gameAssets) GetAssetByUser(ctx code.Context) code.Response {
	caller := ctx.Initiator()
	if caller == "" {
		return code.Error(utils.ErrMissingCaller)
	}
	if !ga.isAdmin(ctx, caller) {
		return code.Error(utils.ErrPermissionDenied)
	}
	args := struct {
		UserID string `json:"user_id",required:"false"` // TODO why false
	}{}
	if err := utils.Validate(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}

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
		assetId := iter.Key()[len([]byte(userAssetKey)):]
		typeId := iter.Value()
		if len(string(typeId)) <= len(ASSETTYPE) { //TODO delete only set a flag
			continue
		}
		assetTypeKey := string(typeId)
		if assetDesc, err := ctx.GetObject([]byte(assetTypeKey)); err != nil {
			getObjectErr = errors.New("get asset desc error,access type key: " + assetTypeKey)
			break
		} else {
			buf.WriteString("assetId=")
			buf.Write(assetId)
			buf.WriteString(",typeId=")
			buf.Write(typeId[len(ASSETTYPE):])
			buf.WriteString(",assetDesc=")
			buf.Write(assetDesc)
			buf.WriteString("\n")
		}
	}
	if getObjectErr != nil {
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
		UserId  string `json:"user_id" required:"true"`
		TypeId  string `json:"type_id" required:"true"`
		AssetId string `json:"asset_id" required:"true"`
	}{}
	if err := utils.Validate(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}
	assetTypeKey := ASSETTYPE + args.TypeId
	_, err := ctx.GetObject([]byte(assetTypeKey))
	if err != nil {
		return code.Error(fmt.Errorf("asset type %s not found", args.TypeId))
	}
	userAssetKey := USERASSET + args.UserId + "_" + args.AssetId

	if err := ctx.PutObject([]byte(userAssetKey), []byte(assetTypeKey)); err != nil {
		return code.Error(err)
	}
	assetKey := ASSET2USER + args.AssetId
	if _, err := ctx.GetObject([]byte(assetKey)); err == nil {
		return code.Error(fmt.Errorf("asset %s exists", args.AssetId))
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
		To      string `json:"to" required:"true"`
		AssetId string `json:"asset_id" required:"true"`
	}{}
	if err := utils.Validate(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}
	userAssetKey := USERASSET + from + "_" + args.AssetId
	assetType, err := ctx.GetObject([]byte(userAssetKey))
	if err != nil {
		return code.Error(fmt.Errorf("asset %s of user %s not found", args.AssetId, from))
	}
	if err := ctx.DeleteObject([]byte(userAssetKey)); err != nil {
		return code.Error(err)
	}

	assetKey := ASSET2USER + args.AssetId
	newuserAssetKey := USERASSET + args.To + "_" + args.AssetId
	if err := ctx.PutObject([]byte(newuserAssetKey), assetType); err != nil {
		return code.Error(err)
	}
	// 这里对重复key 是如何处理的呢
	if err := ctx.PutObject([]byte(assetKey), []byte(args.To)); err != nil {
		return code.Error(err)
	}
	return code.OK([]byte(args.AssetId))
}

func main() {
	driver.Serve(new(gameAssets))
}

//  还差一个查询自己的资产

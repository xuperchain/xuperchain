package main

import (
	"errors"
	"fmt"

	"github.com/xuperchain/xuperchain/core/contractsdk/go/code"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/driver"
)

type gameAssets struct {
}

const (
	ADMIN      = "admin"
	ASSETTYPE  = "AssetType_"
	USERASSET  = "UserAsset_"
	ASSET2USER = "Asset2User_"
)

type assetType struct {
	TypeID   string `json:"type_id" validate:"required"`
	TypeDesc string `json:"type_desc" validat:"required"`
}

type asset struct {
	Id     string `json:"id"`
	TypeId string `json:"type_id"`
	Desc   string `json:"asset_desc"`
}

func (ga *gameAssets) Initialize(ctx code.Context) code.Response {
	args := struct {
		Admin string `json:"admin" validate:"required"`
	}{}

	if err := code.Unmarshal(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}
	if err := ctx.PutObject([]byte(ADMIN), []byte(args.Admin)); err != nil {
		return code.Error(err)
	} else {
		return code.OK(nil)
	}
}

func (ga *gameAssets) isAdmin(ctx code.Context, initiator string) bool {
	admin, err := ctx.GetObject([]byte(ADMIN))
	if err != nil { // return false if GetObject failed
		return false
	}
	return initiator == string(admin)
}

func (ga *gameAssets) AddAssetType(ctx code.Context) code.Response {
	initiator := ctx.Initiator()
	if initiator == "" {
		return code.Error(code.ErrMissingInitiator)
	}
	if !ga.isAdmin(ctx, initiator) {
		return code.Error(code.ErrPermissionDenied)
	}
	args := assetType{}
	if err := code.Unmarshal(ctx.Args(), &args); err != nil {
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
	iter := ctx.NewIterator(code.PrefixRange([]byte(ASSETTYPE)))
	defer iter.Close()

	assetTypes := []assetType{}
	for iter.Next() {
		assetTypes = append(assetTypes, assetType{
			string(iter.Key()[len([]byte(ASSETTYPE)):]),
			string(iter.Value()),
		})
	}
	return code.JSON(assetTypes)
}

func (ga *gameAssets) NewAssetToUser(ctx code.Context) code.Response {
	initiator := ctx.Initiator()
	if initiator == "" {
		return code.Error(code.ErrMissingInitiator)
	}
	if !ga.isAdmin(ctx, initiator) {
		return code.Error(code.ErrPermissionDenied)
	}
	args := struct {
		UserId  string `json:"user_id" validate:"required"`
		TypeId  string `json:"type_id" validate:"required"`
		AssetId string `json:"asset_id" validate:"required"`
	}{}
	if err := code.Unmarshal(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}
	assetTypeKey := ASSETTYPE + args.TypeId
	_, err := ctx.GetObject([]byte(assetTypeKey))
	if err != nil {
		return code.Error(fmt.Errorf("asset type %s not found", args.TypeId))
	}

	assetKey := ASSET2USER + args.AssetId
	if _, err := ctx.GetObject([]byte(assetKey)); err == nil {
		return code.Error(fmt.Errorf("asset %s exists", args.AssetId))
	}
	if err := ctx.PutObject([]byte(assetKey), []byte(args.UserId)); err != nil {
		return code.Error(err)
	}

	userAssetKey := USERASSET + args.UserId + "_" + args.AssetId
	if err := ctx.PutObject([]byte(userAssetKey), []byte(ASSETTYPE+args.TypeId)); err != nil {
		return code.Error(err)
	}

	return code.OK([]byte(args.AssetId))
}

func (ga *gameAssets) TradeAsset(ctx code.Context) code.Response {
	from := ctx.Initiator()
	if from == "" {
		return code.Error(code.ErrMissingInitiator)
	}
	args := struct {
		To      string `json:"to" validate:"required"`
		AssetId string `json:"asset_id" validate:"required"`
	}{}
	if err := code.Unmarshal(ctx.Args(), &args); err != nil {
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
	if err := ctx.PutObject([]byte(assetKey), []byte(args.To)); err != nil {
		return code.Error(err)
	}
	return code.OK([]byte(args.AssetId))
}

func (ga *gameAssets) GetAssetByUser(ctx code.Context) code.Response {
	initiator := ctx.Initiator()
	if initiator == "" {
		return code.Error(code.ErrMissingInitiator)
	}
	userId := initiator
	args := struct {
		UserID string `json:"user_id" validate:"required"`
	}{}

	if ga.isAdmin(ctx, initiator) {
		if err := code.Unmarshal(ctx.Args(), &args); err == nil {
			userId = args.UserID
		}
	}

	userAssetKey := USERASSET + userId + "_"
	iter := ctx.NewIterator(code.PrefixRange([]byte(userAssetKey)))
	defer iter.Close()

	assets := []asset{}
	var getObjectErr error

	for iter.Next() {
		assetId := iter.Key()[len([]byte(userAssetKey)):]
		typeId := iter.Value()
		if len(string(typeId)) <= len(ASSETTYPE) { // delete only set a flag
			continue
		}
		assetTypeKey := string(typeId)
		if assetDesc, err := ctx.GetObject(typeId); err != nil {
			getObjectErr = errors.New("get asset desc error,access type key: " + assetTypeKey)
			break
		} else {
			assets = append(assets, asset{
				Id:     string(assetId),
				TypeId: string(typeId[len(ASSETTYPE):]),
				Desc:   string(assetDesc),
			})
		}
	}
	if getObjectErr != nil {
		return code.Error(getObjectErr)
	}
	if err := iter.Error(); err != nil {
		return code.Error(err)
	}

	return code.JSON(assets)
}
func main() {
	driver.Serve(new(gameAssets))
}

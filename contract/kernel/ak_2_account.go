package kernel

import (
	"encoding/json"
	"errors"

	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/permission/acl/utils"
)

func updateThresholdWithDel(ctx *KContext, aksWeight map[string]float64, accountName string) error {
	for address := range aksWeight {
		key := utils.MakeAK2AccountKey(address, accountName)
		err := ctx.ModelCache.Del(utils.GetAK2AccountBucket(), []byte(key))
		if err != nil {
			return err
		}
	}
	return nil
}

func updateThresholdWithPut(ctx *KContext, aksWeight map[string]float64, accountName string) error {
	for address := range aksWeight {
		key := utils.MakeAK2AccountKey(address, accountName)
		err := ctx.ModelCache.Put(utils.GetAK2AccountBucket(), []byte(key), []byte("true"))
		if err != nil {
			return err
		}
	}
	return nil
}

func updateAkSetWithDel(ctx *KContext, sets map[string]*pb.AkSet, accountName string) error {
	for _, akSets := range sets {
		for _, ak := range akSets.GetAks() {
			key := utils.MakeAK2AccountKey(ak, accountName)
			err := ctx.ModelCache.Del(utils.GetAK2AccountBucket(), []byte(key))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func updateAkSetWithPut(ctx *KContext, sets map[string]*pb.AkSet, accountName string) error {
	for _, akSets := range sets {
		for _, ak := range akSets.GetAks() {
			key := utils.MakeAK2AccountKey(ak, accountName)
			err := ctx.ModelCache.Put(utils.GetAK2AccountBucket(), []byte(key), []byte("true"))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func updateForThreshold(ctx *KContext, aksWeight map[string]float64, accountName string, method string) error {
	switch method {
	case "Del":
		return updateThresholdWithDel(ctx, aksWeight, accountName)
	case "Put":
		return updateThresholdWithPut(ctx, aksWeight, accountName)
	default:
		return errors.New("unexpected error, method only for Del or Put")
	}
}

func updateForAKSet(ctx *KContext, akSets *pb.AkSets, accountName string, method string) error {
	sets := akSets.GetSets()
	switch method {
	case "Del":
		return updateAkSetWithDel(ctx, sets, accountName)
	case "Put":
		return updateAkSetWithPut(ctx, sets, accountName)
	default:
		return errors.New("unexpected error, method only for Del or Put")
	}
}

func update(ctx *KContext, aclJSON []byte, accountName string, method string) error {
	if aclJSON == nil {
		return nil
	}
	acl := &pb.Acl{}
	json.Unmarshal(aclJSON, acl)
	akSets := acl.GetAkSets()
	aksWeight := acl.GetAksWeight()
	permissionRule := acl.GetPm().GetRule()

	switch permissionRule {
	case pb.PermissionRule_SIGN_THRESHOLD:
		return updateForThreshold(ctx, aksWeight, accountName, method)
	case pb.PermissionRule_SIGN_AKSET:
		return updateForAKSet(ctx, akSets, accountName, method)
	default:
		return errors.New("update ak to account reflection failed, permission model is not found")
	}
	return nil
}

func updateAK2AccountReflection(ctx *KContext, aclOldJSON []byte, aclNewJSON []byte, accountName string) error {
	if err := update(ctx, aclOldJSON, accountName, "Del"); err != nil {
		return err
	}
	if err := update(ctx, aclNewJSON, accountName, "Put"); err != nil {
		return err
	}
	return nil
}

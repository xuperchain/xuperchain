package kernel

import (
	"encoding/json"
	"fmt"

	"github.com/xuperchain/xuperunion/contract"
	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/permission/acl"
	"github.com/xuperchain/xuperunion/permission/acl/utils"
	"github.com/xuperchain/xuperunion/xmodel"
)

const (
	newAccountGasAmount = 1000
	setACLGasAmount     = 10
)

// NewAccountMethod define NewAccountMethod type
type NewAccountMethod struct {
}

// SetAccountACLMethod define SetAccountACLMethod type
type SetAccountACLMethod struct {
}

// SetMethodACLMethod define SetMethodACLMethod type
type SetMethodACLMethod struct {
}

func validACL(acl *pb.Acl) error {
	// param absence check
	if acl == nil {
		return fmt.Errorf("valid acl failed, arg of acl is nil")
	}

	// permission model check
	if permissionModel := acl.GetPm(); permissionModel != nil {
		permissionRule := permissionModel.GetRule()
		akSets := acl.GetAkSets()
		aksWeight := acl.GetAksWeight()
		if akSets == nil && aksWeight == nil {
			return fmt.Errorf("invoke NewAccount failed, permission model is not valid")
		}
		// aks limitation check
		if permissionRule == pb.PermissionRule_SIGN_THRESHOLD {
			if aksWeight == nil || len(aksWeight) > utils.GetAkLimit() {
				return fmt.Errorf("valid acl failed, aksWeight is nil or size of aksWeight is very big")
			}
		} else if permissionRule == pb.PermissionRule_SIGN_AKSET {
			if akSets != nil {
				sets := akSets.GetSets()
				if sets == nil || len(sets) > utils.GetAkLimit() {
					return fmt.Errorf("valid acl failed, Sets is nil or size of Sets is very big")
				}
			} else {
				return fmt.Errorf("valid acl failed, akSets is nil")
			}
		} else {
			return fmt.Errorf("valid acl failed, permission model is not found")
		}
	} else {
		return fmt.Errorf("valid acl failed, lack of argument of permission model")
	}

	return nil
}

// Invoke NewAccount method implementation
func (na *NewAccountMethod) Invoke(ctx *KContext, args map[string][]byte) (*contract.Response, error) {
	if ctx.ResourceLimit.XFee < newAccountGasAmount {
		return nil, fmt.Errorf("gas not enough, expect no less than %d", newAccountGasAmount)
	}
	// json -> pb.Acl
	accountName := args["account_name"]
	aclJSON := args["acl"]
	aclBuf := &pb.Acl{}
	json.Unmarshal(aclJSON, aclBuf)

	if accountName == nil {
		return nil, fmt.Errorf("Invoke NewAccount failed, warn: account name is empty")
	}
	accountStr := string(accountName)
	if validErr := acl.ValidRawAccount(accountStr); validErr != nil {
		return nil, validErr
	}

	bcname := ctx.ModelCache.GetBcname()
	if bcname == "" {
		return nil, fmt.Errorf("block name is empty")
	}
	accountStr = utils.MakeAccountKey(bcname, accountStr)

	if validErr := validACL(aclBuf); validErr != nil {
		return nil, validErr
	}

	oldAccount, err := ctx.ModelCache.Get(utils.GetAccountBucket(), []byte(accountStr))
	if err != nil && err != xmodel.ErrNotFound {
		return nil, err
	}
	if oldAccount != nil {
		return nil, fmt.Errorf("account already exists: %s", accountName)
	}
	err = ctx.ModelCache.Put(utils.GetAccountBucket(), []byte(accountStr), aclJSON)
	if err != nil {
		return nil, err
	}

	// add ak -> account reflection
	err = updateAK2AccountReflection(ctx, nil, aclJSON, accountStr)
	if err != nil {
		return nil, err
	}

	ctx.AddXFeeUsed(newAccountGasAmount)

	return &contract.Response{
		Status: contract.StatusOK,
		Body:   aclJSON,
	}, nil
}

// Invoke SetAccountACL method implementation
func (saa *SetAccountACLMethod) Invoke(ctx *KContext, args map[string][]byte) (*contract.Response, error) {
	if ctx.ResourceLimit.XFee < setACLGasAmount {
		return nil, fmt.Errorf("gas not enough, expect no less than %d", setACLGasAmount)
	}
	// json -> pb.Acl
	accountName := args["account_name"]
	aclJSON := args["acl"]
	aclBuf := &pb.Acl{}
	json.Unmarshal(aclJSON, aclBuf)
	if validErr := validACL(aclBuf); validErr != nil {
		return nil, validErr
	}

	//aclOldJSON, err := ctx.ModelCache.Get(utils.GetAccountBucket(), accountName)
	versionData, err := ctx.ModelCache.Get(utils.GetAccountBucket(), accountName)
	if err != nil {
		return nil, err
	}
	// delete ak -> account reflection
	// add ak -> account reflection
	aclOldJSON := versionData.GetPureData().GetValue()
	err = updateAK2AccountReflection(ctx, aclOldJSON, aclJSON, string(accountName))
	if err != nil {
		return nil, err
	}

	err = ctx.ModelCache.Put(utils.GetAccountBucket(), accountName, aclJSON)
	if err != nil {
		return nil, err
	}

	ctx.AddXFeeUsed(setACLGasAmount)

	return &contract.Response{
		Status: contract.StatusOK,
		Body:   aclJSON,
	}, nil
}

// Invoke SetMethodACL method implementation
func (sma *SetMethodACLMethod) Invoke(ctx *KContext, args map[string][]byte) (*contract.Response, error) {
	if ctx.ResourceLimit.XFee < setACLGasAmount {
		return nil, fmt.Errorf("gas not enough, expect no less than %d", setACLGasAmount)
	}
	contractNameBuf := args["contract_name"]
	methodNameBuf := args["method_name"]
	if contractNameBuf == nil || methodNameBuf == nil {
		return nil, fmt.Errorf("set method acl failed, contract name is nil or method name is nil")
	}

	// json -> pb.Acl
	contractName := string(contractNameBuf)
	methodName := string(methodNameBuf)
	aclJSON := args["acl"]
	aclBuf := &pb.Acl{}
	json.Unmarshal(aclJSON, aclBuf)

	if validErr := validACL(aclBuf); validErr != nil {
		return nil, validErr
	}
	key := utils.MakeContractMethodKey(contractName, methodName)
	err := ctx.ModelCache.Put(utils.GetContractBucket(), []byte(key), aclJSON)
	if err != nil {
		return nil, err
	}

	ctx.AddXFeeUsed(setACLGasAmount)
	return &contract.Response{
		Status: contract.StatusOK,
		Body:   aclJSON,
	}, nil
}

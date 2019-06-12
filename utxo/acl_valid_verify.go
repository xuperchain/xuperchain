package utxo

import (
	"bytes"
	"errors"

	"github.com/xuperchain/xuperunion/pb"
	pm "github.com/xuperchain/xuperunion/permission"
	"github.com/xuperchain/xuperunion/permission/acl/utils"
	"github.com/xuperchain/xuperunion/utxo/txhash"
)

func (uv *UtxoVM) verifyAccountACLValid(accountName string, tx *pb.Transaction) (bool, error) {
	digestHash, dhErr := txhash.MakeTxDigestHash(tx)
	if dhErr != nil {
		return false, dhErr
	}
	return pm.IdentifyAccount(accountName, tx.AuthRequire, tx.AuthRequireSigns, digestHash, uv.aclMgr)
}

func (uv *UtxoVM) verifyContractACLValid(contractName string, tx *pb.Transaction) (bool, error) {
	digestHash, dhErr := txhash.MakeTxDigestHash(tx)
	if dhErr != nil {
		return false, dhErr
	}
	versionData, err := uv.model3.Get(utils.GetContract2AccountBucket(), []byte(contractName))
	if err != nil || versionData == nil {
		return false, err
	}
	pureData := versionData.GetPureData()
	confirmed := versionData.GetConfirmed()
	if pureData == nil || confirmed == false {
		return false, errors.New("pure data is nil or unconfirmed")
	}
	accountName := pureData.GetValue()
	return pm.IdentifyAccount(string(accountName), tx.AuthRequire, tx.AuthRequireSigns, digestHash, uv.aclMgr)
}

func (uv *UtxoVM) verifyRWACLValid(tx *pb.Transaction) (bool, error) {
	req := tx.GetContractRequests()
	// if not contract, pass directly
	if req == nil {
		return true, nil
	}
	env, err := uv.model3.PrepareEnv(tx)
	if err != nil {
		return false, err
	}
	writeSet := env.GetOutputs()
	for _, ele := range writeSet {
		bucket := ele.GetBucket()
		key := ele.GetKey()
		switch bucket {
		case utils.GetAccountBucket():
			ok, accountErr := uv.verifyAccountACLValid(string(key), tx)
			if !ok {
				uv.xlog.Warn("tx info ", "AuthRequire ", tx.AuthRequire, "AuthRequireSigns ", tx.AuthRequireSigns)
				return ok, accountErr
			}
		case utils.GetContractBucket():
			separator := utils.GetACLSeparator()
			idx := bytes.Index(key, []byte(separator))
			if idx < 0 {
				return false, errors.New("invalid raw key")
			}
			contractName := string(key[:idx])
			ok, contractErr := uv.verifyContractACLValid(contractName, tx)
			if !ok {
				uv.xlog.Warn("tx info ", "AuthRequire ", tx.AuthRequire, "AuthRequireSigns ", tx.AuthRequireSigns)
				return ok, contractErr
			}
		case utils.GetContract2AccountBucket():
			accountName := ele.GetValue()
			if accountName == nil {
				return false, errors.New("account name is empty")
			}
			ok, accountErr := uv.verifyAccountACLValid(string(accountName), tx)
			if !ok {
				uv.xlog.Warn("tx info ", "AuthRequire ", tx.AuthRequire, "AuthRequireSigns ", tx.AuthRequireSigns)
				return ok, accountErr
			}
		}
	}
	return true, nil
}

func (uv *UtxoVM) verifyContractValid(tx *pb.Transaction) (bool, error) {
	req := tx.GetContractRequests()
	if req == nil {
		return true, nil
	}
	digestHash, dhErr := txhash.MakeTxDigestHash(tx)
	if dhErr != nil {
		return false, dhErr
	}

	for i := 0; i < len(req); i++ {
		tmpReq := req[i]
		contractName := tmpReq.GetContractName()
		methodName := tmpReq.GetMethodName()

		ok, contractErr := pm.CheckContractMethodPerm(tx.AuthRequire, tx.AuthRequireSigns, digestHash, contractName, methodName, uv.aclMgr)
		if !ok {
			uv.xlog.Warn("tx info ", "AuthRequire ", tx.AuthRequire, "AuthRequireSigns ", tx.AuthRequireSigns)
		}
		if contractErr != nil {
			return ok, ErrRWAclNotEnough
		}
	}
	return true, nil
}

package impl

import (
	"encoding/json"
	"errors"

	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/permission/acl/utils"
	"github.com/xuperchain/xuperunion/xmodel"
)

// Manager manages all ACL releated data, providing read/write interface for ACL table
type Manager struct {
	// some members here
	model3 *xmodel.XModel
}

// NewACLManager create instance of ACLManager
func NewACLManager(model *xmodel.XModel) (*Manager, error) {
	return &Manager{
		model3: model,
	}, nil
}

// GetAccountACL get acl of an account
func (mgr *Manager) GetAccountACL(accountName string) (*pb.Acl, error) {
	acl, confirmed, err := mgr.GetAccountACLWithConfirmed(accountName)
	if err != nil {
		return nil, err
	}
	if acl != nil && !confirmed {
		return nil, errors.New("acl in unconfirmed")
	}
	return acl, nil
}

// GetContractMethodACL get acl of contract method
func (mgr *Manager) GetContractMethodACL(contractName string, methodName string) (*pb.Acl, error) {
	acl, confirmed, err := mgr.GetContractMethodACLWithConfirmed(contractName, methodName)
	if err != nil {
		return nil, err
	}
	if acl != nil && !confirmed {
		return nil, errors.New("acl in unconfirmed")
	}
	return acl, nil
}

// GetAccountACLWithConfirmed implements reading ACL of an account with confirmed state
func (mgr *Manager) GetAccountACLWithConfirmed(accountName string) (*pb.Acl, bool, error) {
	versionData, confirmed, err := mgr.model3.GetWithTxStatus(utils.GetAccountBucket(), []byte(accountName))
	if err != nil || versionData == nil {
		return nil, false, err
	}
	// 反序列化
	acl := &pb.Acl{}
	pureData := versionData.GetPureData()
	if pureData == nil {
		return nil, false, errors.New("pureData is nil")
	}
	jsonBuf := pureData.GetValue()
	if len(jsonBuf) == 0 {
		// no acl data found of this key
		return nil, false, nil
	}
	json.Unmarshal(jsonBuf, acl)

	return acl, confirmed, nil
}

// GetContractMethodACLWithConfirmed implements reading ACL of a contract method with confirmed state
func (mgr *Manager) GetContractMethodACLWithConfirmed(contractName string, methodName string) (*pb.Acl, bool, error) {
	key := utils.MakeContractMethodKey(contractName, methodName)
	versionData, confirmed, err := mgr.model3.GetWithTxStatus(utils.GetContractBucket(), []byte(key))
	if err != nil || versionData == nil {
		return nil, false, err
	}
	// 反序列化
	acl := &pb.Acl{}
	pureData := versionData.GetPureData()
	if pureData == nil {
		return nil, false, errors.New("pureData is nil")
	}
	//pbBuf := pureData.GetValue()
	jsonBuf := pureData.GetValue()
	if len(jsonBuf) == 0 {
		// no acl data found of this key
		return nil, false, nil
	}
	json.Unmarshal(jsonBuf, acl)

	return acl, confirmed, nil
}

// GetAccountAddresses get the addresses belongs to contract account
func (mgr *Manager) GetAccountAddresses(accountName string) ([]string, error) {
	acl, err := mgr.GetAccountACL(accountName)
	if err != nil {
		return nil, err
	}

	return mgr.getAddressesByACL(acl)
}

func (mgr *Manager) getAddressesByACL(acl *pb.Acl) ([]string, error) {
	addresses := make([]string, 0)

	switch acl.GetPm().GetRule() {
	case pb.PermissionRule_SIGN_THRESHOLD:
		for ak := range acl.GetAksWeight() {
			addresses = append(addresses, ak)
		}
	case pb.PermissionRule_SIGN_AKSET:
		for _, set := range acl.GetAkSets().GetSets() {
			aks := set.GetAks()
			addresses = append(addresses, aks...)
		}
	default:
		return nil, errors.New("Unknown permission rule")
	}

	return addresses, nil
}

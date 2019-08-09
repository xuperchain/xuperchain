package permission

import (
	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/permission/acl"
	"github.com/xuperchain/xuperunion/permission/ptree"
	"github.com/xuperchain/xuperunion/permission/rule"
	"github.com/xuperchain/xuperunion/permission/utils"

	"errors"
	"fmt"
)

// IdentifyAK checks whether the sign matches the given ak,
// akuri is the address(or address uri like AccountA/AddressB)
// Return true if ak the sign match
func IdentifyAK(akuri string, sign *pb.SignatureInfo, msg []byte) (bool, error) {
	if sign == nil {
		return false, errors.New("sign is nil")
	}
	akpath := utils.SplitAccountURI(akuri)
	if len(akpath) < 1 {
		return false, errors.New("Invalid address")
	}
	ak := akpath[len(akpath)-1]
	return utils.VerifySign(ak, sign, msg)
}

// IdentifyAccount checks whether the aks could represent the given account via account's ACL strategy.
// Return true if the signatures match aks and aks could represent the account.
func IdentifyAccount(account string, aksuri []string, aclMgr acl.ManagerInterface) (bool, error) {
	// aks and signs could have zero length for permission rule Null
	if aclMgr == nil {
		return false, fmt.Errorf("Invalid Param, aclMgr=%v", aclMgr)
	}

	// build perm tree
	pnode, err := ptree.BuildAccountPermTree(aclMgr, account, aksuri)
	if err != nil {
		return false, err
	}

	return validatePermTree(pnode, true)
}

// CheckContractMethodPerm checks whether the aks satisfy the ACL of a contract method <contractName, methodName>.
// Return true if the signatures match aks and satisfy the ACL.
func CheckContractMethodPerm(aksuri []string, contractName string, methodName string,
	aclMgr acl.ManagerInterface) (bool, error) {
	// aks and signs could have zero length for permission rule Null
	if aclMgr == nil {
		return false, fmt.Errorf("Invalid Param, aclMgr=%v", aclMgr)
	}

	// build perm tree
	pnode, err := ptree.BuildMethodPermTree(aclMgr, contractName, methodName, aksuri)
	if err != nil {
		return false, err
	}

	// validate perm tree
	return validatePermTree(pnode, false)
}

func validatePermTree(root *ptree.PermNode, isAccount bool) (bool, error) {
	if root == nil {
		return false, errors.New("Root is null")
	}

	// get BFS list of perm tree
	plist, err := ptree.GetPermTreeList(root)
	if err != nil {
		return false, err
	}
	listlen := len(plist)
	vf := &rule.ACLValidatorFactory{}

	// reverse travel the perm tree
	for i := listlen - 1; i >= 0; i-- {
		pnode := plist[i]
		nameCheck := acl.IsAccount(pnode.Name)
		// 0 means AK, 1 means Account, otherwise invalid
		if nameCheck < 0 || nameCheck > 1 {
			return false, errors.New("Invalid account/ak name")
		}

		// for non-account perm tree, the root node is not account name
		if i == 0 && !isAccount {
			nameCheck = 1
		}

		checkResult := false
		if nameCheck == 0 {
			// current node is AK, signature should be validated before
			checkResult = true
		} else if nameCheck == 1 {
			// current node is Account, so validation using ACLValidator
			if pnode.ACL == nil {
				// empty ACL means everyone could pass ACL validation
				checkResult = true
			} else {
				if pnode.ACL.Pm == nil {
					return false, errors.New("Acl has empty Pm field")
				}

				// get ACLValidator by ACL type
				validator, err := vf.GetACLValidator(pnode.ACL.Pm.Rule)
				if err != nil {
					return false, err
				}
				checkResult, err = validator.Validate(pnode)
				if err != nil {
					return false, err
				}
			}
		}

		// set validation status
		if checkResult {
			pnode.Status = ptree.Success
		} else {
			pnode.Status = ptree.Failed
		}
	}
	return (root.Status == ptree.Success), nil
}

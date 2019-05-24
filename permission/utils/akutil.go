package utils

import (
	"strings"

	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/permission/acl"
)

// SplitAccountURI split a nested ak into account and address
// e.g. "bob/alice/address1" split to ["bob", "alice", "address1"]
// the last string should be an address, other strings should be account
func SplitAccountURI(akuri string) []string {
	// TODO: validate the URI, and check address and account are legal
	ids := strings.Split(akuri, "/")
	return ids
}

// GetAccountACL return account acl
func GetAccountACL(aclMgr acl.ManagerInterface, account string) (*pb.Acl, error) {
	return aclMgr.GetAccountACL(account)
}

// GetContractMethodACL return contract method acl
func GetContractMethodACL(aclMgr acl.ManagerInterface, contractName string, methodName string) (*pb.Acl, error) {
	return aclMgr.GetContractMethodACL(contractName, methodName)
}

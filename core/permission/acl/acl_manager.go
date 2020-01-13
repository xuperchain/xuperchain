package acl

import (
	"github.com/xuperchain/xuperchain/core/pb"
)

// AccountACL is interface to read/write accounts' ACL
type AccountACL interface {
	GetAccountACL(accountName string) (*pb.Acl, error)
	GetAccountACLWithConfirmed(accountName string) (*pb.Acl, bool, error)
	GetAccountAddresses(accountName string) ([]string, error)
}

// ContractACL is interface to read/write contracts' ACL
type ContractACL interface {
	GetContractMethodACL(contractName string, methodName string) (*pb.Acl, error)
	GetContractMethodACLWithConfirmed(contractName string, methodName string) (*pb.Acl, bool, error)
}

// ManagerInterface defines the interface of ACL Manager
type ManagerInterface interface {
	AccountACL
	ContractACL
}

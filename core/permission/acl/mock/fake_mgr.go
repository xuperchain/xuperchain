package mock

import (
	"github.com/xuperchain/xuperchain/core/pb"
)

// FakeACLManager is a mock up of ACL Manager for test purpose
type FakeACLManager struct {
	accountACL  map[string]*pb.Acl
	contractACL map[string]*pb.Acl
}

// NewFakeACLManager create instance of FakeACLManager
func NewFakeACLManager() (*FakeACLManager, error) {
	return &FakeACLManager{
		accountACL:  make(map[string]*pb.Acl),
		contractACL: make(map[string]*pb.Acl),
	}, nil
}

// SetAccountACL save account acl in memory
func (fm *FakeACLManager) SetAccountACL(accountName string, acl *pb.Acl) error {
	fm.accountACL[accountName] = acl
	return nil
}

// GetAccountACL get account acl from memory
func (fm *FakeACLManager) GetAccountACL(accountName string) (*pb.Acl, error) {
	acl, _ := fm.accountACL[accountName]
	return acl, nil
}

// GetAccountACLWithConfirmed not used in this mockup
func (fm *FakeACLManager) GetAccountACLWithConfirmed(accountName string) (*pb.Acl, bool, error) {
	return nil, true, nil
}

// SetContractMethodACL save contract method acl in memory
func (fm *FakeACLManager) SetContractMethodACL(contractName string, methodName string, acl *pb.Acl) error {
	fm.accountACL[contractName+methodName] = acl
	return nil
}

// GetContractMethodACL get contract method acl from memory
func (fm *FakeACLManager) GetContractMethodACL(contractName string, methodName string) (*pb.Acl, error) {
	acl, _ := fm.contractACL[contractName+methodName]
	return acl, nil
}

// GetContractMethodACLWithConfirmed not used in this mockup
func (fm *FakeACLManager) GetContractMethodACLWithConfirmed(contractName string, methodName string) (*pb.Acl, bool, error) {
	return nil, true, nil
}

// GetAccountAddresses not used in this mockup
func (fm *FakeACLManager) GetAccountAddresses(accountName string) ([]string, error) {
	return nil, nil
}

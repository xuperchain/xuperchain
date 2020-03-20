package xchain

import (
	"errors"

	"github.com/xuperchain/xuperchain/core/pb"
)

var (
	errUnimplemented = errors.New("unimplemented")
)

type chainCore struct {
}

// GetAccountAddress get addresses associated with account name
func (c *chainCore) GetAccountAddresses(accountName string) ([]string, error) {
	return []string{}, nil
}

// VerifyContractPermission verify permission of calling contract
func (c *chainCore) VerifyContractPermission(initiator string, authRequire []string, contractName, methodName string) (bool, error) {
	return true, nil
}

// VerifyContractOwnerPermission verify contract ownership permisson
func (c *chainCore) VerifyContractOwnerPermission(contractName string, authRequire []string) error {
	return nil
}

// QueryTransaction query confirmed tx
func (c *chainCore) QueryTransaction(txid []byte) (*pb.Transaction, error) {
	return new(pb.Transaction), nil
}

// QueryBlock query block
func (c *chainCore) QueryBlock(blockid []byte) (*pb.InternalBlock, error) {
	return new(pb.InternalBlock), nil
}

// CrossQuery query contract from otherchain
func (c *chainCore) ResolveChain(chainName string) (*pb.CrossQueryMeta, error) {
	return new(pb.CrossQueryMeta), nil
}

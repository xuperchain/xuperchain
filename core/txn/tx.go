// package txn deals with tx data
package txn

import (
	"errors"
	"math/big"

	"github.com/xuperchain/xuperchain/core/pb"
)

// Transaction is the internal represents of transaction
type Transaction struct {
	*pb.Transaction
}

func ParseContractTransferRequest(requests []*pb.InvokeRequest) (string, *big.Int, error) {
	// found is the flag of whether the contract already carries the amount parameter
	var found bool
	amount := new(big.Int)
	var contractName string
	for _, req := range requests {
		amountstr := req.GetAmount()
		if amountstr == "" {
			continue
		}
		if found {
			return "", nil, errors.New("duplicated contract transfer amount")
		}
		_, ok := amount.SetString(amountstr, 10)
		if !ok {
			return "", nil, errors.New("bad amount in request")
		}
		found = true
		contractName = req.GetContractName()
	}
	return contractName, amount, nil
}

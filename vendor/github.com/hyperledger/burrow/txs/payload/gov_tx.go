package payload

import (
	"fmt"

	"github.com/hyperledger/burrow/acm/balance"
	"github.com/hyperledger/burrow/crypto"
	spec "github.com/hyperledger/burrow/genesis/spec"
	permission "github.com/hyperledger/burrow/permission"
)

// GovernanceTx contains functionality for altering permissions, token distribution, consensus parameters,
// validators, and network forks.

func (tx *GovTx) Type() Type {
	return TypeGovernance
}

func (tx *GovTx) GetInputs() []*TxInput {
	return tx.Inputs
}

func (tx *GovTx) String() string {
	return fmt.Sprintf("GovTx{%v -> %v}", tx.Inputs, tx.AccountUpdates)
}

func (tx *GovTx) Any() *Any {
	return &Any{
		GovTx: tx,
	}
}

// TODO:
// - Set validator power
// - Set account amount(s)
// - Set account permissions
// - Set global permissions
// - Set ConsensusParams
// Future considerations:
// - Handle network forks/termination/merging/replacement ?
// - Provide transaction in stasis/sudo (voting?)
// - Handle bonding by other means (e.g. pre-shared key permitting n bondings)
// - Network administered proxies (i.e. instead of keys have password authentication for identities - allow calls to originate as if from address without key?)
// Subject to:
// - Less than 1/3 validator power change per block

// Creates a GovTx that alters the validator power of id to the power passed
func AlterPowerTx(inputAddress crypto.Address, id crypto.Addressable, power uint64) *GovTx {
	return AlterBalanceTx(inputAddress, id, balance.New().Power(power))
}

func AlterBalanceTx(inputAddress crypto.Address, id crypto.Addressable, bal balance.Balances) *GovTx {
	publicKey := id.GetPublicKey()
	return UpdateAccountTx(inputAddress, &spec.TemplateAccount{
		PublicKey: &publicKey,
		Amounts:   bal,
	})
}

func AlterPermissionsTx(inputAddress crypto.Address, id crypto.Addressable, perms permission.PermFlag) *GovTx {
	address := id.GetAddress()
	return UpdateAccountTx(inputAddress, &spec.TemplateAccount{
		Address:     &address,
		Permissions: permission.PermFlagToStringList(perms),
	})
}

func UpdateAccountTx(inputAddress crypto.Address, updates ...*spec.TemplateAccount) *GovTx {
	return &GovTx{
		Inputs: []*TxInput{{
			Address: inputAddress,
		}},
		AccountUpdates: updates,
	}
}

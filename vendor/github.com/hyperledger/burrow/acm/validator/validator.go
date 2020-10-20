package validator

import (
	"fmt"
	"math/big"

	"github.com/hyperledger/burrow/crypto"

	"github.com/hyperledger/burrow/acm"
)

func New(publicKey crypto.PublicKey, power *big.Int) *Validator {
	v := &Validator{
		PublicKey: publicKey,
		Power:     power.Uint64(),
	}
	v.FillAddress()
	return v
}

func (v *Validator) String() string {
	return fmt.Sprintf("Validator{Address: %v, PublicKey: %v, Power: %v}", v.Address, v.PublicKey, v.Power)
}

func (v *Validator) FillAddress() {
	if v.Address == nil {
		address := v.PublicKey.GetAddress()
		v.Address = &address
	}
}

func (v *Validator) BigPower() *big.Int {
	return new(big.Int).SetUint64(v.Power)
}

func (v *Validator) GetAddress() crypto.Address {
	return *v.Address
}

func FromAccount(acc *acm.Account, power uint64) *Validator {
	pubKey := acc.GetPublicKey()
	address, _ := crypto.AddressFromBytes(pubKey.TendermintAddress().Bytes())
	return &Validator{
		Address:   &address,
		PublicKey: pubKey,
		Power:     power,
	}
}

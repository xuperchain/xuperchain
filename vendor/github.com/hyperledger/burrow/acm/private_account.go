// Copyright Monax Industries Limited
// SPDX-License-Identifier: Apache-2.0

package acm

import (
	"fmt"

	"encoding/json"

	"github.com/hyperledger/burrow/crypto"
)

type AddressableSigner interface {
	crypto.Addressable
	crypto.Signer
}

type PrivateAccount struct {
	concretePrivateAccount *ConcretePrivateAccount
}

func (pa *PrivateAccount) GetAddress() crypto.Address {
	return pa.concretePrivateAccount.Address
}

func (pa *PrivateAccount) GetPublicKey() crypto.PublicKey {
	return pa.concretePrivateAccount.PublicKey
}

func (pa *PrivateAccount) Sign(msg []byte) (*crypto.Signature, error) {
	return pa.concretePrivateAccount.PrivateKey.Sign(msg)
}

func (pa PrivateAccount) MarshalJSON() ([]byte, error) {
	return json.Marshal(pa.concretePrivateAccount)
}

func (pa *PrivateAccount) UnmarshalJSON(bytes []byte) error {
	err := json.Unmarshal(bytes, &pa.concretePrivateAccount)
	if err != nil {
		return err
	}
	return nil
}

func (pa *PrivateAccount) PrivateKey() crypto.PrivateKey {
	return pa.concretePrivateAccount.PrivateKey
}

func (pa *PrivateAccount) ConcretePrivateAccount() *ConcretePrivateAccount {
	cpa := *pa.concretePrivateAccount
	return &cpa
}

func (pa *PrivateAccount) String() string {
	return fmt.Sprintf("PrivateAccount{%v}", pa.GetAddress())
}

type ConcretePrivateAccount struct {
	Address    crypto.Address
	PublicKey  crypto.PublicKey
	PrivateKey crypto.PrivateKey
}

func (cpa *ConcretePrivateAccount) String() string {
	return fmt.Sprintf("ConcretePrivateAccount{%v}", cpa.Address)
}

func (cpa ConcretePrivateAccount) PrivateAccount() *PrivateAccount {
	return &PrivateAccount{
		concretePrivateAccount: &cpa,
	}
}

func PrivateAccountFromPrivateKey(privateKey crypto.PrivateKey) *PrivateAccount {
	publicKey := privateKey.GetPublicKey()
	return &PrivateAccount{
		concretePrivateAccount: &ConcretePrivateAccount{
			PrivateKey: privateKey,
			PublicKey:  publicKey,
			Address:    publicKey.GetAddress(),
		},
	}
}

// Convert slice of ConcretePrivateAccounts to slice of SigningAccounts
func SigningAccounts(concretePrivateAccounts []*PrivateAccount) []AddressableSigner {
	signingAccounts := make([]AddressableSigner, len(concretePrivateAccounts))
	for i, cpa := range concretePrivateAccounts {
		signingAccounts[i] = cpa
	}
	return signingAccounts
}

// Generates a new account with private key.
func GeneratePrivateAccount(ct crypto.CurveType) (*PrivateAccount, error) {
	privateKey, err := crypto.GeneratePrivateKey(nil, ct)
	if err != nil {
		return nil, err
	}
	publicKey := privateKey.GetPublicKey()
	return ConcretePrivateAccount{
		Address:    publicKey.GetAddress(),
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}.PrivateAccount(), nil
}

func privateAccount(privateKey crypto.PrivateKey) *PrivateAccount {
	publicKey := privateKey.GetPublicKey()
	return ConcretePrivateAccount{
		Address:    publicKey.GetAddress(),
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}.PrivateAccount()
}

// Generates a new account with private key from SHA256 hash of a secret
func GeneratePrivateAccountFromSecret(secret string) *PrivateAccount {
	return privateAccount(crypto.PrivateKeyFromSecret(secret, crypto.CurveTypeEd25519))

}

func GenerateEthereumAccountFromSecret(secret string) *PrivateAccount {
	return privateAccount(crypto.PrivateKeyFromSecret(secret, crypto.CurveTypeSecp256k1))
}

package genesis

import (
	"fmt"
	"math/big"
	"math/rand"
	"time"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/permission"
)

type deterministicGenesis struct {
	random *rand.Rand
}

// Generates deterministic pseudo-random genesis state
func NewDeterministicGenesis(seed int64) *deterministicGenesis {
	return &deterministicGenesis{
		random: rand.New(rand.NewSource(seed)),
	}
}

func (dg *deterministicGenesis) GenesisDoc(numAccounts int, numValidators int) (*GenesisDoc, []*acm.PrivateAccount, []*acm.PrivateAccount) {
	accounts := make([]Account, numAccounts+numValidators)
	privAccounts := make([]*acm.PrivateAccount, numAccounts)
	defaultPerms := permission.DefaultAccountPermissions
	for i := 0; i < numAccounts; i++ {
		account, privAccount := dg.Account(9999999)
		acc := Account{
			BasicAccount: BasicAccount{
				Address: account.GetAddress(),
				Amount:  account.Balance.Uint64(),
			},
			Permissions: defaultPerms.Clone(), // This will get copied into each state.Account.
		}
		acc.Permissions.Base.Set(permission.Root, true)
		accounts[i] = acc
		privAccounts[i] = privAccount
	}
	validators := make([]Validator, numValidators)
	privValidators := make([]*acm.PrivateAccount, numValidators)
	for i := 0; i < numValidators; i++ {
		validator := acm.GeneratePrivateAccountFromSecret(fmt.Sprintf("val_%v", i))
		privValidators[i] = validator
		basicAcc := BasicAccount{
			Address:   validator.GetAddress(),
			PublicKey: validator.GetPublicKey(),
			// Avoid max validator cap
			Amount: uint64(dg.random.Int63()/16 + 1),
		}
		fullAcc := Account{
			BasicAccount: basicAcc,
			Permissions:  defaultPerms.Clone(),
		}
		accounts[numAccounts+i] = fullAcc
		validators[i] = Validator{
			BasicAccount: basicAcc,
			UnbondTo: []BasicAccount{
				{
					Address: validator.GetAddress(),
					Amount:  uint64(dg.random.Int63()),
				},
			},
		}
	}
	return &GenesisDoc{
		ChainName:   "TestChain",
		GenesisTime: time.Unix(1506172037, 0).UTC(),
		Accounts:    accounts,
		Validators:  validators,
	}, privAccounts, privValidators

}

func (dg *deterministicGenesis) Account(minBalance uint64) (*acm.Account, *acm.PrivateAccount) {
	privateKey, err := crypto.GeneratePrivateKey(dg.random, crypto.CurveTypeEd25519)
	if err != nil {
		panic(fmt.Errorf("could not generate private key deterministically"))
	}
	privAccount := &acm.ConcretePrivateAccount{
		PublicKey:  privateKey.GetPublicKey(),
		PrivateKey: privateKey,
		Address:    privateKey.GetPublicKey().GetAddress(),
	}
	perms := permission.DefaultAccountPermissions
	acc := &acm.Account{
		Address:     privAccount.Address,
		PublicKey:   privAccount.PublicKey,
		Sequence:    uint64(dg.random.Int()),
		Balance:     big.NewInt(int64(minBalance)),
		Permissions: perms,
	}
	acc.Balance = acc.Balance.Add(acc.Balance, big.NewInt(int64(dg.random.Int())))
	return acc, privAccount.PrivateAccount()
}

package native

import (
	"bytes"
	"math/big"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/deploy/compile"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/txs/payload"
	"golang.org/x/crypto/sha3"
)

func CreateAccount(st acmstate.ReaderWriter, address crypto.Address) error {
	acc, err := st.GetAccount(address)
	if err != nil {
		return err
	}
	if acc != nil {
		if acc.NativeName != "" {
			return errors.Errorf(errors.Codes.ReservedAddress,
				"cannot create account at %v because that address is reserved for a native contract '%s'",
				address, acc.NativeName)
		}
		return errors.Errorf(errors.Codes.DuplicateAddress,
			"tried to create an account at an address that already exists: %v", address)
	}
	return st.UpdateAccount(&acm.Account{Address: address})
}

func InitEVMCode(st acmstate.ReaderWriter, address crypto.Address, code []byte) error {
	return initEVMCode(st, address, nil, code)
}

func InitChildCode(st acmstate.ReaderWriter, address crypto.Address, parent crypto.Address, code []byte) error {
	return initEVMCode(st, address, &parent, code)
}

func initEVMCode(st acmstate.ReaderWriter, address crypto.Address, parent *crypto.Address, code []byte) error {
	acc, err := mustAccount(st, address)
	if err != nil {
		return err
	}
	if acc.EVMCode != nil || acc.WASMCode != nil {
		return errors.Errorf(errors.Codes.IllegalWrite,
			"tried to initialise code for a contract that already has code: %v", address)
	}

	acc.EVMCode = code

	// keccak256 hash of a contract's code
	hash := sha3.NewLegacyKeccak256()
	hash.Write(code)
	codehash := hash.Sum(nil)

	forebear := &address
	metamap := acc.ContractMeta
	if parent != nil {
		// find our ancestor, i.e. the initial contract that was deployed, from which this contract descends
		ancestor, err := st.GetAccount(*parent)
		if err != nil {
			return err
		}
		if ancestor == nil {
			return errors.Errorf(errors.Codes.NonExistentAccount,
				"parent %v of account %v does not exist", *parent, address)
		}
		if ancestor.Forebear != nil {
			ancestor, err = st.GetAccount(*ancestor.Forebear)
			if err != nil {
				return err
			}
			if ancestor == nil {
				return errors.Errorf(errors.Codes.NonExistentAccount,
					"forebear %v of account %v does not exist", *ancestor.Forebear, *parent)
			}
			forebear = ancestor.Forebear
		} else {
			forebear = parent
		}
		metamap = ancestor.ContractMeta
	}

	// If we have a list of ABIs for this contract, we also know what contract code it is allowed to create
	// For compatibility with older contracts, allow any contract to be created if we have no mappings
	if metamap != nil && len(metamap) > 0 {
		found := codehashPermitted(codehash, metamap)

		// Libraries lie about their deployed bytecode
		if !found {
			deployCodehash := compile.GetDeployCodeHash(code, address)
			found = codehashPermitted(deployCodehash, metamap)
		}

		if !found {
			return errors.Errorf(errors.Codes.InvalidContractCode,
				"could not find code with code hash: %X", codehash)
		}
	}

	acc.CodeHash = codehash
	acc.Forebear = forebear

	return st.UpdateAccount(acc)
}

func codehashPermitted(codehash []byte, metamap []*acm.ContractMeta) bool {
	for _, m := range metamap {
		if bytes.Equal(codehash, m.CodeHash) {
			return true
		}
	}

	return false
}

func InitWASMCode(st acmstate.ReaderWriter, address crypto.Address, code []byte) error {
	acc, err := mustAccount(st, address)
	if err != nil {
		return err
	}
	if acc.EVMCode != nil || acc.WASMCode != nil {
		return errors.Errorf(errors.Codes.IllegalWrite,
			"tried to re-initialise code for contract %v", address)
	}

	acc.WASMCode = code
	// keccak256 hash of a contract's code
	hash := sha3.NewLegacyKeccak256()
	hash.Write(code)
	acc.CodeHash = hash.Sum(nil)
	return st.UpdateAccount(acc)
}

func Transfer(st acmstate.ReaderWriter, from, to crypto.Address, amount uint64) error {
	if amount == 0 {
		return nil
	}
	acc, err := mustAccount(st, from)
	if err != nil {
		return err
	}
	if acc.Balance.Cmp(big.NewInt(int64(amount))) == -1 {
		return errors.Codes.InsufficientBalance
	}
	err = UpdateAccount(st, from, func(account *acm.Account) error {
		return account.SubtractFromBalance(big.NewInt(int64(amount)))
	})
	if err != nil {
		return err
	}
	return UpdateAccount(st, to, func(account *acm.Account) error {
		return account.AddToBalance(big.NewInt(int64(amount)))
	})
}

func UpdateContractMeta(st acmstate.ReaderWriter, metaSt acmstate.MetadataWriter, address crypto.Address, payloadMeta []*payload.ContractMeta) error {
	if len(payloadMeta) == 0 {
		return nil
	}
	acc, err := mustAccount(st, address)
	if err != nil {
		return err
	}

	contractMeta := make([]*acm.ContractMeta, len(payloadMeta))
	for i, abi := range payloadMeta {
		metahash := acmstate.GetMetadataHash(abi.Meta)
		contractMeta[i] = &acm.ContractMeta{
			MetadataHash: metahash[:],
			CodeHash:     abi.CodeHash,
		}
		err = metaSt.SetMetadata(metahash, abi.Meta)
		if err != nil {
			return errors.Errorf(errors.Codes.IllegalWrite,
				"cannot update metadata for %v: %v", address, err)
		}
	}
	acc.ContractMeta = contractMeta
	return st.UpdateAccount(acc)
}

func RemoveAccount(st acmstate.ReaderWriter, address crypto.Address) error {
	acc, err := st.GetAccount(address)
	if err != nil {
		return err
	}
	if acc == nil {
		return errors.Errorf(errors.Codes.DuplicateAddress,
			"tried to remove an account at an address that does not exist: %v", address)
	}
	return st.RemoveAccount(address)
}

func UpdateAccount(st acmstate.ReaderWriter, address crypto.Address, updater func(acc *acm.Account) error) error {
	acc, err := mustAccount(st, address)
	if err != nil {
		return err
	}
	err = updater(acc)
	if err != nil {
		return err
	}
	return st.UpdateAccount(acc)
}

func mustAccount(st acmstate.Reader, address crypto.Address) (*acm.Account, error) {
	acc, err := st.GetAccount(address)
	if err != nil {
		return nil, err
	}
	if acc == nil {
		return nil, errors.Errorf(errors.Codes.NonExistentAccount,
			"account %v does not exist", address)
	}
	return acc, nil
}

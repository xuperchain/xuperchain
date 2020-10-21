package acmstate

import (
	"fmt"
	"math/big"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"golang.org/x/crypto/sha3"

	"github.com/hyperledger/burrow/permission"
	"github.com/tmthrgd/go-hex"
)

// MetadataHash is the keccak hash for the metadata. This is to make the metadata content-addressed
type MetadataHash [32]byte

func (h *MetadataHash) Bytes() []byte {
	b := make([]byte, 32)
	copy(b, h[:])
	return b
}

func (ch *MetadataHash) UnmarshalText(hexBytes []byte) error {
	bs, err := hex.DecodeString(string(hexBytes))
	if err != nil {
		return err
	}
	copy(ch[:], bs)
	return nil
}

func (ch MetadataHash) MarshalText() ([]byte, error) {
	return []byte(ch.String()), nil
}

func (ch MetadataHash) String() string {
	return hex.EncodeUpperToString(ch[:])
}

func GetMetadataHash(metadata string) (metahash MetadataHash) {
	hash := sha3.NewLegacyKeccak256()
	hash.Write([]byte(metadata))
	copy(metahash[:], hash.Sum(nil))
	return
}

// CodeHash is the keccak hash for the code for an account. This is used for the EVM CODEHASH opcode, and to find the
// correct Metadata for a contract
type CodeHash [32]byte

func (h *CodeHash) Bytes() []byte {
	b := make([]byte, 32)
	copy(b, h[:])
	return b
}

func (ch *CodeHash) UnmarshalText(hexBytes []byte) error {
	bs, err := hex.DecodeString(string(hexBytes))
	if err != nil {
		return err
	}
	copy(ch[:], bs)
	return nil
}

func (ch CodeHash) MarshalText() ([]byte, error) {
	return []byte(ch.String()), nil
}

func (ch CodeHash) String() string {
	return hex.EncodeUpperToString(ch[:])
}

type AccountGetter interface {
	// Get an account by its address return nil if it does not exist (which should not be an error)
	GetAccount(address crypto.Address) (*acm.Account, error)
}

type AccountIterable interface {
	// Iterates through accounts calling passed function once per account, if the consumer
	// returns true the iteration breaks and returns true to indicate it iteration
	// was escaped
	IterateAccounts(consumer func(*acm.Account) error) (err error)
}

type AccountUpdater interface {
	// Updates the fields of updatedAccount by address, creating the account
	// if it does not exist
	UpdateAccount(updatedAccount *acm.Account) error
	// Remove the account at address
	RemoveAccount(address crypto.Address) error
	// Transfer
	Transfer(from, to crypto.Address, amount *big.Int) error
}

type StorageGetter interface {
	// Retrieve a 32-byte value stored at key for the account at address, return Zero256 if key does not exist but
	// error if address does not
	GetStorage(address crypto.Address, key binary.Word256) (value []byte, err error)
}

type StorageSetter interface {
	// Store a 32-byte value at key for the account at address, setting to Zero256 removes the key
	SetStorage(address crypto.Address, key binary.Word256, value []byte) error
}

type StorageIterable interface {
	// Iterates through the storage of account ad address calling the passed function once per account,
	// if the iterator function returns true the iteration breaks and returns true to indicate it iteration
	// was escaped
	IterateStorage(address crypto.Address, consumer func(key binary.Word256, value []byte) error) (err error)
}

type MetadataReader interface {
	// Get an Metadata by its hash. This is content-addressed
	GetMetadata(metahash MetadataHash) (string, error)
}

type MetadataWriter interface {
	// Set an Metadata according to it keccak-256 hash.
	SetMetadata(metahash MetadataHash, metadata string) error
}

type AccountStats struct {
	AccountsWithCode    uint64
	AccountsWithoutCode uint64
}

type AccountStatsGetter interface {
	GetAccountStats() AccountStats
}

// Compositions

// Read-only account and storage state
type Reader interface {
	AccountGetter
	StorageGetter
}

type Iterable interface {
	AccountIterable
	StorageIterable
}

// Read and list account and storage state
type IterableReader interface {
	Iterable
	Reader
}

type IterableStatsReader interface {
	Iterable
	Reader
	AccountStatsGetter
}

type Writer interface {
	AccountUpdater
	StorageSetter
}

// Read and write account and storage state
type ReaderWriter interface {
	Reader
	Writer
}

type MetadataReaderWriter interface {
	MetadataReader
	MetadataWriter
}

type IterableReaderWriter interface {
	Iterable
	Reader
	Writer
}

// Get global permissions from the account at GlobalPermissionsAddress
func GlobalAccountPermissions(getter AccountGetter) (permission.AccountPermissions, error) {
	acc, err := getter.GetAccount(acm.GlobalPermissionsAddress)
	if err != nil {
		return permission.AccountPermissions{}, err
	}
	if acc == nil {
		return permission.AccountPermissions{}, fmt.Errorf("global permissions account is not defined but must be")
	}
	return acc.Permissions, nil
}

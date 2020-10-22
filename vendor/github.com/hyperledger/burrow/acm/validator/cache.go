package validator

import (
	"math/big"

	"github.com/hyperledger/burrow/crypto"
)

// Cache is just a Ring with no memory
type Cache struct {
	*Bucket
}

func NewCache(backend Iterable) *Cache {
	return &Cache{
		Bucket: NewBucket(backend),
	}
}

func (vc *Cache) Reset(backend Iterable) {
	vc.Bucket = NewBucket(backend)
}

func (vc *Cache) Sync(output Writer) error {
	err := vc.Delta.IterateValidators(func(id crypto.Addressable, power *big.Int) error {
		_, err := output.SetPower(id.GetPublicKey(), power)
		return err
	})
	if err != nil {
		return err
	}
	return nil
}

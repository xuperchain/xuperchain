package keys

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/crypto"
)

type MemoryKeyStore struct {
	keyByAddress map[crypto.Address]crypto.PrivateKey
	keyByName    map[string]crypto.PrivateKey
}

func NewMemoryKeyStore(privateAccounts ...*acm.PrivateAccount) *MemoryKeyStore {
	mks := &MemoryKeyStore{
		keyByAddress: make(map[crypto.Address]crypto.PrivateKey),
		keyByName:    make(map[string]crypto.PrivateKey),
	}
	for _, pa := range privateAccounts {
		mks.keyByAddress[pa.GetAddress()] = pa.PrivateKey()
	}
	return mks
}

func (mks *MemoryKeyStore) GetAddressForKeyName(keyName string) (crypto.Address, error) {
	key, ok := mks.keyByName[keyName]
	if !ok {
		return crypto.Address{}, fmt.Errorf("could not find key with name %s", keyName)
	}
	return key.GetPublicKey().GetAddress(), nil
}

func (mks *MemoryKeyStore) GenerateKey(ctx context.Context, in *GenRequest) (*GenResponse, error) {
	curveType, err := crypto.CurveTypeFromString(in.CurveType)
	if err != nil {
		return nil, fmt.Errorf("unknown curve type '%s'", in.CurveType)
	}
	key, err := crypto.GeneratePrivateKey(rand.Reader, curveType)
	if err != nil {
		return nil, fmt.Errorf("could not generate key: %w", err)
	}

	address := key.GetPublicKey().GetAddress()
	mks.keyByAddress[address] = key
	if in.KeyName != "" {
		mks.keyByName[in.KeyName] = key
	}

	return &GenResponse{
		Address: address.String(),
	}, nil
}

func (mks *MemoryKeyStore) PublicKey(ctx context.Context, in *PubRequest) (*PubResponse, error) {
	key, err := mks.getKey(in.Name, in.Address)
	if err != nil {
		return nil, err
	}
	return &PubResponse{
		CurveType: key.CurveType.String(),
		PublicKey: key.PublicKey,
	}, nil
}

func (mks *MemoryKeyStore) Sign(ctx context.Context, in *SignRequest) (*SignResponse, error) {
	key, err := mks.getKey(in.Name, in.Address)
	if err != nil {
		return nil, err
	}
	signature, err := key.Sign(in.Message)
	if err != nil {
		return nil, fmt.Errorf("could not sign message: %w", err)
	}
	return &SignResponse{
		Signature: signature,
	}, nil
}

// Get a stringly referenced key first by name, then by address
func (mks *MemoryKeyStore) getKey(name string, addressHex string) (*crypto.PrivateKey, error) {
	key, ok := mks.keyByName[name]
	if !ok {
		address, err := crypto.AddressFromHexString(addressHex)
		if err != nil {
			return nil, fmt.Errorf("could not get PublicKey: %w", err)
		}
		key, ok = mks.keyByAddress[address]
		if !ok {
			return nil, fmt.Errorf("could not find key with address %v: %w", address, err)
		}
	}
	return &key, nil
}

package keys

import (
	"context"

	"github.com/hyperledger/burrow/crypto"
)

const (
	scryptN       = 1 << 18
	scryptr       = 8
	scryptp       = 1
	scryptdkLen   = 32
	CryptoNone    = "none"
	CryptoAESGCM  = "scrypt-aes-gcm"
	HashEd25519   = "go-crypto-0.5.0"
	HashSecp256k1 = "btc"
)

type KeyStore interface {
	GetAddressForKeyName(keyName string) (keyAddress crypto.Address, err error)
	GenerateKey(ctx context.Context, in *GenRequest) (*GenResponse, error)
	PublicKey(ctx context.Context, in *PubRequest) (*PubResponse, error)
	Sign(ctx context.Context, in *SignRequest) (*SignResponse, error)
}

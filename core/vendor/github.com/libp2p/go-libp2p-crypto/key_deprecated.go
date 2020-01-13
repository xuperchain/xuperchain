// Deprecated: use github.com/libp2p/go-libp2p-core/crypto instead.
package crypto

import (
	"io"

	core "github.com/libp2p/go-libp2p-core/crypto"
)

const (
	// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.RSA instead.
	RSA = core.RSA
	// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.Ed25519 instead.
	Ed25519 = core.Ed25519
	// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.Secp256k1 instead.
	Secp256k1 = core.Secp256k1
	// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.ECDSA instead.
	ECDSA = core.ECDSA
)

var (
	// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.ErrBadKeyType instead.
	ErrBadKeyType = core.ErrBadKeyType
	// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.KeyTypes instead.
	KeyTypes = core.KeyTypes
)

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.PubKeyUnmarshaller instead.
type PubKeyUnmarshaller = core.PubKeyUnmarshaller

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.PrivKeyUnmarshaller instead.
type PrivKeyUnmarshaller = core.PrivKeyUnmarshaller

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.PubKeyUnmarshallers instead.
var PubKeyUnmarshallers = core.PubKeyUnmarshallers

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.PrivKeyUnmarshallers instead.
var PrivKeyUnmarshallers = core.PrivKeyUnmarshallers

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.Key instead.
type Key = core.Key

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.PrivKey instead.
type PrivKey = core.PrivKey

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.PubKey instead.
type PubKey = core.PubKey

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.GenSharedKey instead.
type GenSharedKey = core.GenSharedKey

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.GenerateKeyPair instead.
func GenerateKeyPair(typ, bits int) (PrivKey, PubKey, error) {
	return core.GenerateKeyPair(typ, bits)
}

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.GenerateKeyPairWithReader instead.
func GenerateKeyPairWithReader(typ, bits int, src io.Reader) (PrivKey, PubKey, error) {
	return core.GenerateKeyPairWithReader(typ, bits, src)
}

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.GenerateEKeyPair instead.
func GenerateEKeyPair(curveName string) ([]byte, GenSharedKey, error) {
	return core.GenerateEKeyPair(curveName)
}

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.GenSharedKey instead.
type StretchedKeys = core.StretchedKeys

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.KeyStretcher instead.
func KeyStretcher(cipherType string, hashType string, secret []byte) (StretchedKeys, StretchedKeys) {
	return core.KeyStretcher(cipherType, hashType, secret)
}

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.UnmarshalPublicKey instead.
func UnmarshalPublicKey(data []byte) (PubKey, error) {
	return core.UnmarshalPublicKey(data)
}

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.MarshalPublicKey instead.
func MarshalPublicKey(k PubKey) ([]byte, error) {
	return core.MarshalPublicKey(k)
}

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.UnmarshalPrivateKey instead.
func UnmarshalPrivateKey(data []byte) (PrivKey, error) {
	return core.UnmarshalPrivateKey(data)
}

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.MarshalPrivateKey instead.
func MarshalPrivateKey(k PrivKey) ([]byte, error) {
	return core.MarshalPrivateKey(k)
}

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.ConfigDecodeKey instead.
func ConfigDecodeKey(b string) ([]byte, error) {
	return core.ConfigDecodeKey(b)
}

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.ConfigEncodeKey instead.
func ConfigEncodeKey(b []byte) string {
	return core.ConfigEncodeKey(b)
}

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.KeyEqual instead.
func KeyEqual(k1, k2 Key) bool {
	return core.KeyEqual(k1, k2)
}

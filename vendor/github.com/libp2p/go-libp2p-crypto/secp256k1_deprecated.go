package crypto

import (
	"io"

	core "github.com/libp2p/go-libp2p-core/crypto"
)

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.Secp256k1PrivateKey instead.
type Secp256k1PrivateKey = core.Secp256k1PrivateKey

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.Secp256k1PublicKey instead.
type Secp256k1PublicKey = core.Secp256k1PublicKey

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.GenerateSecp256k1Key instead.
func GenerateSecp256k1Key(src io.Reader) (PrivKey, PubKey, error) {
	return core.GenerateSecp256k1Key(src)
}

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.UnmarshalSecp256k1PrivateKey instead.
func UnmarshalSecp256k1PrivateKey(data []byte) (PrivKey, error) {
	return core.UnmarshalSecp256k1PrivateKey(data)
}

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.UnmarshalSecp256k1PublicKey instead.
func UnmarshalSecp256k1PublicKey(data []byte) (PubKey, error) {
	return core.UnmarshalSecp256k1PublicKey(data)
}

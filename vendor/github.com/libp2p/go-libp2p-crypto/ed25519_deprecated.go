package crypto

import (
	"io"

	core "github.com/libp2p/go-libp2p-core/crypto"
)

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.Ed25519PrivateKey instead.
type Ed25519PrivateKey = core.Ed25519PrivateKey

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.Ed25519PublicKey instead.
type Ed25519PublicKey = core.Ed25519PublicKey

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.GenerateEd25519Key instead.
func GenerateEd25519Key(src io.Reader) (PrivKey, PubKey, error) {
	return core.GenerateEd25519Key(src)
}

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.UnmarshalEd25519PublicKey instead.
func UnmarshalEd25519PublicKey(data []byte) (PubKey, error) {
	return core.UnmarshalEd25519PublicKey(data)
}

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.UnmarshalEd25519PrivateKey instead.
func UnmarshalEd25519PrivateKey(data []byte) (PrivKey, error) {
	return core.UnmarshalEd25519PrivateKey(data)
}

package crypto

import (
	"io"

	core "github.com/libp2p/go-libp2p-core/crypto"
)

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.ErrRsaKeyTooSmall instead.
var ErrRsaKeyTooSmall = core.ErrRsaKeyTooSmall

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.RsaPrivateKey instead.
type RsaPrivateKey = core.RsaPrivateKey

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.RsaPublicKey instead.
type RsaPublicKey = core.RsaPublicKey

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.GenerateRSAKeyPair instead.
func GenerateRSAKeyPair(bits int, src io.Reader) (PrivKey, PubKey, error) {
	return core.GenerateRSAKeyPair(bits, src)
}

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.UnmarshalRsaPrivateKey instead.
func UnmarshalRsaPrivateKey(b []byte) (PrivKey, error) {
	return core.UnmarshalRsaPrivateKey(b)
}

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.UnmarshalRsaPublicKey instead.
func UnmarshalRsaPublicKey(b []byte) (PubKey, error) {
	return core.UnmarshalRsaPublicKey(b)
}

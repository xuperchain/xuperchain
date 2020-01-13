package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"io"

	core "github.com/libp2p/go-libp2p-core/crypto"
)

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.ECDSAPrivateKey instead.
type ECDSAPrivateKey = core.ECDSAPrivateKey

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.ECDSAPublicKey instead.
type ECDSAPublicKey = core.ECDSAPublicKey

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.ECDSASig instead.
type ECDSASig = core.ECDSASig

var (
	// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.ErrNotECDSAPubKey instead.
	ErrNotECDSAPubKey = core.ErrNotECDSAPubKey
	// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.ErrNilSig instead.
	ErrNilSig = core.ErrNilSig
	// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.ErrNilPrivateKey instead.
	ErrNilPrivateKey = core.ErrNilPrivateKey
	// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.ECDSACurve instead.
	ECDSACurve = core.ECDSACurve
)

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.GenerateECDSAKeyPair instead.
func GenerateECDSAKeyPair(src io.Reader) (PrivKey, PubKey, error) {
	return core.GenerateECDSAKeyPair(src)
}

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.GenerateECDSAKeyPairWithCurve instead.
func GenerateECDSAKeyPairWithCurve(curve elliptic.Curve, src io.Reader) (PrivKey, PubKey, error) {
	return core.GenerateECDSAKeyPairWithCurve(curve, src)
}

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.ECDSAKeyPairFromKey instead.
func ECDSAKeyPairFromKey(priv *ecdsa.PrivateKey) (PrivKey, PubKey, error) {
	return core.ECDSAKeyPairFromKey(priv)
}

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.MarshalECDSAPrivateKey instead.
func MarshalECDSAPrivateKey(ePriv ECDSAPrivateKey) ([]byte, error) {
	return core.MarshalECDSAPrivateKey(ePriv)
}

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.MarshalECDSAPublicKey instead.
func MarshalECDSAPublicKey(ePub ECDSAPublicKey) ([]byte, error) {
	return core.MarshalECDSAPublicKey(ePub)
}

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.UnmarshalECDSAPrivateKey instead.
func UnmarshalECDSAPrivateKey(data []byte) (PrivKey, error) {
	return core.UnmarshalECDSAPrivateKey(data)
}

// Deprecated: use github.com/libp2p/go-libp2p-core/crypto.UnmarshalECDSAPublicKey instead.
func UnmarshalECDSAPublicKey(data []byte) (PubKey, error) {
	return core.UnmarshalECDSAPublicKey(data)
}

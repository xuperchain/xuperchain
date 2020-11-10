package crypto

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	hex "github.com/tmthrgd/go-hex"
	"golang.org/x/crypto/ed25519"
)

func SignatureFromBytes(bs []byte, curveType CurveType) (*Signature, error) {
	switch curveType {
	case CurveTypeEd25519:
		if len(bs) != ed25519.SignatureSize {
			return nil, fmt.Errorf("bytes passed have length %v by ed25519 signatures have %v bytes",
				len(bs), ed25519.SignatureSize)
		}
	case CurveTypeSecp256k1:
		// TODO: validate?
	}

	return &Signature{CurveType: curveType, Signature: bs}, nil
}

func (sig *Signature) RawBytes() []byte {
	return sig.Signature
}

func (sig *Signature) String() string {
	return hex.EncodeUpperToString(sig.Signature)
}

func CompressedSignatureFromParams(v uint64, r, s []byte) []byte {
	bitlen := (btcec.S256().BitSize + 7) / 8
	sig := make([]byte, 1+bitlen*2)
	sig[0] = byte(v)
	copy(sig[1:bitlen+1], r)
	copy(sig[bitlen+1:], s)
	return sig
}

func UncompressedSignatureFromParams(r, s []byte) []byte {
	// <0x30> <length of whole message>
	// <0x02> <length of R> <R>
	// <0x2> <length of S> <S>

	rr := append([]byte{0x02, byte(len(r))}, r...)
	ss := append([]byte{0x2, byte(len(s))}, s...)
	rrss := append(rr, ss...)

	return append([]byte{0x30, byte(len(rrss))}, rrss...)
}

// PublicKeyFromSignature verifies an ethereum compact signature and returns the public key if valid
func PublicKeyFromSignature(sig, hash []byte) (*PublicKey, error) {
	pub, _, err := btcec.RecoverCompact(btcec.S256(), sig, hash)
	if err != nil {
		return nil, err
	}

	publicKey, err := PublicKeyFromBytes(pub.SerializeCompressed(), CurveTypeSecp256k1)
	if err != nil {
		return nil, err
	}

	return &publicKey, nil
}

package crypto

import (
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/btcec"
	abci "github.com/tendermint/tendermint/abci/types"
	tmCrypto "github.com/tendermint/tendermint/crypto"
	tmEd25519 "github.com/tendermint/tendermint/crypto/ed25519"
	tmSecp256k1 "github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/p2p"
)

func NodeIDFromAddress(id Address) p2p.ID {
	return p2p.ID(strings.ToLower(id.String()))
}

func PublicKeyFromTendermintPubKey(pubKey tmCrypto.PubKey) (PublicKey, error) {
	switch pk := pubKey.(type) {
	case tmEd25519.PubKeyEd25519:
		return PublicKeyFromBytes(pk[:], CurveTypeEd25519)
	case tmSecp256k1.PubKeySecp256k1:
		return PublicKeyFromBytes(pk[:], CurveTypeSecp256k1)
	default:
		return PublicKey{}, fmt.Errorf("unrecognised tendermint public key type: %v", pk)
	}

}
func PublicKeyFromABCIPubKey(pubKey abci.PubKey) (PublicKey, error) {
	switch pubKey.Type {
	case CurveTypeEd25519.ABCIType():
		return PublicKey{
			CurveType: CurveTypeEd25519,
			PublicKey: pubKey.Data,
		}, nil
	case CurveTypeSecp256k1.ABCIType():
		return PublicKey{
			CurveType: CurveTypeSecp256k1,
			PublicKey: pubKey.Data,
		}, nil
	}
	return PublicKey{}, fmt.Errorf("did not recognise ABCI PubKey type: %s", pubKey.Type)
}

// PublicKey extensions

// Return the ABCI PubKey. See Tendermint protobuf.go for the go-crypto conversion this is based on
func (p PublicKey) ABCIPubKey() abci.PubKey {
	return abci.PubKey{
		Type: p.CurveType.ABCIType(),
		Data: p.PublicKey,
	}
}

func (p PublicKey) TendermintPubKey() tmCrypto.PubKey {
	switch p.CurveType {
	case CurveTypeEd25519:
		pk := tmEd25519.PubKeyEd25519{}
		copy(pk[:], p.PublicKey)
		return pk
	case CurveTypeSecp256k1:
		pk := tmSecp256k1.PubKeySecp256k1{}
		copy(pk[:], p.PublicKey)
		return pk
	default:
		return nil
	}
}

func (p PublicKey) TendermintAddress() tmCrypto.Address {
	switch p.CurveType {
	case CurveTypeEd25519:
		return tmCrypto.Address(p.GetAddress().Bytes())
	case CurveTypeSecp256k1:
		// Tendermint represents addresses like Bitcoin
		return tmCrypto.Address(RIPEMD160(SHA256(p.PublicKey[:])))
	default:
		panic(fmt.Sprintf("unknown CurveType %d", p.CurveType))
	}
}

// Signature extensions

func (sig Signature) TendermintSignature() []byte {
	switch sig.CurveType {
	case CurveTypeSecp256k1:
		sig, err := btcec.ParseDERSignature(sig.GetSignature(), btcec.S256())
		if err != nil {
			return nil
		}
		// https://github.com/tendermint/tendermint/blob/master/crypto/secp256k1/secp256k1_nocgo.go#L62
		r := sig.R.Bytes()
		s := sig.S.Bytes()
		data := make([]byte, 64)
		copy(data[32-len(r):32], r)
		copy(data[64-len(s):64], s)
		return data
	}
	return sig.Signature
}

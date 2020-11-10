package keys

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/burrow/crypto"
	"github.com/tmthrgd/go-hex"
)

type Key struct {
	CurveType  crypto.CurveType
	Address    crypto.Address
	PublicKey  crypto.PublicKey
	PrivateKey crypto.PrivateKey
}

// json encodings - addresses should be hex encoded
type keyJSON struct {
	CurveType   string
	Address     string
	PublicKey   string
	AddressHash string
	PrivateKey  privateKeyJSON
}

type privateKeyJSON struct {
	Crypto     string
	Plain      string `json:",omitempty"`
	Salt       []byte `json:",omitempty"`
	Nonce      []byte `json:",omitempty"`
	CipherText []byte `json:",omitempty"`
}

func NewKey(typ crypto.CurveType) (*Key, error) {
	privKey, err := crypto.GeneratePrivateKey(nil, typ)
	if err != nil {
		return nil, err
	}
	pubKey := privKey.GetPublicKey()
	return &Key{
		CurveType:  typ,
		PublicKey:  pubKey,
		Address:    pubKey.GetAddress(),
		PrivateKey: privKey,
	}, nil
}

func NewKeyFromPub(curveType crypto.CurveType, PubKeyBytes []byte) (*Key, error) {
	pubKey, err := crypto.PublicKeyFromBytes(PubKeyBytes, curveType)
	if err != nil {
		return nil, err
	}

	return &Key{
		CurveType: curveType,
		PublicKey: pubKey,
		Address:   pubKey.GetAddress(),
	}, nil
}

func NewKeyFromPriv(curveType crypto.CurveType, PrivKeyBytes []byte) (*Key, error) {
	privKey, err := crypto.PrivateKeyFromRawBytes(PrivKeyBytes, curveType)

	if err != nil {
		return nil, err
	}

	pubKey := privKey.GetPublicKey()

	return &Key{
		CurveType:  curveType,
		Address:    pubKey.GetAddress(),
		PublicKey:  pubKey,
		PrivateKey: privKey,
	}, nil
}

func (k *Key) Pubkey() []byte {
	return k.PublicKey.PublicKey
}

func (k *Key) MarshalJSON() (j []byte, err error) {
	jStruct := keyJSON{
		CurveType:   k.CurveType.String(),
		Address:     hex.EncodeUpperToString(k.Address[:]),
		PublicKey:   hex.EncodeUpperToString(k.Pubkey()),
		AddressHash: k.PublicKey.AddressHashType(),
		PrivateKey:  privateKeyJSON{Crypto: CryptoNone, Plain: hex.EncodeUpperToString(k.PrivateKey.RawBytes())},
	}
	j, err = json.Marshal(jStruct)
	return j, err
}

func (k *Key) UnmarshalJSON(j []byte) (err error) {
	keyJ := new(keyJSON)
	err = json.Unmarshal(j, &keyJ)
	if err != nil {
		return err
	}
	if len(keyJ.PrivateKey.Plain) == 0 {
		return fmt.Errorf("no private key")
	}
	curveType, err := crypto.CurveTypeFromString(keyJ.CurveType)
	if err != nil {
		curveType = crypto.CurveTypeEd25519
	}
	privKey, err := hex.DecodeString(keyJ.PrivateKey.Plain)
	if err != nil {
		return err
	}
	k2, err := NewKeyFromPriv(curveType, privKey)
	if err != nil {
		return err
	}

	k.Address = k2.Address
	k.CurveType = curveType
	k.PublicKey = k2.PrivateKey.GetPublicKey()
	k.PrivateKey = k2.PrivateKey

	return nil
}

// returns the address if valid, nil otherwise
func isValidKeyJson(j []byte) []byte {
	j1 := new(keyJSON)
	e1 := json.Unmarshal(j, &j1)
	if e1 == nil {
		addr, _ := hex.DecodeString(j1.Address)
		return addr
	}
	return nil
}

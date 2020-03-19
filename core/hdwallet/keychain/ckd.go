// modified from https://github.com/btcsuite/btcutil

// References:
//   [BIP32]: BIP0032 - Hierarchical Deterministic Wallets
//   https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki

package keychain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"math/big"

	"github.com/btcsuite/btcutil/base58"
	"github.com/xuperchain/xuperchain/core/crypto/account"
	"github.com/xuperchain/xuperchain/core/crypto/config"
	"github.com/xuperchain/xuperchain/core/crypto/hash"
	"github.com/xuperchain/xuperchain/core/crypto/utils"
)

// Hierarchical Deterministic Wallets - child drivation function 相关常量定义
// Reference - https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki
const (
	// HardenedKeyStart is the index at which a hardended key starts.
	// Each extended key has 2^31 normal child keys, and 2^31 hardned child keys.
	// Each of these child keys has an index.
	// The range for normal child keys is [0, 2^31 - 1]
	// The range for hardened child keys is [2^31, 2^32 - 1].
	HardenedKeyStart = 0x80000000 // 2^31 -- 16进制

	// maxUint8 is the max positive integer which can be serialized in a uint8
	MaxUint8 = 1<<8 - 1

	pubkeyUncompressed byte = 0x4 // x coord + y coord

	PubKeyBytesLenUncompressed = 65
)

var (
	// BIP32 hierarchical deterministic extended key magics
	// (i.e. for BTC, mainnet: 0x0488B21E public, 0x0488ADE4 private; testnet: 0x043587CF public, 0x04358394 private)
	HDPrivateKeyID = []byte{0x04, 0x20, 0xb9, 0x00} // starts with sprv
	HDPublicKeyID  = []byte{0x04, 0x20, 0xbd, 0x3a} // starts with spub

	// Prefix used when derive the hardened child private key.
	// I = HMAC-SHA512(Key = cpar, Data = 0x00 || ser256(kpar) || ser32(i)).
	// Note: The 0x00 pads the private key to make it 33 bytes long.
	PrefixDeriveHardenedPrivateKey = []byte{0x00}

	// masterKey is the master key used along with a random seed used to
	// generate the master node in the hierarchical tree.
	MasterKey = []byte("XuperCrypto seed")
)

// 错误相关定义
var (
	// The caller should simply ignore the invalid index for generateing a child extended key at this time.
	// Just move to the next index. (increase the old index by 1)
	ErrInvalidIndexForGenerateChildKey = errors.New("this index is invalid for generating a child key.")

	// The error happens when the provided seed is not usable due to
	// the derived key falling outside of the valid range for the private keys.
	// This error indicates the caller must choose another seed.
	ErrUnusableSeed = errors.New("unusable seed")

	// Seed length must between 128 and 512 bits. Which means between 16 and 64 bytes.
	ErrInvalidSeedLength = errors.New("invalid seed length, must between 128 and 512 bits")

	// 密码学算法暂未被支持
	// Cryptography required for generating Mnemonic has not been supported yet.
	ErrCryptographyNotSupported = fmt.Errorf("this cryptography has not been supported yet.")

	// A hardened child extended key cannot be derived from a public extended key.
	ErrDeriveHardenKeyFromPublicKey = errors.New("cannot derive a hardened child extended key from a public extended key")

	// Cannot derive a key with more than 255 depth from a root key.
	ErrDeriveBeyondMaxDepth = errors.New("cannot derive a key with more than 255 depth in its path from a root key")

	// Cannot create a private ecdsa key from a public extended key.
	ErrNotPrivExtKey = errors.New("cannot to create a private ecdsa key from a public extended key")

	// The checksum contained in the 58-encoded string(a serialized extended key) is not right.
	ErrWrongChecksum = errors.New("wrong extended key checksum")

	// The pubKeyStr contained in the extended key is not right.
	ErrWrongPubKeyStr = errors.New("wrong PubKeyStr")

	// Wrong param for deriving corresponding private child key from child public key.
	ErrWrongParamForDerivingPrivateChildKey = errors.New("wrong param when trying to derive child private key from child public key.")
)

// ExtendedKey has all the information needed to support a hierarchical deterministic extended key.
// <key format>
// Extended public and private keys have follow contents:
// 4 byte: version bytes (i.e. for BTC, mainnet: 0x0488B21E public, 0x0488ADE4 private; testnet: 0x043587CF public, 0x04358394 private)
// 1 byte: depth: 0x00 for master nodes, 0x01 for level-1 derived keys, ....
// 4 bytes: the fingerprint of the parent's key (0x00000000 if master key)
// 4 bytes: child number. This is ser32(i) for i in xi = xpar/i, with xi the key being serialized. (0x00000000 if master key)
// 32 bytes: the chain code
// 33 bytes: the public key or private key data (serP(K) for public keys, 0x00 || ser256(k) for private keys)
type ExtendedKey struct {
	Version   []byte // version bytes (i.e. for BTC, mainnet: 0x0488B21E public, 0x0488ADE4 private; testnet: 0x043587CF public, 0x04358394 private)
	Depth     uint8  // depth: 0x00 for master nodes, 0x01 for level-1 derived keys, ....
	ParentFP  []byte // the fingerprint of the parent's key (0x00000000 if master key)
	ChildNum  uint32 // child number. This is ser32(i) for i in xi = xpar/i, with xi the key being serialized. (0x00000000 if master key)
	ChainCode []byte // the chain code
	Key       []byte // the public key or private key data (serP(K) for public keys, 0x00 || ser256(k) for private keys)
	PubKey    []byte // this will only be set for extended priv keys

	AccountNum map[uint8]uint32 // shows the bloodline of this key

	Cryptography uint8 // the cryptography type: 1 for Nist P-256, 2 for Gm SM-2, ...

	IsPrivate bool // Whether an extended key is a private or public extended key can be determined with the isPrivate
}

// NewMaster creates a new master node. The seed must be between 128 and 512 bits and
// should be generated by a cryptographically secure random generation source.
//
// NOTE: Here rand.GenerateSeedWithStrengthAndKeyLen(strength int, keyLength int) ([]byte, error)
// should be used to generate the seed.
//
// NOTE: There is an extremely small chance (< 1 in 2^127) the provided seed
// will derive to an unusable secret key.  The ErrUnusableSeed error will be
// returned if this happens. The caller must check the error and generate a
// new seed if necessary.
//
// BIP32: Master key generation:
// 1. Generate a seed byte sequence S of a chosen length (between 128 and 512 bits; 256 bits is advised) from a (P)RNG.
// 2. Calculate I = HMAC-SHA512(Key = "Bitcoin seed", Data = S)
// 3. Split I into two 32-byte sequences, IL and IR.
// 4. Use parse256(IL) as master secret key, and IR as master chain code.
func NewMaster(seed []byte, cryptography uint8) (*ExtendedKey, error) {
	// 1. Generate a seed byte sequence S of a chosen length (between 128 and 512 bits; 256 bits is advised)
	if len(seed) < 16 || len(seed) > 64 {
		return nil, ErrInvalidSeedLength
	}

	curve := elliptic.P256()
	switch cryptography {
	case config.Nist: // NIST
	case config.Gm: // 国密
		//		curve = sm2.P256Sm2()
	default: // 不支持的密码学类型
		return nil, ErrCryptographyNotSupported
	}

	// 2. Calculate I = HMAC-SHA512(Key = "Bitcoin seed", Data = S)
	// i is a 64-byte sequence
	I := hash.HashUsingHmac512(seed, MasterKey)

	// Split "I" into two 32-byte sequences
	// master secret key, a 32-byte sequence
	secretKey := I[:len(I)/2]
	// master chain code, a 32-byte sequence
	chainCode := I[len(I)/2:]

	// interprets a 32-byte sequence as a 256-bit number.
	secretKeyNum := new(big.Int).SetBytes(secretKey)
	// In case secretKey ≥ n or secretKey = 0, the master key is invalid,
	// and the caller should proceed with the next value for i.
	// Note: this has probability lower than 1 in 2^127.
	if secretKeyNum.Cmp(curve.Params().N) >= 0 || secretKeyNum.Sign() == 0 {
		return nil, ErrUnusableSeed
	}

	// parentFP: the fingerprint of the parent's key (0x00000000 if master key)
	parentFP := []byte{0x00, 0x00, 0x00, 0x00}

	// depth: 0x00 for master nodes, 0x01 for level-1 derived keys, ....
	depth := uint8(0)

	// childNum: 0x00000000 if master key
	childNum := uint32(0)

	// isPrivate: whether it's a private key
	isPrivate := true

	var AccountNum map[uint8]uint32 = map[uint8]uint32{}
	AccountNum[depth] = childNum

	extendedKey := &ExtendedKey{
		Key:          secretKey,
		ChainCode:    chainCode,
		Depth:        depth,
		ParentFP:     parentFP,
		ChildNum:     childNum,
		Version:      HDPrivateKeyID,
		AccountNum:   AccountNum,
		Cryptography: cryptography,
		IsPrivate:    isPrivate,
	}

	return extendedKey, nil
}

// Derive corresponding(the same depth and childNum) private child key from child public key
func (k *ExtendedKey) CorrespondingPrivateChild(childPublicKey *ExtendedKey) (*ExtendedKey, error) {
	if childPublicKey.Depth == 0 {
		return nil, ErrWrongParamForDerivingPrivateChildKey
	}

	if childPublicKey.Depth <= k.Depth {
		return nil, ErrWrongParamForDerivingPrivateChildKey
	}

	//	childPrivateKey, err := k.Child(childPublicKey.AccountNum[1])
	childPrivateKey, err := k.Child(childPublicKey.AccountNum[k.Depth+1])
	if err != nil {
		return nil, err
	}

	//	for i := uint8(2); i <= childPublicKey.Depth; i++ {
	for i := uint8(k.Depth + 2); i <= childPublicKey.Depth; i++ {
		childPrivateKey, err = childPrivateKey.Child(childPublicKey.AccountNum[i])
		if err != nil {
			return nil, err
		}
	}

	return childPrivateKey, nil
}

// Child key derivation (CKD) functions
// This function returns a derived child extended key at the given index.
// A new private extended key will be derived from a private extended key.
// And a new public extended key will be derived from a public extended key.
//
// When the index is between [0, 2^31 - 1], normal child extended keys will be derived.
// When the index is between [2^31, 2^32 - 1], hardended child extended keys will be derived.
//
// A hardended extended key can only be derived from a private extended key.
//
// NOTE: There is an extremely small chance (< 1 in 2^127) the specific child
// index faild to derive a valid child.  The ErrInvalidChild error will be
// returned if this should occur, and the caller is expected to ignore the
// invalid child and simply increment to the next index.
//
// Given a parent extended key and an index i, it is possible to compute the corresponding child extended key.
// The algorithm to do so depends on whether the child is a hardened key or not (or, equivalently, whether i ≥ 2^31),
// and whether we're talking about private or public keys.
//
// 1. Private parent key → private child key.
// 2. Public parent key → public child key. Note: It is only defined for non-hardened child keys. i.e. i < HardenedKeyStart.
// 3. Private parent key → public child key. Use Neuter func instead. No need to consider in this func.
// 4. Public parent key → private child key. Note: This is impossible! No need to consider in this func.
func (k *ExtendedKey) Child(i uint32) (*ExtendedKey, error) {
	// Prevent derivation of children beyond the max allowed depth.
	if k.Depth == MaxUint8 {
		return nil, ErrDeriveBeyondMaxDepth
	}

	curve := elliptic.P256()
	switch k.Cryptography {
	case config.Nist: // NIST
	case config.Gm: // 国密 GMSM
		//		curve = sm2.P256Sm2()
	default: // 不支持的密码学类型 Unsupported cryptography type
		return nil, ErrCryptographyNotSupported
	}

	// Note:
	// 1. serP(P): serializes the coordinate pair P = (x,y) as a byte sequence
	// 	using SEC1's compressed form: (0x02 or 0x03) || ser256(x),
	// 	where the header byte depends on the parity of the omitted y coordinate.
	// 2. ser256(p): serializes the integer p as a 32-byte(256-bit) sequence
	// 3. point(p): returns the coordinate pair resulting from EC point multiplication.
	// 	i.e. repeated application of the EC group operation of the curve base point with the integer p.

	var isPrivate bool
	var childKey []byte
	var isChildHardened bool
	var childChainCode []byte
	// 1. If private parent key → private child key
	if k.IsPrivate {
		// Using a 4-bytes slice to represent ser32(i)
		var indexBytes [4]byte
		binary.BigEndian.PutUint32(indexBytes[:], i)

		var data = []byte{}

		isChildHardened = i >= HardenedKeyStart
		if isChildHardened {
			// If hardened child: let I = HMAC-SHA512(Key = cpar, Data = 0x00 || ser256(kpar) || ser32(i)).
			// Note: The 0x00 pads the private key to make it 33 bytes long.
			// So let us compute Data = 0x00 || ser256(kpar) || ser32(i)
			data = utils.BytesCombine(PrefixDeriveHardenedPrivateKey, k.Key, indexBytes[:])
		} else {
			// If normal child: let I = HMAC-SHA512(Key = cpar, Data = serP(point(kpar)) || ser32(i)).
			// So let us compute Data = serP(point(kpar)) || ser32(i)
			//point(kpar) = kpar * G
			pubkeyBytes, err := k.pubKeyBytes()
			if err != nil {
				return nil, ErrCryptographyNotSupported
			}
			data = utils.BytesCombine(pubkeyBytes, indexBytes[:])
		}

		// Set Hash func = sha512
		// Set key = cpar(chaincode of parent), i.e. key = k.chainCode
		// Now compute I = HMAC-SHA512(Data = data, Key = cpar)
		I := hash.HashUsingHmac512(data, k.ChainCode)

		// Split I into two 32-byte sequences, IL and IR.
		// IR is the returned child chain code.
		childChainCode = I[len(I)/2:]

		// Now we get IL = intermediate key used to derive the child.
		// The returned child key ki is parse256(IL) + kpar (mod n).
		IL := I[:len(I)/2]

		// There is a small chance il ≥ n or il = 0,, and in that case,
		// a child extended key can't be created for this index
		// and the caller should proceed with the next value for i.
		// Note: this has probability lower than 1 in 2^127.
		intIL := new(big.Int).SetBytes(IL)
		if intIL.Cmp(curve.Params().N) >= 0 || intIL.Sign() == 0 {
			return nil, ErrInvalidIndexForGenerateChildKey
		}

		// The returned child key ki is parse256(IL) + kpar (mod n).
		intKey := new(big.Int).SetBytes(k.Key)
		intIL.Add(intIL, intKey)
		intIL.Mod(intIL, curve.Params().N)
		childKey = intIL.Bytes()
		isPrivate = true
	} else {
		// 2. If Public parent key → public child key
		// It is only defined for non-hardened child keys.
		// If so hardened child: return failure(ErrDeriveHardenKeyFromPublicKey)
		if isChildHardened {
			return nil, ErrDeriveHardenKeyFromPublicKey
		}

		// If normal child: let I = HMAC-SHA512(Key = cpar, Data = serP(Kpar) || ser32(i)).
		// Note:
		// 1. serP(P): serializes the coordinate pair P = (x,y) as a byte sequence
		// 	using SEC1's compressed form: (0x02 or 0x03) || ser256(x),

		// Using a 4-bytes slice to represent ser32(i)
		var indexBytes [4]byte
		binary.BigEndian.PutUint32(indexBytes[:], i)

		var data = []byte{}
		pubkeyBytes, err := k.pubKeyBytes()
		if err != nil {
			return nil, ErrCryptographyNotSupported
		}
		data = utils.BytesCombine(pubkeyBytes, indexBytes[:])

		// Set Hash func = sha512
		// Set key = cpar(chaincode of parent), i.e. key = k.chainCode
		// Now compute I = HMAC-SHA512(Data = data, Key = cpar)
		I := hash.HashUsingHmac512(data, k.ChainCode)

		// Split I into two 32-byte sequences, IL and IR.
		// IR is the returned child chain code.
		childChainCode = I[len(I)/2:]

		// Now we get IL = intermediate key used to derive the child.
		// The returned child key Ki is point(parse256(IL)) + Kpar.
		IL := I[:len(I)/2]

		ILx, ILy := curve.ScalarBaseMult(IL)
		if ILx.Sign() == 0 || ILy.Sign() == 0 {
			return nil, ErrInvalidIndexForGenerateChildKey
		}

		// Convert the serialized compressed parent public key into X
		// and Y coordinates so it can be added to the intermediate public key.
		pubKey, err := parsePubKey(k.Key)
		if err != nil {
			return nil, err
		}

		// The returned child key Ki is point(parse256(IL)) + Kpar.
		//
		// Note:
		// point(p): returns the coordinate pair resulting from EC point multiplication.
		// 	i.e. repeated application of the EC group operation of the curve base point with the integer p.
		childX, childY := curve.Add(ILx, ILy, pubKey.X, pubKey.Y)
		pk := ecdsa.PublicKey{Curve: curve, X: childX, Y: childY}
		childKey, err = serializeUncompressed(&pk)
		if err != nil {
			return nil, err
		}
	}

	// The fingerprint of the parent for the derived child is the first 4
	// bytes of the RIPEMD160(SHA256(parentPubKey)).
	pubKeyBytes, err := k.pubKeyBytes()
	if err != nil {
		return nil, err
	}
	parentFP := hash.UsingRipemd160(pubKeyBytes)[:4]

	var accountNum map[uint8]uint32 = map[uint8]uint32{}
	if len(k.AccountNum) != 0 {
		accountNum = k.AccountNum
	}
	accountNum[k.Depth+1] = i

	extendedKey := &ExtendedKey{
		Key:       childKey,
		ChainCode: childChainCode,
		Depth:     k.Depth + 1,
		ParentFP:  parentFP,
		// TODO: 也许这个地方可以升级，改成是父childNum和子childNum组合的序列化
		// TODO: This is ser32(i) for i in xi = xpar/i, with xi the key being serialized. (0x00000000 if master key)
		ChildNum:     i,
		Version:      k.Version,
		AccountNum:   accountNum,
		IsPrivate:    isPrivate,
		Cryptography: k.Cryptography, // Done: 标记使用的曲线及相关的密码学算法
	}

	return extendedKey, nil
}

// Neuter returns a new extended public key from this extended private key.  The
// same extended key will be returned unaltered if it is already an extended
// public key.
//
// As the name implies, an extended public key does not have access to the
// private key, so it is not capable of signing transactions or deriving
// child extended private keys.  However, it is capable of deriving further
// child extended public keys.
func (k *ExtendedKey) Neuter() (*ExtendedKey, error) {
	// Already an extended public key.
	if !k.IsPrivate {
		return k, nil
	}

	// Get the associated public extended key version bytes.
	version := HDPublicKeyID

	// Convert it to an extended public key.  The key for the new extended
	// key will simply be the pubkey of the current extended private key.
	//
	// This is the function N((k,c)) -> (K, c) from [BIP32].
	pubKeyBytes, err := k.pubKeyBytes()
	if err != nil {
		return nil, err
	}
	extendedKey := &ExtendedKey{
		Key:          pubKeyBytes,
		ChainCode:    k.ChainCode,
		Depth:        k.Depth,
		ParentFP:     k.ParentFP,
		ChildNum:     k.ChildNum,
		Version:      version,
		AccountNum:   k.AccountNum,
		IsPrivate:    false,
		Cryptography: k.Cryptography,
	}

	return extendedKey, nil
}

// ECPublicKey converts the extended key to a ecdsa public key and returns it.
func (k *ExtendedKey) ECPublicKey() (*ecdsa.PublicKey, error) {
	pubKeyBytes, err := k.pubKeyBytes()
	if err != nil {
		return nil, err
	}
	pubKey, err := parsePubKey(pubKeyBytes)
	if err != nil {
		return nil, err
	}

	return pubKey, nil
}

// ECPrivateKey converts the extended key to a ecdsa private key and returns it.
// This is only possible if the extended key is a private extended key.
func (k *ExtendedKey) ECPrivateKey() (*ecdsa.PrivateKey, error) {
	if !k.IsPrivate {
		return nil, ErrNotPrivExtKey
	}

	curve := elliptic.P256()
	switch k.Cryptography {
	case config.Nist: // NIST
	case config.Gm: // 国密
		//		curve = sm2.P256Sm2()
	default: // 不支持的密码学类型
		return nil, ErrCryptographyNotSupported
	}

	// 通过D计算x和y
	x, y := curve.ScalarBaseMult(k.Key)

	privKey := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: curve,
			X:     x,
			Y:     y,
		},
		D: new(big.Int).SetBytes(k.Key),
	}

	return privKey, nil
}

// Return the address of an extended Key.
func (k *ExtendedKey) Address() (string, error) {
	pubKey, err := k.ECPublicKey()
	if err != nil {
		return account.GetAddressFromPublicKey(pubKey)
	}

	return "", err
}

// Convert an extended key to a base58-encoded String.
func (k *ExtendedKey) ToString() string {
	var childNumBytes [4]byte
	binary.BigEndian.PutUint32(childNumBytes[:], k.ChildNum)

	depthBytes := []byte{byte(k.Depth)}

	cryptographyBytes := []byte{byte(k.Cryptography)}

	serializedBytes := utils.BytesCombine(k.Version, depthBytes, k.ParentFP, childNumBytes[:], k.ChainCode, cryptographyBytes)

	if k.IsPrivate {
		serializedBytes = append(serializedBytes, 0x00)
		serializedBytes = paddedAppend(32, serializedBytes, k.Key)
	} else {
		pubkeyBytes, err := k.pubKeyBytes()
		if err != nil {
			return ""
		}
		serializedBytes = utils.BytesCombine(serializedBytes, pubkeyBytes)
	}

	checkCode := hash.DoubleSha256(serializedBytes)
	checkSum := checkCode[:4]

	//	serializedBytes = append(serializedBytes, checkSum...)
	serializedBytes = utils.BytesCombine(serializedBytes, checkSum)

	return base58.Encode(serializedBytes)
}

// Retrieve an extended key from a base58-encoded String.
func NewKeyFromString(key string) (*ExtendedKey, error) {
	// The base58-decoded extended key consists of a serialized payload plus an additional 4 bytes checksum.
	decoded := base58.Decode(key)

	// The serialized format is:
	//   version (4) || depth (1) || parent fingerprint (4)) ||
	//   child num (4) || chain code (32) || key data (33) || checksum (4)

	// Split the payload and checksum up and ensure the checksum matches.
	payload := decoded[:len(decoded)-4]
	checkSum := decoded[len(decoded)-4:]
	checkCode := hash.DoubleSha256(payload)
	expectedCheckSum := checkCode[:4]

	if !bytes.Equal(checkSum, expectedCheckSum) {
		return nil, ErrWrongChecksum
	}

	// Deserialize each of the payload fields.
	version := payload[:4]
	depth := payload[4:5][0]
	parentFP := payload[5:9]
	childNum := binary.BigEndian.Uint32(payload[9:13])
	chainCode := payload[13:45]

	cryptography := payload[45:46][0]
	//	keyData := payload[45:78]
	keyData := payload[46:]

	curve := elliptic.P256()
	switch cryptography {
	case config.Nist: // NIST
	case config.Gm: // 国密
		//		curve = sm2.P256Sm2()
	default: // 不支持的密码学类型
		return nil, ErrCryptographyNotSupported
	}

	// The key data is a private key if it starts with 0x00.  Serialized
	// compressed pubkeys either start with 0x02 or 0x03.
	isPrivate := keyData[0] == 0x00

	if isPrivate {
		// Ensure the private key is valid.  It must be within the range
		// of the order of the secp256k1 curve and not be 0.
		keyData = keyData[1:]
		keyNum := new(big.Int).SetBytes(keyData)
		if keyNum.Cmp(curve.Params().N) >= 0 || keyNum.Sign() == 0 {
			return nil, ErrUnusableSeed
		}
	} else {
		// Ensure the public key parses correctly and is actually on the
		// secp256k1 curve.
		_, err := parsePubKey(keyData)
		if err != nil {
			log.Printf("parsePubKey faild and keyData is: %v", keyData)
			return nil, err
		}
	}

	extendedKey := &ExtendedKey{
		Key:          keyData,
		ChainCode:    chainCode,
		Depth:        depth,
		ParentFP:     parentFP,
		ChildNum:     childNum,
		Version:      version,
		IsPrivate:    isPrivate,
		Cryptography: cryptography,
	}

	return extendedKey, nil
}

func parsePubKey(pubKeyStr []byte) (key *ecdsa.PublicKey, err error) {
	format := pubKeyStr[0]
	if format == pubkeyUncompressed && len(pubKeyStr) == PubKeyBytesLenUncompressed+1 {
		pubkey := ecdsa.PublicKey{}

		cryptography := pubKeyStr[1]
		curve := elliptic.P256()
		switch cryptography {
		case config.Nist: // NIST
		case config.Gm: // 国密
			//			curve = sm2.P256Sm2()
		default: // 不支持的密码学类型
			return nil, ErrCryptographyNotSupported
		}

		pubkey.Curve = curve
		pubkey.X = new(big.Int).SetBytes(pubKeyStr[2:34])
		pubkey.Y = new(big.Int).SetBytes(pubKeyStr[34:])

		return &pubkey, nil
	}

	return nil, ErrWrongPubKeyStr
}

// pubKeyBytes returns bytes for the serialized compressed public key associated
// with this extended key in an efficient manner including memoization as necessary.
//
// When the extended key is already a public key, the key is simply returned as
// is since it's already in the correct form.
//
// However, when the extended key is a private key, the public key will be calculated
// and memorized so future accesses can simply return the cached result.
func (k *ExtendedKey) pubKeyBytes() ([]byte, error) {
	// Just return the key if it's already an extended public key.
	if !k.IsPrivate {
		return k.Key, nil
	}

	// This is a private extended key, so calculate and memorize the public
	// key if needed.
	if len(k.PubKey) == 0 {
		curve := elliptic.P256()
		switch k.Cryptography {
		case config.Nist: // NIST
		case config.Gm: // 国密
			//			curve = sm2.P256Sm2()
		default: // 不支持的密码学类型
			return nil, ErrCryptographyNotSupported
		}

		pkx, pky := curve.ScalarBaseMult(k.Key)
		pubKey := ecdsa.PublicKey{Curve: curve, X: pkx, Y: pky}
		strPubKey, err := serializeUncompressed(&pubKey)
		if err != nil {
			return nil, err
		}

		k.PubKey = []byte(strPubKey)
	}

	return k.PubKey, nil
}

// SerializeUncompressed serializes a public key in a 65-byte uncompressed
// format.
func serializeUncompressed(pubKey *ecdsa.PublicKey) ([]byte, error) {
	cryptography := config.Nist
	switch pubKey.Params().Name {
	case config.CurveNist: // NIST
	case config.CurveGm: // 国密
		cryptography = config.Gm
	default: // 不支持的密码学类型
		return nil, fmt.Errorf("This cryptography[%v] has not been supported yet.", pubKey.Params().Name)
	}

	b := make([]byte, 0, PubKeyBytesLenUncompressed+1)
	b = append(b, pubkeyUncompressed)
	b = append(b, byte(cryptography))
	b = paddedAppend(32, b, pubKey.X.Bytes())
	return paddedAppend(32, b, pubKey.Y.Bytes()), nil
}

// paddedAppend appends the src byte slice to dst, returning the new slice.
// If the length of the source is smaller than the passed size, leading zero
// bytes are appended to the dst slice before appending src.
func paddedAppend(size uint, dst, src []byte) []byte {
	for i := 0; i < int(size)-len(src); i++ {
		dst = append(dst, 0)
	}
	return append(dst, src...)
}

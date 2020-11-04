package keys

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"hash"
	"strings"

	"github.com/hyperledger/burrow/crypto"
	hex "github.com/tmthrgd/go-hex"
	"golang.org/x/crypto/ripemd160"
	"google.golang.org/grpc"
)

//------------------------------------------------------------------------
// all cli commands pass through the http KeyStore
// the KeyStore process also maintains the unlocked accounts

func StandAloneServer(keysDir string, AllowBadFilePermissions bool) *grpc.Server {
	grpcServer := grpc.NewServer()
	RegisterKeysServer(grpcServer, NewFilesystemKeyStore(keysDir, AllowBadFilePermissions))
	return grpcServer
}

//------------------------------------------------------------------------
// handlers

func (k *FilesystemKeyStore) GenerateKey(ctx context.Context, in *GenRequest) (*GenResponse, error) {
	curveT, err := crypto.CurveTypeFromString(in.CurveType)
	if err != nil {
		return nil, err
	}

	key, err := k.Gen(in.Passphrase, curveT)
	if err != nil {
		return nil, fmt.Errorf("error generating key %s %s", curveT, err)
	}

	addrH := key.Address.String()
	if in.KeyName != "" {
		err = coreNameAdd(k.keysDirPath, in.KeyName, addrH)
		if err != nil {
			return nil, err
		}
	}

	return &GenResponse{Address: addrH}, nil
}

func (k *FilesystemKeyStore) Export(ctx context.Context, in *ExportRequest) (*ExportResponse, error) {
	addr, err := getNameAddr(k.keysDirPath, in.GetName(), in.GetAddress())

	if err != nil {
		return nil, err
	}

	addrB, err := crypto.AddressFromHexString(addr)
	if err != nil {
		return nil, err
	}

	// No phrase needed for public key. I hope.
	key, err := k.GetKey(in.GetPassphrase(), addrB.Bytes())
	if err != nil {
		return nil, err
	}

	return &ExportResponse{
		Address:    addrB[:],
		CurveType:  key.CurveType.String(),
		Publickey:  key.PublicKey.PublicKey[:],
		Privatekey: key.PrivateKey.PrivateKey[:],
	}, nil
}

func (k *FilesystemKeyStore) PublicKey(ctx context.Context, in *PubRequest) (*PubResponse, error) {
	addr, err := getNameAddr(k.keysDirPath, in.GetName(), in.GetAddress())
	if err != nil {
		return nil, err
	}

	addrB, err := crypto.AddressFromHexString(addr)
	if err != nil {
		return nil, err
	}

	// No phrase needed for public key. I hope.
	key, err := k.GetKey("", addrB.Bytes())
	if key == nil {
		return nil, err
	}

	return &PubResponse{CurveType: key.CurveType.String(), PublicKey: key.Pubkey()}, nil
}

func (k *FilesystemKeyStore) Sign(ctx context.Context, in *SignRequest) (*SignResponse, error) {
	addr, err := getNameAddr(k.keysDirPath, in.GetName(), in.GetAddress())
	if err != nil {
		return nil, err
	}

	addrB, err := crypto.AddressFromHexString(addr)
	if err != nil {
		return nil, err
	}

	key, err := k.GetKey(in.GetPassphrase(), addrB[:])
	if err != nil {
		return nil, err
	}

	sig, err := key.PrivateKey.Sign(in.GetMessage())
	if err != nil {
		return nil, err
	}
	return &SignResponse{Signature: sig}, err
}

func (k *FilesystemKeyStore) Verify(ctx context.Context, in *VerifyRequest) (*VerifyResponse, error) {
	if in.GetPublicKey() == nil {
		return nil, fmt.Errorf("must provide a pubkey")
	}
	if in.GetMessage() == nil {
		return nil, fmt.Errorf("must provide a message")
	}
	if in.GetSignature() == nil {
		return nil, fmt.Errorf("must provide a signature")
	}

	sig := in.GetSignature()
	pubkey, err := crypto.PublicKeyFromBytes(in.GetPublicKey(), sig.GetCurveType())
	if err != nil {
		return nil, err
	}
	err = pubkey.Verify(in.GetMessage(), sig)
	if err != nil {
		return nil, err
	}

	return &VerifyResponse{}, nil
}

func (k *FilesystemKeyStore) Hash(ctx context.Context, in *HashRequest) (*HashResponse, error) {
	var hasher hash.Hash
	switch in.GetHashtype() {
	case "ripemd160":
		hasher = ripemd160.New()
	case "sha256":
		hasher = sha256.New()
	// case "sha3":
	default:
		return nil, fmt.Errorf("unknown hash type %v", in.GetHashtype())
	}

	hasher.Write(in.GetMessage())

	return &HashResponse{Hash: hex.EncodeUpperToString(hasher.Sum(nil))}, nil
}

func (k *FilesystemKeyStore) ImportJSON(ctx context.Context, in *ImportJSONRequest) (*ImportResponse, error) {
	keyJSON := []byte(in.GetJSON())
	addr := isValidKeyJson(keyJSON)
	if addr != nil {
		_, err := writeKey(k.keysDirPath, addr, keyJSON)
		if err != nil {
			return nil, err
		}
	} else {
		j1 := new(struct {
			CurveType   string
			Address     string
			PublicKey   string
			AddressHash string
			PrivateKey  string
		})

		err := json.Unmarshal([]byte(in.GetJSON()), &j1)
		if err != nil {
			return nil, err
		}

		addr, err = hex.DecodeString(j1.Address)
		if err != nil {
			return nil, err
		}

		curveT, err := crypto.CurveTypeFromString(j1.CurveType)
		if err != nil {
			return nil, err
		}

		privKey, err := hex.DecodeString(j1.PrivateKey)
		if err != nil {
			return nil, err
		}

		key, err := NewKeyFromPriv(curveT, privKey)
		if err != nil {
			return nil, err
		}

		// store the new key
		if err = k.StoreKey(in.GetPassphrase(), key); err != nil {
			return nil, err
		}
	}
	return &ImportResponse{Address: hex.EncodeUpperToString(addr)}, nil
}

func (k *FilesystemKeyStore) Import(ctx context.Context, in *ImportRequest) (*ImportResponse, error) {
	curveT, err := crypto.CurveTypeFromString(in.GetCurveType())
	if err != nil {
		return nil, err
	}
	key, err := NewKeyFromPriv(curveT, in.GetKeyBytes())
	if err != nil {
		return nil, err
	}

	// store the new key
	if err = k.StoreKey(in.GetPassphrase(), key); err != nil {
		return nil, err
	}

	if in.GetName() != "" {
		if err := coreNameAdd(k.keysDirPath, in.GetName(), key.Address.String()); err != nil {
			return nil, err
		}
	}
	return &ImportResponse{Address: hex.EncodeUpperToString(key.Address[:])}, nil
}

func (k *FilesystemKeyStore) List(ctx context.Context, in *ListRequest) (*ListResponse, error) {
	byname, err := coreNameList(k.keysDirPath)
	if err != nil {
		return nil, err
	}

	var list []*KeyID

	if in.KeyName != "" {
		if addr, ok := byname[in.KeyName]; ok {
			list = append(list, &KeyID{KeyName: getAddressNames(addr, byname), Address: addr})
		} else {
			if addr, err := crypto.AddressFromHexString(in.KeyName); err == nil {
				_, err := k.GetKey("", addr[:])
				if err == nil {
					address := addr.String()
					list = append(list, &KeyID{Address: address, KeyName: getAddressNames(address, byname)})
				}
			}
		}
	} else {
		// list all address

		datadir, err := returnDataDir(k.keysDirPath)
		if err != nil {
			return nil, err
		}
		addrs, err := getAllAddresses(datadir)
		if err != nil {
			return nil, err
		}

		for _, addr := range addrs {
			list = append(list, &KeyID{KeyName: getAddressNames(addr, byname), Address: addr})
		}
	}

	return &ListResponse{Key: list}, nil
}

func getAddressNames(address string, byname map[string]string) []string {
	names := make([]string, 0)

	for name, addr := range byname {
		if address == addr {
			names = append(names, name)
		}
	}

	return names
}

func (k *FilesystemKeyStore) RemoveName(ctx context.Context, in *RemoveNameRequest) (*RemoveNameResponse, error) {
	if in.GetKeyName() == "" {
		return nil, fmt.Errorf("please specify a name")
	}

	return &RemoveNameResponse{}, coreNameRm(k.keysDirPath, in.GetKeyName())
}

func (k *FilesystemKeyStore) AddName(ctx context.Context, in *AddNameRequest) (*AddNameResponse, error) {
	if in.GetKeyname() == "" {
		return nil, fmt.Errorf("please specify a name")
	}

	if in.GetAddress() == "" {
		return nil, fmt.Errorf("please specify an address")
	}

	return &AddNameResponse{}, coreNameAdd(k.keysDirPath, in.GetKeyname(), strings.ToUpper(in.GetAddress()))
}

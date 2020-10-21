package keys

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/hyperledger/burrow/crypto"
	"github.com/tmthrgd/go-hex"
	"golang.org/x/crypto/scrypt"
)

type FilesystemKeyStore struct {
	sync.Mutex
	AllowBadFilePermissions bool
	keysDirPath             string
}

var _ KeyStore = &FilesystemKeyStore{}
var _ KeysServer = &FilesystemKeyStore{}

func NewFilesystemKeyStore(dir string, AllowBadFilePermissions bool) *FilesystemKeyStore {
	return &FilesystemKeyStore{
		keysDirPath:             dir,
		AllowBadFilePermissions: AllowBadFilePermissions,
	}
}

func (ks *FilesystemKeyStore) Gen(passphrase string, curveType crypto.CurveType) (key *Key, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("GenerateNewKey error: %v", r)
		}
	}()
	key, err = NewKey(curveType)
	if err != nil {
		return nil, err
	}
	err = ks.StoreKey(passphrase, key)
	return key, err
}

func (ks *FilesystemKeyStore) GetKey(passphrase string, keyAddr []byte) (*Key, error) {
	ks.Lock()
	defer ks.Unlock()
	dataDirPath, err := returnDataDir(ks.keysDirPath)
	if err != nil {
		return nil, err
	}
	fileContent, err := ks.GetKeyFile(dataDirPath, keyAddr)
	if err != nil {
		return nil, err
	}
	key := new(keyJSON)
	if err = json.Unmarshal(fileContent, key); err != nil {
		return nil, err
	}

	if len(key.PrivateKey.CipherText) > 0 {
		return DecryptKey(passphrase, key)
	} else {
		key := new(Key)
		err = key.UnmarshalJSON(fileContent)
		return key, err
	}
}

func (ks *FilesystemKeyStore) AllKeys() ([]*Key, error) {
	dataDirPath, err := returnDataDir(ks.keysDirPath)
	if err != nil {
		return nil, err
	}
	addrs, err := getAllAddresses(dataDirPath)
	if err != nil {
		return nil, err
	}

	var list []*Key

	for _, addr := range addrs {
		addrB, err := crypto.AddressFromHexString(addr)
		if err != nil {
			return nil, err
		}
		k, err := ks.GetKey("", addrB[:])
		if err != nil {
			return nil, err
		}
		list = append(list, k)
	}

	return list, nil
}

func DecryptKey(passphrase string, keyProtected *keyJSON) (*Key, error) {
	salt := keyProtected.PrivateKey.Salt
	nonce := keyProtected.PrivateKey.Nonce
	cipherText := keyProtected.PrivateKey.CipherText

	curveType, err := crypto.CurveTypeFromString(keyProtected.CurveType)
	if err != nil {
		return nil, err
	}
	authArray := []byte(passphrase)
	derivedKey, err := scrypt.Key(authArray, salt, scryptN, scryptr, scryptp, scryptdkLen)
	if err != nil {
		return nil, err
	}
	aesBlock, err := aes.NewCipher(derivedKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return nil, err
	}
	pubKey, err := hex.DecodeString(keyProtected.PublicKey)
	if err != nil {
		return nil, err
	}
	plainText, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		pkey, _ := NewKeyFromPub(curveType, pubKey)
		return pkey, err
	}
	address, err := crypto.AddressFromHexString(keyProtected.Address)
	if err != nil {
		return nil, err
	}
	k, err := NewKeyFromPriv(curveType, plainText)
	if err != nil {
		return nil, err
	}
	if address != k.Address {
		return nil, fmt.Errorf("address does not match")
	}
	return k, nil
}

func (ks *FilesystemKeyStore) GetAllAddresses() (addresses []string, err error) {
	ks.Lock()
	defer ks.Unlock()

	dir, err := returnDataDir(ks.keysDirPath)
	if err != nil {
		return nil, err
	}
	return getAllAddresses(dir)
}

func (ks *FilesystemKeyStore) StoreKey(passphrase string, key *Key) error {
	ks.Lock()
	defer ks.Unlock()
	if passphrase != "" {
		return ks.StoreKeyEncrypted(passphrase, key)
	} else {
		return ks.StoreKeyPlain(key)
	}
}

func (ks *FilesystemKeyStore) StoreKeyPlain(key *Key) (err error) {
	keyJSON, err := json.Marshal(key)
	if err != nil {
		return err
	}
	dataDirPath, err := returnDataDir(ks.keysDirPath)
	if err != nil {
		return err
	}
	err = WriteKeyFile(key.Address[:], dataDirPath, keyJSON)
	return err
}

func (ks *FilesystemKeyStore) StoreKeyEncrypted(passphrase string, key *Key) error {
	authArray := []byte(passphrase)
	salt := make([]byte, 32)
	_, err := rand.Read(salt)
	if err != nil {
		return err
	}

	derivedKey, err := scrypt.Key(authArray, salt, scryptN, scryptr, scryptp, scryptdkLen)
	if err != nil {
		return err
	}

	toEncrypt := key.PrivateKey.RawBytes()

	AES256Block, err := aes.NewCipher(derivedKey)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(AES256Block)
	if err != nil {
		return err
	}

	// XXX: a GCM nonce may only be used once per key ever!
	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		return err
	}

	// (dst, nonce, plaintext, extradata)
	cipherText := gcm.Seal(nil, nonce, toEncrypt, nil)

	cipherStruct := privateKeyJSON{
		Crypto: CryptoAESGCM, Salt: salt, Nonce: nonce, CipherText: cipherText,
	}
	keyStruct := keyJSON{
		CurveType:   key.CurveType.String(),
		Address:     hex.EncodeUpperToString(key.Address[:]),
		PublicKey:   hex.EncodeUpperToString(key.Pubkey()),
		AddressHash: key.PublicKey.AddressHashType(),
		PrivateKey:  cipherStruct,
	}
	keyJSON, err := json.Marshal(keyStruct)
	if err != nil {
		return err
	}
	dataDirPath, err := returnDataDir(ks.keysDirPath)
	if err != nil {
		return err
	}

	return WriteKeyFile(key.Address[:], dataDirPath, keyJSON)
}

func (ks *FilesystemKeyStore) DeleteKey(passphrase string, keyAddr []byte) (err error) {
	dataDirPath, err := returnDataDir(ks.keysDirPath)
	if err != nil {
		return err
	}
	keyDirPath := path.Join(dataDirPath, strings.ToUpper(hex.EncodeToString(keyAddr))+".json")
	return os.Remove(keyDirPath)
}

func (ks *FilesystemKeyStore) GetKeyFile(dataDirPath string, keyAddr []byte) (fileContent []byte, err error) {
	filename := path.Join(dataDirPath, strings.ToUpper(hex.EncodeToString(keyAddr))+".json")
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}
	if (uint32(fileInfo.Mode()) & 0077) != 0 {
		if !ks.AllowBadFilePermissions {
			return nil, fmt.Errorf("file %s should be accessible by user only", filename)
		}
	}
	return ioutil.ReadFile(filename)
}

func WriteKeyFile(addr []byte, dataDirPath string, content []byte) (err error) {
	addrHex := strings.ToUpper(hex.EncodeToString(addr))
	keyFilePath := path.Join(dataDirPath, addrHex+".json")
	err = os.MkdirAll(dataDirPath, 0700) // read, write and dir search for user
	if err != nil {
		return err
	}
	return ioutil.WriteFile(keyFilePath, content, 0600) // read, write for user
}

func (ks *FilesystemKeyStore) GetAddressForKeyName(name string) (crypto.Address, error) {
	const errHeader = "GetAddressForKeyName"
	nameAddressLookup, err := coreNameList(ks.keysDirPath)
	if err != nil {
		return crypto.Address{}, fmt.Errorf("%s: could not get names list from filesysetm: %w",
			errHeader, err)
	}

	addressHex, ok := nameAddressLookup[name]
	if !ok {
		return crypto.Address{}, fmt.Errorf("%s: could not find key named '%s'", errHeader, name)
	}

	address, err := crypto.AddressFromHexString(addressHex)
	if err != nil {
		return crypto.Address{}, fmt.Errorf("%s: could not parse key address: %v", errHeader, err)
	}
	return address, nil
}

func (ks *FilesystemKeyStore) GetAllNames() (map[string]string, error) {
	return coreNameList(ks.keysDirPath)
}

func getAllAddresses(dataDirPath string) (addresses []string, err error) {
	fileInfos, err := ioutil.ReadDir(dataDirPath)
	if err != nil {
		return nil, err
	}
	addresses = make([]string, len(fileInfos))
	for i, fileInfo := range fileInfos {
		addr := strings.TrimSuffix(fileInfo.Name(), ".json")
		addresses[i] = addr
	}
	return addresses, err
}

package evm

import (
	"errors"
	"strings"

	"github.com/btcsuite/btcutil/base58"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/xuperchain/xuperchain/core/crypto/hash"
)

const (
	contractNamePrefix    = "\t"
	contractAccountPrefix = "\n"

	contractNamePrefixs    = "\t\t\t\t"
	contractAccountPrefixs = "\n\n\n\n"

	xchainAddrType      = "xchain"
	contractNameType    = "contract-name"
	contractAccountType = "contract-account"
)

// transfer xchain address to evm address
func XchainToEVMAddress(addr string) (crypto.Address, error) {
	rawAddr := base58.Decode(addr)
	if len(rawAddr) < 21 {
		return crypto.ZeroAddress, errors.New("bad address")
	}
	ripemd160Hash := rawAddr[1:21]
	return crypto.AddressFromBytes(ripemd160Hash)
}

// transfer evm address to xchain address
func EVMAddressToXchain(evmAddress crypto.Address) (string, error) {
	addrType := 1
	nVersion := uint8(addrType)
	bufVersion := []byte{byte(nVersion)}

	outputRipemd160 := evmAddress.Bytes()

	strSlice := make([]byte, len(bufVersion)+len(outputRipemd160))
	copy(strSlice, bufVersion)
	copy(strSlice[len(bufVersion):], outputRipemd160)

	checkCode := hash.DoubleSha256(strSlice)
	simpleCheckCode := checkCode[:4]
	slice := make([]byte, len(strSlice)+len(simpleCheckCode))
	copy(slice, strSlice)
	copy(slice[len(strSlice):], simpleCheckCode)

	return base58.Encode(slice), nil
}

func ContractAddress(name string) (crypto.Address, error) {
	rawAddr := hash.UsingRipemd160([]byte(name))
	return crypto.AddressFromBytes(rawAddr)
}

// transfer contract name to evm address
func ContractNameToEVMAddress(contractName string) (crypto.Address, error) {
	contractNameLength := len(contractName)
	var prefixStr string
	for i := 0; i < binary.Word160Length-contractNameLength; i++ {
		prefixStr += contractNamePrefix
	}
	contractName = prefixStr + contractName
	return crypto.AddressFromBytes([]byte(contractName))
}

// transfer evm address to contract name
func EVMAddressToContractName(evmAddr crypto.Address) (string, error) {
	contractNameWithPrefix := evmAddr.Bytes()
	contractNameStrWithPrefix := string(contractNameWithPrefix)
	prefixIndex := strings.LastIndex(contractNameStrWithPrefix, contractNamePrefix)
	return contractNameStrWithPrefix[prefixIndex+1:], nil
}

// transfer contract account to evm address
func ContractAccountToEVMAddress(contractAccount string) (crypto.Address, error) {
	contractAccountLength := 16
	contractAccountValid := contractAccount[2:18]
	var prefixStr string
	for i := 0; i < binary.Word160Length-contractAccountLength; i++ {
		prefixStr += contractAccountPrefix
	}
	contractAccountValid = prefixStr + contractAccountValid
	return crypto.AddressFromBytes([]byte(contractAccountValid))
}

// transfer evm address to contract account
func EVMAddressToContractAccount(evmAddr crypto.Address) (string, error) {
	contractNameWithPrefix := evmAddr.Bytes()
	contractNameStrWithPrefix := string(contractNameWithPrefix)
	prefixIndex := strings.LastIndex(contractNameStrWithPrefix, contractAccountPrefix)
	return "XC" + contractNameStrWithPrefix[prefixIndex+1:] + "@xuper", nil
}

// determine whether it is a contract account
func DetermineContractAccount(account string) bool {
	return strings.Index(account, "@xuper") != -1
}

// determine an EVM address
func DetermineEVMAddress(evmAddr crypto.Address) (string, string, error) {
	evmAddrWithPrefix := evmAddr.Bytes()
	evmAddrStrWithPrefix := string(evmAddrWithPrefix)

	var addr, addrType string
	var err error
	if strings.Index(evmAddrStrWithPrefix, contractAccountPrefixs) != -1 {
		addr, err = EVMAddressToContractAccount(evmAddr)
		addrType = contractAccountType
	} else if strings.Index(evmAddrStrWithPrefix, contractNamePrefixs) != -1 {
		addr, err = EVMAddressToContractName(evmAddr)
		addrType = contractNameType
	} else {
		addr, err = EVMAddressToXchain(evmAddr)
		addrType = xchainAddrType
	}
	if err != nil {
		return "", "", err
	}

	return addr, addrType, nil
}

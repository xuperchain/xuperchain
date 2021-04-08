package evm

import (
	"errors"
	"fmt"
	"strings"

	"github.com/btcsuite/btcutil/base58"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"

	"github.com/xuperchain/xuperchain/core/common"
	"github.com/xuperchain/xuperchain/core/crypto/hash"
	"github.com/xuperchain/xuperchain/core/permission/acl"
	"github.com/xuperchain/xuperchain/core/permission/acl/utils"
)

const (
	evmAddressFiller = "-"

	contractNamePrefixs    = "1111"
	contractAccountPrefixs = "1112"

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

// transfer contract name to evm address
func ContractNameToEVMAddress(contractName string) (crypto.Address, error) {
	contractNameLength := len(contractName)
	var prefixStr string
	for i := 0; i < binary.Word160Length-contractNameLength-utils.GetContractNameMinSize(); i++ {
		prefixStr += evmAddressFiller
	}
	contractName = prefixStr + contractName
	contractName = contractNamePrefixs + contractName
	return crypto.AddressFromBytes([]byte(contractName))
}

// transfer evm address to contract name
func EVMAddressToContractName(evmAddr crypto.Address) (string, error) {
	contractNameWithPrefix := evmAddr.Bytes()
	contractNameStrWithPrefix := string(contractNameWithPrefix)
	prefixIndex := strings.LastIndex(contractNameStrWithPrefix, evmAddressFiller)
	if prefixIndex == -1 {
		return contractNameStrWithPrefix[4:], nil
	}
	return contractNameStrWithPrefix[prefixIndex+1:], nil
}

// transfer contract account to evm address
func ContractAccountToEVMAddress(contractAccount string) (crypto.Address, error) {
	contractAccountValid := contractAccount[2:18]
	contractAccountValid = contractAccountPrefixs + contractAccountValid
	return crypto.AddressFromBytes([]byte(contractAccountValid))
}

// transfer evm address to contract account
func EVMAddressToContractAccount(evmAddr crypto.Address) (string, error) {
	contractNameWithPrefix := evmAddr.Bytes()
	contractNameStrWithPrefix := string(contractNameWithPrefix)
	return utils.GetAccountPrefix() + contractNameStrWithPrefix[4:] + "@xuper", nil
}

// determine whether it is a contract account
func DetermineContractAccount(account string) bool {
	if acl.IsAccount(account) != 1 {
		return false
	}
	return strings.Index(account, "@xuper") != -1
}

// determine whether it is a contract name
func DetermineContractName(contractName string) error {
	return common.ValidContractName(contractName)
}

// determine whether it is a contract name
func DetermineContractNameFromEVM(evmAddr crypto.Address) (string, error) {
	var addr string
	var err error

	evmAddrWithPrefix := evmAddr.Bytes()
	evmAddrStrWithPrefix := string(evmAddrWithPrefix)
	if evmAddrStrWithPrefix[0:4] != contractNamePrefixs {
		return "", fmt.Errorf("not a valid contract name from evm")
	} else {
		addr, err = EVMAddressToContractName(evmAddr)
	}

	if err != nil {
		return "", err
	}

	return addr, nil
}

// determine an EVM address
func DetermineEVMAddress(evmAddr crypto.Address) (string, string, error) {
	evmAddrWithPrefix := evmAddr.Bytes()
	evmAddrStrWithPrefix := string(evmAddrWithPrefix)

	var addr, addrType string
	var err error
	if evmAddrStrWithPrefix[0:4] == contractAccountPrefixs {
		addr, err = EVMAddressToContractAccount(evmAddr)
		addrType = contractAccountType
	} else if evmAddrStrWithPrefix[0:4] == contractNamePrefixs {
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

// determine an xchain address
func DetermineXchainAddress(xAddr string) (string, string, error) {
	var addr crypto.Address
	var addrType string
	var err error
	if DetermineContractAccount(xAddr) {
		addr, err = ContractAccountToEVMAddress(xAddr)
		addrType = contractAccountType
	} else if DetermineContractName(xAddr) == nil {
		addr, err = ContractNameToEVMAddress(xAddr)
		addrType = contractNameType
	} else {
		addr, err = XchainToEVMAddress(xAddr)
		addrType = xchainAddrType
	}
	if err != nil {
		return "", "", err
	}

	return addr.String(), addrType, nil
}

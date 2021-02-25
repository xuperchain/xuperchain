package evm

import (
	"fmt"
	"testing"

	"github.com/hyperledger/burrow/crypto"
)

func TestXchainToEVMAddress(t *testing.T) {
	// jSPJQSAR3NWoKcSFMxYGfcY8KVskvNMtm  D1824C1050F55CA7E564243CE087706CACF1C687
	// dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN  93F86A462A3174C7AD1281BCF400A9F18D244E06
	xchainAddr := "jSPJQSAR3NWoKcSFMxYGfcY8KVskvNMtm"
	xchainEvmAddr, err := XchainToEVMAddress(xchainAddr)
	if err != nil {
		t.Error(err)
	}
	fmt.Println("evm addr:", xchainEvmAddr.String())

	evmAddr := "D1824C1050F55CA7E564243CE087706CACF1C687"
	if xchainEvmAddr.String() != evmAddr {
		t.Errorf("expect %s got %s", evmAddr, xchainEvmAddr.String())
	}

	xchainFromEVMAddr, err := EVMAddressToXchain(xchainEvmAddr)
	if err != nil {
		t.Error(err)
	}
	//fmt.Println("xchain addr from evm:", xchainFromEVMAddr)

	if xchainFromEVMAddr != xchainAddr {
		t.Errorf("expect %s got %s", xchainAddr, xchainFromEVMAddr)
	}
}

func TestContractNameToEVMAddress(t *testing.T) {
	// testtoken1    0x313131312D2D2D2D2D2D74657374746F6B656E31
	// storagedata11    0x313131312D2D2D73746F72616765646174613131
	//contractName := "storagedata11"
	contractName := "mike_test_sol_31"
	contractNameEvmAddr, err := ContractNameToEVMAddress(contractName)
	if err != nil {
		t.Error(err)
	}

	// 0x313131312D2D2D73746F72616765646174613131
	//evmAddr := "313131312D2D2D73746F72616765646174613131"
	evmAddr := "313131316D696B655F746573745F736F6C5F3331"
	if contractNameEvmAddr.String() != evmAddr {
		t.Errorf("expect %s got %s", evmAddr, contractNameEvmAddr.String())
	}

	contractNameFromEVMAddr, err := EVMAddressToContractName(contractNameEvmAddr)
	if err != nil {
		t.Error(err)
	}

	if contractNameFromEVMAddr != contractName {
		t.Errorf("expect %s got %s", contractName, contractNameFromEVMAddr)
	}
}

func TestContractAccountToEVMAddress(t *testing.T) {
	contractAccount := "XC1111111111111113@xuper"
	contractAccountEvmAddr, err := ContractAccountToEVMAddress(contractAccount)
	if err != nil {
		t.Error(err)
	}

	// 0x3131313231313131313131313131313131313133
	evmAddr := "3131313231313131313131313131313131313133"
	if contractAccountEvmAddr.String() != evmAddr {
		t.Errorf("expect %s got %s", evmAddr, contractAccountEvmAddr.String())
	}

	contractAccountFromEVMAddr, err := EVMAddressToContractAccount(contractAccountEvmAddr)
	if err != nil {
		t.Error(err)
	}

	if contractAccountFromEVMAddr != contractAccount {
		t.Errorf("expect %s got %s", contractAccount, contractAccountFromEVMAddr)
	}
}

func TestDetermineEVMAddress(t *testing.T) {
	// contract account
	evmAddrHex := "3131313231313131313131313131313131313133"
	contractAccount := "XC1111111111111113@xuper"
	evmAddr, _ := crypto.AddressFromHexString(evmAddrHex)

	contractAccountFromEVMAddr, addrType, err := DetermineEVMAddress(evmAddr)
	if err != nil {
		t.Error(err)
	}
	if contractAccountFromEVMAddr != contractAccount {
		t.Errorf("expect %s got %s", contractAccount, contractAccountFromEVMAddr)
	}
	if addrType != contractAccountType {
		t.Errorf("expect %s got %s", contractAccountType, addrType)
	}

	// contract name
	evmAddrHex = "313131312D2D2D73746F72616765646174613131"
	contractName := "storagedata11"
	evmAddr, _ = crypto.AddressFromHexString(evmAddrHex)

	contractNameFromEVMAddr, addrType, err := DetermineEVMAddress(evmAddr)
	if err != nil {
		t.Error(err)
	}
	if contractNameFromEVMAddr != contractName {
		t.Errorf("expect %s got %s", contractName, contractNameFromEVMAddr)
	}
	if addrType != contractNameType {
		t.Errorf("expect %s got %s", contractNameType, addrType)
	}

	// xchain addr
	evmAddrHex = "D1824C1050F55CA7E564243CE087706CACF1C687"
	xchainAddr := "jSPJQSAR3NWoKcSFMxYGfcY8KVskvNMtm"
	evmAddr, _ = crypto.AddressFromHexString(evmAddrHex)

	xchainFromEVMAddr, addrType, err := DetermineEVMAddress(evmAddr)
	if err != nil {
		t.Error(err)
	}
	if xchainFromEVMAddr != xchainAddr {
		t.Errorf("expect %s got %s", xchainAddr, xchainFromEVMAddr)
	}
	if addrType != xchainAddrType {
		t.Errorf("expect %s got %s", xchainAddrType, addrType)
	}
}

func TestDetermineXchainAddress(t *testing.T) {
	// contract account
	evmAddrHex := "3131313231313131313131313131313131313133"
	contractAccount := "XC1111111111111113@xuper"

	contractAccountFromXchain, addrType, err := DetermineXchainAddress(contractAccount)
	if err != nil {
		t.Error(err)
	}
	if contractAccountFromXchain != evmAddrHex {
		t.Errorf("expect %s got %s", evmAddrHex, contractAccountFromXchain)
	}
	if addrType != contractAccountType {
		t.Errorf("expect %s got %s", contractAccountType, addrType)
	}

	// contract name
	evmAddrHex = "313131312D2D2D73746F72616765646174613131"
	contractName := "storagedata11"

	contractNameFromXchain, addrType, err := DetermineXchainAddress(contractName)
	if err != nil {
		t.Error(err)
	}
	if contractNameFromXchain != evmAddrHex {
		t.Errorf("expect %s got %s", evmAddrHex, contractNameFromXchain)
	}
	if addrType != contractNameType {
		t.Errorf("expect %s got %s", contractNameType, addrType)
	}

	// xchain addr
	evmAddrHex = "D1824C1050F55CA7E564243CE087706CACF1C687"
	xchainAddr := "jSPJQSAR3NWoKcSFMxYGfcY8KVskvNMtm"

	xchainFromXchain, addrType, err := DetermineXchainAddress(xchainAddr)
	if err != nil {
		t.Error(err)
	}
	if xchainFromXchain != evmAddrHex {
		t.Errorf("expect %s got %s", evmAddrHex, xchainFromXchain)
	}
	if addrType != xchainAddrType {
		t.Errorf("expect %s got %s", xchainAddrType, addrType)
	}
}

func TestDetermineContractNameFromEVM(t *testing.T) {
	// contract name
	evmAddrHex := "313131312D2D2D73746F72616765646174613131"
	contractName := "storagedata11"
	evmAddr, _ := crypto.AddressFromHexString(evmAddrHex)

	contractNameFromEVMAddr, err := DetermineContractNameFromEVM(evmAddr)
	if err != nil {
		t.Error(err)
	}
	if contractNameFromEVMAddr != contractName {
		t.Errorf("expect %s got %s", contractName, contractNameFromEVMAddr)
	}
}

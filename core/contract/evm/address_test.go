package evm

import (
	"fmt"
	"testing"
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
	// testtoken1    0909090909090909090974657374746F6B656E31
	// storagedata11    0x0909090909090973746F72616765646174613131
	contractName := "storagedata11"
	contractNameEvmAddr, err := ContractNameToEVMAddress(contractName)
	if err != nil {
		t.Error(err)
	}

	// 0x0909090909090973746F72616765646174613131
	evmAddr := "0909090909090973746F72616765646174613131"
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

	// 0x0A0A0A0A31313131313131313131313131313133
	evmAddr := "0A0A0A0A31313131313131313131313131313133"
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

package evm

import (
	"testing"

	"github.com/xuperchain/xuperchain/core/contract/bridge"
)

func TestEvmCodeKey(t *testing.T) {
	contractName := "test1234"
	contractCodeKey := "test1234.code"

	contractCodeKeyBytes := evmCodeKey(contractName)

	if string(contractCodeKeyBytes) != contractCodeKey {
		t.Errorf("expect %s got %s", contractCodeKey, string(contractCodeKeyBytes))
	}
}

func TestEvmAbiKey(t *testing.T) {
	contractName := "test1234"
	contractAbiKey := "test1234.abi"

	contractCodeKeyBytes := evmAbiKey(contractName)

	if string(contractCodeKeyBytes) != contractAbiKey {
		t.Errorf("expect %s got %s", contractAbiKey, string(contractCodeKeyBytes))
	}
}

func TestNewEvmCreator(t *testing.T) {
	creator, err := newEvmCreator(nil)
	if err != nil {
		t.Errorf("newEvmCreator error %v", err)
	}

	var cp bridge.ContractCodeProvider
	instance, err := creator.CreateInstance(&bridge.Context{
		ContractName: "contractName",
		Method:       "initialize",
	}, cp)

	instance.Abort("test")

	instance.Release()

	instance.ResourceUsed()

	creator.RemoveCache("contractName")

}

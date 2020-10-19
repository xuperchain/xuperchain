package bridge

import (
	"testing"

	"github.com/xuperchain/xuperchain/core/pb"
)

func TestContractCodeDescKey(t *testing.T) {
	contractName := "test1234"
	contractCodeDesc := "test1234.desc"

	contractCodeDescKeyBytes := ContractCodeDescKey(contractName)

	if string(contractCodeDescKeyBytes) != contractCodeDesc {
		t.Errorf("expect %s got %s", contractCodeDesc, string(contractCodeDescKeyBytes))
	}
}

func TestContractCodeKey(t *testing.T) {
	contractName := "test1234"
	contractCodeKeyStr := "test1234.code"

	contractCodeKeyBytes := contractCodeKey(contractName)

	if string(contractCodeKeyBytes) != contractCodeKeyStr {
		t.Errorf("expect %s got %s", contractCodeKeyStr, string(contractCodeKeyBytes))
	}
}

func TestContractAbiKey(t *testing.T) {
	contractName := "test1234"
	contractAbiKeyStr := "test1234.abi"

	contractAbiKeyBytes := contractAbiKey(contractName)

	if string(contractAbiKeyBytes) != contractAbiKeyStr {
		t.Errorf("expect %s got %s", contractAbiKeyStr, string(contractAbiKeyBytes))
	}
}

func TestGetContractType(t *testing.T) {
	descpb := new(pb.WasmCodeDesc)
	descpb.ContractType = string(TypeEvm)

	contractType, err := getContractType(descpb)
	if err != nil {
		t.Errorf("getContractType error %v", err)
	}

	if contractType != TypeEvm {
		t.Errorf("expect %s got %s", TypeEvm, contractType)
	}
}

package evm

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hyperledger/burrow/execution/exec"
	abi2 "github.com/hyperledger/burrow/execution/evm/abi"

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

func TestUnpackEventFromAbi(t *testing.T){
	abi := `[{"inputs":[],"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"string","name":"key","type":"string"},{"indexed":false,"internalType":"uint256","name":"value","type":"uint256"}],"name":"increaseEvent","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"string","name":"key","type":"string"},{"indexed":false,"internalType":"uint256","name":"value","type":"uint256"}],"name":"increaseEvent1","type":"event"},{"inputs":[{"internalType":"string","name":"key","type":"string"}],"name":"get","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getOwner","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"string","name":"key","type":"string"}],"name":"increase","outputs":[],"stateMutability":"payable","type":"function"}]`
	contractName := "increaseEvent"

	//==== 以下为组装log
	eventAbi := `{"anonymous":false,"inputs":[{"indexed":false,"internalType":"string","name":"key","type":"string"},{"indexed":false,"internalType":"uint256","name":"value","type":"uint256"}],"name":"increaseEvent","type":"event"}`	//
	type args struct {
		Key string
		Value  int64
	}
	in := &args{
		Key:"test",
		Value:12,
	}
	eventSpec := new(abi2.EventSpec)
	err := json.Unmarshal([]byte(eventAbi), eventSpec)
	if err != nil {
		t.Error(err)
	}
	topics,data,err := abi2.PackEvent(eventSpec,in)
	log := &exec.LogEvent{}
	log.Topics = topics
	log.Data = data
	//====

	event,err := unpackEventFromAbi([]byte(abi),contractName,log)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%+v\n",event)
}
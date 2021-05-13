package evm

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/hyperledger/burrow/execution/exec"

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

func TestDecodeRespWithAbiForEVM(t *testing.T) {
	abi := `[{"inputs":[{"internalType":"uint256","name":"num","type":"uint256"}],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"constant":false,"inputs":[],"name":"retrieve","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"num","type":"uint256"}],"name":"store","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"}]`
	out, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000005")

	result, err := decodeRespWithAbiForEVM(abi, "retrieve", out)
	if err != nil {
		t.Error(err)
	}
	fmt.Println("success:", string(result))
	if string(result) != `[{"0":"5"}]` {
		t.Error("decodeRespWithAbiForEVM failed")
	}
}

func TestEncodeInvokeInput(t *testing.T) {
	inputBytes := []byte(`{"num":"1"}`)
	ei := &evmInstance{
		ctx: &bridge.Context{
			ContractName: "contractName",
			Method:       "store",
			Args:         map[string][]byte{"input": inputBytes},
		},
		abi: []byte(`[{"inputs":[{"internalType":"uint256","name":"num","type":"uint256"}],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"constant":false,"inputs":[],"name":"retrieve","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"num","type":"uint256"}],"name":"store","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"}]`),
	}

	_, err := ei.encodeInvokeInput()
	if err != nil {
		t.Error(err)
	}
}

func TestEncodeDeployInput(t *testing.T) {
	inputBytes := []byte(`{"num":"1"}`)
	ei := &evmInstance{
		ctx: &bridge.Context{
			ContractName: "contractName",
			Method:       "initialize",
			Args:         map[string][]byte{"input": inputBytes},
		},
		abi: []byte(`[{"inputs":[{"internalType":"uint256","name":"num","type":"uint256"}],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"constant":false,"inputs":[],"name":"retrieve","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"num","type":"uint256"}],"name":"store","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"}]`),
	}

	_, err := ei.encodeDeployInput()
	if err != nil {
		t.Error(err)
	}
}

func TestUnpackEventFromAbi(t *testing.T) {
	testCases := []struct {
		abiJson      string
		contractName string
		eventAbi     string
	}{
		{
			abiJson:      `[{"inputs":[],"stateMutability"{"indexed":false,"internalType":"string","name":"key","type":"string"},{"indexed":false,"internalType":"uint256","name":"value","type":"uint256"}],"name":"increaseEvent","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"string","name":"key","type":"string"},{"indexed":false,"internalType":"uint256","name":"value","type":"uint256"}],"name":"increaseEvent1","type":"event"},{"inputs":[{"internalType":"string","name":"key","type":"string"}],"name":"get","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getOwner","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"string","name":"key","type":"string"}],"name":"increase","outputs":[],"stateMutability":"payable","type":"function"}]`,
			contractName: "increaseEvent",
			eventAbi:     `{"anonymous":false,"inputs":[{"indexed":false,"internalType":"string","name":"key","type":"string"},{"indexed":false,"internalType":"uint256","name":"value","type":"uint256"}],"name":"increaseEvent","type":"event"}`, //

		},
		{
			abiJson:      `[{"inputs":[],"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"string","name":"key","type":"string"},{"indexed":false,"internalType":"uint256","name":"value","type":"uint256"}],"name":"increaseEvent","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"string","name":"key","type":"string"},{"indexed":false,"internalType":"uint256","name":"value","type":"uint256"}],"name":"increaseEvent1","type":"event"},{"inputs":[{"internalType":"string","name":"key","type":"string"}],"name":"get","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getOwner","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"string","name":"key","type":"string"}],"name":"increase","outputs":[],"stateMutability":"payable","type":"function"}]`,
			contractName: "",
			eventAbi:     `{"anonymous":false,"inputs":[{"indexed":false,"internalType":"string","name":"key","type":"string"},{"indexed":false,"internalType":"uint256","name":"value","type":"uint256"}],"name":"increaseEvent","type":"event"}`, //

		}, {
			abiJson:      `[{"inputs":[],"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"string","name":"key","type":"string"},{"indexed":false,"internalType":"uint256","name":"value","type":"uint256"}],"name":"increaseEvent","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"string","name":"key","type":"string"},{"indexed":false,"internalType":"uint256","name":"value","type":"uint256"}],"name":"increaseEvent1","type":"event"},{"inputs":[{"internalType":"string","name":"key","type":"string"}],"name":"get","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getOwner","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"string","name":"key","type":"string"}],"name":"increase","outputs":[],"stateMutability":"payable","type":"function"}]`,
			contractName: "increaseEvent",
			eventAbi:     `{"anonymous":false,"inputs":[{"indexed":false,"internalType":"string","name":"key","type":"string"},{"indexed":false,"internalType":"uint256","name":"value","type":"uint256"}],"name":"increaseEvent","type":"event"}`, //

		}, {
			abiJson:      `[{"inputs":[],"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"string","name":"key","type":"string"},{"indexed":false,"internalType":"uint256","name":"value","type":"uint256"}],"name":"increaseEvent","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"string","name":"key","type":"string"},{"indexed":false,"internalType":"uint256","name":"value","type":"uint256"}],"name":"increaseEvent1","type":"event"},{"inputs":[{"internalType":"string","name":"key","type":"string"}],"name":"get","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getOwner","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"string","name":"key","type":"string"}],"name":"increase","outputs":[],"stateMutability":"payable","type":"function"}]`,
			contractName: "increaseEvent",
			eventAbi:     `{"anonymous":false,"inputs":[{"indexed":false,"internalType":"string","name":"key","type":"string"},{"indexed":false,"internalType":"uint256","name":"value","type":"uint256"}],"name":"increaseEvent","type":"event"}`, //

		},
	}

	type args struct {
		Key   string
		Value int64
	}
	in := &args{
		Key:   "test",
		Value: 12,
	}

	for k, param := range testCases {
		eventSpec := new(abi.EventSpec)
		err := json.Unmarshal([]byte(param.eventAbi), eventSpec)
		if err != nil {
			t.Error(err)
		}
		topics, data, err := abi.PackEvent(eventSpec, in)
		log := &exec.LogEvent{}
		log.Topics = topics
		log.Data = data

		event, err := unpackEventFromAbi([]byte(param.abiJson), param.contractName, log)
		if err != nil && k == len(testCases)-1 {
			t.Error(err)
		} else {
			fmt.Println("expect err:", err)
		}
		fmt.Printf("%+v\n", event)
	}
}

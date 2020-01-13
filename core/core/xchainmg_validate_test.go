package xchaincore

import (
	"fmt"
	"testing"

	"encoding/json"
	"github.com/xuperchain/xuperchain/core/contract"
	"github.com/xuperchain/xuperchain/core/pb"
)

func TestValidatePostTx(t *testing.T) {
	testCases := map[string]struct {
		in *pb.TxStatus
	}{
		"test validate TxStatus.Bcname": {
			in: &pb.TxStatus{
				Bcname: "",
				Txid:   []byte("todo"),
			},
		},
		"test validate TxStatus.Txid": {
			in: &pb.TxStatus{
				Bcname: "Bcname",
			},
		},
	}
	for testName, testCase := range testCases {
		err := validatePostTx(testCase.in)
		fmt.Println(err.Error())
		if err == nil {
			t.Errorf("%s expected: %v, actual: %v", testName, "not nil", "nil")
		}
	}
}

func TestCheckContractAuthority(t *testing.T) {
	contractWhiteList := map[string]map[string]bool{
		"kernel": map[string]bool{
			"bob":   true,
			"alice": false,
		},
	}
	desc := contract.TxDesc{
		Module: "kernel",
	}
	strDesc, _ := json.Marshal(desc)
	fakeTx := &pb.Transaction{
		TxInputs: []*pb.TxInput{
			&pb.TxInput{
				FromAddr: []byte("bob"),
			},
		},
		Desc: []byte(strDesc),
	}
	state, err := checkContractAuthority(contractWhiteList, fakeTx)
	if err != nil {
		t.Error("checkContractAuthority error ", err.Error())
	} else {
		t.Log("checkContractAuthority state ", state)
	}
}

func TestValidateSendBlock(t *testing.T) {
	testCases := map[string]struct {
		in *pb.Block
	}{
		"test validate Block.Block.Blockid": {
			in: &pb.Block{},
		},
		"test validate Block.Block": {
			in: &pb.Block{
				Blockid: []byte("123"),
			},
		},
	}
	for testName, testCase := range testCases {
		err := validateSendBlock(testCase.in)
		fmt.Println(err.Error())
		if err == nil {
			t.Errorf("%s expected: %v, actual: %v", testName, "not nil", "nil")
		}
	}
}

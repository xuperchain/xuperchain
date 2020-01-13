package server

import (
	"fmt"
	"github.com/xuperchain/xuperchain/core/pb"
	"testing"
)

func TestValidateSendBlock(t *testing.T) {
	testCases := map[string]struct {
		in *pb.Block
	}{
		"test validate Block.Blockid": {
			in: &pb.Block{},
		},
		"test validate Block.Block": {
			in: &pb.Block{
				Blockid: []byte("todo"),
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

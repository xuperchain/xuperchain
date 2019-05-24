// +build wasm

package wasm

import (
	"fmt"
	"os"

	"github.com/xuperchain/xuperunion/contractsdk/go/code"
	pb "github.com/xuperchain/xuperunion/contractsdk/go/litepb"
)

func returnResponse(resp *code.Response) {
	syscall(methodOutput, &pb.SetOutputRequest{
		Response: &pb.Response{
			Status:  int32(resp.Status),
			Message: resp.Message,
			Body:    resp.Body,
		},
	}, new(pb.SetOutputResponse))
}

func fatal(x interface{}) {
	var msg string
	switch e := x.(type) {
	case error:
		msg = e.Error()
	case string:
		msg = e
	default:
		msg = fmt.Sprintf("%s", e)
	}
	returnResponse(&code.Response{
		Status:  500,
		Message: msg,
	})
	os.Exit(0)
}

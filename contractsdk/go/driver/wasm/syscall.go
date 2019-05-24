// +build wasm

package wasm

import (
	"errors"

	"github.com/golang/protobuf/proto"
)

func callMethod(method string, request []byte) uint64
func fetchResponse(response []byte) uint64

func syscall(method string, request proto.Message, response proto.Message) error {
	buf, _ := proto.Marshal(request)
	responseLen := callMethod(method, buf)
	if responseLen == 0 {
		return nil
	}
	responseBuf := make([]byte, responseLen)
	success := fetchResponse(responseBuf)
	if success == 0 {
		return errors.New(string(responseBuf))
	}
	return proto.Unmarshal(responseBuf, response)
}

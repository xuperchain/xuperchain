package xvm

import (
	"context"
	"encoding/binary"
	"fmt"

	"github.com/xuperchain/xuperunion/contract/bridge"
	"github.com/xuperchain/xuperunion/contract/bridge/memrpc"
	"github.com/xuperchain/xuperunion/xvm/exec"
)

const (
	contextIDKey = "ctxid"
	responseKey  = "callResponse"
)

type responseDesc struct {
	Body  []byte
	Error bool
}

type syscallResolver struct {
	syscall   *bridge.SyscallService
	rpcserver *memrpc.Server
}

func newSyscallResolver(syscall *bridge.SyscallService) exec.Resolver {
	return &syscallResolver{
		syscall:   syscall,
		rpcserver: memrpc.NewServer(syscall),
	}
}

func (s *syscallResolver) ResolveGlobal(module, name string) (float64, bool) {
	return 0, false
}

func (s *syscallResolver) ResolveFunc(module, name string) (interface{}, bool) {
	fullname := module + "." + name
	switch fullname {
	case "go.github.com/xuperchain/xuperunion/contractsdk/go/driver/wasm.callMethod":
		return s.goCallMethod, true
	case "go.github.com/xuperchain/xuperunion/contractsdk/go/driver/wasm.fetchResponse":
		return s.goFetchResponse, true
	case "env._call_method":
		return s.cCallMethod, true
	case "env._fetch_response":
		return s.cFetchResponse, true
	default:
		return nil, false
	}
}

func (s *syscallResolver) goCallMethod(ctx *exec.Context, sp uint32) uint32 {
	codec := exec.NewCodec(ctx)
	ctxid := ctx.GetUserData(contextIDKey).(int64)
	method := codec.GoString(sp + 8)
	requestBuf := codec.GoBytes(sp + 24)
	responseBuf, err := s.rpcserver.CallMethod(context.TODO(), ctxid, method, requestBuf)
	var responseDesc responseDesc
	if err != nil {
		responseDesc.Error = true
		responseDesc.Body = []byte(err.Error())
	} else {
		responseDesc.Body = responseBuf
	}
	binary.LittleEndian.PutUint64(codec.Bytes(sp+48, 8), uint64(len(responseDesc.Body)))
	ctx.SetUserData(responseKey, responseDesc)
	return 0
}

func (s *syscallResolver) goFetchResponse(ctx *exec.Context, sp uint32) uint32 {
	codec := exec.NewCodec(ctx)
	iresponse := ctx.GetUserData(responseKey)
	if iresponse == nil {
		exec.Throw(exec.NewTrap("call fetchResponse on nil value"))
	}
	response := iresponse.(responseDesc)
	userbuf := codec.GoBytes(sp + 8)
	if len(response.Body) != len(userbuf) {
		exec.Throw(exec.NewTrap(fmt.Sprintf("call fetchResponse with bad length, got %d, expect %d", len(userbuf), len(response.Body))))
	}
	copy(userbuf, response.Body)
	success := uint64(1)
	if response.Error {
		success = 0
	}
	binary.LittleEndian.PutUint64(codec.Bytes(sp+32, 8), success)
	ctx.SetUserData(responseKey, nil)
	return 0
}

func (s *syscallResolver) cCallMethod(ctx *exec.Context, methodAddr, methodLen, requestAddr, requestLen uint32) uint32 {
	codec := exec.NewCodec(ctx)
	ctxid := ctx.GetUserData(contextIDKey).(int64)
	method := codec.String(methodAddr, methodLen)
	requestBuf := codec.Bytes(requestAddr, requestLen)
	responseBuf, err := s.rpcserver.CallMethod(context.TODO(), ctxid, method, requestBuf)
	var responseDesc responseDesc
	if err != nil {
		responseDesc.Error = true
		responseDesc.Body = []byte(err.Error())
	} else {
		responseDesc.Body = responseBuf
	}
	ctx.SetUserData(responseKey, responseDesc)
	return uint32(len(responseDesc.Body))
}

func (s *syscallResolver) cFetchResponse(ctx *exec.Context, userBuf, userLen uint32) uint32 {
	codec := exec.NewCodec(ctx)
	iresponse := ctx.GetUserData(responseKey)
	if iresponse == nil {
		exec.Throw(exec.NewTrap("call fetchResponse on nil value"))
	}
	response := iresponse.(responseDesc)
	userbuf := codec.Bytes(userBuf, userLen)
	if len(response.Body) != len(userbuf) {
		exec.Throw(exec.NewTrap(fmt.Sprintf("call fetchResponse with bad length, got %d, expect %d", len(userbuf), len(response.Body))))
	}
	copy(userbuf, response.Body)
	success := uint32(1)
	if response.Error {
		success = 0
	}
	ctx.SetUserData(responseKey, nil)
	return success
}

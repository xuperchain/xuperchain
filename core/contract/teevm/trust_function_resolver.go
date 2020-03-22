package teevm

import (
	"encoding/json"
	"errors"
	"plugin"

	"github.com/xuperchain/xuperchain/core/common/config"
	"github.com/xuperchain/xuperchain/core/common/log"
	"github.com/xuperchain/xuperchain/core/xvm/exec"
	"github.com/xuperchain/xuperchain/core/xvm/runtime/emscripten"
)

// TrustFunctionResolver
type TrustFunctionResolver struct {
	handler *plugin.Plugin
	runFunc func([]byte) ([]byte, error)
}

var _ exec.Resolver = (*TrustFunctionResolver)(nil)

//
//
func NewTrustFunctionResolver(conf *config.TEEConfig) (*TrustFunctionResolver, error) {
	if conf.Enable == false {
		return nil, errors.New("private ledger is not enabled")
	}
	data, err := json.Marshal(conf)
	if err != nil {
		return nil, err
	}
	p, err := plugin.Open(conf.PluginPath)
	if err != nil {
		return nil, err
	}
	initFunc, err := p.Lookup("Init")
	if err != nil {
		return nil, err
	}
	if err := initFunc.(func(string) error)(string(data)); err != nil {
		return nil, err
	}
	runFunc, err := p.Lookup("Run")
	if err != nil {
		return nil, err
	}
	return &TrustFunctionResolver{handler: p, runFunc: runFunc.(func([]byte) ([]byte, error))}, nil
}

func (tf *TrustFunctionResolver) ResolveGlobal(module, name string) (int64, bool) {
	return 0, false
}

func (tf *TrustFunctionResolver) ResolveFunc(module, name string) (interface{}, bool) {
	fullname := module + "." + name
	switch fullname {
	case "env._xvm_tfcall":
		return tf.tfcall, true
	default:
		return nil, false
	}
}

func (tf *TrustFunctionResolver) tfcall(ctx exec.Context, inptr, inlen, outpptr, outlenptr uint32) uint32 {
	var (
		retCode uint32 = 0
	)
	codec := exec.NewCodec(ctx)
	requestBuf := codec.Bytes(inptr, inlen)
	responseBuf, err := tf.runFunc(requestBuf)

	if err != nil {
		log.Error("contract trust function call error", "error", err)
		copy(responseBuf, []byte(err.Error()))
		retCode = 1
	}
	codec.SetUint32(outpptr, bytesdup(ctx, responseBuf))
	codec.SetUint32(outlenptr, uint32(len(responseBuf)))
	return retCode
}

//copied from https://github.com/xuperchain/xuperchain/blob/master/core/contract/wasm/vm/xvm/builtin_resolver.go#L180, TODO refer not copy
func bytesdup(ctx exec.Context, b []byte) uint32 {
	codec := exec.NewCodec(ctx)
	memptr, err := emscripten.Malloc(ctx, len(b))
	if err != nil {
		exec.ThrowError(err)
	}
	mem := codec.Bytes(memptr, uint32(len(b)))
	copy(mem, b)
	return memptr
}

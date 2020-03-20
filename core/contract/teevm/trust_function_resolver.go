package teevm

import (
	"encoding/json"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/xuperdata/teesdk"

	"github.com/xuperchain/xuperchain/core/common/log"
	"github.com/xuperchain/xuperchain/core/contract/teevm/pb"
	"github.com/xuperchain/xuperchain/core/xvm/exec"
	"github.com/xuperchain/xuperchain/core/xvm/runtime/emscripten"
)

// TrustFunctionResolver
type TrustFunctionResolver struct {
	client *teesdk.TEEClient
	config *teesdk.TEEConfig
}

var _ exec.Resolver = (*TrustFunctionResolver)(nil)

func NewTrustFunctionResolver(cfg *teesdk.TEEConfig) *TrustFunctionResolver {
	client := teesdk.NewTEEClient(cfg.Uid,
		cfg.Token,
		cfg.Auditors[0].PublicDer,
		cfg.Auditors[0].Sign,
		cfg.Auditors[0].EnclaveInfoConfig,
		cfg.TMSPort,
		cfg.TDFSPort)
	return &TrustFunctionResolver{
		client: client,
		config: cfg}
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
		responseBuf, tmpbuf []byte
		err                 error
		kvs                 *pb.TrustFunctionCallResponse_Kvs
		k, v, tmpbufstr     string
		plainMap            map[string]string
		retCode             uint32 = 0
	)
	codec := exec.NewCodec(ctx)
	requestBuf := codec.Bytes(inptr, inlen)
	in := &pb.TrustFunctionCallRequest{}
	if err = proto.Unmarshal(requestBuf, in); err != nil {
		goto ret
	}
	if tf.config != nil && !tf.config.Enable || tf.client == nil {
		err = fmt.Errorf("IsTFCEnabled is false, this node doest not enable TEE")
		goto ret
	}
	if tmpbuf, err = json.Marshal(teesdk.FuncCaller{
		Method: in.Method, Args: in.Args, Svn: in.Svn,
		Address: in.Address}); err != nil {
		goto ret
	}
	if tmpbufstr, err = tf.client.Submit("xchaintf", string(tmpbuf)); err != nil {
		goto ret
	}
	if err = json.Unmarshal([]byte(tmpbufstr), &plainMap); err != nil {
		goto ret
	}
	kvs = &pb.TrustFunctionCallResponse_Kvs{
		Kvs: &pb.KVPairs{},
	}
	for k, v = range plainMap {
		kvs.Kvs.Kv = append(kvs.Kvs.Kv, &pb.KVPair{Key: k, Value: v})
	}
	responseBuf, err = proto.Marshal(&pb.TrustFunctionCallResponse{Results: kvs})

ret:
	if err != nil {
		err = fmt.Errorf("TrustFunctionCall: " + err.Error())
		log.Error("contract trust function call error", "method", in.Method, "error", err)
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

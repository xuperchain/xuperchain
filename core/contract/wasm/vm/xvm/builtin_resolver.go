package xvm

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"unsafe"

	"github.com/golang/protobuf/proto"
	"github.com/xuperchain/xuperchain/core/contract/bridge"
	"github.com/xuperchain/xuperchain/core/crypto/account"
	"github.com/xuperchain/xuperchain/core/crypto/sign"
	"github.com/xuperchain/xuperchain/core/pb"
	"github.com/xuperchain/xuperchain/core/utxo/txhash"
	"github.com/xuperchain/xuperchain/core/xvm/exec"
	"github.com/xuperchain/xuperchain/core/xvm/runtime/emscripten"
)

func touint32(n int32) uint32 {
	return *(*uint32)(unsafe.Pointer(&n))
}

func hashFunc(name string) hash.Hash {
	switch name {
	case "sha256":
		return sha256.New()
	default:
		return nil
	}
}

func xvmHash(ctx exec.Context,
	nameptr uint32,
	inputptr uint32, inputlen uint32,
	outputptr uint32, outputlen uint32) uint32 {

	codec := exec.NewCodec(ctx)
	name := codec.CString(nameptr)
	input := codec.Bytes(inputptr, inputlen)
	output := codec.Bytes(outputptr, outputlen)

	hasher := hashFunc(name)
	if hasher == nil {
		exec.ThrowMessage(fmt.Sprintf("hash %s not found", name))
	}
	hasher.Write(input)
	out := hasher.Sum(nil)
	copy(output, out[:])
	return 0
}

type codec interface {
	Encode(in []byte) []byte
	Decode(in []byte) ([]byte, error)
}

func getCodec(name string) codec {
	switch name {
	case "hex":
		return hexCodec{}
	default:
		return nil
	}
}

type hexCodec struct{}

func (h hexCodec) Encode(in []byte) []byte {
	out := make([]byte, hex.EncodedLen(len(in)))
	hex.Encode(out, in)
	return out
}
func (h hexCodec) Decode(in []byte) ([]byte, error) {
	out := make([]byte, hex.DecodedLen(len(in)))
	_, err := hex.Decode(out, in)
	return out, err
}

func xvmEncode(ctx exec.Context,
	nameptr uint32,
	inputptr uint32, inputlen uint32,
	outputpptr uint32, outputLenPtr uint32) uint32 {

	codec := exec.NewCodec(ctx)
	name := codec.CString(nameptr)
	input := codec.Bytes(inputptr, inputlen)

	c := getCodec(name)
	if c == nil {
		exec.ThrowMessage(fmt.Sprintf("codec %s not found", name))
	}
	out := c.Encode(input)

	codec.SetUint32(outputpptr, bytesdup(ctx, out))
	codec.SetUint32(outputLenPtr, uint32(len(out)))
	return 0
}

func xvmDecode(ctx exec.Context,
	nameptr uint32,
	inputptr uint32, inputlen uint32,
	outputpptr uint32, outputLenPtr uint32) uint32 {

	codec := exec.NewCodec(ctx)
	name := codec.CString(nameptr)
	input := codec.Bytes(inputptr, inputlen)

	c := getCodec(name)
	if c == nil {
		exec.ThrowMessage(fmt.Sprintf("codec %s not found", name))
	}
	out, err := c.Decode(input)
	if err != nil {
		return 1
	}

	codec.SetUint32(outputpptr, bytesdup(ctx, out))
	codec.SetUint32(outputLenPtr, uint32(len(out)))
	return 0
}

func xvmECVerify(ctx exec.Context,
	pubptr, publen,
	sigptr, siglen, hashptr, hashlen uint32) uint32 {
	codec := exec.NewCodec(ctx)

	pubkeyJSON := codec.Bytes(pubptr, publen)
	sig := codec.Bytes(sigptr, siglen)
	hash := codec.Bytes(hashptr, hashlen)
	pubkey, err := account.GetEcdsaPublicKeyFromJSON(pubkeyJSON)
	if err != nil {
		return touint32(-1)
	}

	ok, _ := sign.VerifyECDSA(pubkey, sig, hash)
	if ok {
		return 0
	}
	return touint32(-1)
}

func xvmMakeTx(ctx exec.Context, txptr, txlen, outpptr, outlenPtr uint32) uint32 {
	codec := exec.NewCodec(ctx)
	txbuf := codec.Bytes(txptr, txlen)
	tx := new(pb.Transaction)
	err := proto.Unmarshal(txbuf, tx)
	if err != nil {
		return touint32(-1)
	}
	txid, err := txhash.MakeTransactionID(tx)
	if err != nil {
		return touint32(-1)
	}
	outpb := bridge.ConvertTxToSDKTx(tx)
	outpb.Txid = hex.EncodeToString(txid)

	buf, _ := proto.Marshal(outpb)
	codec.SetUint32(outpptr, bytesdup(ctx, buf))
	codec.SetUint32(outlenPtr, uint32(len(buf)))
	return 0
}

func xvmAddressFromPubkey(ctx exec.Context, pubptr, publen uint32) uint32 {
	codec := exec.NewCodec(ctx)
	pubkeystr := codec.Bytes(pubptr, publen)
	pubkey, err := account.GetEcdsaPublicKeyFromJSON(pubkeystr)
	if err != nil {
		return 0
	}
	addr, err := account.GetAddressFromPublicKey(pubkey)
	if err != nil {
		return 0
	}
	return strdup(ctx, addr)
}

// Returns a pointer to a bytes, which is a duplicate of b.
// The returned pointer must be passed to free to avoid a memory leak
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

// Returns a pointer to a null-terminated string, which is a duplicate of the string s.
// The returned pointer must be passed to free to avoid a memory leak
func strdup(ctx exec.Context, s string) uint32 {
	codec := exec.NewCodec(ctx)
	memptr, err := emscripten.Malloc(ctx, len(s)+1)
	if err != nil {
		exec.ThrowError(err)
	}
	mem := codec.Bytes(memptr, uint32(len(s)+1))
	copy(mem, s)
	mem[len(s)] = 0
	return memptr
}

var builtinResolver = exec.MapResolver(map[string]interface{}{
	"env._xvm_hash":             xvmHash,
	"env._xvm_encode":           xvmEncode,
	"env._xvm_decode":           xvmDecode,
	"env._xvm_ecverify":         xvmECVerify,
	"env._xvm_make_tx":          xvmMakeTx,
	"env._xvm_addr_from_pubkey": xvmAddressFromPubkey,
})

package xvm

import (
	"crypto/sha256"
	"encoding/hex"
	"hash"

	"github.com/xuperchain/xuperchain/core/xvm/exec"
)

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
	outputptr uint32, outputlen uint32,
	hexEncoding uint32) uint32 {

	codec := exec.NewCodec(ctx)
	name := codec.CString(nameptr)
	input := codec.Bytes(inputptr, inputlen)
	output := codec.Bytes(outputptr, outputlen)

	hasher := hashFunc(name)
	if hasher == nil {
		return 1
	}
	hasher.Write(input)
	out := hasher.Sum(nil)
	if hexEncoding == 0 {
		copy(output, out[:])
		return 0
	}
	hex.Encode(output, out[:])
	return 0
}

var builtinResolver = exec.MapResolver(map[string]interface{}{
	"env._xvm_hash": xvmHash,
})

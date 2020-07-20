package txhash

import (
	"crypto/sha256"
	"encoding/binary"
	"io"
	"log"
	"sort"

	"github.com/xuperchain/xuperchain/core/pb"
)

type encoder struct {
	intbuf [8]byte
	w      io.Writer
}

func newEncoder(w io.Writer) *encoder {
	return &encoder{
		w: w,
	}
}

func (e *encoder) EncodeInt32(x int32) {
	buf := e.intbuf[:4]
	binary.BigEndian.PutUint32(buf, uint32(x))
	e.w.Write(buf)
}

func (e *encoder) EncodeInt64(x int64) {
	buf := e.intbuf[:8]
	binary.BigEndian.PutUint64(buf, uint64(x))
	e.w.Write(buf)
}

func (e *encoder) EncodeString(s string) {
	if len(s) == 0 {
		return
	}
	io.WriteString(e.w, s)
}

func (e *encoder) EncodeBytes(s []byte) {
	e.EncodeInt32(int32(len(s)))
	if len(s) == 0 {
		return
	}
	e.w.Write(s)
}

func (e *encoder) EncodeMap(m map[string][]byte) {
	length := len(m)
	e.EncodeInt32(int32(length))
	if length == 0 {
		return
	}
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	e.Encode(len(m))
	for _, key := range keys {
		e.EncodeString(key)
		e.EncodeBytes(m[key])
	}
}

func (e *encoder) Encode(x interface{}) {
	switch v := x.(type) {
	case bool:
		if v {
			e.EncodeInt32(int32(1))
		} else {
			e.EncodeInt32(int32(0))
		}
	case int:
		e.EncodeInt32(int32(v))
	case int32:
		e.EncodeInt32(v)
	case int64:
		e.EncodeInt64(v)
	case string:
		e.EncodeString(v)
	case []byte:
		e.EncodeBytes(v)
	case map[string][]byte:
		e.EncodeMap(v)
	default:
		log.Panicf("not supported type:%T", x)
	}
}

// txDigestHashV2 make tx hash using double sha256
func txDigestHashV2(tx *pb.Transaction, includeSigns bool) []byte {
	h := sha256.New()
	enc := newEncoder(h)

	// encode TxInputs
	enc.Encode(len(tx.TxInputs))
	for _, input := range tx.TxInputs {
		enc.Encode(input.RefTxid)
		enc.Encode(input.RefOffset)
		enc.Encode(input.FromAddr)
		enc.Encode(input.Amount)
		enc.Encode(input.FrozenHeight)
	}

	// encode TxOutputs
	enc.Encode(len(tx.TxOutputs))
	for _, output := range tx.TxOutputs {
		enc.Encode(output.Amount)
		enc.Encode(output.ToAddr)
		enc.Encode(output.FrozenHeight)
	}

	enc.Encode(tx.Desc)
	enc.Encode(tx.Nonce)
	enc.Encode(tx.Timestamp)
	enc.Encode(tx.Version)

	// encode TxInputsExt
	enc.Encode(len(tx.TxInputsExt))
	for _, input := range tx.TxInputsExt {
		enc.Encode(input.Bucket)
		enc.Encode(input.Key)
		enc.Encode(input.RefTxid)
		enc.Encode(input.RefOffset)
	}

	// encode TxOutputsExt
	enc.Encode(len(tx.TxOutputsExt))
	for _, output := range tx.TxOutputsExt {
		enc.Encode(output.Bucket)
		enc.Encode(output.Key)
		enc.Encode(output.Value)
	}

	// encode ContractRequests
	enc.Encode(len(tx.ContractRequests))
	for _, req := range tx.ContractRequests {
		enc.Encode(req.ModuleName)
		enc.Encode(req.ContractName)
		enc.Encode(req.MethodName)
		enc.Encode(req.Args)

		enc.Encode(len(req.ResourceLimits))
		for _, limit := range req.ResourceLimits {
			enc.Encode(int32(limit.Type))
			enc.Encode(limit.Limit)
		}
		enc.Encode(req.Amount)
	}

	enc.Encode(tx.Initiator)
	enc.Encode(len(tx.AuthRequire))
	for _, addr := range tx.AuthRequire {
		enc.Encode(addr)
	}

	encSigs := func(sigs []*pb.SignatureInfo) {
		enc.Encode(len(sigs))
		for _, sig := range sigs {
			enc.Encode(sig.PublicKey)
			enc.Encode(sig.Sign)
		}
	}
	if includeSigns {
		encSigs(tx.InitiatorSigns)
		encSigs(tx.AuthRequireSigns)
		// encode PublicKeys
		xuperSign := tx.GetXuperSign()
		enc.Encode(len(xuperSign.GetPublicKeys()))
		for _, pubkey := range xuperSign.GetPublicKeys() {
			enc.Encode(pubkey)
		}
		enc.Encode(tx.GetXuperSign().GetSignature())
	}

	enc.Encode(tx.Coinbase)
	enc.Encode(tx.Autogen)

	enc.Encode(tx.GetHDInfo().GetHdPublicKey())
	enc.Encode(tx.GetHDInfo().GetOriginalHash())

	sum := sha256.Sum256(h.Sum(nil))
	return sum[:]
}

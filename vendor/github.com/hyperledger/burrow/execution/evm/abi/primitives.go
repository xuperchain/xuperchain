package abi

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"unsafe" // just for Sizeof

	"github.com/hyperledger/burrow/crypto"
)

// EVM Solidity calls and return values are packed into
// pieces of 32 bytes, including a bool (wasting 255 out of 256 bits)
const ElementSize = 32

type EVMType interface {
	GetSignature() string
	getGoType() interface{}
	pack(v interface{}) ([]byte, error)
	unpack(data []byte, offset int, v interface{}) (int, error)
	Dynamic() bool
	ImplicitCast(o EVMType) bool
}

var _ EVMType = (*EVMBool)(nil)

type EVMBool struct {
}

func (e EVMBool) String() string {
	return "EVMBool"
}

func (e EVMBool) GetSignature() string {
	return "bool"
}

func (e EVMBool) getGoType() interface{} {
	return new(bool)
}

func (e EVMBool) pack(v interface{}) ([]byte, error) {
	var b bool
	arg := reflect.ValueOf(v)
	if arg.Kind() == reflect.String {
		val := arg.String()
		if strings.EqualFold(val, "true") || val == "1" {
			b = true
		} else if strings.EqualFold(val, "false") || val == "0" {
			b = false
		} else {
			return nil, fmt.Errorf("%s is not a valid value for EVM Bool type", val)
		}
	} else if arg.Kind() == reflect.Bool {
		b = arg.Bool()
	} else {
		return nil, fmt.Errorf("%s cannot be converted to EVM Bool type", arg.Kind().String())
	}
	res := make([]byte, ElementSize)
	if b {
		res[ElementSize-1] = 1
	}
	return res, nil
}

func (e EVMBool) unpack(data []byte, offset int, v interface{}) (int, error) {
	if len(data)-offset < 32 {
		return 0, fmt.Errorf("%v: not enough data", e)
	}
	data = data[offset:]
	switch v := v.(type) {
	case *string:
		if data[ElementSize-1] == 1 {
			*v = "true"
		} else if data[ElementSize-1] == 0 {
			*v = "false"
		} else {
			return 0, fmt.Errorf("unexpected value for EVM bool")
		}
	case *int8:
		*v = int8(data[ElementSize-1])
	case *int16:
		*v = int16(data[ElementSize-1])
	case *int32:
		*v = int32(data[ElementSize-1])
	case *int64:
		*v = int64(data[ElementSize-1])
	case *int:
		*v = int(data[ElementSize-1])
	case *uint8:
		*v = uint8(data[ElementSize-1])
	case *uint16:
		*v = uint16(data[ElementSize-1])
	case *uint32:
		*v = uint32(data[ElementSize-1])
	case *uint64:
		*v = uint64(data[ElementSize-1])
	case *uint:
		*v = uint(data[ElementSize-1])
	case *bool:
		*v = data[ElementSize-1] == 1
	default:
		return 0, fmt.Errorf("cannot set type %s for EVM bool", reflect.ValueOf(v).Kind().String())
	}
	return 32, nil
}

func (e EVMBool) Dynamic() bool {
	return false
}

func (e EVMBool) ImplicitCast(o EVMType) bool {
	return false
}

var _ EVMType = (*EVMUint)(nil)

type EVMUint struct {
	M uint64
}

func (e EVMUint) GetSignature() string {
	return fmt.Sprintf("uint%d", e.M)
}

func (e EVMUint) getGoType() interface{} {
	switch e.M {
	case 8:
		return new(uint8)
	case 16:
		return new(uint16)
	case 32:
		return new(uint32)
	case 64:
		return new(uint64)
	default:
		return new(big.Int)
	}
}

func (e EVMUint) pack(v interface{}) ([]byte, error) {
	n := new(big.Int)

	arg := reflect.ValueOf(v)
	switch arg.Kind() {
	case reflect.String:
		_, ok := n.SetString(arg.String(), 0)
		if !ok {
			return nil, fmt.Errorf("Failed to parse `%s", arg.String())
		}
		if n.Sign() < 0 {
			return nil, fmt.Errorf("negative value not allowed for uint%d", e.M)
		}
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		fallthrough
	case reflect.Uint:
		n.SetUint64(arg.Uint())
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		fallthrough
	case reflect.Int:
		x := arg.Int()
		if x < 0 {
			return nil, fmt.Errorf("negative value not allowed for uint%d", e.M)
		}
		n.SetInt64(x)
	default:
		t := reflect.TypeOf(new(uint64))
		if reflect.TypeOf(v).ConvertibleTo(t) {
			n.SetUint64(reflect.ValueOf(v).Convert(t).Uint())
		} else {
			return nil, fmt.Errorf("cannot convert type %s to uint%d", arg.Kind().String(), e.M)
		}
	}

	b := n.Bytes()
	if uint64(len(b)) > e.M {
		return nil, fmt.Errorf("value to large for int%d", e.M)
	}
	return pad(b, ElementSize, true), nil
}

func (e EVMUint) unpack(data []byte, offset int, v interface{}) (int, error) {
	if len(data)-offset < ElementSize {
		return 0, fmt.Errorf("%v: not enough data", e)
	}

	data = data[offset:]
	empty := 0
	for empty = 0; empty < ElementSize; empty++ {
		if data[empty] != 0 {
			break
		}
	}

	length := ElementSize - empty

	switch v := v.(type) {
	case *string:
		b := new(big.Int)
		b.SetBytes(data[empty:ElementSize])
		*v = b.String()
	case *big.Int:
		b := new(big.Int)
		*v = *b.SetBytes(data[0:ElementSize])
	case *uint64:
		maxLen := int(unsafe.Sizeof(*v))
		if length > maxLen {
			return 0, fmt.Errorf("value to large for uint64")
		}
		*v = binary.BigEndian.Uint64(data[ElementSize-maxLen : ElementSize])
	case *uint32:
		maxLen := int(unsafe.Sizeof(*v))
		if length > maxLen {
			return 0, fmt.Errorf("value to large for uint64")
		}
		*v = binary.BigEndian.Uint32(data[ElementSize-maxLen : ElementSize])
	case *uint16:
		maxLen := int(unsafe.Sizeof(*v))
		if length > maxLen {
			return 0, fmt.Errorf("value to large for uint16")
		}
		*v = binary.BigEndian.Uint16(data[ElementSize-maxLen : ElementSize])
	case *uint8:
		maxLen := 1
		if length > maxLen {
			return 0, fmt.Errorf("value to large for uint8")
		}
		*v = uint8(data[31])
	case *int64:
		maxLen := int(unsafe.Sizeof(*v))
		if length > maxLen || (data[ElementSize-maxLen]&0x80) != 0 {
			return 0, fmt.Errorf("value to large for int64")
		}
		*v = int64(binary.BigEndian.Uint64(data[ElementSize-maxLen : ElementSize]))
	case *int32:
		maxLen := int(unsafe.Sizeof(*v))
		if length > maxLen || (data[ElementSize-maxLen]&0x80) != 0 {
			return 0, fmt.Errorf("value to large for int64")
		}
		*v = int32(binary.BigEndian.Uint32(data[ElementSize-maxLen : ElementSize]))
	case *int16:
		maxLen := int(unsafe.Sizeof(*v))
		if length > maxLen || (data[ElementSize-maxLen]&0x80) != 0 {
			return 0, fmt.Errorf("value to large for int16")
		}
		*v = int16(binary.BigEndian.Uint16(data[ElementSize-maxLen : ElementSize]))
	case *int8:
		maxLen := 1
		if length > maxLen || (data[ElementSize-maxLen]&0x80) != 0 {
			return 0, fmt.Errorf("value to large for int8")
		}
		*v = int8(data[ElementSize-1])
	default:
		return 0, fmt.Errorf("unable to convert %s to %s", e.GetSignature(), reflect.ValueOf(v).Kind().String())
	}

	return 32, nil
}

func (e EVMUint) Dynamic() bool {
	return false
}

func (e EVMUint) String() string {
	return fmt.Sprintf("EVMUInt{%v}", e.M)
}

var _ EVMType = (*EVMInt)(nil)

type EVMInt struct {
	M uint64
}

func (e EVMInt) String() string {
	return fmt.Sprintf("EVMInt{%v}", e.M)
}

func (e EVMInt) getGoType() interface{} {
	switch e.M {
	case 8:
		return new(int8)
	case 16:
		return new(int16)
	case 32:
		return new(int32)
	case 64:
		return new(int64)
	default:
		return new(big.Int)
	}
}

func (e EVMInt) ImplicitCast(o EVMType) bool {
	i, ok := o.(EVMInt)
	return ok && i.M >= e.M
}

func (e EVMInt) GetSignature() string {
	return fmt.Sprintf("int%d", e.M)
}

func (e EVMInt) pack(v interface{}) ([]byte, error) {
	n := new(big.Int)

	switch arg := v.(type) {
	case *big.Int:
		n.Set(arg)
	case string:
		_, ok := n.SetString(arg, 0)
		if !ok {
			return nil, fmt.Errorf("failed to parse `%s", arg)
		}
	case uint:
		n.SetUint64(uint64(arg))
	case uint8:
		n.SetUint64(uint64(arg))
	case uint16:
		n.SetUint64(uint64(arg))
	case uint32:
		n.SetUint64(uint64(arg))
	case uint64:
		n.SetUint64(arg)
	case int:
		n.SetInt64(int64(arg))
	case int8:
		n.SetInt64(int64(arg))
	case int16:
		n.SetInt64(int64(arg))
	case int32:
		n.SetInt64(int64(arg))
	case int64:
		n.SetInt64(arg)
	default:
		t := reflect.TypeOf(new(int64))
		if reflect.TypeOf(v).ConvertibleTo(t) {
			n.SetInt64(reflect.ValueOf(v).Convert(t).Int())
		} else {
			return nil, fmt.Errorf("cannot convert type %v to int%d", v, e.M)
		}
	}

	b := n.Bytes()
	if uint64(len(b)) > e.M {
		return nil, fmt.Errorf("value to large for int%d", e.M)
	}
	res := pad(b, ElementSize, true)
	if (res[0] & 0x80) != 0 {
		return nil, fmt.Errorf("value to large for int%d", e.M)
	}
	if n.Sign() < 0 {
		// One's complement; i.e. 0xffff is -1, not 0.
		n.Add(n, big.NewInt(1))
		b := n.Bytes()
		res = pad(b, ElementSize, true)
		for i := 0; i < len(res); i++ {
			res[i] = ^res[i]
		}
	}
	return res, nil
}

func (e EVMInt) unpack(data []byte, offset int, v interface{}) (int, error) {
	if len(data)-offset < ElementSize {
		return 0, fmt.Errorf("%v: not enough data", e)
	}

	data = data[offset:]
	sign := (data[0] & 0x80) != 0

	empty := 0
	for empty = 0; empty < ElementSize; empty++ {
		if (sign && data[empty] != 255) || (!sign && data[empty] != 0) {
			break
		}
	}

	length := ElementSize - empty
	inv := make([]byte, ElementSize)
	for i := 0; i < ElementSize; i++ {
		if sign {
			inv[i] = ^data[i]
		} else {
			inv[i] = data[i]
		}
	}

	switch v := v.(type) {
	case **big.Int:
		b := new(big.Int).SetBytes(inv[empty:ElementSize])
		if sign {
			*v = b.Sub(big.NewInt(-1), b)
		} else {
			*v = b
		}
	case *string:
		b := new(big.Int)
		b.SetBytes(inv[empty:ElementSize])
		if sign {
			*v = b.Sub(big.NewInt(-1), b).String()
		} else {
			*v = b.String()
		}
	case *big.Int:
		b := new(big.Int)
		b.SetBytes(inv[0:ElementSize])
		if sign {
			*v = *b.Sub(big.NewInt(-1), b)
		} else {
			*v = *b
		}
	case *uint64:
		if sign {
			return 0, fmt.Errorf("cannot convert negative EVM int to %T", *v)
		}
		maxLen := int(unsafe.Sizeof(*v))
		if length > maxLen {
			return 0, fmt.Errorf("value to large for %T", *v)
		}
		*v = binary.BigEndian.Uint64(data[ElementSize-maxLen : ElementSize])
	case *uint32:
		if sign {
			return 0, fmt.Errorf("cannot convert negative EVM int to %T", *v)
		}
		maxLen := int(unsafe.Sizeof(*v))
		if length > maxLen {
			return 0, fmt.Errorf("value to large for %T", *v)
		}
		*v = binary.BigEndian.Uint32(data[ElementSize-maxLen : ElementSize])
	case *uint16:
		if sign {
			return 0, fmt.Errorf("cannot convert negative EVM int to %T", *v)
		}
		maxLen := int(unsafe.Sizeof(*v))
		if length > maxLen {
			return 0, fmt.Errorf("value to large for %T", *v)
		}
		*v = binary.BigEndian.Uint16(data[ElementSize-maxLen : ElementSize])
	case *uint8:
		if sign {
			return 0, fmt.Errorf("cannot convert negative EVM int to %T", *v)
		}
		if length > 1 {
			return 0, fmt.Errorf("value to large for %T", *v)
		}
		*v = data[ElementSize-1]
	case *int64:
		maxLen := int(unsafe.Sizeof(*v))
		if length > maxLen || (inv[ElementSize-maxLen]&0x80) != 0 {
			return 0, fmt.Errorf("value to large for %T", *v)
		}
		*v = int64(binary.BigEndian.Uint64(data[ElementSize-maxLen : ElementSize]))
	case *int32:
		maxLen := int(unsafe.Sizeof(*v))
		if length > maxLen || (inv[ElementSize-maxLen]&0x80) != 0 {
			return 0, fmt.Errorf("value to large for %T", *v)
		}
		*v = int32(binary.BigEndian.Uint32(data[ElementSize-maxLen : ElementSize]))
	case *int16:
		maxLen := int(unsafe.Sizeof(*v))
		if length > maxLen || (inv[ElementSize-maxLen]&0x80) != 0 {
			return 0, fmt.Errorf("value to large for %T", *v)
		}
		*v = int16(binary.BigEndian.Uint16(data[ElementSize-maxLen : ElementSize]))
	case *int8:
		if length > 1 || (inv[ElementSize-1]&0x80) != 0 {
			return 0, fmt.Errorf("value to large for %T", *v)
		}
		*v = int8(data[ElementSize-1])
	default:
		return 0, fmt.Errorf("unable to convert %s to %T", e.GetSignature(), v)
	}

	return ElementSize, nil
}

func (e EVMInt) Dynamic() bool {
	return false
}

func (e EVMUint) ImplicitCast(o EVMType) bool {
	u, ok := o.(EVMUint)
	return ok && u.M >= e.M
}

var _ EVMType = (*EVMAddress)(nil)

type EVMAddress struct {
}

func (e EVMAddress) String() string {
	return "EVMAddress"
}

func (e EVMAddress) getGoType() interface{} {
	return new(crypto.Address)
}

func (e EVMAddress) GetSignature() string {
	return "address"
}

func (e EVMAddress) pack(v interface{}) ([]byte, error) {
	var bs []byte
	switch a := v.(type) {
	case crypto.Address:
		bs = a[:]
	case *crypto.Address:
		bs = (*a)[:]
	case string:
		address, err := crypto.AddressFromHexString(a)
		if err != nil {
			return nil, fmt.Errorf("could not convert '%s' to address: %v", a, err)
		}
		bs = address[:]
	case []byte:
		address, err := crypto.AddressFromBytes(a)
		if err != nil {
			return nil, fmt.Errorf("could not convert byte 0x%X to address: %v", a, err)
		}
		bs = address[:]
	default:
		return nil, fmt.Errorf("cannot map from %s to EVM address", reflect.ValueOf(v).Kind().String())
	}
	return pad(bs, ElementSize, true), nil
}

func (e EVMAddress) unpack(data []byte, offset int, v interface{}) (int, error) {
	if len(data)-offset < ElementSize {
		return 0, fmt.Errorf("%v: not enough data", e)
	}
	addr, err := crypto.AddressFromBytes(data[offset+ElementSize-crypto.AddressLength : offset+ElementSize])
	if err != nil {
		return 0, err
	}
	switch v := v.(type) {
	case *string:
		*v = addr.String()
	case *crypto.Address:
		*v = addr
	case *([]byte):
		*v = data[offset+ElementSize-crypto.AddressLength : offset+ElementSize]
	default:
		return 0, fmt.Errorf("cannot map EVM address to %s", reflect.ValueOf(v).Kind().String())
	}

	return ElementSize, nil
}

func (e EVMAddress) Dynamic() bool {
	return false
}

func (e EVMAddress) ImplicitCast(o EVMType) bool {
	return false
}

var _ EVMType = (*EVMBytes)(nil)

type EVMBytes struct {
	M uint64
}

func (e EVMBytes) String() string {
	if e.M == 0 {
		return "EVMBytes"
	}
	return fmt.Sprintf("EVMBytes[%v]", e.M)
}

func (e EVMBytes) getGoType() interface{} {
	v := make([]byte, e.M)
	return &v
}

func (e EVMBytes) pack(v interface{}) ([]byte, error) {
	b, ok := v.([]byte)
	if !ok {
		s, ok := v.(string)
		if ok {
			b = []byte(s)
		} else {
			return nil, fmt.Errorf("cannot map from %s to EVM bytes", reflect.ValueOf(v).Kind().String())
		}
	}

	if e.M > 0 {
		if uint64(len(b)) > e.M {
			return nil, fmt.Errorf("[%d]byte to long for %s", len(b), e.GetSignature())
		}
		return pad(b, ElementSize, false), nil
	} else {
		length := EVMUint{M: 256}
		p, err := length.pack(len(b))
		if err != nil {
			return nil, err
		}
		for i := 0; i < len(b); i += ElementSize {
			a := b[i:]
			if len(a) == 0 {
				break
			}
			p = append(p, pad(a, ElementSize, false)...)
		}

		return p, nil
	}
}

func (e EVMBytes) unpack(data []byte, offset int, v interface{}) (int, error) {
	if len(data)-offset < ElementSize {
		return 0, fmt.Errorf("%v: not enough data", e)
	}
	if e.M == 0 {
		s := EVMString{}

		return s.unpack(data, offset, v)
	}

	v2 := reflect.ValueOf(v).Elem()
	switch v2.Type().Kind() {
	case reflect.String:
		start := 0
		end := int(e.M)

		for start < ElementSize-1 && data[offset+start] == 0 && start < end {
			start++
		}
		for end > start && data[offset+end-1] == 0 {
			end--
		}
		v2.SetString(string(data[offset+start : offset+end]))
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		v2.SetBytes(data[offset : offset+int(e.M)])
	default:
		return 0, fmt.Errorf("cannot map EVM %s to %v", e.GetSignature(), reflect.ValueOf(v).Kind())
	}

	return ElementSize, nil
}

func (e EVMBytes) Dynamic() bool {
	return e.M == 0
}

func (e EVMBytes) GetSignature() string {
	if e.M > 0 {
		return fmt.Sprintf("bytes%d", e.M)
	} else {
		return "bytes"
	}
}

func (e EVMBytes) ImplicitCast(o EVMType) bool {
	return false
}

var _ EVMType = (*EVMString)(nil)

type EVMString struct {
}

func (e EVMString) String() string {
	return "EVMString"
}

func (e EVMString) GetSignature() string {
	return "string"
}

func (e EVMString) getGoType() interface{} {
	return new(string)
}

func (e EVMString) pack(v interface{}) ([]byte, error) {
	b := EVMBytes{M: 0}

	return b.pack(v)
}

func (e EVMString) unpack(data []byte, offset int, v interface{}) (int, error) {
	lenType := EVMInt{M: 64}
	var length int64
	l, err := lenType.unpack(data, offset, &length)
	if err != nil {
		return 0, fmt.Errorf("could not unpack string length prefix: %v", err)
	}
	offset += l

	switch v := v.(type) {
	case *string:
		*v = string(data[offset : offset+int(length)])
	case *[]byte:
		*v = data[offset : offset+int(length)]
	default:
		return 0, fmt.Errorf("cannot map EVM string to %s", reflect.ValueOf(v).Kind().String())
	}

	return ElementSize, nil
}

func (e EVMString) Dynamic() bool {
	return true
}

func (e EVMString) ImplicitCast(o EVMType) bool {
	return false
}

var _ EVMType = (*EVMFixed)(nil)

type EVMFixed struct {
	N, M   uint64
	signed bool
}

func (e EVMFixed) getGoType() interface{} {
	// This is not right, obviously
	return new(big.Float)
}

func (e EVMFixed) GetSignature() string {
	if e.signed {
		return fmt.Sprintf("fixed%dx%d", e.M, e.N)
	} else {
		return fmt.Sprintf("ufixed%dx%d", e.M, e.N)
	}
}

func (e EVMFixed) pack(v interface{}) ([]byte, error) {
	// The ABI spec does not describe how this should be packed; go-ethereum abi does not implement this
	// need to dig in solidity to find out how this is packed
	return nil, fmt.Errorf("packing of %s not implemented, patches welcome", e.GetSignature())
}

func (e EVMFixed) unpack(data []byte, offset int, v interface{}) (int, error) {
	// The ABI spec does not describe how this should be packed; go-ethereum abi does not implement this
	// need to dig in solidity to find out how this is packed
	return 0, fmt.Errorf("unpacking of %s not implemented, patches welcome", e.GetSignature())
}

func (e EVMFixed) Dynamic() bool {
	return false
}

func (e EVMFixed) ImplicitCast(o EVMType) bool {
	return false
}

// quick helper padding
func pad(input []byte, size int, left bool) []byte {
	if len(input) >= size {
		return input[:size]
	}
	padded := make([]byte, size)
	if left {
		copy(padded[size-len(input):], input)
	} else {
		copy(padded, input)
	}
	return padded
}

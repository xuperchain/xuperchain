package rlp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/bits"
	"reflect"
)

const (
	EmptyString = 0x80
	EmptySlice  = 0xC0
)

type Code uint32

const (
	ErrUnknown Code = iota
	ErrNoInput
	ErrInvalid
)

func (c Code) Error() string {
	switch c {
	case ErrNoInput:
		return "no input"
	case ErrInvalid:
		return "input not valid RLP encoding"
	default:
		return "unknown error"
	}
}

func encodeUint8(input uint8) ([]byte, error) {
	if input == 0 {
		return []byte{EmptyString}, nil
	} else if input >= 0x00 && input <= 0x7f {
		return []byte{input}, nil
	} else if input >= 0x80 && input <= 0xff {
		return []byte{0x81, input}, nil
	}
	return []byte{EmptyString}, nil
}

func encodeUint64(i uint64) ([]byte, error) {
	size := bits.Len64(i)/8 + 1
	if size == 1 {
		return encodeUint8(uint8(i))
	}
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(i))
	return encodeString(b[8-size:])
}

func encodeLength(n, offset int) []byte {
	if n <= 55 {
		return []uint8{uint8(n + offset)}
	}

	i := uint64(n)
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, i)
	size := bits.Len64(i)/8 + 1
	return append([]byte{uint8(0xb7 + len(string(n)))}, b[8-size:]...)
}

func encodeString(input []byte) ([]byte, error) {
	if len(input) == 0 {
		return []byte{EmptyString}, nil
	} else if len(input) == 1 {
		return encodeUint8(input[0])
	} else {
		return append(encodeLength(len(input), EmptyString), []byte(input)...), nil
	}
}

func encodeList(val reflect.Value) ([]byte, error) {
	if val.Len() == 0 {
		return []byte{EmptySlice}, nil
	}

	out := make([][]byte, 0)
	for i := 0; i < val.Len(); i++ {
		data, err := encode(val.Index(i).Interface())
		if err != nil {
			return nil, err
		}
		out = append(out, data)
	}

	sum := bytes.Join(out, []byte{})
	return append(encodeLength(len(sum), EmptySlice), sum...), nil
}

func encodeStruct(val reflect.Value) ([]byte, error) {
	out := make([][]byte, 0)

	for i := 0; i < val.NumField(); i++ {
		data, err := encode(val.Field(i).Interface())
		if err != nil {
			return nil, err
		}
		out = append(out, data)
	}
	sum := bytes.Join(out, []byte{})
	return append(encodeLength(len(sum), EmptySlice), sum...), nil
}

func encode(input interface{}) ([]byte, error) {
	val := reflect.ValueOf(input)
	typ := reflect.TypeOf(input)

	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i := val.Int()
		if i < 0 {
			return nil, fmt.Errorf("cannot rlp encode negative integer")
		}
		return encodeUint64(uint64(i))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return encodeUint64(val.Uint())
	case reflect.Bool:
		if val.Bool() {
			return []byte{0x01}, nil
		}
		return []byte{EmptyString}, nil
	case reflect.String:
		return encodeString([]byte(reflect.ValueOf(input).String()))
	case reflect.Slice:
		switch typ.Elem().Kind() {
		case reflect.Uint8:
			return encodeString(reflect.ValueOf(input).Bytes())
		default:
			return encodeList(val)
		}
	case reflect.Struct:
		return encodeStruct(val)
	default:
		return []byte{EmptyString}, nil
	}
}

func Encode(input interface{}) ([]byte, error) {
	return encode(input)
}

type fields struct {
	fields [][]byte
}

func (f *fields) add(element []byte) {
	f.fields = append(f.fields, element)
}

func decode(in []byte, out *fields) error {
	if len(in) == 0 {
		return nil
	}

	offset, length, typ, err := decodeLength(in)
	if err != nil {
		return err
	}

	switch typ {
	case reflect.String:
		out.add(in[offset:length])
	case reflect.Slice:
		err = decode(in[offset:length], out)
		if err != nil {
			return err
		}
	}

	return decode(in[length:], out)
}

func decodeLength(input []byte) (uint8, uint8, reflect.Kind, error) {
	length := len(input)

	if length == 0 {
		return 0, 0, reflect.Invalid, ErrNoInput
	}

	prefix := input[0]

	if prefix <= 0x7f {
		// single byte
		return 0, 1, reflect.String, nil

	} else if length > int(prefix-0x80) && prefix <= 0xb7 {
		// short string
		strLen := prefix - 0x80
		if strLen == 1 && uint8(input[1]) <= 0x7f {
			return 0, 0, reflect.Invalid, fmt.Errorf("single byte below 128 must be encoded as itself")
		}
		return 1, strLen + 1, reflect.String, nil

	} else if length > int(prefix-0xb7) && prefix <= 0xbf {
		// long string
		next, err := getLength(input[1 : (prefix-0xb7)+1])
		if err != nil {
			return 0, 0, reflect.Invalid, err
		} else if length > int(prefix-0xb7+next) {
			lenOfStrLen := prefix - 0xb7
			if input[1] == 0 {
				return 0, 0, reflect.Invalid, fmt.Errorf("multi-byte length must have no leading zero")
			}
			strLen, err := getLength(input[1 : lenOfStrLen+1])
			if err != nil {
				return 0, 0, reflect.Invalid, err
			} else if strLen < 56 {
				return 0, 0, reflect.Invalid, fmt.Errorf("length below 56 must be encoded in one byte")
			}
			return lenOfStrLen + 1, lenOfStrLen + strLen, reflect.String, nil
		}

	} else if length > int(prefix-0xc0) && prefix <= 0xf7 {
		// short list
		lenOfList := prefix - 0xc0
		return 1, lenOfList + 1, reflect.Slice, nil

	} else if prefix <= 0xff && length > int(prefix-0xf7) {
		// long list
		lenOfListLen := (prefix - 0xf7) + 1
		next, err := getLength(input[1:lenOfListLen])
		if err != nil {
			return 0, 0, reflect.Invalid, err
		} else if length > int(prefix-0xf7+next) {
			if input[1] == 0 {
				return 0, 0, reflect.Invalid, fmt.Errorf("multi-byte length must have no leading zero")
			}
			listLen, err := getLength(input[1:lenOfListLen])
			if err != nil {
				return 0, 0, reflect.Invalid, err
			} else if listLen < 56 {
				return 0, 0, reflect.Invalid, fmt.Errorf("length below 56 must be encoded in one byte")
			}
			return lenOfListLen, lenOfListLen + listLen, reflect.Slice, nil
		}
	}

	return 0, 0, reflect.Invalid, ErrInvalid
}

func getLength(data []byte) (uint8, error) {
	length := len(data)
	if length == 0 {
		return 0, ErrNoInput
	} else if length == 1 {
		return uint8(data[0]), nil
	} else {
		next, err := getLength(data[0 : len(data)-1])
		return uint8(data[len(data)-1]) + next, err
	}
}

func decodeStruct(in reflect.Value, fields [][]byte) error {
	if in.NumField() != len(fields) {
		return fmt.Errorf("wrong number of fields; have %d, want %d", len(fields), in.NumField())
	}
	for i := 0; i < in.NumField(); i++ {
		val := in.Field(i)
		typ := in.Field(i).Type()
		switch val.Kind() {
		case reflect.String:
			val.SetString(string(fields[i]))
		case reflect.Uint64:
			out := make([]byte, 8)
			for j := range fields[i] {
				out[len(out)-(len(fields[i])-j)] = fields[i][j]
			}
			val.SetUint(binary.BigEndian.Uint64(out))
		case reflect.Slice:
			if typ.Elem().Kind() != reflect.Uint8 {
				continue
			}
			out := make([]byte, len(fields[i]))
			for i, b := range fields[i] {
				out[i] = b
			}
			val.SetBytes(out)
		}
	}
	return nil
}

func Decode(src []byte, dst interface{}) error {
	dec := new(fields)
	err := decode(src, dec)
	if err != nil {
		return err
	}

	val := reflect.ValueOf(dst)
	typ := reflect.TypeOf(dst)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Slice:
		switch typ.Elem().Kind() {
		case reflect.Uint8:
			out, ok := dst.([]byte)
			if !ok {
				return fmt.Errorf("cannot decode into type %s", val.Type())
			}
			found := bytes.Join(dec.fields, []byte(""))
			if len(out) < len(found) {
				return fmt.Errorf("cannot decode %d bytes into slice of size %d", len(found), len(out))
			}
			for i, b := range found {
				out[i] = b
			}
			return nil
		case reflect.Slice:
			out, ok := dst.([][]byte)
			if !ok {
				return fmt.Errorf("cannot decode into type %s", val.Type())
			}
			for i := range out {
				out[i] = dec.fields[i]
			}
			return nil
		}
	case reflect.Struct:
		return decodeStruct(val, dec.fields)
	}

	return fmt.Errorf("cannot decode into unsupported type %v", reflect.TypeOf(dst))
}

package abi

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/hyperledger/burrow/binary"
)

func Pack(argSpec []Argument, args ...interface{}) ([]byte, error) {
	getArg, err := argGetter(argSpec, args, false)
	if err != nil {
		return nil, err
	}
	return pack(argSpec, getArg)
}

func Unpack(argSpec []Argument, data []byte, args ...interface{}) error {
	getArg, err := argGetter(argSpec, args, true)
	if err != nil {
		return err
	}
	return unpack(argSpec, data, getArg)
}

func PackEvent(eventSpec *EventSpec, args ...interface{}) ([]binary.Word256, []byte, error) {
	getArg, err := argGetter(eventSpec.Inputs, args, false)
	if err != nil {
		return nil, nil, err
	}
	data, err := pack(eventSpec.Inputs, getArg)
	if err != nil {
		return nil, nil, err
	}
	topics, err := packTopics(eventSpec, getArg)
	return topics, data, err
}

func UnpackEvent(eventSpec *EventSpec, topics []binary.Word256, data []byte, args ...interface{}) error {
	getArg, err := argGetter(eventSpec.Inputs, args, true)
	if err != nil {
		return err
	}
	err = unpack(eventSpec.Inputs, data, getArg)
	if err != nil {
		return err
	}
	return unpackTopics(eventSpec, topics, getArg)
}

// UnpackRevert decodes the revert reason if a contract called revert. If no
// reason was given, message will be nil else it will point to the string
func UnpackRevert(data []byte) (message *string, err error) {
	if len(data) > 0 {
		var msg string
		err = revertAbi.UnpackWithID(data, &msg)
		message = &msg
	}
	return
}

// revertAbi exists to decode reverts. Any contract function call fail using revert(), assert() or require().
// If a function exits this way, the this hardcoded ABI will be used.
var revertAbi *Spec

func init() {
	var err error
	revertAbi, err = ReadSpec([]byte(`[{"name":"Error","type":"function","outputs":[{"type":"string"}],"inputs":[{"type":"string"}]}]`))
	if err != nil {
		panic(fmt.Sprintf("internal error: failed to build revert abi: %v", err))
	}
}

func argGetter(argSpec []Argument, args []interface{}, ptr bool) (func(int) interface{}, error) {
	if len(args) == 1 {
		rt := reflect.TypeOf(args[0])
		if rt.String() == "*big.Int"{
			return func(i int) interface{} { return args[i] }, nil
		}
		rv := reflect.ValueOf(args[0])
		if rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		} else if ptr {
			return nil, fmt.Errorf("struct pointer required in order to set values, but got %v", rv.Kind())
		}
		if rv.Kind() != reflect.Struct {
			if len(args) == 1 {
				// Treat s single arg
				return func(i int) interface{} { return args[i] }, nil
			}
			return nil, fmt.Errorf("expected single argument to be struct but got %v", rv.Kind())
		}
		fields := rv.NumField()
		if fields != len(argSpec) {
			return nil, fmt.Errorf("%d arguments in struct expected, %d received", len(argSpec), fields)
		}
		if ptr {
			return func(i int) interface{} {
				return rv.Field(i).Addr().Interface()
			}, nil
		}
		return func(i int) interface{} {
			return rv.Field(i).Interface()
		}, nil
	}
	if len(args) == len(argSpec) {
		return func(i int) interface{} {
			return args[i]
		}, nil
	}
	return nil, fmt.Errorf("%d arguments expected, %d received", len(argSpec), len(args))
}

func packTopics(eventSpec *EventSpec, getArg func(int) interface{}) ([]binary.Word256, error) {
	topics := make([]binary.Word256, 0, 5)
	if !eventSpec.Anonymous {
		topics = append(topics, binary.Word256(eventSpec.ID))
	}
	for i, a := range eventSpec.Inputs {
		if a.Indexed {
			data, err := a.EVM.pack(getArg(i))
			if err != nil {
				return nil, err
			}
			var topic binary.Word256
			copy(topic[:], data)
			topics = append(topics, topic)
		}
	}
	return topics, nil
}

// Unpack event topics
func unpackTopics(eventSpec *EventSpec, topics []binary.Word256, getArg func(int) interface{}) error {
	// First unpack the topic fields
	topicIndex := 0
	if !eventSpec.Anonymous {
		topicIndex++
	}

	for i, a := range eventSpec.Inputs {
		if a.Indexed {
			_, err := a.EVM.unpack(topics[topicIndex][:], 0, getArg(i))
			if err != nil {
				return err
			}
			topicIndex++
		}
	}
	return nil
}

func pack(argSpec []Argument, getArg func(int) interface{}) ([]byte, error) {
	packed := make([]byte, 0)
	var packedDynamic []byte
	fixedSize := 0
	// Anything dynamic is stored after the "fixed" block. For the dynamic types, the fixed
	// block contains byte offsets to the data. We need to know the length of the fixed
	// block, so we can calcute the offsets
	for _, as := range argSpec {
		if as.Indexed {
			continue
		}
		if as.IsArray {
			if as.ArrayLength > 0 {
				fixedSize += ElementSize * int(as.ArrayLength)
			} else {
				fixedSize += ElementSize
			}
		} else {
			fixedSize += ElementSize
		}
	}

	addArg := func(v interface{}, a Argument) error {
		var b []byte
		var err error
		if a.EVM.Dynamic() {
			offset := EVMUint{M: 256}
			b, _ = offset.pack(fixedSize)
			d, err := a.EVM.pack(v)
			if err != nil {
				return err
			}
			fixedSize += len(d)
			packedDynamic = append(packedDynamic, d...)
		} else {
			b, err = a.EVM.pack(v)
			if err != nil {
				return err
			}
		}
		packed = append(packed, b...)
		return nil
	}

	for i, as := range argSpec {
		if as.Indexed {
			continue
		}
		a := getArg(i)
		if as.IsArray {
			s, ok := a.(string)
			if ok && s[0:1] == "[" && s[len(s)-1:] == "]" {
				a = strings.Split(s[1:len(s)-1], ",")
			}

			val := reflect.ValueOf(a)
			if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
				return nil, fmt.Errorf("argument %d should be array or slice, not %s", i, val.Kind().String())
			}

			if as.ArrayLength > 0 {
				if as.ArrayLength != uint64(val.Len()) {
					return nil, fmt.Errorf("argumment %d should be array of %d, not %d", i, as.ArrayLength, val.Len())
				}

				for n := 0; n < val.Len(); n++ {
					err := addArg(val.Index(n).Interface(), as)
					if err != nil {
						return nil, err
					}
				}
			} else {
				// dynamic array
				offset := EVMUint{M: 256}
				b, _ := offset.pack(fixedSize)
				packed = append(packed, b...)
				fixedSize += len(b)

				// store length
				b, _ = offset.pack(val.Len())
				packedDynamic = append(packedDynamic, b...)
				for n := 0; n < val.Len(); n++ {
					d, err := as.EVM.pack(val.Index(n).Interface())
					if err != nil {
						return nil, err
					}
					packedDynamic = append(packedDynamic, d...)
				}
			}
		} else {
			err := addArg(a, as)
			if err != nil {
				return nil, err
			}
		}
	}

	return append(packed, packedDynamic...), nil
}

func unpack(argSpec []Argument, data []byte, getArg func(int) interface{}) error {
	offset := 0
	offType := EVMInt{M: 64}

	getPrimitive := func(e interface{}, a Argument) error {
		if a.EVM.Dynamic() {
			var o int64
			l, err := offType.unpack(data, offset, &o)
			if err != nil {
				return err
			}
			offset += l
			_, err = a.EVM.unpack(data, int(o), e)
			if err != nil {
				return err
			}
		} else {
			l, err := a.EVM.unpack(data, offset, e)
			if err != nil {
				return err
			}
			offset += l
		}

		return nil
	}

	for i, as := range argSpec {
		if as.Indexed {
			continue
		}

		arg := getArg(i)
		if as.IsArray {
			var array *[]interface{}

			array, ok := arg.(*[]interface{})
			if !ok {
				if _, ok := arg.(*string); ok {
					// We have been asked to return the value as a string; make intermediate
					// array of strings; we will concatenate after
					intermediate := make([]interface{}, as.ArrayLength)
					for i := range intermediate {
						intermediate[i] = new(string)
					}
					array = &intermediate
				} else {
					return fmt.Errorf("argument %d should be array, slice or string", i)
				}
			}

			if as.ArrayLength > 0 {
				if int(as.ArrayLength) != len(*array) {
					return fmt.Errorf("argument %d should be array or slice of %d elements", i, as.ArrayLength)
				}

				for n := 0; n < len(*array); n++ {
					err := getPrimitive((*array)[n], as)
					if err != nil {
						return err
					}
				}
			} else {
				var o int64
				var length int64

				l, err := offType.unpack(data, offset, &o)
				if err != nil {
					return err
				}

				offset += l
				s, err := offType.unpack(data, int(o), &length)
				if err != nil {
					return err
				}
				o += int64(s)

				intermediate := make([]interface{}, length)

				if _, ok := arg.(*string); ok {
					// We have been asked to return the value as a string; make intermediate
					// array of strings; we will concatenate after
					for i := range intermediate {
						intermediate[i] = new(string)
					}
				} else {
					for i := range intermediate {
						intermediate[i] = as.EVM.getGoType()
					}
				}

				for i := 0; i < int(length); i++ {
					l, err = as.EVM.unpack(data, int(o), intermediate[i])
					if err != nil {
						return err
					}
					o += int64(l)
				}

				array = &intermediate
			}

			// If we were supposed to return a string, convert it back
			if ret, ok := arg.(*string); ok {
				s := "["
				for i, e := range *array {
					if i > 0 {
						s += ","
					}
					s += *(e.(*string))
				}
				s += "]"
				*ret = s
			}
		} else {
			err := getPrimitive(arg, as)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

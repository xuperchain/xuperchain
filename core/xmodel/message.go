package xmodel

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"

	"github.com/golang/protobuf/proto"
)

var (
	protoIface = reflect.TypeOf((*proto.Message)(nil)).Elem()
)

// MarshalMessages marshal protobuf message slice
func MarshalMessages(msgs interface{}) ([]byte, error) {
	if msgs == nil {
		return nil, nil
	}
	value := reflect.ValueOf(msgs)
	tp := value.Type()
	if tp.Kind() != reflect.Slice {
		return nil, errors.New("bad slice type")
	}
	if !tp.Elem().Implements(protoIface) {
		return nil, errors.New("elem of slice must be protobuf message")
	}
	if value.Len() == 0 {
		return nil, nil
	}

	var buf proto.Buffer
	buf.EncodeVarint(uint64(value.Len()))
	for i := 0; i < value.Len(); i++ {
		msg := value.Index(i).Interface().(proto.Message)
		err := buf.EncodeMessage(msg)
		if err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

// UnmsarshalMessages unmarshal protobuf messages to slice, x must be a pointer to message slice
func UnmsarshalMessages(p []byte, x interface{}) error {
	if p == nil {
		return nil
	}
	tp := reflect.TypeOf(x)
	// x must be a pointer to message slice
	if tp.Kind() != reflect.Ptr {
		return errors.New("must be slice ptr")
	}
	tp = tp.Elem()
	if tp.Kind() != reflect.Slice {
		return errors.New("must be slice ptr")
	}
	// element of slice must be proto.Message
	if !tp.Elem().Implements(protoIface) {
		return errors.New("elem of slice must be protobuf message")
	}
	// element of slice must be ptr type
	if tp.Elem().Kind() != reflect.Ptr {
		return errors.New("elem of slice must be ptr type")
	}
	// if tp is []*pb.TxInput then elemtp is pb.TxInput
	elemtp := tp.Elem().Elem()
	value := reflect.ValueOf(x).Elem()

	buf := proto.NewBuffer(p)
	total, err := buf.DecodeVarint()
	if err != nil {
		return fmt.Errorf("error while read message length:%s", err)
	}

	value.Set(reflect.MakeSlice(tp, int(total), int(total)))
	for i := 0; i < int(total); i++ {
		v := reflect.New(elemtp)
		m := v.Interface().(proto.Message)
		err = buf.DecodeMessage(m)
		if err != nil {
			return fmt.Errorf("error while unmsarshal message:%s", err)
		}
		value.Index(i).Set(v)
	}
	return nil
}

func isMsgEqual(reqHead, reqIncome proto.Message) bool {
	encodeHead, err := encodeMsg(reqHead)
	if err != nil {
		return false
	}
	encodeIncome, err := encodeMsg(reqIncome)
	if err != nil {
		return false
	}
	return bytes.Equal(encodeHead, encodeIncome)
}

func encodeMsg(req proto.Message) ([]byte, error) {
	var buf proto.Buffer
	buf.SetDeterministic(true)
	err := buf.EncodeMessage(req)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

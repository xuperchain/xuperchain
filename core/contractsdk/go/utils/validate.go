package utils

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"
)

var (
	typeBigInt   = reflect.TypeOf(big.NewInt(0))
	typeBigFloat = reflect.TypeOf(big.NewFloat(0))
	typeFloat    = reflect.TypeOf(0.1)
	typeInt      = reflect.TypeOf(1)
	typeString   = reflect.TypeOf("")
)

func Validate(data map[string][]byte, inStructPtr interface{}) error {
	rType := reflect.TypeOf(inStructPtr)
	rVal := reflect.ValueOf(inStructPtr)
	if rType.Kind() == reflect.Ptr {
		rType = rType.Elem()
		rVal = rVal.Elem()
	} else {
		return errors.New("inStructPtr must be ptr to struct")
	}

	for i := 0; i < rType.NumField(); i++ {
		t := rType.Field(i)
		f := rVal.Field(i)
		found := false // TODO 局部变量？？
		for k, v := range data {
			if t.Tag.Get("json") != k {
				continue
			}
			found = true
			switch t.Type {
			case typeBigFloat:
				value, succ := big.NewFloat(0).SetString(string(v))
				if !succ {
					return errors.New("errorrrrr")
				}
				f.Set(reflect.ValueOf(value))
			case typeBigInt:

				value, succ := big.NewInt(0).SetString(string(v), 10)
				if !succ {
					return errors.New("errorrrrr")
				}
				f.Set(reflect.ValueOf(value))

			case typeString:
				f.Set(reflect.ValueOf(string(v)))
			default:
				return errors.New("errorrrrr")
			}

			// TODO @fengjin
			//ltStr, ok := t.Tag.Lookup("lt")
			//gtStr, ok := t.Tag.Lookup(("gt"))
			//leStr, ok := t.Tag.Lookup("le")
			// length, ok = t.Tag.Lookup("length")
			//geStr, ok := t.Tag.Lookup("ge")
		}
		if !found && t.Tag.Get("reired") != "false" {
			if f.IsZero() {
				return fmt.Errorf("missing %s", t.Tag.Get("json"))
			}
		}
	}
	return nil
}

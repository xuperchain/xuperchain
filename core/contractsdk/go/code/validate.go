package code

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"math/big"
	"reflect"
	"strings"
)

var (
	typeBigInt = reflect.TypeOf(big.NewInt(0))
	typeString = reflect.TypeOf("")
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	validate.RegisterStructValidation(func(sl validator.StructLevel) {
		rvalue := sl.Top().Elem()
		rtype := sl.Top().Elem().Type()

		for i := 0; i < rvalue.NumField(); i++ {
			t := rtype.Field(i)
			tag := t.Tag.Get("validate")
			for _, check := range strings.Split(tag, ",") {
				// required works as default
				if strings.Contains(check, "=") {
					conditions := strings.Split(check, "=")
					condition := conditions[0]
					param := conditions[1]

					if fn, ok := bigIntFunctions[condition]; ok {
						if !fn(rvalue.Field(i).Interface().(*big.Int), param) {
							sl.ReportError(t.Name, t.Name, t.Name, condition, param)
						}
					}
				}
			}

		}

	}, big.Int{})
}

func Unmarshal(input map[string][]byte, output interface{}) error {
	rType := reflect.TypeOf(output)
	rVal := reflect.ValueOf(output)
	if rType.Kind() == reflect.Ptr {
		rType = rType.Elem()
		rVal = rVal.Elem()
	}
	for i := 0; i < rType.NumField(); i++ {
		t := rType.Field(i)
		f := rVal.Field(i)
		for k, v := range input {
			if t.Tag.Get("json") != k {
				continue
			}
			switch t.Type {
			case typeBigInt:
				value, succ := big.NewInt(0).SetString(string(v), 10)
				if !succ {
					return fmt.Errorf("failed to parse %s as int", string(v))
				}
				f.Set(reflect.ValueOf(value))
			case typeString:
				f.Set(reflect.ValueOf(string(v)))
			default:
				return fmt.Errorf("type %s of %s not supported", t.Type, t.Tag.Get("json"))
			}
		}
	}

	if err := validate.Struct(output); err != nil {
		return err
	}

	return nil
}

package utils

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"
)

var (
	typeBigInt   = reflect.TypeOf(big.NewInt(0))
	typeBigFloat = reflect.TypeOf(big.NewFloat(0))
	// typeFloat    = reflect.TypeOf(0.1)
	// typeInt      = reflect.TypeOf(1)
	typeString = reflect.TypeOf("")
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
		found := false
		for k, v := range data {
			if t.Tag.Get("json") != k {
				continue
			}
			found = true
			switch t.Type {
			case typeBigFloat:
				value, succ := big.NewFloat(0).SetString(string(v))
				if !succ {
					return fmt.Errorf("failed to parse %s as float", string(v))
				}

				if ltStr, ok := t.Tag.Lookup("lt"); ok {
					lt, succ := big.NewFloat(0).SetString(ltStr)
					if !succ {
						return fmt.Errorf("failed to parse %s as float", ltStr)
					}
					if !(value.Cmp(lt) < 0) {
						return fmt.Errorf("%s must be less than %s", t.Tag.Get("json"), ltStr)
					}
				}
				if gtStr, ok := t.Tag.Lookup("gt"); ok {
					gt, succ := big.NewFloat(0).SetString(gtStr)
					if !succ {
						return fmt.Errorf("failed to parse %s as float", gtStr)
					}
					if !(value.Cmp(gt) > 0) {
						return fmt.Errorf("%s must be greater than %s", t.Tag.Get("json"), gtStr)
					}
				}
				if leStr, ok := t.Tag.Lookup("le"); ok {
					le, succ := big.NewFloat(0).SetString(leStr)
					if !succ {
						return fmt.Errorf("failed to parse %s as float", leStr)
					}
					if !(value.Cmp(le) <= 0) {
						return fmt.Errorf("%s must be less than or equal to %s", t.Tag.Get("json"), leStr)
					}
				}
				if geStr, ok := t.Tag.Lookup("ge"); ok {
					lt, succ := big.NewFloat(0).SetString(geStr)
					if !succ {
						return fmt.Errorf("failed to parse %s as float", geStr)
					}
					if !(value.Cmp(lt) >= 0) {
						return fmt.Errorf("%s must be greater than or equal to %s", t.Tag.Get("json"), geStr)
					}
				}

				f.Set(reflect.ValueOf(value))
			case typeBigInt:
				value, succ := big.NewInt(0).SetString(string(v), 10)
				if !succ {
					return fmt.Errorf("failed to parse %s as int", string(v))
				}
				if ltStr, ok := t.Tag.Lookup("lt"); ok {
					lt, succ := big.NewInt(0).SetString(ltStr, 10)
					if !succ {
						return fmt.Errorf("failed to parse %s as int", ltStr)
					}
					if !(value.Cmp(lt) < 0) {
						return fmt.Errorf("%s must be less than %s", t.Tag.Get("json"), ltStr)
					}
				}
				if gtStr, ok := t.Tag.Lookup("gt"); ok {
					lt, succ := big.NewInt(0).SetString(gtStr, 10)
					if !succ {
						return fmt.Errorf("failed to parse %s as int", gtStr)
					}
					if !(value.Cmp(lt) > 0) {
						return fmt.Errorf("%s must be less than %s", t.Tag.Get("json"), gtStr)
					}
				}
				if leStr, ok := t.Tag.Lookup("le"); ok {
					lt, succ := big.NewInt(0).SetString(leStr, 10)
					if !succ {
						return fmt.Errorf("failed to parse %s as int", leStr)
					}
					if !(value.Cmp(lt) <= 0) {
						return fmt.Errorf("%s must be less than %s", t.Tag.Get("json"), leStr)
					}
				}
				if geStr, ok := t.Tag.Lookup("ge"); ok {
					lt, succ := big.NewInt(0).SetString(geStr, 10)
					if !succ {
						return fmt.Errorf("failed to parse %s as int", geStr)
					}
					if !(value.Cmp(lt) >= 0) {
						return fmt.Errorf("%s must be less than %s", t.Tag.Get("json"), geStr)
					}
				}
				f.Set(reflect.ValueOf(value))
			case typeString:
				if lengthStr, ok := t.Tag.Lookup("length"); ok {
					if strings.Contains(lengthStr, "-") {
						lengthArray := strings.Split((lengthStr), "-")
						if len(lengthArray) != 2 {
							return fmt.Errorf("string length formt %s is not valid", lengthStr)
						}
					} else {
						length, err := strconv.ParseInt(lengthStr, 10, 64)
						if err != nil {
							return fmt.Errorf("string length formt %s is not valid", lengthStr)
						}
						if len(v) != int(length) {
							return fmt.Errorf("%s must be a string of length %d", t.Tag.Get("json"), length)
						}
					}
				} else {
					if len(v) == 0 {
						return fmt.Errorf("%s can not be empty", t.Tag.Get("json"))
					}
				}
				f.Set(reflect.ValueOf(string(v)))
			default:
				return fmt.Errorf("type of %s not supported", t.Tag.Get("json"))
			}
		}
		if !found && t.Tag.Get("required") != "false" {
			if f.IsZero() {
				return fmt.Errorf("missing %s", t.Tag.Get("json"))
			}
		}
	}
	return nil
}

package js

import (
	"fmt"
	"math"
)

// Value is the internal representing of a js object
type Value struct {
	name  string // for debug
	value interface{}
	ref   Ref
}

// String return the string representing of a value
func (v *Value) String() string {
	return fmt.Sprintf("%s", v.value)
}

// Name is used for debugging
func (v *Value) Name() string {
	return v.name
}

// predefined values
var (
	valueUndefined = &Value{
		name:  "Undefined",
		value: "undefined",
		ref:   0,
	}
	valueNaN = &Value{
		name:  "NaN",
		value: math.NaN(),
		ref:   ValueNaN,
	}
	valueZero = &Value{
		name:  "Zero",
		value: 0,
		ref:   ValueZero,
	}
	valueNull = &Value{
		name:  "Null",
		value: (*int)(nil),
		ref:   ValueNull,
	}

	valueTrue = &Value{
		name:  "True",
		value: true,
		ref:   ValueTrue,
	}
	valueFalse = &Value{
		name:  "False",
		value: false,
		ref:   ValueFalse,
	}
)

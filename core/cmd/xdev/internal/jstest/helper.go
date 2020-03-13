package jstest

import (
	"errors"

	"github.com/robertkrimen/otto"
)

// Throw throw go error in js vm as an Exception
func Throw(err error) {
	v, _ := otto.ToValue("Exception: " + err.Error())
	panic(v)
}

// Throws throw go string in js vm as an Exception
func Throws(msg string) {
	Throw(errors.New(msg))
}

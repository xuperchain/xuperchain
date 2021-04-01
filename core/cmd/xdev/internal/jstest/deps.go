package jstest

import (
	"errors"
	"io"
	"regexp"
)

var errUnimplemented = errors.New("unimplemented")

type testDeps struct{}

func (t testDeps) MatchString(pat, str string) (bool, error) {
	reg, err := regexp.Compile(pat)
	if err != nil {
		return false, err
	}
	return reg.MatchString(str), nil
}

func (t testDeps) StartCPUProfile(w io.Writer) error           { return nil }
func (t testDeps) StopCPUProfile()                             {}
func (t testDeps) WriteProfileTo(string, io.Writer, int) error { return nil }
func (t testDeps) ImportPath() string                          { return "" }
func (t testDeps) StartTestLog(io.Writer)                      {}
func (t testDeps) StopTestLog() error                          { return nil }
func (t testDeps) SetPanicOnExit0(bool)                        {}

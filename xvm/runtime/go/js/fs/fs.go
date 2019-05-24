package fs

import (
	"os"

	"github.com/xuperchain/xuperunion/xvm/runtime/go/js"
)

// Constants is the constants in fs
type Constants struct {
	O_WRONLY int
	O_RDWR   int
	O_CREAT  int
	O_TRUNC  int
	O_APPEND int
	O_EXCL   int
}

// NewConstants instances a Constants
func NewConstants() *Constants {
	return &Constants{
		O_WRONLY: os.O_WRONLY,
		O_RDWR:   os.O_RDWR,
		O_CREAT:  os.O_CREATE,
		O_TRUNC:  os.O_TRUNC,
		O_APPEND: os.O_APPEND,
		O_EXCL:   os.O_EXCL,
	}

}

// FS represents the fs namespace in js
type FS struct {
	Constants *Constants
}

// NewFS instances the fs namespace
func NewFS() *FS {
	return &FS{
		Constants: NewConstants(),
	}
}

// GetProperty implements js.PropertyGetter interface
func (f *FS) GetProperty(name string) (interface{}, bool) {
	return func(args []interface{}) interface{} {
		return js.ExceptionNoSys
	}, true
}

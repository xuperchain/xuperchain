// +build wasm

package wasm

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/xuperchain/xuperunion/contractsdk/go/code"
)

type driver struct {
}

// New returns a wasm driver
func New() code.Driver {
	return new(driver)
}

func (d *driver) Serve(contract code.Contract) {
	ctx, err := newHandlerContext()
	if err != nil {
		fatal(err)
	}

	defer func() {
		err := recover()
		if err != nil {
			buf := make([]byte, 64<<10)
			n := runtime.Stack(buf, true)
			fatal(fmt.Sprintf("%s\n%s", err, buf[:n]))
		}
	}()

	var resp code.Response
	methodName := ctx.Method()
	contractv := reflect.ValueOf(contract)
	methodv := contractv.MethodByName(strings.Title(methodName))
	if !methodv.IsValid() {
		resp = code.Errors("bad method " + methodName)
		returnResponse(&resp)
		return
	}
	method, ok := methodv.Interface().(func(code.Context) code.Response)
	if !ok {
		resp = code.Errors("bad method type " + methodName)
		returnResponse(&resp)
		return
	}
	resp = method(ctx)
	returnResponse(&resp)
}

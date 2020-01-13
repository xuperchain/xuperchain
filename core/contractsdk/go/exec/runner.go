package exec

import (
	"reflect"
	"runtime"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/code"
)

type BridgeCallFunc func(method string, request proto.Message, response proto.Message) error

func RunContract(ctxid int64, contract code.Contract, bridgeCall BridgeCallFunc) {
	ctx := newContractContext(ctxid, bridgeCall)

	err := ctx.Init()
	if err != nil {
		resp := code.Error(err)
		ctx.SetOutput(&resp)
		return
	}

	defer func() {
		err := recover()
		if err != nil {
			buf := make([]byte, 64<<10)
			n := runtime.Stack(buf, true)
			ctx.SetOutput(&code.Response{
				Status:  code.StatusError,
				Message: string(buf[:n]),
			})
		}
	}()

	var resp code.Response
	methodName := ctx.Method()
	contractv := reflect.ValueOf(contract)
	methodv := contractv.MethodByName(strings.Title(methodName))
	if !methodv.IsValid() {
		resp = code.Errors("bad method " + methodName)
		ctx.SetOutput(&resp)
		return
	}
	method, ok := methodv.Interface().(func(code.Context) code.Response)
	if !ok {
		resp = code.Errors("bad method type " + methodName)
		ctx.SetOutput(&resp)
		return
	}
	resp = method(ctx)
	ctx.SetOutput(&resp)
}

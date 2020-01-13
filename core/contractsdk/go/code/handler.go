package code

import "encoding/json"

const (
	// StatusOK is used when contract successfully ends.
	StatusOK = 200
	// StatusErrorThreshold is the status dividing line for the normal operation of the contract
	StatusErrorThreshold = 400
	// StatusError is used when contract fails.
	StatusError = 500
)

// Response is the result of the contract run
type Response struct {
	// Status 用于反映合约的运行结果，如果status值超过CodeErrorThreshold则认为合约执行失败
	// 所有的操作将会被回滚
	Status int `json:"status"`
	// Message 用于携带一些有用的debug信息
	Message string `json:"message"`
	// Data 字段用于存储合约执行的结果
	Body []byte `json:"body"`
}

// JSON is used to assist in generating a response in which the body is in JSON format
func JSON(body interface{}) Response {
	out, _ := json.Marshal(body)
	return Response{
		Status: StatusOK,
		Body:   out,
	}
}

// OK generates a response with StatusOK and the given body
func OK(body []byte) Response {
	return Response{
		Status: StatusOK,
		Body:   body,
	}
}

// Errors generates a response with StatusError and the given message
func Errors(err string) Response {
	return Response{
		Status:  StatusError,
		Message: err,
	}
}

// Error generates a response with StatusError and uses err as message
func Error(err error) Response {
	return Response{
		Status:  StatusError,
		Message: err.Error(),
	}
}

// IsStatusError is used to determine if the given status is error
func IsStatusError(status int) bool {
	return status >= StatusErrorThreshold
}

// Driver 接口用于抽象执行合约的框架
type Driver interface {
	Serve(contract Contract)
}

// Contract is the interface of contract
type Contract interface {
	Initialize(ctx Context) Response
}

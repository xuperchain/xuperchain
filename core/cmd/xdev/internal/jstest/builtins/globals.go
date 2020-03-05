package builtins

func bytesToString(s []byte) string {
	return string(s)
}

// Globals will register to global object
var Globals = map[string]interface{}{
	"string": bytesToString,
}

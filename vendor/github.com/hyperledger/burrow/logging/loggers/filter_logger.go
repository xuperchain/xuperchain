package loggers

import (
	"github.com/go-kit/kit/log"
	"github.com/hyperledger/burrow/logging/structure"
)

// Filter logger allows us to filter lines logged to it before passing on to underlying
// output logger
// Creates a logger that removes lines from output when the predicate evaluates true
func FilterLogger(outputLogger log.Logger, predicate func(keyvals []interface{}) bool) log.Logger {
	return log.LoggerFunc(func(keyvals ...interface{}) error {
		// Always forward signals
		if structure.Signal(keyvals) != "" || !predicate(keyvals) {
			return outputLogger.Log(keyvals...)
		}
		return nil
	})
}

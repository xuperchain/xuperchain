// Copyright Monax Industries Limited
// SPDX-License-Identifier: Apache-2.0

package loggers

import (
	"github.com/go-kit/kit/log"
	"github.com/hyperledger/burrow/logging/errors"
)

// This represents an 'AND' type logger. When logged to it will log to each of
// the loggers in the slice.
type MultipleOutputLogger []log.Logger

var _ log.Logger = MultipleOutputLogger(nil)

func (mol MultipleOutputLogger) Log(keyvals ...interface{}) error {
	var errs []error
	for _, logger := range mol {
		err := logger.Log(keyvals...)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errors.CombineErrors(errs)
}

// Creates a logger that forks log messages to each of its outputLoggers
func NewMultipleOutputLogger(outputLoggers ...log.Logger) log.Logger {
	moLogger := make(MultipleOutputLogger, 0, len(outputLoggers))
	// Flatten any MultipleOutputLoggers
	for _, ol := range outputLoggers {
		if ls, ok := ol.(MultipleOutputLogger); ok {
			moLogger = append(moLogger, ls...)
		} else {
			moLogger = append(moLogger, ol)
		}
	}
	return moLogger
}

package ssdp

import "log"

// Logger is default logger for SSDP module.
var Logger *log.Logger

func logf(s string, a ...interface{}) {
	if Logger != nil {
		Logger.Printf(s, a...)
	}
}

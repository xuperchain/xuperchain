// +build wasm

package wasm

import "log"

type debugWriter struct {
}

func (w *debugWriter) Write(p []byte) (int, error) {
	print(string(p))
	return len(p), nil
}
func initDebugLog() {
	log.SetFlags(0)
	log.SetOutput(new(debugWriter))
}

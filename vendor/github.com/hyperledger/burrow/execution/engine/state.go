package engine

import (
	"github.com/hyperledger/burrow/execution/exec"
)

type State struct {
	*CallFrame
	Blockchain
	exec.EventSink
}

// wrapper for consensus-tdpos plugin
package main

import (
	"github.com/xuperchain/xuperunion/consensus/tdpos"
)

// GetInstance : implement plugin framework
func GetInstance() interface{} {
	tdposIns := tdpos.TDpos{}
	tdposIns.Init()
	return &tdposIns
}

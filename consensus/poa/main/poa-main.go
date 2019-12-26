// wrapper for consensus-poa plugin
package main

import (
	"github.com/xuperchain/xuperunion/consensus/poa"
)

// GetInstance : implement plugin framework
func GetInstance() interface{} {
	poaIns := poa.Poa{}
	poaIns.Init()
	return &poaIns
}

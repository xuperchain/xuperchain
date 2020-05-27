// wrapper for consensus-xpoa plugin
package main

import (
	"github.com/xuperchain/xuperchain/core/consensus/xpoa"
)

// GetInstance : implement plugin framework
func GetInstance() interface{} {
	return &xpoa.XPoa{}
}

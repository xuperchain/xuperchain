package xendorser

import (
	"errors"

	"github.com/xuperchain/xuperchain/core/pluginmgr"
)

const (
	// PluginName is the family name of XEndorser plugins
	PluginName = "xendorser"
)

// GetXEndorser get new instance of given XEndorser plugin
func GetXEndorser(name string) (XEndorser, error) {
	pluginMgr, err := pluginmgr.GetPluginMgr()
	if err != nil {
		return nil, errors.New("GetXEndorser: get plugin mgr failed " + err.Error())
	}

	pluginIns, err := pluginMgr.PluginMgr.CreatePluginInstance(PluginName, name)
	if err != nil {
		errmsg := "GetXEndorser: create plugin failed! name=" + name
		return nil, errors.New(errmsg)
	}

	endorserIns, ok := pluginIns.(XEndorser)
	if !ok {
		return nil, errors.New("Invalid XEndorser plugin, type error")
	}
	return endorserIns, nil
}

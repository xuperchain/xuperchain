package contract

import (
	"github.com/xuperchain/xuperchain/core/pluginmgr"

	"errors"
)

const (
	// ContractPluginName is the name of contract plugin
	ContractPluginName = "contract"
)

// CreateContractInstance create contract driver from plugin manager
func CreateContractInstance(subtype string, extParams map[string]interface{}) (ContractInterface, error) {
	// load contract plugin
	pluginMgr, err := pluginmgr.GetPluginMgr()
	if err != nil {
		return nil, errors.New("CreateContractInstance: get plugin mgr failed")
	}

	pluginIns, err := pluginMgr.PluginMgr.CreatePluginInstance(ContractPluginName, subtype)
	if err != nil {
		errmsg := "CreateContractInstance: create plugin failed! name=" + subtype
		return nil, errors.New(errmsg)
	}

	switch pluginIns.(type) {
	case ContractExtInterface:
		consIns := pluginIns.(ContractExtInterface)
		err := consIns.Init(extParams)
		if err != nil {
			errmsg := "CreateContractInstance: init failed! name=" + subtype
			return nil, errors.New(errmsg)
		}
	}

	conIns := pluginIns.(ContractInterface)

	return conIns, nil
}

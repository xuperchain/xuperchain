package factory

import (
	"errors"

	log "github.com/xuperchain/log15"
	"github.com/xuperchain/xuperchain/core/common/config"
	p2p_base "github.com/xuperchain/xuperchain/core/p2p/base"
	"github.com/xuperchain/xuperchain/core/pluginmgr"
)

const (
	// P2PCategory is the plugin module name of p2p server
	P2PCategory = "p2p"
)

// GetP2PServer create a p2p instance by plugin name
func GetP2PServer(name string, cfg config.P2PConfig, log log.Logger, extra map[string]interface{}) (p2p_base.P2PServer, error) {
	pluginMgr, err := pluginmgr.GetPluginMgr()
	if err != nil {
		return nil, errors.New("GetP2PServer: get plugin mgr failed " + err.Error())
	}

	pluginIns, err := pluginMgr.PluginMgr.CreatePluginInstance(P2PCategory, name)
	if err != nil {
		errmsg := "GetP2PServer: create plugin failed! name=" + name + ", err=" + err.Error()
		return nil, errors.New(errmsg)
	}
	server := pluginIns.(p2p_base.P2PServer)
	if err = server.Init(cfg, log, extra); err != nil {
		errmsg := "GetP2PServer: Init P2PServer failed, name=" + name + ", err=" + err.Error()
		return nil, errors.New(errmsg)
	}
	return server, err
}

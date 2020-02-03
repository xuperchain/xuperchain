package factory

import (
	"fmt"

	log "github.com/xuperchain/log15"
	"github.com/xuperchain/xuperchain/core/common/config"
	p2p_base "github.com/xuperchain/xuperchain/core/p2p/base"
	p2pv2 "github.com/xuperchain/xuperchain/core/p2p/p2pv2"
)

const (
	// P2PModuleName is the plugin module name of p2p server
	P2PModuleName = "p2p"
)

// GetP2PServer create a p2p instance by plugin name
func GetP2PServer(name string, cfg config.P2PConfig, log log.Logger, extra map[string]interface{}) (p2p_base.P2PServer, error) {
	if name == "p2pv2" {
		server := p2pv2.NewP2PServerV2()
		err := server.Init(cfg, log, extra)
		return server, err
	}
	log.Error("unknown p2p plugin name", "name", name)
	return nil, fmt.Errorf("unknown p2p plugin name")
}

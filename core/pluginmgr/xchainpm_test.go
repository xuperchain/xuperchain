package pluginmgr

import (
	"github.com/xuperchain/xuperchain/core/common/config"

	"testing"
)

func TestGetPluginMgr(t *testing.T) {
	cfg := config.NewNodeConfig()
	cfg.PluginConfPath = "./conf/plugins.conf"
	err := Init(cfg)
	if err != nil {
		t.Error("Init failed, err=", err)
	}

	xpm, err := GetPluginMgr()
	if err != nil {
		t.Error("GetPluginMgr failed, err=", err)
	}

	if xpm == nil || xpm.PluginMgr == nil {
		t.Error("GetPluginMgr return empty value, err=", err)
	}

	t.Log("PluginMgr create successfully")

	_, err = xpm.PluginMgr.CreatePluginInstance("crypto", "default")
	if err != nil {
		t.Error("CreatePluginInstance failed, err=", err)
	}

	t.Log("create plugin successfully")
}

func TestDefaultGetPluginMgr(t *testing.T) {
	_ = config.NewNodeConfig()

	xpm, err := GetPluginMgr()
	if err != nil {
		t.Error("GetPluginMgr failed, err=", err)
	}

	if xpm == nil || xpm.PluginMgr == nil {
		t.Error("GetPluginMgr return empty value, err=", err)
	}

	t.Log("PluginMgr create successfully")

	_, err = xpm.PluginMgr.CreatePluginInstance("crypto", "default")
	if err != nil {
		t.Error("CreatePluginInstance failed, err=", err)
	}

	t.Log("create plugin successfully")
}

// Package pluginmgr is a plugin manager which keeps all plugins' instance
// Notice: any plugin using this framework should implements
//         a func named 'GetInstance' to return there instance
package pluginmgr

import (
	"encoding/json"
	"errors"
	"os"
	"path"
	"path/filepath"
	"plugin"
	"sync"

	"github.com/xuperchain/xuperchain/core/common/log"
)

// PluginMgr defines the data struct of plugin manager
type PluginMgr struct {
	pluginConf map[string]map[string]confNode
	xlog       log.Logger
	rootFolder string
	sync.Mutex
}

type confNode struct {
	SubType  string `json:"subtype"`
	Path     string `json:"path"`
	Version  string `json:"version"`
	OnDemand bool   `json:"ondemand"`
}

// PluginMeta is the meta info for plugins
type PluginMeta struct {
	PluginID string `json:"pluginid"`
	Type     string `json:"type"`
	SubType  string `json:"subtype"`
	Version  string `json:"version"`
}

// public functions start from here

// CreateMgr returns instance of PluginMgr
func CreateMgr(rootFolder string, confPath string, autoloadPath string, logger log.Logger) (pm *PluginMgr, err error) {
	pm = new(PluginMgr)
	// init config struct
	pm.pluginConf = make(map[string]map[string]confNode)
	pm.xlog = logger
	pm.rootFolder = rootFolder

	// Read conf file to get all plugins
	err = pm.readPluginConfig(confPath)
	if err != nil {
		return
	}

	err = pm.autoloadPlugins(autoloadPath)
	return
}

// CreatePluginInstance always create new plugin instance
func (pm *PluginMgr) CreatePluginInstance(name string, subtype string) (pluginInstance interface{}, err error) {
	if _, ok := pm.pluginConf[name]; !ok {
		pm.xlog.Warn("Invalid plugin name", "name", name)
		return nil, errors.New("Invalid plugin name")
	}

	if _, ok := pm.pluginConf[name][subtype]; !ok {
		pm.xlog.Warn("Invalid plugin subtype", "name", name, "subtype", subtype)
		return nil, errors.New("Invalid plugin subtype")
	}
	return pm.loadOnePlugin(name, subtype)
}

// internal functions start from here
func (pm *PluginMgr) readPluginConfig(confPath string) error {
	// get config file
	confPath = path.Join(pm.rootFolder, confPath)
	f, err := os.Open(confPath)
	if err != nil {
		pm.xlog.Warn("load plugin conf file failed, cannot open file", "confPath", confPath)
		return err
	}
	defer f.Close()

	tmpConf := make(map[string][]confNode)

	// read config
	decoder := json.NewDecoder(f)
	err = decoder.Decode(&tmpConf)
	if err != nil {
		pm.xlog.Warn("load plugin conf file failed while parsing")
		return err
	}

	pluginCount := 0
	for pname, pconfs := range tmpConf {
		if pm.pluginConf[pname] == nil {
			pm.pluginConf[pname] = make(map[string]confNode)
		}

		for _, pconf := range pconfs {
			pm.pluginConf[pname][pconf.SubType] = pconf
			pluginCount++
		}
	}

	pm.xlog.Trace("Plugin conf load successfully!", "pluginCount", pluginCount)
	return nil
}

func (pm *PluginMgr) loadOnePlugin(name string, subtype string) (pi interface{}, err error) {
	// open plugin
	conf := pm.pluginConf[name][subtype]
	pluginPath := path.Join(pm.rootFolder, conf.Path)
	pg, err := plugin.Open(pluginPath)
	if err != nil {
		pm.xlog.Warn("Warn: plugin open failed!", "pluginname", name, "err", err)
		return nil, err
	}

	// get instance
	iSymbol, err := pg.Lookup("GetInstance")
	if err != nil {
		pm.xlog.Warn("Warn: plugin don't have func named GetInstance!", "pluginname", name)
		err = errors.New("Invalid plugin, it doesn't meet our requirements")
		return
	}
	pi = iSymbol.(func() interface{})()
	pm.xlog.Trace("Load a plugin successfully", "name", name, "subtype", subtype)
	// TODO: verify the plugin's signature, make sure it's authorized by us

	return
}

// auto-load plugins in given path
func (pm *PluginMgr) autoloadPlugins(autoloadPath string) error {
	pluginPath := path.Join(pm.rootFolder, autoloadPath)
	pm.xlog.Trace("start autoloadPlugins", "pluginPath", pluginPath)
	err := filepath.Walk(pluginPath, func(path string, info os.FileInfo, err error) error {
		if info == nil || info.IsDir() {
			return nil
		}
		pg, err := plugin.Open(path)
		if err != nil {
			pm.xlog.Warn("found file in autoload folder but not a plugin", "name", info.Name(), "path", path)
			return nil
		}
		getmetaFunc, err := pg.Lookup("GetMeta")
		if err != nil {
			pm.xlog.Warn("plugin not support GetMeta, ignore it", "name", info.Name(), "path", path)
			return nil
		}

		meta := &PluginMeta{}
		metaStr := getmetaFunc.(func() string)()
		err = json.Unmarshal([]byte(metaStr), meta)
		if err != nil {
			pm.xlog.Warn("plugin meta unmarshal failed, ignore it", "name", info.Name(), "path", path, "meta", metaStr)
			return nil
		}

		pm.Lock()
		defer pm.Unlock()

		pm.pluginConf[meta.Type][meta.SubType] = confNode{
			SubType: meta.SubType,
			Path:    path,
			Version: meta.Version,
		}
		pm.xlog.Warn("Found one autoload plugin", "name", info.Name(), "path", path, "meta", metaStr)

		return nil
	})

	if err != nil {
		pm.xlog.Warn("auto load plugin failed", "error", err)
		return err
	}
	return nil
}

package pluginmgr

import (
	"github.com/xuperchain/xuperchain/core/common/config"
	"github.com/xuperchain/xuperchain/core/common/log"

	"fmt"
	"os"
	"path"
	"path/filepath"
	"sync"
)

// XchainPM is the plugin manager wrapper for XuperChain
type XchainPM struct {
	PluginMgr    *PluginMgr
	xlog         log.Logger
	confPath     string
	autoloadPath string
}

var xpm *XchainPM
var locker sync.Mutex

// const for default valuess
const (
	DefaultConfPath     = "./conf/plugins.conf" // default config path
	DefaultAutoloadPath = "./plugins/autoload/" // default plugin auto-load path
)

// Init is not necessary, there is a default value
func Init(nc *config.NodeConfig) error {
	if xpm == nil {
		locker.Lock()
		defer locker.Unlock()

		if xpm == nil {
			confPath := nc.PluginConfPath
			autoloadPath := nc.PluginLoadPath
			logConfig := getDefaultLogConfig()
			return createXchainPM(getXchainRoot(), confPath, logConfig, autoloadPath)
		}
	}
	return nil
}

func createXchainPM(rootFolder string, confPath string, logConf *config.LogConfig, autoloadPath string) error {
	logger, err := log.OpenLog(logConf)
	if err != nil {
		fmt.Println("Init pluginmgr log failed!")
		return err
	}

	if confPath == "" {
		confPath = DefaultConfPath
	}
	if autoloadPath == "" {
		autoloadPath = DefaultAutoloadPath
	}

	pluginMgr, err := CreateMgr(rootFolder, confPath, autoloadPath, logger)
	if err != nil {
		logger.Warn("Init pluginmgr failed!")
		return err
	}

	xpm = &XchainPM{
		xlog:         logger,
		PluginMgr:    pluginMgr,
		confPath:     confPath,
		autoloadPath: autoloadPath,
	}

	return nil
}

func createDefaultXchianPM() error {
	pluginConf := DefaultConfPath
	autoloadPath := DefaultAutoloadPath
	logConfig := getDefaultLogConfig()

	return createXchainPM(getXchainRoot(), pluginConf, logConfig, autoloadPath)
}

func getDefaultLogConfig() *config.LogConfig {
	logFolder, err := makeFullPath("logs")
	if err != nil {
		logFolder = "./logs"
	}
	logConfig := &config.LogConfig{
		Module:         "pluginmgr",
		Filepath:       logFolder,
		Filename:       "pluginmgr",
		Fmt:            "logfmt",
		Console:        false,
		Level:          "trace",
		Async:          false,
		RotateInterval: 60 * 24, // rotate every 1 day
		RotateBackups:  7,       // keep old log files for 7 days
	}
	return logConfig
}

// GetPluginMgr return plugin manager instance
func GetPluginMgr() (*XchainPM, error) {
	// if not initialized XchainPM, use default value to init
	if xpm == nil {
		err := createDefaultXchianPM()
		return xpm, err
	}

	return xpm, nil
}

func makeFullPath(relativePath string) (string, error) {
	xchainRoot := getXchainRoot()
	if xchainRoot != "" {
		return path.Join(xchainRoot, relativePath), nil
	}

	return filepath.Abs(relativePath)
}

func getXchainRoot() string {
	return os.Getenv("XCHAIN_ROOT")
}

package cmd

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	sconf "github.com/xuperchain/xuperchain/common/config"
	"github.com/xuperchain/xuperchain/service"
	econf "github.com/xuperchain/xupercore/kernel/common/xconfig"
	"github.com/xuperchain/xupercore/kernel/engines"
	"github.com/xuperchain/xupercore/kernel/engines/xuperos/common"
	"github.com/xuperchain/xupercore/lib/logs"

	// import要使用的内核核心组件驱动
	_ "github.com/xuperchain/xupercore/bcs/consensus/pow"
	_ "github.com/xuperchain/xupercore/bcs/consensus/single"
	_ "github.com/xuperchain/xupercore/bcs/consensus/tdpos"
	_ "github.com/xuperchain/xupercore/bcs/consensus/xpoa"
	_ "github.com/xuperchain/xupercore/bcs/contract/evm"
	_ "github.com/xuperchain/xupercore/bcs/contract/native"
	_ "github.com/xuperchain/xupercore/bcs/contract/xvm"
	_ "github.com/xuperchain/xupercore/bcs/network/p2pv1"
	_ "github.com/xuperchain/xupercore/bcs/network/p2pv2"
	_ "github.com/xuperchain/xupercore/kernel/contract/kernel"
	_ "github.com/xuperchain/xupercore/kernel/contract/manager"
	_ "github.com/xuperchain/xupercore/lib/crypto/client"
	_ "github.com/xuperchain/xupercore/lib/storage/kvdb/leveldb"

	"github.com/spf13/cobra"
)

type StartupCmd struct {
	BaseCmd
}

func GetStartupCmd() *StartupCmd {
	startupCmdIns := new(StartupCmd)

	// 定义命令行参数变量
	var envCfgPath string

	startupCmdIns.Cmd = &cobra.Command{
		Use:           "startup",
		Short:         "Start up the blockchain node service.",
		Example:       "xchain startup --conf /home/rd/xchain/conf/env.yaml",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return StartupXchain(envCfgPath)
		},
	}

	// 设置命令行参数并绑定变量
	startupCmdIns.Cmd.Flags().StringVarP(&envCfgPath, "conf", "c", "",
		"engine environment config file path")

	return startupCmdIns
}

// 启动节点
func StartupXchain(envCfgPath string) error {
	// 加载基础配置
	envConf, servConf, err := loadConf(envCfgPath)
	if err != nil {
		return err
	}

	// 初始化日志
	logs.InitLog(envConf.GenConfFilePath(envConf.LogConf), envConf.GenDirAbsPath(envConf.LogDir))

	// 实例化区块链引擎
	engine, err := engines.CreateBCEngine(common.BCEngineName, envConf)
	if err != nil {
		return err
	}
	// 实例化service
	serv, err := service.NewServMG(servConf, engine)
	if err != nil {
		return err
	}

	// 启动服务和区块链引擎
	wg := &sync.WaitGroup{}
	wg.Add(2)
	engChan := runEngine(engine)
	servChan := runServ(serv)

	// 阻塞等待进程退出指令
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		// 退出调用幂等
		for {
			select {
			case <-engChan:
				wg.Done()
				serv.Exit()
			case <-servChan:
				wg.Done()
				engine.Exit()
			case <-sigChan:
				serv.Exit()
				engine.Exit()
			}
		}
	}()

	// 等待异步任务全部退出
	wg.Wait()
	return nil
}

func loadConf(envCfgPath string) (*econf.EnvConf, *sconf.ServConf, error) {
	// 加载环境配置
	envConf, err := econf.LoadEnvConf(envCfgPath)
	if err != nil {
		return nil, nil, err
	}

	// 加载服务配置
	servConf, err := sconf.LoadServConf(envConf.GenConfFilePath(envConf.ServConf))
	if err != nil {
		return nil, nil, err
	}

	return envConf, servConf, nil
}

func runEngine(engine engines.BCEngine) <-chan bool {
	exitCh := make(chan bool)

	// 启动引擎，监听退出信号
	go func() {
		engine.Run()
		exitCh <- true
	}()

	return exitCh
}

func runServ(servMG *service.ServMG) <-chan error {
	exitCh := make(chan error)

	// 启动服务，监听退出信号
	go func() {
		err := servMG.Run()
		exitCh <- err
	}()

	return exitCh
}

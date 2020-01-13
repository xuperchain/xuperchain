package native

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/docker/go-connections/sockets"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/golang/protobuf/proto"
	log "github.com/xuperchain/log15"

	"github.com/xuperchain/xuperunion/common/config"
	"github.com/xuperchain/xuperunion/contract"
	"github.com/xuperchain/xuperunion/kv/kvdb"
	ledgerlib "github.com/xuperchain/xuperunion/ledger"
	xpb "github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/pluginmgr"
)

const (
	nativeCodePrefix        = "N"
	nativeCodeHistoryPrefix = "H"
)

// New instances a new GeneralSCFramework
func New(cfg *config.NativeConfig, rootpath string, xlog log.Logger, otherPaths []string, kvEngineType string) (*GeneralSCFramework, error) {
	rootpath, _ = filepath.Abs(rootpath)
	xlog.Info("native.New", "rootpath", rootpath)
	dbpath := rootpath + "/framework"
	plgMgr, plgErr := pluginmgr.GetPluginMgr()
	if plgErr != nil {
		xlog.Warn("fail to get plugin manager")
		return nil, plgErr
	}
	var db kvdb.Database
	soInst, err := plgMgr.PluginMgr.CreatePluginInstance("kv", kvEngineType)
	if err != nil {
		xlog.Warn("fail to create plugin instance", "kvtype", kvEngineType)
		return nil, err
	}
	db = soInst.(kvdb.Database)
	err = db.Open(dbpath, map[string]interface{}{
		"cache":     ledgerlib.MemCacheSize,
		"fds":       ledgerlib.FileHandlersCacheSize,
		"dataPaths": otherPaths,
	})
	if err != nil {
		xlog.Warn("fail to open db", "db_path", dbpath)
		return nil, err
	}
	socket3Path := filepath.Join(rootpath, "chain3.sock")
	framework := &GeneralSCFramework{
		running:        true,
		cfg:            cfg,
		nativecodes:    make(map[string]*versionedStandardNativeContract),
		rootpath:       rootpath,
		driverpath:     filepath.Join(rootpath, "driver"),
		chainSock3Path: socket3Path,
		db:             db,
		versionTable:   kvdb.NewTable(db, nativeCodeHistoryPrefix),
		wg:             new(sync.WaitGroup),
		Logger:         xlog.New("module", "scf"),
	}
	err = framework.initEnv()
	if err != nil {
		return nil, err
	}

	uid, gid := os.Getuid(), os.Getgid()
	relpath, err := RelPathOfCWD(socket3Path)
	if err != nil {
		return nil, err
	}
	listener, err := sockets.NewUnixSocketWithOpts(relpath, sockets.WithChown(uid, gid), sockets.WithChmod(0660))
	if err != nil {
		framework.Error("NewUnixSocketWithOpts error", "error", err, "chainSockPath", socket3Path)
		return nil, err
	}
	framework.syscallListener = listener

	framework.register()
	return framework, nil
}

// GeneralSCFramework manage native contracts
type GeneralSCFramework struct {
	cfg            *config.NativeConfig
	mutex          sync.Mutex
	running        bool
	nativecodes    map[string]*versionedStandardNativeContract
	rootpath       string
	driverpath     string
	chainSock3Path string
	db             kvdb.Database
	context        *contract.TxContext
	//持久化版本信息
	versionTable    kvdb.Database
	wg              *sync.WaitGroup
	dockerClient    *docker.Client
	syscallListener net.Listener

	log.Logger
}

//TODO 需要进一步封装下
type versionedStandardNativeContract struct {
	//记录每个版本对应的合约
	sncMap map[string]*standardNativeContract
	//前一个版本 以及当前版本
	curVersion string
	gscf       *GeneralSCFramework
}

// getSNC 返回的是一个启动就绪的合约实例
func (vsnc *versionedStandardNativeContract) GetSNC(name, version string) (*standardNativeContract, error) {
	if snc, ok := vsnc.sncMap[version]; ok {
		vsnc.gscf.Info("getSNC from sncMap", "name", name, "version", version)
		return snc, nil
	}
	// 从历史库里面取
	descRaw, err := vsnc.gscf.versionTable.Get(makeNativeCodeNameRaw(name, version))
	vsnc.gscf.Info("getSNC from versionTable", "name", name, "version", version)
	if err != nil {
		return nil, err
	}

	var desc xpb.NativeCodeDesc
	if err := proto.Unmarshal(descRaw, &desc); err != nil {
		return nil, err
	}

	snc, err := vsnc.gscf.launchOne(&desc, statusRegistered)
	if err != nil {
		return nil, err
	}
	return snc, nil
}

// Upgrade的前提是这个升级的合约已经部署了， 也就是说已经存在于vnc
func (vsnc *versionedStandardNativeContract) upgrade(snc *standardNativeContract) error {
	version := snc.desc.GetVersion()
	if vsnc.curVersion == version {
		return nil
	}
	if _, ok := vsnc.sncMap[version]; !ok {
		//几乎不会走到的分支
		return fmt.Errorf("version is not initialized, name %s, version %s", snc.name, version)
	}
	pv := vsnc.curVersion
	vsnc.curVersion = version
	psnc := vsnc.sncMap[pv]

	//存在状态的更新，因此这里要从新更新下
	vsnc.sncMap[version] = snc

	vc := vsnc.sncMap[version]
	//必须是被激活的，才能成为更新链上的版本
	if psnc.status == statusReady {
		vc.desc.PrevVersion = pv
	}

	// 写入versionTable
	descbuf, _ := proto.Marshal(vc.desc)
	return vsnc.gscf.versionTable.Put(makeNativeCodeName(vc.desc), descbuf)
}

// deactivate 标记合约状态为注册状态，然后序列化到历史库
func (vsnc *versionedStandardNativeContract) serialize(snc *standardNativeContract) error {
	version := snc.desc.GetVersion()
	vsnc.sncMap[version] = snc
	//更新内存中snc状态
	descbuf, _ := proto.Marshal(snc.desc)
	//更新数据库中snc状态
	vsnc.gscf.Info("serialize native code", "desc", fmt.Sprintf("%#v", snc))
	return vsnc.gscf.versionTable.Put(makeNativeCodeName(snc.desc), descbuf)
}

func (gscf *GeneralSCFramework) has(desc *xpb.NativeCodeDesc) (bool, error) {
	// 从历史库里面看有没有这个版本
	return gscf.versionTable.Has(makeNativeCodeName(desc))
}

func (gscf *GeneralSCFramework) initEnv() error {
	if !gscf.cfg.Docker.Enable {
		return nil
	}
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return fmt.Errorf("Native code: init docker client:%s", err)
	}
	imageName := gscf.cfg.Docker.ImageName
	_, err = client.InspectImage(imageName)
	if err != nil {
		return fmt.Errorf("Native code: find docker image: %s", err)
	}
	gscf.dockerClient = client
	return nil
}

//为了兼容，这个表永远存储当前生效的合约
func getNativeCodeKey(name string) []byte {
	return []byte(fmt.Sprintf("%s%s", nativeCodePrefix, name))
}

func makeNativeCodeName(desc *xpb.NativeCodeDesc) []byte {
	return makeNativeCodeNameRaw(desc.GetName(), desc.GetVersion())
}

func makeNativeCodeNameRaw(name, version string) []byte {
	return []byte(fmt.Sprintf("%s_%s", name, version))
}

func (gscf *GeneralSCFramework) getVSNC(name string) *versionedStandardNativeContract {
	return gscf.nativecodes[name]
}

func (gscf *GeneralSCFramework) getSNC(name, version string) (*standardNativeContract, error) {
	vsnc := gscf.getVSNC(name)
	if vsnc == nil {
		return nil, fmt.Errorf("native code `%s` not found", name)
	}
	return vsnc.GetSNC(name, version)
}

func (gscf *GeneralSCFramework) register() error {
	//restore the nativecode
	it := gscf.db.NewIteratorWithPrefix([]byte(nativeCodePrefix))
	defer it.Release()
	for it.Next() {
		name := string(it.Key())
		name = name[1:] // trim the prefix
		var desc xpb.NativeCodeDesc
		err := proto.Unmarshal(it.Value(), &desc)
		if err != nil {
			gscf.Error("unmarshal native code desc error", "error", err)
			continue
		}
		//依然标识为注册状态
		_, err = gscf.launchOne(&desc, statusRegistered)
		if err != nil {
			gscf.Error("load plugins get plugin failed", "name", name, "error", err)
			continue
		}
		//直接激活, 存在冗余操作： 持久化desc
		gscf.activate(name, desc.GetVersion())
	}
	go gscf.monitorNativeCode()
	return nil
}

// purgeLocalNativeCode实现清理本地的旧版本数据的功能
func (gscf *GeneralSCFramework) purgeLocalNativeCode(desc *xpb.NativeCodeDesc) {
}

//Deploy 部署新版本到当前节点，并且启动合约进程
func (gscf *GeneralSCFramework) Deploy(desc *xpb.NativeCodeDesc, code []byte) error {
	gscf.mutex.Lock()
	defer gscf.mutex.Unlock()
	name := desc.GetName()
	//如果已经存在了，直接报错
	if has, err := gscf.has(desc); err != nil {
		return err
	} else if has {
		return fmt.Errorf("name %s version %s exists", desc.GetName(), desc.GetVersion())
	}
	//部署代码到指定位置
	err := gscf.deployCode(desc, code)
	if err != nil {
		return err
	}
	//启动合约
	_, err = gscf.launchOne(desc, statusRegistered)
	if err != nil {
		gscf.Error("launch native code failed", "name", name, "error", err)
		return err
	}
	gscf.Info("launch native code", "name", name)
	return nil
}

// Status returns status of all the native contracts
func (gscf *GeneralSCFramework) Status() []*xpb.NativeCodeStatus {
	gscf.mutex.Lock()
	defer gscf.mutex.Unlock()
	var result []*xpb.NativeCodeStatus
	for _, vsnc := range gscf.nativecodes {
		for _, snc := range vsnc.sncMap {
			status := &xpb.NativeCodeStatus{
				Desc:    snc.desc,
				Status:  int32(snc.status),
				Healthy: !snc.lostBeatheart,
			}
			result = append(result, status)
		}
	}
	return result
}

// launchOne 启动一个合约, 并且将其添加到vsnc列表
// directory structue of a contract:
// basedir: data/blockchain/xuper/native/driver
// tree:
// .
// |____math_1.0
// | |____bin
// | | |____nativecode
// | | |____nativecode.old
// |____math
// | |____bin
// | | |____nativecode
// | | |____nativecode.old
func (gscf *GeneralSCFramework) launchOne(desc *xpb.NativeCodeDesc, status nativeCodeStatus) (*standardNativeContract, error) {
	gscf.wg.Add(1)
	defer gscf.wg.Done()
	name := desc.GetName()
	syscallSockpath := gscf.chainSock3Path
	snc := &standardNativeContract{
		name:          name,
		version:       desc.Version,
		status:        status,
		basedir:       filepath.Join(gscf.rootpath, "driver", string(makeNativeCodeName(desc))),
		chainSockPath: syscallSockpath,
		Logger:        gscf.Logger.New("code", name),
		mutex:         new(sync.Mutex),
		mgr:           gscf,
		desc:          desc,
		lostBeatheart: true,
		dockerClient:  gscf.dockerClient,
	}
	var err error
	err = snc.Init()
	if err != nil {
		return nil, fmt.Errorf("init driver %s error:%s", name, err)
	}
	digest, err := snc.GetNativeCodeDigest()
	if err != nil {
		gscf.Error("get native code digest error", "error", err, "name", name)
	} else {
		if !bytes.Equal(digest, desc.GetDigest()) {
			gscf.Warn(fmt.Sprintf("native load plugins get digest %x, not equal one in desc %x", digest, desc.Digest))
		}
	}

	err = snc.Start()
	if err != nil {
		if status != statusReady {
			return nil, fmt.Errorf("start driver %s error:%s", name, err)
		}
		// 如果是xchain重启后的启动失败，只打日志来标明错误，需要重新部署或者手动检查错误
		gscf.Error("start driver error", "error", err)
	}
	if err := gscf.addSNC(snc); err != nil {
		return nil, err
	}

	gscf.Info("native load plugin", "name", name, "digest", hex.EncodeToString(desc.Digest))
	return snc, nil
}

//addSNC 添加一个新版本的StandardNativeContract, 如果是第一个的话，默认创建一个versionedStandardNativeContract
func (gscf *GeneralSCFramework) addSNC(snc *standardNativeContract) error {
	//加载前一个版本的信息，保证回滚没有问题
	desc := snc.desc
	if tmpsnc, ok := gscf.nativecodes[snc.name]; !ok || tmpsnc == nil {
		vsnc := &versionedStandardNativeContract{
			curVersion: desc.GetVersion(),
			sncMap: map[string]*standardNativeContract{
				desc.GetVersion(): snc,
			},
			gscf: gscf,
		}
		snc.vsnc = vsnc
		gscf.Debug("Init GeneralSCFramework", "name", desc.Name)
		gscf.nativecodes[desc.GetName()] = vsnc
	} else {
		//检查版本是不是已经有了
		if _, ok := tmpsnc.sncMap[snc.desc.GetVersion()]; ok {
			return errors.New("can't deploy contract again with same version and name")
		}
		snc.vsnc = tmpsnc
		tmpsnc.sncMap[snc.desc.GetVersion()] = snc
		gscf.nativecodes[snc.name] = tmpsnc
	}
	return nil
}

// deployCode 实现代码按照指定版本进行部署的功能
func (gscf *GeneralSCFramework) deployCode(desc *xpb.NativeCodeDesc, code []byte) error {
	bindpath := makeNativeCodeName(desc)
	bindir := filepath.Join(gscf.driverpath, string(bindpath), "bin")
	err := os.MkdirAll(bindir, 0755)
	if err != nil {
		return err
	}
	binpath := filepath.Join(bindir, "nativecode")
	if _, err := os.Stat(binpath); err == nil {
		os.Rename(binpath, binpath+".old")
	}
	return ioutil.WriteFile(binpath, code, 0755)
}

func (gscf *GeneralSCFramework) monitorNativeCode() {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	for gscf.running {
		<-ticker.C
		for name, vsnc := range gscf.nativecodes {
			snc := vsnc.sncMap[vsnc.curVersion]
			func() {
				if !snc.lostBeatheart {
					return
				}
				gscf.Info("monitorNativeCode restart", "name", snc.name)
				err := snc.Restart()
				if err != nil {
					gscf.Error("restart native code error", "error", err, "name", name)
				}
			}()
		}
	}
}

func (gscf *GeneralSCFramework) activate(name string, version string) error {
	gscf.mutex.Lock()
	defer gscf.mutex.Unlock()
	var err error
	vsnc := gscf.nativecodes[name]
	if vsnc == nil {
		return fmt.Errorf("contract %s is missing, aim to activate version %s", name, version)
	}
	snc, err := vsnc.GetSNC(name, version)
	if err != nil {
		return fmt.Errorf("plugin %s does not exist", name)
	}
	oldsnc := vsnc.sncMap[vsnc.curVersion]
	//判断当前版本的状态是否已经已经就绪
	if snc.status != statusRegistered {
		return fmt.Errorf("can not transfer from %d to %d", snc.status, statusReady)
	}

	// set state of the plugin to ready
	snc.status = statusReady

	//修改版本信息
	if err := vsnc.upgrade(snc); err != nil {
		gscf.Error(fmt.Sprintf("version upgrade error %s", err))
		return err
	}

	if oldsnc.desc.GetVersion() != version {
		oldsnc.status = statusInvalid
		// 写入versionTable
		descbuf, _ := proto.Marshal(oldsnc.desc)
		vsnc.gscf.versionTable.Put(makeNativeCodeName(oldsnc.desc), descbuf)
	}

	//serialize the state into table, and value should include the md5 of the file
	descbuf, err := proto.Marshal(snc.desc)
	if err != nil {
		gscf.Error("proto marshal error", "error", err)
		return err
	}
	err = gscf.db.Put(getNativeCodeKey(name), descbuf)
	gscf.Debug("native activate", "name", name, "snc", fmt.Sprintf("%#v", snc), "key", string(getNativeCodeKey(name)), "dbput error", err)
	return err
}

// 注销合约
func (gscf *GeneralSCFramework) deactivate(name string, version string) error {
	gscf.mutex.Lock()
	defer gscf.mutex.Unlock()
	//set state of the plugin to registered
	vsnc := gscf.nativecodes[name]
	if vsnc == nil {
		return nil
	}
	snc, err := vsnc.GetSNC(name, version)
	if err != nil {
		gscf.Error(fmt.Sprintf("plugin %s does not exist", name), "error", err)
		return nil
	}
	if snc.status != statusReady {
		gscf.Error("status is not statusReady, there is no need to deactivate", "name", name, "version", version, "status", snc.status)
		return nil
	}
	//从内核模块中注销
	gscf.Debug("native deactivate", "name", name, "snc", fmt.Sprintf("%#v", snc), "key", string(getNativeCodeKey(name)))
	//标记为注册无效状态
	snc.status = statusInvalid
	//记录到历史库, 并且从vsnc里面删除snc
	if err := vsnc.serialize(snc); err != nil {
		gscf.Error("version deactivate error", "error", err)
		return nil
	}
	//从当前合约最新版本库删除
	gscf.db.Delete(getNativeCodeKey(name))
	//停止进程
	snc.Stop()
	return nil
}

// Run implements ContractInterface
// 升级流程： 直接部署新版本，然后投票指定高度激活即可
func (gscf *GeneralSCFramework) Run(desc *contract.TxDesc) error {
	nameObj, ok := desc.Args["pluginName"]
	if !ok || nameObj == nil {
		return fmt.Errorf("name [%s] is invalid", nameObj)
	}
	name := nameObj.(string)
	versionObj, ok := desc.Args["version"]
	if !ok || versionObj == nil {
		return fmt.Errorf("version [%s] is invalid", versionObj)
	}
	version := versionObj.(string)

	//	key := getTxVersionUpgraeKey(desc.Tx.Txid, name)
	switch desc.Method {
	case "activate":
		return gscf.activate(name, version)
	case "deactivate":
		return gscf.deactivate(name, version)
	}
	return nil
}

// Rollback implements ContractInterface
func (gscf *GeneralSCFramework) Rollback(desc *contract.TxDesc) error {
	nameObj, ok := desc.Args["pluginName"]
	if !ok || nameObj == nil {
		return fmt.Errorf("name [%s] is invalid", nameObj)
	}
	name := nameObj.(string)
	versionObj, ok := desc.Args["version"]
	if !ok || versionObj == nil {
		return fmt.Errorf("version [%s] is invalid", versionObj)
	}
	version := versionObj.(string)
	switch desc.Method {
	case "activate":
		//如果vsnc发生了变化
		vsnc := gscf.nativecodes[name]
		if vsnc == nil {
			gscf.Warn("cann't find contract", "name", name)
			return nil
		}
		if version == vsnc.curVersion {
			//说明在已经激活成功（至少部分成功）
			snc := vsnc.sncMap[vsnc.curVersion]
			prevVersion := snc.desc.PrevVersion
			if err := gscf.activate(name, prevVersion); err != nil {
				gscf.Error("Deactive error, log only", "name", name, "version", version)
			}
		}
		return nil
	case "deactivate":
		//注销的时候，几乎不需要回滚
		return nil
	}
	return nil
}

// Finalize implements ContractInterface
func (gscf *GeneralSCFramework) Finalize(blockid []byte) error {
	return nil
}

// Stop stop all the native contracts process
// It blocks until all the process ends
func (gscf *GeneralSCFramework) Stop() {
	gscf.running = false
	gscf.wg.Wait()
	gscf.db.Close()
}

// SetContext implements ContractInterface
func (gscf *GeneralSCFramework) SetContext(context *contract.TxContext) error {
	gscf.context = context
	return nil
}

// ReadOutput implements ContractInterface
func (gscf *GeneralSCFramework) ReadOutput(desc *contract.TxDesc) (contract.ContractOutputInterface, error) {
	return nil, nil
}

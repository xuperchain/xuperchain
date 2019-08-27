package consensus

import (
	"errors"
	"fmt"
	"github.com/xuperchain/xuperunion/p2pv2"
	"os"
	"strconv"
	"strings"
	"sync"

	log "github.com/xuperchain/log15"
	"github.com/xuperchain/xuperunion/common/config"
	cons_base "github.com/xuperchain/xuperunion/consensus/base"
	"github.com/xuperchain/xuperunion/contract"
	crypto_client "github.com/xuperchain/xuperunion/crypto/client"
	crypto_base "github.com/xuperchain/xuperunion/crypto/client/base"
	"github.com/xuperchain/xuperunion/global"
	"github.com/xuperchain/xuperunion/ledger"
	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/pluginmgr"
	"github.com/xuperchain/xuperunion/utxo"
	"github.com/xuperchain/xuperunion/vat"
)

// Question: how to support unknown consensus? when does the consensus update?
const (
	updateConsensusMethod = "update_consensus"
	ConsensusPluginName   = "consensus"
	ConsensusTypeTdpos    = "tdpos"
	ConsensusTypePow      = "pow"
	ConsensusTypeSingle   = "single"
)

// StepConsensus is the struct stored the consensus instance
type StepConsensus struct {
	StartHeight int64
	Txid        []byte
	Conn        cons_base.ConsensusInterface
}

// PluggableConsensus is the struct stored the pluggable consensus of a chain
type PluggableConsensus struct {
	xlog         log.Logger
	cfg          *config.NodeConfig
	bcname       string
	ledger       *ledger.Ledger
	utxoVM       *utxo.UtxoVM
	context      *contract.TxContext
	cons         []*StepConsensus
	cryptoClient crypto_base.CryptoClient
	mutex        *sync.RWMutex
	p2psvr       p2pv2.P2PServer
}

func genPlugConsKey(height int64, timestamp int64) string {
	return fmt.Sprintf("%020d_%d", height, timestamp)
}

func genPlugConsKeyWithPrefix(height int64, timestamp int64) string {
	baseKey := genPlugConsKey(height, timestamp)
	return pb.PlugConsPrefix + baseKey
}

func parsePlugConsKeyWithPrefix(key string) (height int64, timestamp int64, err error) {
	baseKey := strings.TrimPrefix(key, pb.PlugConsPrefix)
	subKeys := strings.Split(baseKey, "_")
	if len(subKeys) != 2 {
		return 0, 0, errors.New("parse height and timestamp error")
	}

	height, err = strconv.ParseInt(subKeys[0], 10, 64)
	if err != nil {
		return 0, 0, err
	}

	timestamp, err = strconv.ParseInt(subKeys[1], 10, 64)
	if err != nil {
		return 0, 0, err
	}

	return height, timestamp, nil
}

func (pc *PluggableConsensus) makeFirstCons(xlog log.Logger, cfg *config.NodeConfig, gCon map[string]interface{}) (*StepConsensus, error) {
	name, consConf, err := pc.validateUpdateConsensus(gCon)
	if err != nil {
		return nil, err
	}
	height := int64(0)
	timestamp := int64(0)
	if name == ConsensusTypeTdpos {
		if consConf["timestamp"] == nil {
			return nil, errors.New("Genious consensus tdpos's timestamp can not be null")
		}
		tmpTime, err := strconv.ParseInt(consConf["timestamp"].(string), 10, 64)
		if err != nil {
			return nil, err
		}
		timestamp = tmpTime
	}
	return pc.newUpdateConsensus(name, height, timestamp, consConf, nil)
}

// NewPluggableConsensus create the PluggableConsensus instance
func NewPluggableConsensus(xlog log.Logger, cfg *config.NodeConfig, bcname string,
	ledger *ledger.Ledger, utxoVM *utxo.UtxoVM, gCon map[string]interface{},
	cryptoType string, p2psvr p2pv2.P2PServer) (*PluggableConsensus, error) {
	if xlog == nil {
		xlog = log.New("module", "plug_cons")
		xlog.SetHandler(log.StreamHandler(os.Stderr, log.LogfmtFormat()))
	}
	// create crypto client
	cryptoClient, cryptoErr := crypto_client.CreateCryptoClient(cryptoType)
	if cryptoErr != nil {
		xlog.Warn("Load crypto client failed", "err", cryptoErr)
	}
	pc := &PluggableConsensus{
		xlog:         xlog,
		cfg:          cfg,
		bcname:       bcname,
		ledger:       ledger,
		utxoVM:       utxoVM,
		cryptoClient: cryptoClient,
		mutex:        new(sync.RWMutex),
		p2psvr:       p2psvr,
	}

	first, err := pc.makeFirstCons(xlog, cfg, gCon)
	if err != nil {
		pc.xlog.Warn("NewPluggableConsensus make first cons error!", "error", err.Error())
		return nil, err
	}

	pc.cons = append(pc.cons, first)
	meta := ledger.GetMeta()
	if meta.TrunkHeight == 0 {
		return pc, nil
	}
	blockTip, err := ledger.QueryBlock(meta.TipBlockid)
	if err != nil {
		return nil, err
	}

	it := utxoVM.ScanWithPrefix([]byte(pb.PlugConsPrefix))
	defer it.Release()
	for it.Next() {
		key := string(it.Key())
		height, timestamp, err := parsePlugConsKeyWithPrefix(key)
		txid := it.Value()
		tx, err := ledger.QueryTransaction(txid)
		if err != nil {
			return nil, err
		}
		pc.xlog.Trace("Start to init consensus", "height", height, "timestamp", timestamp,
			"blockTip", fmt.Sprintf("%x", blockTip.Blockid), "txid", fmt.Sprintf("%x", txid))
		desc, err := contract.Parse(string(tx.Desc))
		if err != nil {
			return nil, err
		}
		name, consConf, err := pc.validateUpdateConsensus(desc.Args)
		if err != nil {
			return nil, err
		}
		cons, err := pc.newUpdateConsensus(name, height, timestamp, consConf, blockTip)
		if err != nil {
			return nil, err
		}
		cons.Txid = txid
		pc.cons = append(pc.cons, cons)
	}
	return pc, nil
}

// CompeteMaster confirm whether the node is a miner or not
func (pc *PluggableConsensus) CompeteMaster(height int64) (bool, bool) {
	for i := len(pc.cons) - 1; i >= 0; i-- {
		if height >= pc.cons[i].StartHeight {
			return pc.cons[i].Conn.CompeteMaster(height)
		}
	}
	return false, false
}

// Type return the consensus type of a specific height
func (pc *PluggableConsensus) Type(height int64) string {
	for i := len(pc.cons) - 1; i >= 0; i-- {
		if height >= pc.cons[i].StartHeight {
			return pc.cons[i].Conn.Type()
		}
	}
	return ConsensusTypeSingle
}

// Version return the consensus version of a specific height
func (pc *PluggableConsensus) Version(height int64) int64 {
	for i := len(pc.cons) - 1; i >= 0; i-- {
		if height >= pc.cons[i].StartHeight {
			return pc.cons[i].Conn.Version()
		}
	}
	return 0
}

// CheckMinerMatch check whether the block is valid
func (pc *PluggableConsensus) CheckMinerMatch(header *pb.Header, in *pb.InternalBlock) (bool, error) {
	for i := len(pc.cons) - 1; i >= 0; i-- {
		if in.Height >= pc.cons[i].StartHeight {
			if header == nil {
				header = global.GHeader()
			}
			return pc.cons[i].Conn.CheckMinerMatch(header, in)
		}
	}
	return false, nil
}

// ProcessBeforeMiner preprocessing before mining
func (pc *PluggableConsensus) ProcessBeforeMiner(height int64, timestamp int64) (map[string]interface{}, bool) {
	for i := len(pc.cons) - 1; i >= 0; i-- {
		if height >= pc.cons[i].StartHeight {
			return pc.cons[i].Conn.ProcessBeforeMiner(timestamp)
		}
	}
	return nil, false
}

// ProcessConfirmBlock process after block has been confirmed
func (pc *PluggableConsensus) ProcessConfirmBlock(in *pb.InternalBlock) error {
	for i := len(pc.cons) - 1; i >= 0; i-- {
		if in.Height >= pc.cons[i].StartHeight {
			return pc.cons[i].Conn.ProcessConfirmBlock(in)
		}
	}
	return nil
}

func (pc *PluggableConsensus) postUpdateConsensusActions(name string, sc *StepConsensus) error {
	if name == ConsensusTypeTdpos {
		cons := sc.Conn
		for _, v := range pc.cons {
			if v.Conn.Type() == ConsensusTypeTdpos {
				if v.Conn.Version() == cons.Version() {
					pc.xlog.Warn("This version of tdpos already exist", "version", v.Conn.Version())
					return errors.New("This version of tdpos already exist")
				}
			}
		}

		// 注册tdpos合约
		ci := cons.(contract.ContractInterface)
		pc.utxoVM.UnRegisterVM(ConsensusTypeTdpos, global.VMPrivRing0)
		pc.utxoVM.RegisterVM(ConsensusTypeTdpos, ci, global.VMPrivRing0)
		vat := cons.(vat.VATInterface)
		pc.utxoVM.UnRegisterVAT(ConsensusTypeTdpos)
		pc.utxoVM.RegisterVAT(ConsensusTypeTdpos, vat, vat.GetVATWhiteList())
		pc.xlog.Trace("Register Tdpos utxovm after updateTDPosConsensus", "name", "Tdpos")
	}

	return nil
}

func (pc *PluggableConsensus) updateConsensusByName(name string, height int64,
	consConf map[string]interface{}, extParams map[string]interface{}) (*StepConsensus, error) {
	// load consensus plugin
	pluginMgr, err := pluginmgr.GetPluginMgr()
	if err != nil {
		pc.xlog.Warn("create consensus instance failed", "name", name)
		return nil, err
	}

	pluginIns, err := pluginMgr.PluginMgr.CreatePluginInstance(ConsensusPluginName, name)
	if err != nil {
		pc.xlog.Warn("create consensus instance failed", "name", name)
		if err.Error() == "Invalid plugin subtype" {
			err = errors.New("Consensus not support")
		}
		return nil, err
	}

	consInstance := pluginIns.(cons_base.ConsensusInterface)
	err = consInstance.Configure(pc.xlog, pc.cfg, consConf, extParams)
	if err != nil {
		return nil, err
	}
	return &StepConsensus{
		StartHeight: height,
		Conn:        consInstance,
	}, nil
}

func (pc *PluggableConsensus) newUpdateConsensus(name string, height int64, timestamp int64,
	consConf map[string]interface{}, block *pb.InternalBlock) (*StepConsensus, error) {
	// build extParams
	extParams := make(map[string]interface{})
	extParams["block"] = block
	extParams["crypto_client"] = pc.cryptoClient
	if name == ConsensusTypeTdpos {
		extParams["bcname"] = pc.bcname
		extParams["ledger"] = pc.ledger
		extParams["utxovm"] = pc.utxoVM
		extParams["timestamp"] = timestamp
		extParams["p2psvr"] = pc.p2psvr
		extParams["height"] = height
	} else if name == ConsensusTypePow {
		extParams["ledger"] = pc.ledger
	}
	// create and config consensus instance
	cons, err := pc.updateConsensusByName(name, height, consConf, extParams)
	if err != nil {
		pc.xlog.Error("Update consensus error", "name", name, "error", err.Error())
		return nil, err
	}

	// handle post actions of consensus
	err = pc.postUpdateConsensusActions(name, cons)
	if err != nil {
		return nil, err
	}
	return cons, nil
}

//升级智能合约
func (pc *PluggableConsensus) updateConsensus(name string, consConf map[string]interface{},
	txid []byte, block *pb.InternalBlock) error {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()

	height := int64(0)
	if block.Height == 0 {
		height = pc.ledger.GetMeta().TrunkHeight + 1
	} else {
		height = block.Height
	}
	cons, err := pc.newUpdateConsensus(name, height, block.Timestamp, consConf, block)
	if err != nil {
		pc.xlog.Error("new update consensus error", "error", err.Error())
		return err
	}
	cons.Txid = txid
	pc.cons = append(pc.cons, cons)
	key := genPlugConsKeyWithPrefix(height, block.Timestamp)
	pc.context.UtxoBatch.Put([]byte(key), txid)
	return nil
}

// 回滚智能合约
func (pc *PluggableConsensus) rollbackConsensus(name string, consConf map[string]interface{},
	txid []byte, block *pb.InternalBlock) error {
	if block == nil {
		return errors.New("RollbackConsensus block can not be null")
	}
	pc.mutex.Lock()
	defer pc.mutex.Unlock()

	length := len(pc.cons)

	if length <= 1 {
		pc.xlog.Trace("rollbackConsensus len of plugging cons is nil, don't need to rollback!")
		return nil
	}

	if string(pc.cons[length-1].Txid) != string(txid) {
		pc.xlog.Trace("rollbackConsensus current con txid doesn't match, don't need to rollback!")
		return nil
	}

	height := int64(0)
	if block.Height == 0 {
		height = pc.ledger.GetMeta().TrunkHeight + 1
	} else {
		height = block.Height
	}

	var flagIndex int
	for i := length - 1; i >= 0; i-- {
		if pc.cons[i].StartHeight <= height {
			flagIndex = i
			break
		}
	}
	err := pc.cons[flagIndex-1].Conn.InitCurrent(block)
	if err != nil {
		return err
	}

	if pc.cons[flagIndex-1].Conn.Type() == ConsensusTypeTdpos {
		// 注销 tdpos
		pc.utxoVM.UnRegisterVM(ConsensusTypeTdpos, global.VMPrivRing0)
		pc.utxoVM.UnRegisterVAT(ConsensusTypeTdpos)
	}

	pc.cons = pc.cons[:flagIndex-1]
	if flagIndex > 1 && pc.cons[flagIndex-2].Conn.Type() == ConsensusTypeTdpos {
		con := pc.cons[flagIndex-2].Conn
		// 注册tdpos合约
		vmCons := con.(contract.ContractInterface)
		pc.utxoVM.RegisterVM(ConsensusTypeTdpos, vmCons, global.VMPrivRing0)
		vatCons := con.(vat.VATInterface)
		pc.utxoVM.RegisterVAT(ConsensusTypeTdpos, vatCons, vatCons.GetVATWhiteList())
	}

	key := genPlugConsKeyWithPrefix(height, block.Timestamp)
	pc.context.UtxoBatch.Delete([]byte(key))
	return nil
}

func (pc *PluggableConsensus) validateUpdateConsensus(args map[string]interface{}) (string, map[string]interface{},
	error) {
	if args["name"] == nil {
		return "", nil, errors.New("Consensus name can not be bull")
	}

	if args["config"] == nil {
		return "", nil, errors.New("Consensus config can not be bull")
	}

	name := args["name"].(string)
	consConf := args["config"].(map[string]interface{})
	return name, consConf, nil
}

// Run is the specific implementation of interface contract
func (pc *PluggableConsensus) Run(desc *contract.TxDesc) error {
	pc.xlog.Trace("receive update cons", "module", desc.Method, "method", desc.Method)
	switch desc.Method {
	case updateConsensusMethod:
		name, consConf, err := pc.validateUpdateConsensus(desc.Args)
		if err != nil {
			pc.xlog.Warn("run pluggable consensus contract error", "error", "valid desc error")
			return err
		}
		return pc.updateConsensus(name, consConf, desc.Tx.Txid, pc.context.Block)
	default:
		pc.xlog.Warn("method not defined", "module", desc.Method, "method", desc.Method)
		return errors.New("PluggableConsensus not define this method")
	}
}

// Rollback is the specific implementation of interface contract
func (pc *PluggableConsensus) Rollback(desc *contract.TxDesc) error {
	switch desc.Method {
	case updateConsensusMethod:
		name, consConf, err := pc.validateUpdateConsensus(desc.Args)
		if err != nil {
			pc.xlog.Warn("rollback pluggable consensus contract error", "error", "valid desc error")
			return err
		}
		return pc.rollbackConsensus(name, consConf, desc.Tx.Txid, pc.context.Block)
	default:
		pc.xlog.Warn("method not defined", "module", desc.Method, "method", desc.Method)
		return errors.New("PluggableConsensus not define this method")
	}
}

// Finalize is the specific implementation of interface contract
func (pc *PluggableConsensus) Finalize(blockid []byte) error {
	return nil
}

// SetContext is the specific implementation of interface contract
func (pc *PluggableConsensus) SetContext(context *contract.TxContext) error {
	pc.context = context
	return nil
}

// Stop is the specific implementation of interface contract
func (pc *PluggableConsensus) Stop() {
}

// ReadOutput is the specific implementation of interface contract
func (pc *PluggableConsensus) ReadOutput(desc *contract.TxDesc) (contract.ContractOutputInterface, error) {
	return nil, nil
}

// GetVerifiableAutogenTx is the specific implementation of interface VAT
func (pc *PluggableConsensus) GetVerifiableAutogenTx(blockHeight int64, maxCount int, timestamp int64) ([]*pb.Transaction, error) {
	return nil, nil
}

// GetVATWhiteList the specific implementation of interface VAT
func (pc *PluggableConsensus) GetVATWhiteList() map[string]bool {
	whiteList := map[string]bool{
		updateConsensusMethod: true,
	}
	return whiteList
}

// GetCoreMiners get the information of core miners
func (pc *PluggableConsensus) GetCoreMiners() []*cons_base.MinerInfo {
	currentConsIndex := len(pc.cons) - 1
	res := pc.cons[currentConsIndex].Conn.GetCoreMiners()
	return res
}

// GetStatus get current consensus status
func (pc *PluggableConsensus) GetStatus() *cons_base.ConsensusStatus {
	currentConsIndex := len(pc.cons) - 1
	return pc.cons[currentConsIndex].Conn.GetStatus()
}

package xchaincore

import (
	"container/list"
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	math_rand "math/rand"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	log "github.com/xuperchain/log15"
	"github.com/xuperchain/xuperchain/core/common"
	"github.com/xuperchain/xuperchain/core/common/config"
	"github.com/xuperchain/xuperchain/core/consensus"
	"github.com/xuperchain/xuperchain/core/global"
	"github.com/xuperchain/xuperchain/core/ledger"
	p2p_base "github.com/xuperchain/xuperchain/core/p2p/base"
	xuper_p2p "github.com/xuperchain/xuperchain/core/p2p/pb"
	"github.com/xuperchain/xuperchain/core/pb"
	"github.com/xuperchain/xuperchain/core/utxo"
)

type KeeperStatus int

const DefaultFilterHeight = 10
const (
	Waiting     KeeperStatus = iota // 无任务
	Syncing                         // 自动同步
	Truncateing                     // 手动裁剪
	Appending                       // 主干已为最新，向主干append新区块
)

var (
	// ErrInvalidMsg msg received cannot work on agreement terms the protocol required
	ErrInvalidMsg = errors.New("params are invalid when sync data")
	// ErrUnmarshal unmarshal error
	ErrUnmarshal = errors.New("unmarshall msg error")
	// ErrTargetPeerTimeOut peer timeout
	ErrTargetPeerTimeOut = errors.New("target peer cannot response to the requests on time")
	// ErrInternal p2p internal error
	ErrInternal = errors.New("cannot sent msg because of internal error")
	// ErrHaveNoTargetData the callee cannot find data in main-chain
	ErrTargetDataNotFound = errors.New("cannot search the specific data the caller wants")
	// ErrTargetDataNotEnough the callee cannot find enough data
	ErrTargetDataNotEnough = errors.New("peers cannot search the whole data the caller wants, not enough")
	// ErrPeersInvalid p2p peers invalid
	ErrAllPeersInvalid = errors.New("All the peers are invalid")
)

var zeroStr = fmt.Sprintf("%064d", 0)
var taskMutex sync.Mutex
var taskId int64

const (
	SIZE_OF_UINT32          = 4
	SYNC_BLOCKS_TIMEOUT     = 600000 * time.Millisecond
	HEADER_SYNC_SIZE        = 100
	FAST_CACHE_SIZE         = 1000
	CHAN_SIZE               = 10000
	BLOCK_CACHE_SIZE        = 1000
	CHAN_OF_TASK_SIZE       = 10
	KEEPER_SLEEP_MIL_SECOND = 1000
	MANAGER_TASKS_TYPES     = 2
	MAX_TASK_MN_SIZE        = 20
)

type TasksList struct {
	tList *list.List       // 存放具体task
	tMap  *map[string]bool // tList目标BlockId的Set
}

/* NewTasksList return a link queue TasksList
 * NewTasksList 返回一个包含链表和链表所含元素组成的set的数据结构
 */
func NewTasksList() *TasksList {
	return &TasksList{
		tList: list.New(),
		tMap:  &map[string]bool{},
	}
}

/* Len return the length of list
 * Len 返回TasksList中tList的长度
 */
func (t *TasksList) Len() int {
	return t.tList.Len()
}

/* PushBack push task at the end of taskslist
 * PushBack 向TasksList中的链表做PushBack操作，同时确保push的元素原链表中没有
 */
func (t *TasksList) PushBack(ledgerTask *LedgerTask) {
	if _, ok := (*t.tMap)[ledgerTask.targetBlockId]; ok {
		return
	}
	t.tList.PushBack(ledgerTask)
	(*t.tMap)[ledgerTask.targetBlockId] = true
}

/* Remove remove the target element in tList of taskslist
 * Remove 删除链表中元素，并对应删除set
 */
func (t *TasksList) Remove(e *list.Element) {
	if _, ok := (*t.tMap)[e.Value.(*LedgerTask).targetBlockId]; !ok {
		return
	}
	delete(*t.tMap, e.Value.(*LedgerTask).targetBlockId)
	t.tList.Remove(e)
}

/* Front get the front element in tList of taskslist
 * Front获取链表头元素
 */
func (t *TasksList) Front() *list.Element {
	if t.Len() == 0 {
		return nil
	}
	e := t.tList.Front()
	return e
}

/* fix 删除链表中比目标高度高的所有任务，一般在trunate操作调用
 */
func (t *TasksList) fix(targetHeight int64) {
	e := t.tList.Front()
	for e != nil {
		if e.Value.(*LedgerTask).GetHeight() <= targetHeight {
			e = e.Next()
			continue
		}
		target := e
		e = e.Next()
		t.Remove(target)
	}
}

type SyncTaskManager struct {
	syncingTasks     *TasksList      // 账本同步batch操作
	appendingTasks   *TasksList      // 账本同步追加操作
	syncMaxSize      int             // sync队列最大size
	FilterBlockidMap map[string]bool // 存储试图同步但收到全0的blockid，在下一次检查时直接过滤
	log              log.Logger
	syncMutex        *sync.Mutex
}

/* newSyncTaskManager 生成一个新SyncTaskManager，管理所有任务队列
 * 任务队列包含两个队列，syncingTasks用于记录多个block的同步任务，
 * appendingTasks用于记录收到广播新块，且新块直接为账本下一高度的情况，该任务直接尝试写账本
 */
func newSyncTaskManager(log log.Logger) *SyncTaskManager {
	return &SyncTaskManager{
		syncingTasks:     NewTasksList(),
		appendingTasks:   NewTasksList(),
		syncMaxSize:      MAX_TASK_MN_SIZE,
		FilterBlockidMap: make(map[string]bool),
		syncMutex:        &sync.Mutex{},
		log:              log,
	}
}

/* put 由manager操作，直接操作所handle的各个队列
 */
func (stm *SyncTaskManager) put(ledgerTask *LedgerTask) bool {
	stm.syncMutex.Lock()
	defer stm.syncMutex.Unlock()
	if ledgerTask.GetAction() == Syncing {
		stm.log.Trace("SyncTaskManager put syncingTasks", "len", stm.syncingTasks.Len())
		if stm.syncingTasks.Len() > stm.syncMaxSize {
			stm.log.Trace("SyncTaskManager put task err, too much task, refuse it")
			return false
		}
		stm.syncingTasks.PushBack(ledgerTask)
		return true
	}
	if ledgerTask.GetAction() == Appending {
		stm.log.Trace("SyncTaskManager put appendingTasks", "len", stm.appendingTasks.Len())
		if stm.appendingTasks.Len() > stm.syncMaxSize {
			stm.log.Trace("SyncTaskManager put task err, too much task, refuse it")
			return false
		}
		stm.appendingTasks.PushBack(ledgerTask)
		return true
	}
	return false
}

/* getTaskIndex 优先选取Appending队列中的task，
 * 在同步逻辑中，最新块广播的优先级高于其他
 */
func (stm *SyncTaskManager) getTaskIndex() int {
	if stm.appendingTasks.Len() > 0 {
		return 0
	}
	return 1
}

/* get 由manager操作从所有队列中挑选一个队列拿一个任务
 */
func (stm *SyncTaskManager) get() *LedgerTask {
	stm.syncMutex.Lock()
	defer stm.syncMutex.Unlock()
	index := stm.getTaskIndex()
	switch index {
	case 0:
		e := stm.appendingTasks.Front()
		stm.appendingTasks.Remove(e)
		stm.log.Trace("SyncTaskManager::", "appendingTasks Len", stm.appendingTasks.Len())
		return e.Value.(*LedgerTask)
	case 1:
		e := stm.syncingTasks.Front()
		if e == nil {
			stm.log.Trace("SyncTaskManager::queues' Len == 0")
			return nil
		}
		stm.syncingTasks.Remove(e)
		stm.log.Trace("SyncTaskManager::", "syncingTasks Len", stm.syncingTasks.Len())
		return e.Value.(*LedgerTask)
	}
	return nil
}

// fixTask 清洗队列中失效的任务
func (stm *SyncTaskManager) fixTask(targetHeight int64) {
	stm.syncMutex.Lock()
	defer stm.syncMutex.Unlock()
	stm.syncingTasks.fix(targetHeight)
	stm.appendingTasks.fix(targetHeight)
}

type LedgerKeeper struct {
	P2pSvr              p2p_base.P2PServer
	log                 log.Logger
	syncMsgChan         chan *xuper_p2p.XuperMessage
	peersStatusMap      *sync.Map        // *map[string]bool 更新同步节点的p2p列表活性
	fastFetchBlockCache *common.LRUCache // block header cache, 加速fetchBlock
	maxBlocksMsgSize    int64            // 取最大区块大小
	ledger              *ledger.Ledger
	bcName              string
	keeperStatus        KeeperStatus // 当前状态
	ledgerMutex         *sync.Mutex  // 锁, Sync(异步)和Truncate(同步)抢锁
	syncTaskMg          *SyncTaskManager
	nodeMode            string

	utxovm *utxo.UtxoVM
	con    *consensus.PluggableConsensus
}

type LedgerTask struct {
	taskId        string
	action        KeeperStatus
	targetBlockId string
	targetHeight  int64
	ctx           *LedgerTaskContext
}

type LedgerTaskContext struct {
	extBlocks  *map[string]*SimpleBlock
	preferPeer *[]string
	hd         *global.XContext
}

/* NewLedgerTaskContext make an new task ctx in LedgerKeeper.
 * NewLedgerTaskContext 新建一个账本任务Ctx，extBlocks为appending任务时需要的新block消息，
 * preferPeer指定了发送同步消息的目的节点
 */
func NewLedgerTaskContext(extBlocks *map[string]*SimpleBlock, preferPeer *[]string, hd *global.XContext) *LedgerTaskContext {
	return &LedgerTaskContext{
		extBlocks:  extBlocks,
		preferPeer: preferPeer,
		hd:         hd,
	}
}

type SimpleBlock struct {
	internalBlock *pb.InternalBlock
	header        *pb.Header
}

/* GetTargetBlockId get the target blockid of the task.
 * GetTargetBlockId获取任务目标Blockid
 * SyncingTask的目标Blockid表示请求者的本地最高id
 * AppendingTask的目标Blockid表示请求者的本地最高id
 * TruncateingTask的目标Blockid表示请求者需要裁剪到的id
 */
func (lt *LedgerTask) GetTargetBlockId() string {
	return lt.targetBlockId
}

/* GetAction get the action of the task.
 * GetAction 获取task的目标行为，Waiting/Syncing/Truncateing/Appending
 */
func (lt *LedgerTask) GetAction() KeeperStatus {
	return lt.action
}

/* GetHeight get the height of the task.
 * GetHeight 获取task的targetId的对应高度，此处只是作为manager管理task使用
 * 加入height的目的是减少操作者频繁QueryBlock
 */
func (lt *LedgerTask) GetHeight() int64 {
	return lt.targetHeight
}

/* GetExtBlocks get extBlocks of the task.
 * GetExtBlocks 获取任务的addition存储，在AppendingTask中使用，存入要直接写入账本的block
 */
func (lt *LedgerTask) GetExtBlocks() map[string]*SimpleBlock {
	if lt.ctx == nil || lt.ctx.extBlocks == nil {
		return nil
	}
	return *lt.ctx.extBlocks
}

/* GetPreferPeer get the callee of the task.
 * GetPreferPeer 获取向其余节点发送请求信息时指定的Peer地址
 */
func (lt *LedgerTask) GetPreferPeer() []string {
	if lt.ctx == nil || lt.ctx.preferPeer == nil {
		return nil
	}
	return *lt.ctx.preferPeer
}

// GetXCtx get the context of the task.
func (lt *LedgerTask) GetXCtx() *global.XContext {
	if lt.ctx == nil || lt.ctx.hd == nil {
		return nil
	}
	return lt.ctx.hd
}

/* NewLedgerKeeper create a LedgerKeeper.
 * NewLedgerKeeper 构建一个操作集合类，包含了所有对账本的写操作，外界对账本的写操作都需经过LedgerKeeper对外接口完成
 * LedgerKeeper会管理一组task队列，task为外界对其的请求封装，分为直接追加账本(Appending)、批量同步(Syncing)和裁剪(Truncating)
 */
func NewLedgerKeeper(bcName string, slog log.Logger, p2pV2 p2p_base.P2PServer, ledger *ledger.Ledger, nodeMode string,
	utxovm *utxo.UtxoVM, con *consensus.PluggableConsensus) (*LedgerKeeper, error) {
	if slog == nil { //如果外面没传进来log对象的话
		slog = log.New("module", "syncnode")
		slog.SetHandler(log.StreamHandler(os.Stderr, log.LogfmtFormat()))
	}
	lk := &LedgerKeeper{
		P2pSvr:              p2pV2,
		log:                 slog,
		syncMsgChan:         make(chan *xuper_p2p.XuperMessage, CHAN_SIZE),
		maxBlocksMsgSize:    ledger.GetMaxBlockSize(),
		peersStatusMap:      new(sync.Map),
		fastFetchBlockCache: common.NewLRUCache(BLOCK_CACHE_SIZE),
		ledger:              ledger,
		bcName:              bcName,
		utxovm:              utxovm,
		con:                 con,
		ledgerMutex:         &sync.Mutex{},
		syncTaskMg:          newSyncTaskManager(slog),
		nodeMode:            nodeMode,
	}
	lk.log.Trace("LedgerKeeper Start to Register Subscriber")
	if _, err := lk.P2pSvr.Register(lk.P2pSvr.NewSubscriber(nil, xuper_p2p.XuperMessage_GET_HASHES, lk.handleGetHeadersMsg, "", lk.log)); err != nil {
		return lk, err
	}
	if _, err := lk.P2pSvr.Register(lk.P2pSvr.NewSubscriber(nil, xuper_p2p.XuperMessage_GET_BLOCKS, lk.handleGetDataMsg, "", lk.log)); err != nil {
		return lk, err
	}
	updatePeerStatusMap(lk)
	return lk, nil
}

/* DoTruncateTask Truncate truncate ledger and set tipblock to utxovmLastID
 * 封装原来的ledger Truncate(), xchaincore使用此函数执行裁剪
 * TODO：共识部分也执行了裁剪工作，但使用的是ledger的Truncate()，后续需要使用LedgerKeeper替代
 */
func (lk *LedgerKeeper) DoTruncateTask(utxovmLastID []byte) error {
	lk.ledgerMutex.Lock()
	defer lk.ledgerMutex.Unlock()
	lk.keeperStatus = Truncateing
	return lk.ledger.Truncate(utxovmLastID)
}

// PutTask put a task into LedgerKeeper's queues.
func (lk *LedgerKeeper) PutTask(ledgerTask *LedgerTask) bool {
	return lk.syncTaskMg.put(ledgerTask)
}

// Start start ledgerkeeper's loop
func (lk *LedgerKeeper) Start() {
	go lk.startTaskLoop()
}

// startTaskLoop get tasks from ledgerkeeper's queues in the loop.
func (lk *LedgerKeeper) startTaskLoop() {
	for {
		lk.log.Trace("StartTaskLoop::Start......")
		updatePeerStatusMap(lk)
		task, ok := lk.getTask()
		if !ok {
			time.Sleep(time.Duration(math_rand.Intn(KEEPER_SLEEP_MIL_SECOND)) * time.Millisecond)
			continue
		}
		action := task.GetAction()
		lk.log.Trace("StartTaskLoop::Get a task", "action", action, "targetId", task.GetTargetBlockId())
		method := runKernelFuncMap[action]
		lk.keeperRun(method, task)
	}
}

// getTask get a task from LedgerKeeper's queues.
func (lk *LedgerKeeper) getTask() (*LedgerTask, bool) {
	lk.ledgerMutex.Lock()
	defer lk.ledgerMutex.Unlock()
	// 刚刚完成Truncate操作，需要清除列表中已无效的task
	if lk.keeperStatus == Truncateing {
		lk.syncTaskMg.fixTask(lk.ledger.GetMeta().GetTrunkHeight())
	}
	task := lk.syncTaskMg.get()
	if task == nil {
		lk.keeperStatus = Waiting
		return nil, false
	}
	lk.keeperStatus = task.GetAction()
	return task, true
}

/* updatePeerStatusMap 更新LedgerKeeper的syncMap
 * TODO: 后续完善Peer节点维护逻辑，节点不一定要永久剔除
 */
func updatePeerStatusMap(lk *LedgerKeeper) {
	for _, id := range lk.P2pSvr.GetPeersConnection() {
		lk.log.Trace("Init::", "id", id)
		if id == lk.P2pSvr.GetLocalUrl() {
			continue
		}
		if _, ok := lk.peersStatusMap.Load(id); ok {
			continue
		}
		lk.peersStatusMap.Store(id, true)
	}
}

// generateTaskid generate a random id.
func generateTaskid() string {
	taskMutex.Lock()
	taskId++
	taskMutex.Unlock()
	t := time.Now().UnixNano()
	return fmt.Sprintf("%d_%d", t, taskId)
}

/* NewLedgerTask create a ledger task.
 * NewLedgerTask 新建一个账本任务，需输入目标blockid(必须), 目标blockid所在的高度(必须),
 * 任务类型(必须: Appending追加、Syncing批量同步、Truncating裁剪),
 * 任务上下文(可选), 任务上下文包括	extBlocks *map[string]*SimpleBlock，追加账本时附带的追加block信息
 *	   preferPeer *[]string，询问邻居节点的地址
 *	   hd *global.XContext，计时context
 */
func NewLedgerTask(targetBlockId string, targetHeight int64, action KeeperStatus, ctx *LedgerTaskContext) *LedgerTask {
	return &LedgerTask{
		targetBlockId: targetBlockId,
		targetHeight:  targetHeight,
		taskId:        generateTaskid(),
		action:        action,
		ctx:           ctx,
	}
}

type KeeperMethodFunc func(sn *LedgerKeeper, st *LedgerTask) error

var runKernelFuncMap = map[KeeperStatus]KeeperMethodFunc{
	Syncing:   syncingBlocks,
	Appending: appendingBlock,
}

// keeperRun 任务枚举对应的函数指针
func (lk *LedgerKeeper) keeperRun(method KeeperMethodFunc, lt *LedgerTask) error {
	return method(lk, lt)
}

// syncingBlocks 批量同步操作
func syncingBlocks(lk *LedgerKeeper, lt *LedgerTask) error {
	lk.log.Trace("syncingBlocks::SyncTask by batch", "begin at", lt.GetTargetBlockId())
	syncBlocksWithTipidAndPeers(lk, lt)
	return nil
}

/* appendingBlock 直接在账本尾追加操作
 * 在广播新块时使用，当有新块产生时，且该新块恰好为本地账本的下一区块，此时直接试图向账本进行写操作，无需同步流程
 */
func appendingBlock(lk *LedgerKeeper, lt *LedgerTask) error {
	lk.log.Trace("appendingBlock::Run......")
	if lt.GetExtBlocks() == nil {
		lk.log.Warn("appendingBlock::AppendTask input error, len(ExtBlocks)==0")
		return ErrInvalidMsg
	}
	// 尝试直接写账本
	if len(lt.GetExtBlocks()) != 1 {
		lk.log.Warn("appendingBlock::AppendTask length error", "ExtBlocks", lt.GetExtBlocks())
		return ErrInvalidMsg
	}
	headersInfo := map[int]string{}
	i := 0
	for key, _ := range lt.GetExtBlocks() {
		headersInfo[i] = key
	}
	newBegin, ok, err := lk.confirmBlocks(lt.GetXCtx(), lt.GetTargetBlockId(), lt.GetTargetBlockId(), lt.GetExtBlocks(), headersInfo, true)
	lk.log.Trace("appendingBlock::AppendTask try to append ledger directly", "blockid", lt.GetTargetBlockId(), "newbegin", newBegin, "ok", ok, "error", err)
	return nil

}

/* syncBlocksWithTipIdAndPeers
 * syncBlocksWithTipIdAndPeers 请求消息头同步逻辑
 * 输入请求节点列表和起始区块哈希，迭代完成同步
 * 该函数首先完成区块头同步工作，发送GetHHashesMsg给指定peer，试图获取区间内所有区块哈希值
 * 获取到全部区块哈希列表之后，本节点将列表散列成若干份，并向指定列表节点发起同步具体区块工作，发送GetDataMsg请求，试图获取对应的所有详细区块消息
 * 若上一步并未在指定时间内获取到所有区块，则继续更换节点列表，该过程一直阻塞，直到获得所有区块，或者在超时后退出
 * 在该同步过程中顺便标注错误peer
 * 完成一个迭代后，task会向ledger中写数据，同时判断是否需要切换主干并完成写任务
 */
func syncBlocksWithTipidAndPeers(lk *LedgerKeeper, lt *LedgerTask) error {
	headerBegin := lt.targetBlockId
	nextLoop := true
	// 同步头过程
	id, err := hex.DecodeString(headerBegin)
	if err != nil {
		lk.log.Warn("syncBlocksWithTipidAndPeers::SyncTask parameter err", "task", headerBegin, "err", err)
	}
	block, err := lk.ledger.QueryBlock(id)
	tipHeight := lk.ledger.GetMeta().GetTrunkHeight()
	// 若任务的target高度过小，会直接忽略该任务
	if err != nil || !block.GetInTrunk() || block.GetHeight() < tipHeight-DefaultFilterHeight {
		lk.log.Warn("syncBlocksWithTipidAndPeers::SyncTask old blockid, ignore it", "task", headerBegin)
		return nil
	}
	// 该任务之前无法获取任何值，因此在新的同步开始被筛除
	if _, ok := lk.syncTaskMg.FilterBlockidMap[headerBegin]; ok && global.F(lk.ledger.GetMeta().GetRootBlockid()) != headerBegin {
		lk.log.Warn("syncBlocksWithTipidAndPeers::filterBlockidMap return", "headerBegin", headerBegin)
		return nil
	}
	lk.log.Trace("syncBlocksWithTipidAndPeers::Run......")
	for nextLoop {
		if getValidPeersNumber(lk.peersStatusMap) == 0 {
			lk.log.Warn("syncBlocksWithTipidAndPeers::getValidPeersNumber=0", "task", lt.GetTargetBlockId())
			return ErrAllPeersInvalid
		}

		peer := lt.GetPreferPeer()
		if peer == nil {
			peer, err = randomPickPeersWithNumber(1, lk.peersStatusMap)
			if err != nil {
				lk.log.Warn("syncBlocksWithTipidAndPeers::randomPickPeersWithNumber error", "task", lt.GetTargetBlockId(), "err", err)
				continue
			}
			lk.log.Trace("randomPickPeersWithNumber", "peer", peer[0])
		}

		endFlag, headersInfo, err := getBlockIdsWithGetHeadersMsg(lk, lk.bcName, headerBegin, HEADER_SYNC_SIZE, peer[0])
		lk.log.Info("syncBlocksWithTipidAndPeers::getBlockIdsWithGetHeadersMsg result", "task", lt.GetTargetBlockId(), "peer", peer[0], "err", err, "endFlag", endFlag)
		if err == ErrTargetDataNotFound {
			// beginBlockId疑似无效，需往前回溯，注意:往前回溯可能会导致主干切换
			lk.log.Trace("syncBlocksWithTipidAndPeers::get nothing from peers, begin backtracking...", "task", lt.GetTargetBlockId(), "headerBegin", headerBegin)
			lk.syncTaskMg.FilterBlockidMap[headerBegin] = true
			id, err := hex.DecodeString(headerBegin)
			if err != nil {
				return ErrInvalidMsg
			}
			block, err := lk.ledger.QueryBlock(id)
			if err != nil {
				return ErrInternal
			}
			headerBegin = changeSyncBeginPointBackwards(block)
			lk.log.Info("syncBlocksWithTipidAndPeers::backtrack start point", "task", lt.GetTargetBlockId(), "headerBegin", headerBegin)
			continue
		}
		if err != nil {
			lk.log.Warn("syncBlocksWithTipidAndPeers::delete peer", "task", lt.GetTargetBlockId(), "address", peer[0], "err", err)
			lk.peersStatusMap.Store(peer[0], false)
			continue
		}
		if endFlag {
			nextLoop = false
		}
		blocksMap := lt.blocksDownloadWithHeadersList(lk, headersInfo)
		if len(headersInfo) == 0 || len(blocksMap) == 0 {
			// 此处直接return, 加速task消耗
			return nil
		}
		// 本轮同步结束，开始写账本
		newBegin, _, err := lk.confirmBlocks(lt.GetXCtx(), lt.targetBlockId, headerBegin, blocksMap, headersInfo, endFlag)
		if err != nil {
			lk.log.Warn("syncBlocksWithTipidAndPeers::ConfirmBlocks error", "err", err)
			return nil
		}
		headerBegin = newBegin
	}
	lk.log.Trace("syncBlocksWithTipidAndPeers::Run End......")
	return nil
}

/* getBlockIdsWithGetHeadersMsg
 * getBlockIdsWithGetHeadersMsg 向指定节点发送getHeadersMsg，并根据收到的回复返回相应的error
 * 若未在规定时间内获取任何headers信息则返回超时错误
 * 若收到一个全零的返回，则表示对方节点里没有找到完整的区块头列表(包含该区间不在对方节点主干，该区间并不是对方账本某合法区间两种)
 * 若收到的列表长度小于请求长度，则证明上层整个获取区块头迭代完毕
 */
func getBlockIdsWithGetHeadersMsg(lk *LedgerKeeper, bcName, beginBlockId string, length int64, targetPeerAddr string) (bool, map[int]string, error) {
	body := &pb.GetHashesMsgBody{
		HashesCount:   length,
		HeaderBlockId: beginBlockId,
	}
	bodyBuf, _ := proto.Marshal(body)
	msg, err := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion2, bcName, "", xuper_p2p.XuperMessage_GET_HASHES, bodyBuf, xuper_p2p.XuperMessage_NONE)
	if err != nil {
		lk.log.Warn("getBlockIdsWithGetHeadersMsg::Generate GET_HASHES Message Error", "Error", err)
		return true, nil, ErrInternal
	}
	opts := []p2p_base.MessageOption{
		p2p_base.WithBcName(bcName),
		p2p_base.WithTargetPeerAddrs([]string{targetPeerAddr}),
	}
	lk.log.Trace("getBlockIdsWithGetHeadersMsg::Send GET_HASHES", "HEADER", beginBlockId)
	res, err := lk.P2pSvr.SendMessageWithResponse(context.Background(), msg, opts...)
	if err != nil {
		lk.log.Warn("getBlockIdsWithGetHeadersMsg::Sync Headers P2P Error: local error or target error", "Logid", msg.GetHeader().GetLogid(), "Error", err)
		return true, nil, ErrInternal
	}

	response := res[0]
	headerMsgBody := &pb.HashesMsgBody{}
	err = proto.Unmarshal(response.GetData().GetMsgInfo(), headerMsgBody)
	if err != nil {
		lk.log.Warn("getBlockIdsWithGetHeadersMsg::unmarshal error")
		return true, nil, ErrInvalidMsg
	}
	blockIds := headerMsgBody.GetBlockIds()
	lk.log.Info("getBlockIdsWithGetHeadersMsg::GET HEADERS RESULT", "HEADERS", blockIds)
	// 空值，包含连接错误
	if len(blockIds) == 0 {
		return true, nil, nil
	}
	if int64(len(blockIds)) > length {
		// 返回消息参数非法
		return true, nil, ErrInvalidMsg
	}
	headersInfo := map[int]string{}
	for i, blockId := range blockIds {
		headersInfo[i] = blockId
	}
	// 当前同步头的最后一次同步
	if int64(len(blockIds)) < length {
		// 若当前接受的区块哈希列表为全0，则表示对方无相应数据
		if blockIds[0] == zeroStr {
			return true, nil, ErrTargetDataNotFound
		}
		return true, headersInfo, nil
	}
	return false, headersInfo, nil
}

/* blocksDownloadWithHeadersList
 * blocksDownloadWithHeadersList 在节点列表中随机选取若干节点，并将headersList散列到不同的节点任务中，
 * 输入是一个headersList其包含连续的BlockIds区间
 * 本地节点向对应节点发送GetDataMsg消息，试图获取全部需要的block信息，并将其存入cahche中，
 * 一直循环直到区间被填满
 */
func (lt *LedgerTask) blocksDownloadWithHeadersList(lk *LedgerKeeper, headersList map[int]string) map[string]*SimpleBlock {
	// 同步map，放置连续区间内的所有区块指针
	syncMap := map[string]*SimpleBlock{}
	// cache并发读写操作时使用的锁
	syncBlockMutex := &sync.RWMutex{}
	if len(headersList) == 0 {
		return nil
	}
	// 若不收集齐会一直阻塞直到超时
	for {
		// 在targetPeers中随机选择peers个数
		validPeers := getValidPeersNumber(lk.peersStatusMap)
		if validPeers == 0 {
			lk.log.Warn("syncBlocksWithTipidAndPeers::all peer invalid")
			return nil
		}
		randomLen, err := rand.Int(rand.Reader, big.NewInt(validPeers+1))
		if err != nil {
			lk.log.Warn("blocksDownloadWithHeadersList::generate random numer error", "error", err)
			continue
		}
		// 随机选择Peers数目
		targetPeers, err := randomPickPeersWithNumber(randomLen.Int64(), lk.peersStatusMap)
		if err != nil {
			continue
		}
		// 散列headersList随机向被选取的peer分配BlockIds, 这些任务Blockid有可能部分已经完成同步，仅需关注未同步部分
		peersTask, err := assignTaskRandomly(targetPeers, headersList)
		if err != nil {
			continue
		}
		// 对于单个peer，先查看cache中是否有该区块，选择cache中没有的生成列表，向peer发送GetDataMsg
		wg := sync.WaitGroup{}
		wg.Add(len(peersTask))
		ch := make(chan struct{})
		for peer, headers := range peersTask {
			go func(peer string, headers []string, cache *map[string]*SimpleBlock) {
				defer wg.Done()
				crashFlag, err := lt.peerBlockDownloadTask(lk, peer, headers, cache, syncBlockMutex)
				if crashFlag {
					lk.log.Warn("syncBlocksWithTipidAndPeers::delete peer", "address", peer, "err", err)
					lk.peersStatusMap.Store(peer, false)
					return
				}
				if err != nil {
					lk.log.Warn("blocksDownloadWithHeadersList::peerBlockDownloadTask error", "error", err)
					return
				}
			}(peer, headers, &syncMap)
		}
		wg.Wait()
		close(ch)
		select {
		case <-ch:
			// 若headersList全部被填满，则返回success
			if len(headersList) == len(syncMap) {
				return syncMap
			}
			continue
		case <-time.After(SYNC_BLOCKS_TIMEOUT):
			break
		}
	}
}

/* peerBlockDownloadTask
 * peerBlockDownloadTask 向指定peer拉取指定区块列表，若该peer未返回任何块，则剔除节点，获取到的区块写入cache，上层逻辑判断是否继续拉取未获取的区块
 */
func (lt *LedgerTask) peerBlockDownloadTask(lk *LedgerKeeper, peerAddr string, taskBlockIds []string, cache *map[string]*SimpleBlock, syncBlockMutex *sync.RWMutex) (bool, error) {
	peerSyncMap := map[string]*SimpleBlock{}
	syncBlockMutex.RLock()
	for _, blockId := range taskBlockIds {
		if v, ok := (*cache)[blockId]; ok && v != nil {
			continue
		}
		peerSyncMap[blockId] = nil
	}
	syncBlockMutex.RUnlock()

	err := getBlocksWithGetDataMsg(lk, lk.bcName, peerAddr, &peerSyncMap)
	// 判断是否剔除peer
	if err != nil && err != ErrTargetDataNotEnough {
		return true, err
	}
	syncBlockMutex.Lock()
	for blockId, ptr := range peerSyncMap {
		(*cache)[blockId] = ptr
	}
	syncBlockMutex.Unlock()
	return false, err
}

/* getBlocksWithGetDataMsg
 * getBlocksWithGetDataMsg 输入一个map，该map key包含一个peer需要返回的blockId，和一个空指针，随后向特定peer发送GetDataMsg消息，以获取指定区块信息，
 * 若指定节点并未在规定时间内返回任何区块，则返回节点超时错误
 * 若指定节点仅返回部分区块，则返回缺失提醒
 */
func getBlocksWithGetDataMsg(lk *LedgerKeeper, bcName, targetPeer string, peerSyncMap *map[string]*SimpleBlock) error {
	if len(*peerSyncMap) == 0 {
		return nil
	}
	headersList := []string{}
	for key, _ := range *peerSyncMap {
		headersList = append(headersList, key)
	}
	body := &pb.GetBlocksMsgBody{
		BlockList: headersList,
	}
	bodyBuf, _ := proto.Marshal(body)
	msg, err := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion2, bcName, "", xuper_p2p.XuperMessage_GET_BLOCKS, bodyBuf, xuper_p2p.XuperMessage_NONE)
	if err != nil {
		lk.log.Warn("getBlocksWithGetDataMsg::Generate GET_BLOCKS Message Error", "Error", err)
		return ErrInternal
	}
	opts := []p2p_base.MessageOption{
		p2p_base.WithBcName(bcName),
		p2p_base.WithTargetPeerAddrs([]string{targetPeer}),
	}
	res, err := lk.P2pSvr.SendMessageWithResponse(context.Background(), msg, opts...)
	if err != nil {
		lk.log.Warn("getBlocksWithGetDataMsg::Sync GetBlocks P2P Error, local error or target error", "Logid", msg.GetHeader().GetLogid(), "Error", err)
		return ErrInternal
	}

	response := res[0]
	blocksMsgBody := &pb.BlocksMsgBody{}
	err = proto.Unmarshal(response.GetData().GetMsgInfo(), blocksMsgBody)
	if err != nil {
		return ErrUnmarshal
	}
	lk.log.Info("getBlocksWithGetDataMsg::GET BLOCKS RESULT", "Logid", blocksMsgBody.GetHeader().GetLogid(), "LEN", len(blocksMsgBody.GetBlocksInfo()))
	if len(blocksMsgBody.GetBlocksInfo()) == 0 && len(*peerSyncMap) > 1 {
		// 目标节点完全未找到任何block
		return ErrTargetDataNotFound
	}
	blocks := blocksMsgBody.GetBlocksInfo()
	for _, block := range blocks {
		blockId := global.F(block.GetBlockid())
		mapValue, ok := (*peerSyncMap)[blockId]
		if !ok || mapValue != nil {
			return ErrInvalidMsg
		}
		(*peerSyncMap)[blockId] = &SimpleBlock{
			internalBlock: block,
			header:        blocksMsgBody.GetHeader(),
		}
	}
	if len(blocksMsgBody.GetBlocksInfo()) < len(headersList) {
		// 目标节点并未在其本地找到所有需要的区块，需给上层返回缺失提醒
		return ErrTargetDataNotEnough
	}
	return nil
}

/* checkHeadersSafty
 * checkHeadersSafty对blockIds的安全性证明，例如基于pow的区块链blockids需要满足difficulty公式
 */
func checkHeadersSafty(blockIds []string) bool {
	return true
}

/* handleGetHeadersMsg response to the GetHeadersMsg with a HeadersMsg containing a list of the block-hashes required.
 * As a callee, peer checks whether the interval received is valid in its main-chain and then put the
 * corresponding block-hashes into the HeadersMsg.
 * When the callee cannot search the HEADER_HASH or the STOPPING_HASH of the GetHeadersMsg in its main-chain, it will
 * set the BLOCK_HASHES field to all zeroes to response to the caller.
 * handleGetHeadersMsg 接受GetHeadersMsg消息并返回，若GetHeadersMsg消息的消息区间在主干上，则直接返回区间所有的区块哈希列表，
 * 若不在主干，则返回一个全零消息。
 * 注意: 本次处理暂将消息区间不在主干在分支，以及账本无消息区间作为同样情况返回。
 */
func (lk *LedgerKeeper) handleGetHeadersMsg(ctx context.Context, msg *xuper_p2p.XuperMessage) (*xuper_p2p.XuperMessage, error) {
	bc := msg.GetHeader().GetBcname()
	if !p2p_base.VerifyDataCheckSum(msg) {
		lk.log.Error("handleGetHeadersMsg::verify msg error")
		return nil, ErrInvalidMsg
	}
	bodyBytes := msg.GetData().GetMsgInfo()
	body := &pb.GetHashesMsgBody{}
	if err := proto.Unmarshal(bodyBytes, body); err != nil {
		return nil, ErrUnmarshal
	}
	headersCount := body.GetHashesCount()
	if headersCount <= 0 {
		lk.log.Warn("handleGetHeadersMsg::Invalid headersCount, no service provided", "headersCount", headersCount)
		return nil, ErrInvalidMsg
	}
	nilHeaders := &pb.HashesMsgBody{
		BlockIds: []string{zeroStr},
	}
	nilBuf, err := proto.Marshal(nilHeaders)
	if err != nil {
		return nil, ErrUnmarshal
	}
	nilRes, err := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion2, bc, msg.GetHeader().GetLogid(), xuper_p2p.XuperMessage_HASHES, nilBuf, xuper_p2p.XuperMessage_NONE)
	if err != nil {
		return nil, ErrInternal
	}
	headerBlockId := body.GetHeaderBlockId()
	lk.log.Trace("handleGetHeadersMsg::GET_HEADER_MSG handling...", "BEGIN HEADER", headerBlockId)
	id, err := hex.DecodeString(headerBlockId)
	if err != nil {
		lk.log.Warn("handleGetHeadersMsg::Invalid header", "header", headerBlockId)
		return nil, ErrInvalidMsg
	}
	headerBlock, err := lk.ledger.QueryBlock(id)
	if err != nil {
		lk.log.Warn("handleGetHeadersMsg::internal error", "error", err, "headerBlockId", headerBlockId)
		return nilRes, nil
	}
	if !headerBlock.GetInTrunk() {
		lk.log.Warn("handleGetHeadersMsg::not in trunck", "headerBlock", headerBlockId)
		return nilRes, nil
	}
	// 循环获取下一个block，并批量放入Cache中
	resultBlocks := []*pb.InternalBlock{}
	beginHeight := headerBlock.GetHeight()
	for i := int64(1); i <= headersCount; i++ {
		block, err := lk.ledger.QueryBlockByHeight(beginHeight + i)
		if err != nil {
			lk.log.Warn("handleGetHeadersMsg::QueryBlock error", "error", err)
			break
		}
		resultBlocks = append(resultBlocks, block)
	}
	if len(resultBlocks) == 0 {
		lk.log.Warn("handleGetHeadersMsg::resultBlocks LEN=0")
		return nilRes, nil
	}
	resultHeaders := &pb.HashesMsgBody{
		BlockIds: []string{},
	}
	for _, block := range resultBlocks {
		lk.fastFetchBlockCache.Add(global.F(block.GetBlockid()), block)
		resultHeaders.BlockIds = append(resultHeaders.BlockIds, global.F(block.GetBlockid()))
	}
	lk.log.Info("handleGetHeadersMsg::GET_HEADER_MSG response...", "response res", resultHeaders.BlockIds)
	resBuf, _ := proto.Marshal(resultHeaders)
	res, err := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion2, bc, msg.GetHeader().GetLogid(),
		xuper_p2p.XuperMessage_HASHES, resBuf, xuper_p2p.XuperMessage_SUCCESS)
	return res, err
}

/* handleGetDataMsg response to the GetDataMsg with a BlocksMsg containing the internal-blocks required.
 * Peer set the TOTAL_NUMBERS field to the sum of blocks it has as the form of an answer, which shows the local status of
 * its main-chain when sync blocks and the caller, at the same time, can find out if it could get the whole blocks
 * with a one-time request, otherwise it will send the GetDataMsg to another peer.
 * handleGetDataMsg 接受GetDataMsg，解析访问者需要的blockId列表，并发送相应的blocks回去，处理节点会返回一个TOTAL_MUMBERS作为对这一个GetDataMsg
 * 的整体回应，若访问者需要的单次消息的总区块数为N，而回应者仅有M个区块(M < N)，访问者会向其他节点请求剩余区块。
 * 注意：本次处理仅设定一种错误节点的发现规则：当回应者在规定时间内返回0个区块且这些区块在访问其他节点后获取到值时。以及返回的区块中发现验证错误时。
 */
func (lk *LedgerKeeper) handleGetDataMsg(ctx context.Context, msg *xuper_p2p.XuperMessage) (*xuper_p2p.XuperMessage, error) {
	bc := msg.GetHeader().GetBcname()
	if !p2p_base.VerifyDataCheckSum(msg) {
		return nil, ErrInvalidMsg
	}
	bodyBytes := msg.GetData().GetMsgInfo()
	body := &pb.GetBlocksMsgBody{}
	if err := proto.Unmarshal(bodyBytes, body); err != nil {
		return nil, ErrUnmarshal
	}
	lk.log.Trace("handleGetDataMsg::GET_BLOCKS_MSG handling...", "REQUIRE LIST", body.GetBlockList())
	resultBlocks := []*pb.InternalBlock{}
	for _, blockId := range body.GetBlockList() {
		if value, hit := lk.fastFetchBlockCache.Get(blockId); hit {
			block := value.(*pb.InternalBlock)
			resultBlocks = append(resultBlocks, block)
			continue
		}
		id, err := hex.DecodeString(blockId)
		if err != nil {
			lk.log.Error("handleGetDataMsg::Invalid header", "header", blockId)
			return nil, ErrInvalidMsg
		}
		block, err := lk.ledger.QueryBlock(id)
		if err != nil {
			continue
		}
		resultBlocks = append(resultBlocks, block)
		lk.fastFetchBlockCache.Add(global.F(block.GetBlockid()), block)
	}
	// peer自己通过区块大小切分Data消息返回， 按照尽可能多的返回区块规则选取区块
	resultBlocks = pickBlocksForBlocksMsg(lk.maxBlocksMsgSize, resultBlocks)
	result := &pb.BlocksMsgBody{
		Header: &pb.Header{
			Logid: msg.GetHeader().GetLogid(),
		},
		BlocksInfo: resultBlocks,
	}
	resBuf, _ := proto.Marshal(result)
	lk.log.Info("handleGetDataMsg::GET_BLOCKS_MSG response...", "Logid", result.Header.Logid, "LEN", len(resultBlocks))
	res, err := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion2, bc, msg.GetHeader().GetLogid(), xuper_p2p.XuperMessage_BLOCKS, resBuf, xuper_p2p.XuperMessage_SUCCESS)
	return res, err
}

/* pickBlocksForBlocksMsg
 * pickBlocksForBlocksMsg 根据目标size尽可能多的选取区块，返回区块列表
 */
func pickBlocksForBlocksMsg(maxSize int64, blockList []*pb.InternalBlock) []*pb.InternalBlock {
	list := []int{}
	for _, block := range blockList {
		size := proto.Size(block)
		list = append(list, size)
	}
	indexes := pickIndexesWithTargetSize(maxSize, list)
	result := []*pb.InternalBlock{}
	for _, v := range indexes {
		result = append(result, blockList[v])
	}
	return result
}

/* pickIndexesWithTargetSize
 * pickIndexesWithTargetSize 尽可能选择多的区块返回
 */
func pickIndexesWithTargetSize(targetSize int64, list []int) []int {
	result := make([]int, 0)
	sizeMap := make(map[int][]int, 0)
	indexMap := make(map[int]int, 0)
	for i, value := range list {
		if _, ok := sizeMap[value]; !ok {
			l := []int{i}
			sizeMap[value] = l
			indexMap[value] = 0
			continue
		}
		sizeMap[value] = append(sizeMap[value], i)
	}
	sort.Ints(list)
	for i := 0; i < len(list) && targetSize-int64(list[i]) >= int64(0); i++ {
		index := sizeMap[list[i]][indexMap[list[i]]]
		result = append(result, index)
		targetSize -= int64(list[i])
		indexMap[list[i]]++
	}
	return result
}

/* getValidPeersNumber
 * getValidPeersNumber 返回目前peers列表可用节点总数
 */
func getValidPeersNumber(peers *sync.Map) int64 {
	number := int64(0)
	peers.Range(func(key, value interface{}) bool {
		valid := value.(bool)
		if valid {
			number++
		}
		return true
	})
	return number
}

/* randomPickPeersWithNumber
 * randomPickPeersWithNumber 从现有peersStatusMap中可连接的peers中随机选取number个作为目标节点
 */
func randomPickPeersWithNumber(number int64, peers *sync.Map) ([]string, error) {
	if number == 0 {
		return nil, nil
	}
	originPeers := []string{}
	peers.Range(func(key, value interface{}) bool {
		peer := key.(string)
		valid := value.(bool)
		if valid {
			originPeers = append(originPeers, peer)
		}
		return true
	})
	if number > int64(len(originPeers)) {
		return nil, ErrInvalidMsg
	}
	selection := number
	remains := len(originPeers)
	result := make([]int, number)
	for i := 0; i < len(originPeers); i++ {
		if random, err := rand.Int(rand.Reader, big.NewInt(int64(remains))); err == nil {
			if random.Int64() < selection {
				result[number-selection] = i
				selection--
			}
			remains--
		}
	}
	resultPeers := []string{}
	for _, value := range result {
		resultPeers = append(resultPeers, originPeers[value])
	}
	if len(resultPeers) == 0 {
		return nil, ErrInternal
	}
	return resultPeers, nil
}

/* assignTaskRandomly
 * assignTaskRandomly 随机将需要处理的blockId请求分配给指定的peers
 */
func assignTaskRandomly(targetPeers []string, headersList map[int]string) (map[string][]string, error) {
	if len(targetPeers) == 0 {
		return nil, ErrInvalidMsg
	}
	assignMap := map[uint32][]string{}
	for _, blockId := range headersList {
		input, err := hex.DecodeString(blockId)
		if err != nil {
			return nil, ErrInternal
		}
		var index uint32
		for j := 0; j+SIZE_OF_UINT32 < len(input); j = j + SIZE_OF_UINT32 {
			pos := binary.BigEndian.Uint32(input[j : j+SIZE_OF_UINT32])
			index = index + pos
		}
		index = index % uint32(len(targetPeers))
		if _, ok := assignMap[index]; !ok {
			assignMap[index] = []string{blockId}
			continue
		}
		assignMap[index] = append(assignMap[index], blockId)
	}
	peersTask := map[string][]string{}
	for peerIndex, taskList := range assignMap {
		peersTask[targetPeers[int(peerIndex)]] = taskList
	}
	return peersTask, nil
}

/* changeSyncBeginBlockPoint
 * changeSyncBeginBlockPoint 当当前beginBlockId无法获取同步头列表时，需要通过输入账本回溯获取新的BlockId
 * 当前提供一种方法，向前回溯一个高度，TODO: 二分查找
 */
func changeSyncBeginPointBackwards(beginBlock *pb.InternalBlock) string {
	return global.F(beginBlock.GetPreHash())
}

/* confirmBlocks 原SendBlock逻辑
 * 账本接受blockMap数据，代替原来的PendingBlocks，正向confirm block
 */
func (lk *LedgerKeeper) confirmBlocks(hd *global.XContext, originBegin, headerBegin string, blocksMap map[string]*SimpleBlock, headersInfo map[int]string, endFlag bool) (string, bool, error) {
	// 取这段新链的第一个区块，判断走账本分叉逻辑还是直接账本追加逻辑
	var err error
	newBegin := headerBegin
	index := 0
	beginSimpleBlock := blocksMap[headersInfo[index]]
	if beginSimpleBlock == nil {
		return newBegin, false, ErrTargetDataNotFound
	}
	lk.log.Debug("ConfirmBlocks", "genesis", global.F(lk.ledger.GetMeta().GetRootBlockid()), "utxo", global.F(lk.utxovm.GetLatestBlockid()),
		"len(blocksMap)", len(blocksMap), "len(headersInfo)", len(headersInfo), "cost", hd.Timer.Print())
	listLen := len(headersInfo)
	if listLen == 0 {
		return newBegin, false, ErrInternal
	}
	// 更新的节点是否是账本主干上的末枝，或者是账本上的一个分叉, originBegin表示task初始值，headerBegin表示同步后初始值
	noFork := global.F(beginSimpleBlock.internalBlock.GetPreHash()) == originBegin
	if noFork {
		lk.log.Debug("ConfirmBlocks::Equal The Same", "cost", hd.Timer.Print())
		needVerify := (lk.nodeMode == config.NodeModeFastSync)
		for ; index < listLen; index++ {
			needRepost := (index == listLen-1) && endFlag
			checkBlock := blocksMap[headersInfo[index]]
			if checkBlock == nil {
				break
			}
			nextBlockid := global.F(checkBlock.internalBlock.GetBlockid())
			err = lk.confirmAppendingBlock(checkBlock, needRepost, needVerify)
			if err != nil && err != ErrBlockExist {
				lk.log.Debug("ConfirmBlocks::confirmAppendingBlock error", "err", err, "PreCheckBlock", checkBlock, "cost", hd.Timer.Print())
				break
			}
			newBegin = nextBlockid
		}
		if index < 1 {
			lk.log.Debug("ConfirmBlocks::confirm error", "err", err, "cost", hd.Timer.Print())
			return newBegin, false, err
		}
		simpleBlock := blocksMap[headersInfo[index-1]]
		b := simpleBlock.internalBlock
		err = lk.con.ProcessConfirmBlock(b)
		if err != nil {
			lk.log.Debug("ConfirmBlocks::ProcessConfirmBlock error", "logid", simpleBlock.header.GetLogid(), "error", err, "cost", hd.Timer.Print())
		}
		lk.log.Debug("ConfirmBlocks::Equal The Same, confirm blocks finish", "newBegin", newBegin, "cost", hd.Timer.Print())
		return newBegin, index == listLen, err
	}
	//交点不等于utxo latest block
	lk.log.Debug("XXXXXXXXX The NO Same", "cost", hd.Timer.Print())
	for ; index < listLen; index++ {
		checkBlock := blocksMap[headersInfo[index]]
		if checkBlock == nil {
			break
		}
		nextBlockid := global.F(checkBlock.internalBlock.GetBlockid())
		err, trunkSwitch := lk.confirmForkingBlock(checkBlock)
		if err != nil && err != ErrBlockExist {
			break
		}
		if trunkSwitch {
			err := lk.utxovm.Walk(checkBlock.internalBlock.GetBlockid(), false)
			lk.log.Debug("ConfirmBlocks::Walk Time", "logid", checkBlock.header.GetLogid(), "cost", hd.Timer.Print())
			if err != nil {
				lk.log.Warn("ConfirmBlocks::Walk error", "logid", checkBlock.header.GetLogid(), "err", err, "cost", hd.Timer.Print())
				break
			}
		}
		newBegin = nextBlockid
	}
	// 待块确认后, 共识执行相应的操作
	if index < 1 {
		return newBegin, false, nil
	}
	err = lk.con.ProcessConfirmBlock(blocksMap[headersInfo[index-1]].internalBlock)
	if err != nil {
		lk.log.Debug("ConfirmBlocks::ProcessConfirmBlock error", "error", err, "cost", hd.Timer.Print())
	}
	lk.log.Debug("ConfirmBlocks::XXXXXXXXX The NO Same, confirm blocks finish", "newBegin", newBegin, "cost", hd.Timer.Print())
	return newBegin, index == listLen, err
}

/* confirmAppendingBlock 直接试图在账本尾追加Block
 */
func (lk *LedgerKeeper) confirmAppendingBlock(simpleBlock *SimpleBlock, needRepost, needVerify bool) error {
	block := simpleBlock.internalBlock
	if int64(proto.Size(block)) > lk.maxBlocksMsgSize {
		lk.log.Debug("ConfirmSingleBlockFromBatch:: Large block error", "logid", simpleBlock.header.Logid)
		return ErrInvalidMsg
	}
	// 如果已经存在，则立即返回
	if lk.ledger.ExistBlock(block.GetBlockid()) {
		lk.log.Debug("ConfirmSingleBlockFromBatch::Block exist", "logid", simpleBlock.header.Logid)
		return ErrBlockExist
	}
	for idx, tx := range block.Transactions {
		if !lk.ledger.IsValidTx(idx, tx, block) {
			lk.log.Warn("ConfirmSingleBlockFromBatch::invalid tx got from the block", "logid", simpleBlock.header.Logid, "txid", global.F(tx.Txid), "blkid", global.F(block.GetBlockid()))
			return ErrInvalidMsg
		}
	}
	// 区块加解密有效性检查
	if needVerify {
		if res, err := lk.con.CheckMinerMatch(simpleBlock.header, block); !res {
			lk.log.Warn("ConfirmAppendingBlock::check miner error", "logid", simpleBlock.header.Logid, "error", err)
			return ErrServiceRefused
		}
	}
	cs := lk.ledger.ConfirmBlock(block, false)
	if !cs.Succ {
		lk.log.Warn("ConfirmAppendingBlock::confirm error", "logid", simpleBlock.header.Logid)
		return ErrConfirmBlock
	}
	// 判断是否是最新区块及最长链，若是则最新区块需广播
	err := lk.utxovm.PlayAndRepost(block.Blockid, needRepost, false)
	lk.log.Debug("ConfirmAppendingBlock::Play", "logid", simpleBlock.header.Logid)
	if err != nil {
		lk.log.Warn("ConfirmAppendingBlock::utxo vm play err", "logid", simpleBlock.header.Logid, "err", err)
		return ErrUTXOVMPlay
	}
	return nil
}

/* confirmForkingBlock 分叉逻辑，当获取链为原主干叉链时的确认逻辑
 */
func (lk *LedgerKeeper) confirmForkingBlock(simpleBlock *SimpleBlock) (error, bool) {
	block := simpleBlock.internalBlock
	if int64(proto.Size(block)) > lk.maxBlocksMsgSize {
		lk.log.Debug("ConfirmSingleBlockFromBatch:: Large block error", "logid", simpleBlock.header.Logid)
		return ErrInvalidMsg, false
	}
	// 如果已经存在，则立即返回
	if lk.ledger.ExistBlock(block.GetBlockid()) {
		lk.log.Debug("ConfirmSingleBlockFromBatch::Block exist", "logid", simpleBlock.header.Logid)
		return ErrBlockExist, false
	}
	for idx, tx := range block.Transactions {
		if !lk.ledger.IsValidTx(idx, tx, block) {
			lk.log.Warn("ConfirmSingleBlockFromBatch::invalid tx got from the block", "logid", simpleBlock.header.Logid, "txid", global.F(tx.Txid), "blkid", global.F(block.GetBlockid()))
			return ErrInvalidMsg, false
		}
	}
	if res, err := lk.con.CheckMinerMatch(simpleBlock.header, block); !res {
		lk.log.Warn("ConfirmSingleBlockFromBatch::check miner error", "logid", simpleBlock.header.Logid, "error", err)
		return ErrServiceRefused, false
	}
	cs := lk.ledger.ConfirmBlock(block, false)
	if !cs.Succ {
		lk.log.Warn("ConfirmSingleBlockFromBatch::confirm error", "logid", simpleBlock.header.Logid)
		return ErrConfirmBlock, false
	}
	//是否发生主干切换
	trunkSwitch := (cs.TrunkSwitch || block.InTrunk)
	return nil, trunkSwitch
}

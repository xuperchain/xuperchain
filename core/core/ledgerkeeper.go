package xchaincore

import (
	"bytes"
	"container/list"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	math_rand "math/rand"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	log "github.com/xuperchain/log15"
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
	// ErrTargetPeerInvalid peer timeout
	ErrTargetPeerInvalid = errors.New("target peer cannot response to the requests on time")
	// ErrInternal p2p internal error
	ErrInternal = errors.New("cannot sent msg because of internal error")
	// ErrHaveNoTargetData the callee cannot find data in main-chain
	ErrTargetDataNotFound = errors.New("cannot search the specific data the caller wants")
	// ErrTargetDataNotEnough the callee cannot find enough data
	ErrTargetDataNotEnough = errors.New("peers cannot search the whole data the caller wants, not enough")
	// ErrPeersInvalid p2p peers invalid
	ErrAllPeersInvalid = errors.New("All the peers are invalid")
	// ErrPeerFinish no more new block valid
	ErrPeerFinish = errors.New("Node has become main-chain-holder in the network and no more new blockId is valid.")
)

const (
	SIZE_OF_UINT32          = 4
	SYNC_BLOCKS_TIMEOUT     = 10
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
	tList *list.List  // 存放具体task
	mutex *sync.Mutex // mutex保护tList的并发操作
}

/* NewTasksList return a link queue TasksList
 * NewTasksList 返回一个包含链表和链表所含元素组成的set的数据结构
 */
func NewTasksList() *TasksList {
	return &TasksList{
		tList: list.New(),
		mutex: &sync.Mutex{},
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
func (t *TasksList) PushBack(ledgerTask *LedgerTask) bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.tList.PushBack(ledgerTask)
	return true
}

/* Remove remove the target element in tList of taskslist
 * Remove 删除链表中元素，并对应删除set
 */
func (t *TasksList) Remove(e *list.Element) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.tList.Remove(e)
}

/* Front get the front element in tList of taskslist
 * Front获取链表头元素
 */
func (t *TasksList) Front() *list.Element {
	t.mutex.Lock()
	defer t.mutex.Unlock()
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

type syncTaskManager struct {
	syncingTasks   *TasksList // 账本同步batch操作
	appendingTasks *TasksList // 账本同步追加操作
	syncMaxSize    int        // sync队列最大size
	log            log.Logger
	taskId         int64
}

/* newSyncTaskManager 生成一个新SyncTaskManager，管理所有任务队列
 * 任务队列包含两个队列，syncingTasks用于记录多个block的同步任务，
 * appendingTasks用于记录收到广播新块，且新块直接为账本下一高度的情况，该任务直接尝试写账本
 */
func newSyncTaskManager(log log.Logger) *syncTaskManager {
	return &syncTaskManager{
		syncingTasks:   NewTasksList(),
		appendingTasks: NewTasksList(),
		syncMaxSize:    MAX_TASK_MN_SIZE,
		log:            log,
	}
}

// generateTaskid generate a random id.
func (stm *syncTaskManager) GenerateTaskid() string {
	stm.taskId++
	t := time.Now().UnixNano()
	return fmt.Sprintf("%d_%d", t, stm.taskId)
}

/* put 由manager操作，直接操作所handle的各个队列
 */
func (stm *syncTaskManager) Put(ledgerTask *LedgerTask) bool {
	var q *TasksList
	switch ledgerTask.GetAction() {
	case Syncing:
		q = stm.syncingTasks
	case Appending:
		q = stm.appendingTasks
	default:
		stm.log.Error("SyncTaskManager::task error", "type", ledgerTask.GetAction())
		return false
	}
	stm.log.Trace("SyncTaskManager put task", "len", stm.syncingTasks.Len())
	if q.Len() > stm.syncMaxSize {
		stm.log.Trace("SyncTaskManager put task err, too much task, refuse it")
		return false
	}
	q.PushBack(ledgerTask)
	return true
}

/* get 由manager操作从所有队列中挑选一个队列拿一个任务
 */
func (stm *syncTaskManager) Pop() *LedgerTask {
	var queue *TasksList
	// 在同步逻辑中，最新块广播的优先级高于其他
	if stm.appendingTasks.Len() > 0 {
		queue = stm.appendingTasks
	} else {
		queue = stm.syncingTasks
	}
	if queue.Len() == 0 {
		stm.log.Debug("SyncTaskManager::queues' Len == 0")
		return nil
	}
	element := queue.Front()
	queue.Remove(element)
	stm.log.Debug("SyncTaskManager::", "Len", queue.Len())
	return element.Value.(*LedgerTask)
}

// fixTask 清洗队列中失效的任务
func (stm *syncTaskManager) ClearTask(targetHeight int64) {
	stm.syncingTasks.fix(targetHeight)
	stm.appendingTasks.fix(targetHeight)
}

type LedgerKeeper struct {
	p2pSvr           p2p_base.P2PServer
	log              log.Logger
	syncMsgChan      chan *xuper_p2p.XuperMessage
	peersStatusMap   *sync.Map // map[string]bool 更新同步节点的p2p列表活性
	maxBlocksMsgSize int64     // 取最大区块大小
	ledger           *ledger.Ledger
	bcName           string
	syncTaskMg       *syncTaskManager
	nodeMode         string

	utxovm *utxo.UtxoVM
	con    *consensus.PluggableConsensus
	// 该锁保护同一时间内只有矿工or账本keeper对象中的一个对ledger及utxovm操作
	// ledgledgerKeeper同步块和xchaincore 矿工doMiner抢锁
	coreMutex sync.RWMutex
}

type LedgerTask struct {
	taskId string
	action KeeperStatus
	// targetBlockId []byte
	targetHeight int64
	ctx          *LedgerTaskContext
}

type LedgerTaskContext struct {
	extBlocks  []*SimpleBlock
	preferPeer []string
	hd         *global.XContext
}

type SimpleBlock struct {
	internalBlock *pb.InternalBlock
	logid         string
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
func (lt *LedgerTask) GetExtBlocks() []*SimpleBlock {
	if lt.ctx == nil || lt.ctx.extBlocks == nil {
		return nil
	}
	return lt.ctx.extBlocks
}

/* GetPreferPeer get the callee of the task.
 * GetPreferPeer 获取向其余节点发送请求信息时指定的Peer地址
 */
func (lt *LedgerTask) GetPreferPeer() []string {
	if lt.ctx == nil || lt.ctx.preferPeer == nil {
		return nil
	}
	return lt.ctx.preferPeer
}

func (lt *LedgerTask) setPreferPeer(addr string) {
	lt.ctx.preferPeer = append(lt.ctx.preferPeer, addr)
}

// GetXCtx get the context of the task.
func (lt *LedgerTask) GetXContext() *global.XContext {
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
		p2pSvr:           p2pV2,
		log:              slog,
		syncMsgChan:      make(chan *xuper_p2p.XuperMessage, CHAN_SIZE),
		maxBlocksMsgSize: ledger.GetMaxBlockSize(),
		peersStatusMap:   new(sync.Map),
		ledger:           ledger,
		bcName:           bcName,
		utxovm:           utxovm,
		con:              con,
		syncTaskMg:       newSyncTaskManager(slog),
		nodeMode:         nodeMode,
	}
	lk.log.Trace("LedgerKeeper Start to Register Subscriber")
	if _, err := lk.p2pSvr.Register(lk.p2pSvr.NewSubscriber(nil, xuper_p2p.XuperMessage_GET_BLOCKIDS, lk.handleGetHeadersMsg, "", lk.log)); err != nil {
		return lk, err
	}
	if _, err := lk.p2pSvr.Register(lk.p2pSvr.NewSubscriber(nil, xuper_p2p.XuperMessage_GET_BLOCKS, lk.handleGetDataMsg, "", lk.log)); err != nil {
		return lk, err
	}
	lk.updatePeerStatusMap()
	return lk, nil
}

// CoreLock 锁coreMutex
func (lk *LedgerKeeper) CoreLock() {
	lk.coreMutex.Lock()
}

// CoreUnlock 解锁coreMutex
func (lk *LedgerKeeper) CoreUnlock() {
	lk.coreMutex.Unlock()
}

/* DoTruncateTask Truncate truncate ledger and set tipblock to utxovmLastID
 * 封装原来的ledger Truncate(), xchaincore使用此函数执行裁剪
 * TODO：共识部分也s执行了裁剪工作，但使用的是ledger的Truncate()，后续需要使用LedgerKeeper替代
 */
func (lk *LedgerKeeper) DoTruncateTask(utxovmLastID []byte) error {
	err := lk.ledger.Truncate(utxovmLastID)
	if err != nil {
		return err
	}
	lk.syncTaskMg.ClearTask(lk.ledger.GetMeta().GetTrunkHeight())
	return nil
}

/* PutTask create a ledger task and put a task into LedgerKeeper's queues.
 * PutTask 新建一个账本任务，需输入目标blockid(必须), 目标blockid所在的高度(必须),
 * 任务类型(必须: Appending追加、Syncing批量同步、Truncating裁剪),
 * 任务上下文(可选), 任务上下文包括	extBlocks map[string]*SimpleBlock，追加账本时附带的追加block信息
 *	   preferPeer []string，询问邻居节点的地址
 *	   hd *global.XContext，计时context
 */
func (lk *LedgerKeeper) PutTask(targetHeight int64, action KeeperStatus, ctx *LedgerTaskContext) {
	ledgerTask := &LedgerTask{
		targetHeight: targetHeight,
		action:       action,
		ctx:          ctx,
		taskId:       lk.syncTaskMg.GenerateTaskid(),
	}
	lk.syncTaskMg.Put(ledgerTask)
}

// Start start ledgerkeeper's loop
func (lk *LedgerKeeper) Start() {
	go lk.startTaskLoop()
}

// startTaskLoop get tasks from ledgerkeeper's queues in the loop.
func (lk *LedgerKeeper) startTaskLoop() {
	for {
		lk.log.Trace("StartTaskLoop::Start......")
		lk.updatePeerStatusMap()
		task, ok := lk.getTask()
		if !ok {
			time.Sleep(time.Duration(math_rand.Intn(KEEPER_SLEEP_MIL_SECOND)) * time.Millisecond)
			continue
		}
		action := task.GetAction()
		lk.log.Trace("StartTaskLoop::Get a task", "action", action, "taskId", task.taskId)
		switch action {
		case Syncing:
			lk.handleSyncTask(task)
		case Appending:
			lk.handleAppendTask(task)
		}
	}
}

// getTask get a task from LedgerKeeper's queues.
func (lk *LedgerKeeper) getTask() (*LedgerTask, bool) {
	task := lk.syncTaskMg.Pop()
	if task == nil {
		return nil, false
	}
	return task, true
}

/* updatePeerStatusMap 更新LedgerKeeper的syncMap
 * TODO: 后续完善Peer节点维护逻辑，节点不一定要永久剔除
 */
func (lk *LedgerKeeper) updatePeerStatusMap() {
	for _, id := range lk.p2pSvr.GetPeersConnection() {
		local, loop := lk.p2pSvr.GetLocalUrl()
		if id == local || id == loop {
			continue
		}
		if _, ok := lk.peersStatusMap.Load(id); ok {
			continue
		}
		lk.peersStatusMap.Store(id, true)
		lk.log.Trace("Init::", "id", id)
	}
}

/* appendingBlock 直接在账本尾追加操作
 * 在广播新块时使用，当有新块产生时，且该新块恰好为本地账本的下一区块，此时直接试图向账本进行写操作，无需同步流程
 */
func (lk *LedgerKeeper) handleAppendTask(lt *LedgerTask) error {
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
	newBegin, ok, err := lk.confirmBlocks(lt.GetXContext(), lt.GetExtBlocks(), true)
	lk.log.Trace("appendingBlock::AppendTask try to append ledger directly", "task", lt.taskId, "newbegin", global.F(newBegin), "ok", ok, "error", err)
	return nil
}

/* handleSyncTask
 * handleSyncTask 请求消息头同步逻辑
 * 输入请求节点列表和起始区块哈希，迭代完成同步
 * 该函数首先完成区块头同步工作，发送GetHHashesMsg给指定peer，试图获取区间内所有区块哈希值
 * 获取到全部区块哈希列表之后，本节点将列表散列成若干份，并向指定列表节点发起同步具体区块工作，发送GetDataMsg请求，试图获取对应的所有详细区块消息
 * 若上一步并未在指定时间内获取到所有区块，则继续更换节点列表，该过程一直阻塞，直到获得所有区块，或者在超时后退出
 * 在该同步过程中顺便标注错误peer
 * 完成一个迭代后，task会向ledger中写数据，同时判断是否需要切换主干并完成写任务
 */
func (lk *LedgerKeeper) handleSyncTask(lt *LedgerTask) error {
	nextLoop := true
	// ATTENTION: 此处可能出现脏读，矿工后续更新了一个新区块到账本，而此次向外同步区块任务获取的区块s可能无用，但该操作可容忍
	headerBegin := lk.utxovm.GetLatestBlockid()
	headerBeginStr := global.F(headerBegin)
	// 同步头过程
	lk.log.Trace("handleSyncTask::Run......", "task", lt.taskId, "headerBegin", headerBeginStr)
	for nextLoop {
		if getValidPeersNumber(lk.peersStatusMap) == 0 {
			lk.log.Warn("handleSyncTask::getValidPeersNumber=0", "task", lt.taskId, "headerBegin", headerBeginStr)
			return ErrAllPeersInvalid
		}
		// 先随机选择一个peer进行询问，找其询问最新的blockids列表, 若走回溯逻辑，则直接选取回溯传入的peer
		peer := lt.GetPreferPeer()
		if peer == nil {
			var err error
			peer, err = randomPickPeers(1, lk.peersStatusMap)
			if err != nil {
				lk.log.Warn("handleSyncTask::randomPickPeers error", "task", lt.taskId, "headerBegin", headerBeginStr, "err", err)
				continue
			}
			lk.log.Trace("randomPickPeers", "peer", peer[0], "task", lt.taskId, "headerBegin", headerBeginStr)
		}
		endFlag, blockIds, err := lk.getPeerBlockIds(headerBegin, HEADER_SYNC_SIZE, peer[0])
		lk.log.Info("handleSyncTask::getPeerBlockIds result", "task", lt.taskId, "headerBegin", headerBeginStr, "NEWpeer", peer[0], "err", err, "endFlag", endFlag)
		if err == ErrPeerFinish {
			// 本账本已达到最新区块高度，消解此任务
			return nil
		}
		if err == ErrTargetDataNotFound {
			// beginBlockId疑似无效，需往前回溯，注意:往前回溯可能会导致主干切换
			lk.log.Trace("handleSyncTask::get nothing from peers, begin backtracking...", "task", lt.taskId, "headerBegin", headerBeginStr)
			block, err := lk.ledger.QueryBlock(headerBegin)
			if err != nil {
				return ErrInternal
			}
			headerBegin = changeSyncBeginPointBackwards(block)
			if headerBegin == nil {
				return nil // genesisBlock有问题，暂不解决
			}
			headerBeginStr = global.F(headerBegin)
			// 回溯逻辑直接向全零返回的peer发送GetBlockIdsRequest
			lt.setPreferPeer(peer[0])
			lk.log.Info("handleSyncTask::backtrack start point", "task", lt.taskId, "headerBegin", headerBeginStr)
			continue
		}
		if err == ErrInternal {
			return err
		}
		// ATTENTION:上一次询问失败，因此此处重试
		if err == ErrTargetPeerInvalid {
			lk.log.Info("handleSyncTask::peer failed, try again", "task", lt.taskId, "headerBegin", headerBeginStr, "peer", peer[0])
			continue
		}
		if err != nil {
			lk.log.Warn("handleSyncTask::delete peer", "task", lt.taskId, "headerBegin", headerBeginStr, "address", peer[0], "err", err)
			lk.peersStatusMap.Store(peer[0], false)
			continue
		}
		if len(blockIds) == 0 {
			return nil
		}
		if endFlag {
			nextLoop = false
		}
		// TODO: 后续可加入CheckHeaderSafety()对blockIds的安全性证明，例如基于pow的区块链blockids需要满足difficulty公式
		blocksSlice := lk.downloadPeerBlocks(blockIds) // blocksMap中可能有key存在，但值为nil的值，此值标示本地账本已含数据
		if blocksSlice == nil {
			// 此处直接return, 加速task消耗
			return nil
		}
		// 本轮同步结束，开始写账本
		newBegin, _, err := lk.confirmBlocks(lt.GetXContext(), blocksSlice, endFlag)
		if err != nil {
			lk.log.Warn("handleSyncTask::ConfirmBlocks error", "err", err)
			return nil
		}
		headerBegin = newBegin
	}
	lk.log.Trace("handleSyncTask::Run End......")
	return nil
}

/* getPeerBlockIds
 * getPeerBlockIds 向指定节点发送getHeadersMsg，并根据收到的回复返回相应的error
 * 若未在规定时间内获取任何headers信息则返回超时错误
 * 若收到一个全零的返回，则表示对方节点里没有找到完整的区块头列表(包含该区间不在对方节点主干，该区间并不是对方账本某合法区间两种)
 * 若收到的列表长度小于请求长度，则证明上层整个获取区块头迭代完毕
 */
func (lk *LedgerKeeper) getPeerBlockIds(beginBlockId []byte, length int64, targetPeerAddr string) (bool, [][]byte, error) {
	body := &pb.GetBlockIdsRequest{
		Count:   length,
		BlockId: beginBlockId,
	}
	bodyBuf, _ := proto.Marshal(body)
	msg, err := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion2, lk.bcName, "", xuper_p2p.XuperMessage_GET_BLOCKIDS, bodyBuf, xuper_p2p.XuperMessage_NONE)
	if err != nil {
		lk.log.Warn("getPeerBlockIds::Generate GET_HASHES Message Error", "Error", err)
		return true, nil, ErrInternal
	}
	opts := []p2p_base.MessageOption{
		p2p_base.WithBcName(lk.bcName),
		p2p_base.WithTargetPeerAddrs([]string{targetPeerAddr}),
	}
	lk.log.Trace("getPeerBlockIds::Send GET_HASHES", "HEADER", global.F(beginBlockId))
	res, err := lk.p2pSvr.SendMessageWithResponse(context.Background(), msg, opts...)
	if err != nil {
		lk.log.Warn("getPeerBlockIds::Sync Headers P2P Error: local error or target error", "Logid", msg.GetHeader().GetLogid(), "Error", err)
		return true, nil, ErrInternal
	}

	response := res[0]
	headerMsgBody := &pb.GetBlockIdsResponse{}
	err = proto.Unmarshal(response.GetData().GetMsgInfo(), headerMsgBody)
	if err != nil {
		lk.log.Warn("getPeerBlockIds::unmarshal error")
		return true, nil, ErrInvalidMsg
	}
	blockIds := headerMsgBody.GetBlockIds()
	tip := headerMsgBody.GetTipBlockId()
	printStr := ""
	for _, id := range blockIds {
		printStr += global.F(id) + " "
	}
	lk.log.Info("getPeerBlockIds::GET HEADERS RESULT", "HEADERS", printStr, "TIP", global.F(tip))

	// 空值，包含连接错误
	if len(blockIds) == 0 && tip == nil {
		return false, nil, ErrTargetPeerInvalid // 此处并不endFlag，可以找其他peer拿
	}
	if tip == nil && len(blockIds) != 0 || int64(len(blockIds)) > length {
		// 返回消息参数非法
		return true, nil, ErrInvalidMsg
	}
	// 该beginBlockId已经是对方最高tipId
	if bytes.Equal(tip, beginBlockId) {
		return true, nil, ErrPeerFinish
	}
	// 该beginBlockId不在对方主干上，开启回溯
	if len(blockIds) == 0 {
		return false, nil, ErrTargetDataNotFound
	}
	// 当前同步头的最后一次同步
	if int64(len(blockIds)) < length {
		// 若当前接受的区块哈希列表为全0，则表示对方无相应数据
		return true, blockIds, nil
	}
	return false, blockIds, nil
}

/* downloadPeerBlocks
 * downloadPeerBlocks 在节点列表中随机选取若干节点，并将headersList散列到不同的节点任务中，
 * 输入是一个headersList其包含连续的BlockIds区间
 * 本地节点向对应节点发送GetDataMsg消息，试图获取全部需要的block信息，并将其存入cahche中，
 * 一直循环直到区间被填满
 */
func (lk *LedgerKeeper) downloadPeerBlocks(headersList [][]byte) []*SimpleBlock {
	// 若不收集齐会一直阻塞直到超时
	// 同步map，放置连续区间内的所有区块指针，由于分配给不同peer的blockIds任务随机，该数据结构保证返回时能够按顺序插入
	syncMap := map[string]*SimpleBlock{}
	for {
		// 在targetPeers中随机选择peers个数
		validPeers := getValidPeersNumber(lk.peersStatusMap)
		if validPeers == 0 {
			lk.log.Warn("downloadPeerBlocks::all peer invalid")
			return nil
		}
		randomLen := math_rand.Int63n(validPeers) // [0, validPeers)
		// 随机选择Peers数目
		targetPeers, err := randomPickPeers(randomLen+1, lk.peersStatusMap)
		if err != nil {
			continue
		}
		// 散列headersList随机向被选取的peer分配BlockIds, 这些任务Blockid有可能部分已经完成同步，仅需关注未同步部分
		peersTask, err := assignTaskRandomly(targetPeers, headersList)
		if err != nil {
			continue
		}
		// 对于单个peer，先查看cache中是否有该区块，选择cache中没有的生成列表，向peer发送GetDataMsg
		// cache并发读写操作时使用的锁
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second*SYNC_BLOCKS_TIMEOUT)
		defer cancel()
		ch := make(chan bool, len(peersTask))
		syncBlockMutex := &sync.RWMutex{}
		for peer, headers := range peersTask {
			go func(peer string, headers [][]byte, cache map[string]*SimpleBlock) {
				defer func() {
					ch <- true
				}()
				crashFlag, err := lk.peerBlockDownloadTask(ctx, peer, headers, cache, syncBlockMutex)
				if crashFlag {
					lk.log.Warn("downloadPeerBlocks::delete peer", "address", peer, "err", err)
					lk.peersStatusMap.Store(peer, false)
					return
				}
				if err != nil {
					lk.log.Warn("downloadPeerBlocks::peerBlockDownloadTask error", "error", err)
					return
				}
			}(peer, headers, syncMap)
		}
		for {
			select {
			case <-ch:
				if len(headersList) == len(syncMap) {
					returnBlocks := []*SimpleBlock{}
					for _, id := range headersList {
						returnBlocks = append(returnBlocks, syncMap[global.F(id)])
					}
					return returnBlocks
				}
			case <-ctx.Done():
				goto GetTrash
			}
		}

	}
GetTrash:
	// 看看剩下的cache里面有什么能捡的，找出从开头为始最长的连续存储返回
	returnBlocks := []*SimpleBlock{}
	for _, id := range headersList {
		if v, ok := syncMap[global.F(id)]; ok {
			returnBlocks = append(returnBlocks, v)
			continue
		}
		break
	}
	if len(returnBlocks) == 0 {
		return nil
	}
	return returnBlocks
}

/* peerBlockDownloadTask
 * peerBlockDownloadTask 向指定peer拉取指定区块列表，若该peer未返回任何块，则剔除节点，获取到的区块写入cache，上层逻辑判断是否继续拉取未获取的区块
 */
func (lk *LedgerKeeper) peerBlockDownloadTask(ctx context.Context, peerAddr string, taskBlockIds [][]byte, cache map[string]*SimpleBlock, syncBlockMutex *sync.RWMutex) (bool, error) {
	syncBlockMutex.Lock()             // 锁cache，进行读
	refreshTaskBlockIds := [][]byte{} // 筛除cache中已经从别的peer拿到的block，这些block无需重新传递
	for _, blockId := range taskBlockIds {
		if _, ok := cache[global.F(blockId)]; ok {
			continue
		}
		refreshTaskBlockIds = append(refreshTaskBlockIds, blockId)
	}
	syncBlockMutex.Unlock()
	if len(refreshTaskBlockIds) == 0 {
		return false, nil
	}
	peerSyncMap, err := lk.getBlocks(ctx, peerAddr, refreshTaskBlockIds)
	// 判断是否剔除peer, 目前保守删除
	if err == ErrInvalidMsg {
		return true, err
	}
	syncBlockMutex.Lock() // 锁cache，进行写
	for blockId, block := range peerSyncMap {
		cache[blockId] = block
	}
	syncBlockMutex.Unlock()
	return false, err
}

/* getBlocks
 * getBlocks 输入一个map，该map key包含一个peer需要返回的blockId，和一个空指针，随后向特定peer发送GetDataMsg消息，以获取指定区块信息，
 * 若指定节点并未在规定时间内返回任何区块，则返回节点超时错误
 * 若指定节点仅返回部分区块，则返回缺失提醒
 */
func (lk *LedgerKeeper) getBlocks(ctx context.Context, targetPeer string, blockIds [][]byte) (map[string]*SimpleBlock, error) {
	body := &pb.GetBlocksRequest{
		BlockIds: blockIds,
	}
	bodyBuf, _ := proto.Marshal(body)
	msg, err := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion2, lk.bcName, "", xuper_p2p.XuperMessage_GET_BLOCKS, bodyBuf, xuper_p2p.XuperMessage_NONE)
	if err != nil {
		lk.log.Warn("getBlocks::Generate GET_BLOCKS Message Error", "Error", err)
		return nil, ErrInternal
	}
	opts := []p2p_base.MessageOption{
		p2p_base.WithBcName(lk.bcName),
		p2p_base.WithTargetPeerAddrs([]string{targetPeer}),
	}
	res, err := lk.p2pSvr.SendMessageWithResponse(ctx, msg, opts...)
	if err != nil {
		lk.log.Warn("getBlocks::Sync GetBlocks P2P Error, local error or target error", "Logid", msg.GetHeader().GetLogid(), "Error", err)
		return nil, ErrInternal
	}

	response := res[0]
	blocksMsgBody := &pb.GetBlocksResponse{}
	err = proto.Unmarshal(response.GetData().GetMsgInfo(), blocksMsgBody)
	if err != nil {
		return nil, ErrUnmarshal
	}
	lk.log.Info("getBlocks::GET BLOCKS RESULT", "Logid", msg.GetHeader().GetLogid(), "LEN", len(blocksMsgBody.GetBlocksInfo()))
	if len(blocksMsgBody.GetBlocksInfo()) == 0 && len(blockIds) > 1 {
		// 目标节点完全未找到任何block
		return nil, ErrTargetDataNotFound
	}
	blocks := blocksMsgBody.GetBlocksInfo()
	peerSyncMap := map[string]*SimpleBlock{}
	for _, block := range blocks {
		blockId := global.F(block.GetBlockid())
		_, ok := peerSyncMap[blockId]
		if ok { // 即peer给出了重复了的blocks
			return nil, ErrInvalidMsg
		}
		peerSyncMap[blockId] = &SimpleBlock{
			internalBlock: block,
			logid:         msg.GetHeader().GetLogid() + "_" + msg.GetHeader().GetFrom(),
		}
	}
	if len(blocksMsgBody.GetBlocksInfo()) < len(blockIds) {
		// 目标节点并未在其本地找到所有需要的区块，需给上层返回缺失提醒
		return peerSyncMap, ErrTargetDataNotEnough
	}
	return peerSyncMap, nil
}

/* handleGetHeadersMsg response to the GetHeadersMsg with a HeadersMsg containing a list of the block-hashes required.
 * As a callee, peer checks whether the interval received is valid in its main-chain and then put the
 * corresponding block-hashes into the HeadersMsg.
 * When the callee cannot search the HEADER_HASH or the STOPPING_HASH of the GetHeadersMsg in its main-chain, it will
 * set the BLOCK_HASHES field to all zeroes to response to the caller.
 * handleGetHeadersMsg 接受GetHeadersMsg消息并返回，若GetHeadersMsg消息的消息区间在主干上，则直接返回区间所有的区块哈希列表，
 * 若不在主干，则返回一个空消息，若不存在，则返回本地最高ID。
 * 注意: 本次处理暂将消息区间不在主干在分支，以及账本无消息区间作为同样情况返回。
 */
func (lk *LedgerKeeper) handleGetHeadersMsg(ctx context.Context, msg *xuper_p2p.XuperMessage) (*xuper_p2p.XuperMessage, error) {
	bc := msg.GetHeader().GetBcname()
	if !p2p_base.VerifyDataCheckSum(msg) {
		lk.log.Error("handleGetHeadersMsg::verify msg error")
		return nil, ErrInvalidMsg
	}
	bodyBytes := msg.GetData().GetMsgInfo()
	body := &pb.GetBlockIdsRequest{}
	if err := proto.Unmarshal(bodyBytes, body); err != nil {
		return nil, ErrUnmarshal
	}
	headersCount := body.GetCount()
	if headersCount <= 0 {
		lk.log.Warn("handleGetHeadersMsg::Invalid headersCount, no service provided", "headersCount", headersCount)
		return nil, ErrInvalidMsg
	}
	localTip := lk.utxovm.GetLatestBlockid()
	nilHeaders := &pb.GetBlockIdsResponse{
		TipBlockId: localTip,
		BlockIds:   [][]byte{},
	}
	nilBuf, err := proto.Marshal(nilHeaders)
	if err != nil {
		return nil, ErrUnmarshal
	}
	nilRes, err := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion2, bc, msg.GetHeader().GetLogid(), xuper_p2p.XuperMessage_GET_BLOCKIDS_RES, nilBuf, xuper_p2p.XuperMessage_NONE)
	if err != nil {
		return nil, ErrInternal
	}
	headerBlockId := body.GetBlockId()
	lk.log.Trace("handleGetHeadersMsg::GET_HEADER_MSG handling...", "BEGIN HEADER", global.F(headerBlockId))
	headerBlock, err := lk.ledger.QueryBlock(headerBlockId)
	if err != nil {
		lk.log.Warn("handleGetHeadersMsg::internal error", "error", err, "headerBlockId", global.F(headerBlockId))
		return nilRes, nil
	}
	if !headerBlock.GetInTrunk() {
		lk.log.Warn("handleGetHeadersMsg::not in trunck", "headerBlock", global.F(headerBlockId))
		return nilRes, nil
	}

	resultHeaders := &pb.GetBlockIdsResponse{
		TipBlockId: localTip,
		BlockIds:   [][]byte{},
	}
	// 已经是最高高度，直接返回tipBlockId
	if bytes.Equal(localTip, headerBlockId) {
		resBuf, _ := proto.Marshal(resultHeaders)
		res, err := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion2, bc, msg.GetHeader().GetLogid(),
			xuper_p2p.XuperMessage_GET_BLOCKIDS_RES, resBuf, xuper_p2p.XuperMessage_SUCCESS)
		lk.log.Info("handleGetHeadersMsg::GET_HEADER_MSG response...", "response res: TipBlockId = beginBlockId")
		return res, err
	}

	// 循环获取下一个block，并批量放入Cache中
	b := headerBlock
	for i := int64(1); i <= headersCount; i++ {
		nextHash := b.GetNextHash()
		block, err := lk.ledger.QueryBlockHeader(nextHash)
		if err != nil {
			lk.log.Warn("handleGetHeadersMsg::QueryBlock error", "error", err, "blockId", global.F(nextHash))
			break
		}
		resultHeaders.BlockIds = append(resultHeaders.BlockIds, block.GetBlockid())
		b = block
		if bytes.Equal(localTip, nextHash) {
			break
		}
	}
	resBuf, _ := proto.Marshal(resultHeaders)
	res, err := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion2, bc, msg.GetHeader().GetLogid(),
		xuper_p2p.XuperMessage_GET_BLOCKIDS_RES, resBuf, xuper_p2p.XuperMessage_SUCCESS)
	printStr := ""
	for _, blockId := range resultHeaders.BlockIds {
		printStr += global.F(blockId) + " "
	}
	lk.log.Info("handleGetHeadersMsg::GET_HEADER_MSG response...", "response res", printStr)
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
	body := &pb.GetBlocksRequest{}
	if err := proto.Unmarshal(bodyBytes, body); err != nil {
		return nil, ErrUnmarshal
	}
	resultBlocks := []*pb.InternalBlock{}
	printStr := ""
	leftSize := lk.maxBlocksMsgSize
	for _, blockId := range body.GetBlockIds() {
		printStr += global.F(blockId) + " "
		block, err := lk.ledger.QueryBlock(blockId)
		if err != nil {
			continue
		}
		if leftSize-int64(proto.Size(block)) < 0 {
			break
		}
		resultBlocks = append(resultBlocks, block)
		leftSize -= int64(proto.Size(block))
	}
	lk.log.Trace("handleGetDataMsg::GET_BLOCKS_MSG handling...", "REQUIRE LIST", printStr)
	// peer自己通过区块大小切分Data消息返回， 按照尽可能多的返回区块规则选取区块
	result := &pb.GetBlocksResponse{
		BlocksInfo: resultBlocks,
	}
	resBuf, _ := proto.Marshal(result)
	msg.Header.From, _ = lk.p2pSvr.GetLocalUrl()
	lk.log.Info("handleGetDataMsg::GET_BLOCKS_MSG response...", "Logid", msg.GetHeader().GetLogid(), "LEN", len(resultBlocks))
	res, err := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion2, bc, msg.GetHeader().GetLogid(), xuper_p2p.XuperMessage_GET_BLOCKS_RES, resBuf, xuper_p2p.XuperMessage_SUCCESS)
	return res, err
}

/* confirmBlocks 原SendBlock逻辑
 * 账本接受blockMap数据，代替原来的PendingBlocks，正向confirm block
 */
func (lk *LedgerKeeper) confirmBlocks(hd *global.XContext, blocksSlice []*SimpleBlock, endFlag bool) ([]byte, bool, error) {
	// 取这段新链的第一个区块，判断走账本分叉逻辑还是直接账本追加逻辑
	var err error
	newBegin := lk.utxovm.GetLatestBlockid()
	lk.log.Debug("ConfirmBlocks", "genesis", global.F(lk.ledger.GetMeta().GetRootBlockid()), "utxo", global.F(lk.utxovm.GetLatestBlockid()),
		"len(blocksSlice)", len(blocksSlice), "cost", hd.Timer.Print())
	listLen := len(blocksSlice)
	if listLen == 0 {
		return newBegin, false, ErrInternal
	}
	// 获取整个slice中第一个非账本中的新Block
	index := 0
	for ; index < listLen; index++ {
		block := blocksSlice[index]
		if _, err := lk.ledger.QueryBlockHeader(block.internalBlock.GetBlockid()); err == nil {
			newBegin = block.internalBlock.GetBlockid()
			continue
		}
		break
	}
	// ledger已全部存在的情况
	if index == listLen {
		lk.log.Debug("ConfirmBlocks::all block exist", "cost", hd.Timer.Print())
		return newBegin, false, ErrTargetDataNotFound
	}
	beginSimpleBlock := blocksSlice[index]
	needVerify := (lk.nodeMode == config.NodeModeFastSync)

	/* ledgerkeeper和矿工抢同一把锁，该锁保证了ledgerkeeper在当前确认tx和打包区块不会冲突
	 * 矿工在doMiner会将UnconfirmedTx拿出来，也会进行uxtovm Play，而此时也有可能ledgerkeeper同步块做同样操作
	 * 该锁保护了当前仅有两者中的一个对象进行操作
	 * 原有SendBlock中多个同步进程抢这把锁的问题已经通过同步串行解决
	 */
	lk.coreMutex.Lock()
	defer lk.coreMutex.Unlock()

	if bytes.Compare(beginSimpleBlock.internalBlock.GetPreHash(), lk.ledger.GetMeta().GetTipBlockid()) == 0 {
		lk.log.Debug("ConfirmBlocks::Equal The Same", "cost", hd.Timer.Print())
		for ; index < listLen; index++ {
			needRepost := (index == listLen-1) && endFlag
			checkBlock := blocksSlice[index]
			if checkBlock == nil {
				continue
			}
			nextBlockid := checkBlock.internalBlock.GetBlockid()
			err, _ = lk.checkAndComfirm(needVerify, checkBlock)
			if err == ErrBlockExist {
				continue
			}
			if err != nil {
				lk.log.Warn("ConfirmBlocks::confirmAppendingBlock error", "err", err, "PreCheckBlock", checkBlock, "cost", hd.Timer.Print())
				break
			}
			// 判断是否是最新区块及最长链，若是则最新区块需广播
			err := lk.utxovm.PlayAndRepost(checkBlock.internalBlock.GetBlockid(), needRepost, false)
			lk.log.Debug("ConfirmAppendingBlock::Play", "logid", checkBlock.logid)
			if err != nil {
				lk.log.Warn("ConfirmAppendingBlock::utxo vm play err", "logid", checkBlock.logid, "err", err)
				break
			}
			newBegin = nextBlockid
			err = lk.con.ProcessConfirmBlock(checkBlock.internalBlock)
			if err != nil {
				lk.log.Warn("ConfirmBlocks::ProcessConfirmBlock error", "logid", checkBlock.logid, "error", err, "cost", hd.Timer.Print())
			}
		}
		if index < 1 {
			lk.log.Debug("ConfirmBlocks::confirm error", "err", err, "cost", hd.Timer.Print())
			return newBegin, false, err
		}
		lk.log.Debug("ConfirmBlocks::Equal The Same, confirm blocks finish", "newBegin", global.F(newBegin), "cost", hd.Timer.Print())
		return newBegin, index == listLen, err
	}
	//交点不等于utxo latest block
	lk.log.Debug("XXXXXXXXX The NO Same", "cost", hd.Timer.Print())
	for ; index < listLen; index++ {
		checkBlock := blocksSlice[index]
		if checkBlock == nil {
			continue
		}
		nextBlockid := checkBlock.internalBlock.GetBlockid()
		err, trunkSwitch := lk.checkAndComfirm(needVerify, checkBlock)
		if err != nil && err != ErrBlockExist {
			break
		}
		if trunkSwitch {
			err := lk.utxovm.Walk(checkBlock.internalBlock.GetBlockid(), false)
			lk.log.Debug("ConfirmBlocks::Walk Time", "logid", checkBlock.logid, "cost", hd.Timer.Print())
			if err != nil {
				lk.log.Warn("ConfirmBlocks::Walk error", "logid", checkBlock.logid, "err", err, "cost", hd.Timer.Print())
				break
			}
		}
		newBegin = nextBlockid
		err = lk.con.ProcessConfirmBlock(checkBlock.internalBlock)
		if err != nil {
			lk.log.Warn("ConfirmBlocks::ProcessConfirmBlock error", "error", err, "cost", hd.Timer.Print())
		}
	}
	// 待块确认后, 共识执行相应的操作
	if index < 1 {
		return newBegin, false, nil
	}
	lk.log.Debug("ConfirmBlocks::XXXXXXXXX The NO Same, confirm blocks finish", "newBegin", global.F(newBegin), "cost", hd.Timer.Print())
	return newBegin, index == listLen, err
}

func (lk *LedgerKeeper) checkAndComfirm(needVerify bool, simpleBlock *SimpleBlock) (error, bool) {
	block := simpleBlock.internalBlock
	if int64(proto.Size(block)) > lk.maxBlocksMsgSize {
		lk.log.Warn("checkAndComfirm:: Large block error", "logid", simpleBlock.logid)
		return ErrInvalidMsg, false
	}
	// 如果已经存在，则立即返回
	if lk.ledger.ExistBlock(block.GetBlockid()) {
		lk.log.Debug("checkAndComfirm::Block exist", "logid", simpleBlock.logid)
		return ErrBlockExist, false
	}
	for idx, tx := range block.Transactions {
		if !lk.ledger.IsValidTx(idx, tx, block) {
			lk.log.Warn("checkAndComfirm::invalid tx got from the block", "logid", simpleBlock.logid, "txid", global.F(tx.Txid), "blkid", global.F(block.GetBlockid()))
			return ErrInvalidMsg, false
		}
	}
	// 区块加解密有效性检查
	if needVerify {
		if res, err := lk.con.CheckMinerMatch(&pb.Header{Logid: simpleBlock.logid}, block); !res {
			lk.log.Warn("checkAndComfirm::check miner error", "logid", simpleBlock.logid, "error", err)
			return ErrServiceRefused, false
		}
	}
	cs := lk.ledger.ConfirmBlock(block, false)
	if !cs.Succ {
		lk.log.Warn("checkAndComfirm::confirm error", "logid", simpleBlock.logid)
		return ErrConfirmBlock, false
	}
	//是否发生主干切换
	trunkSwitch := (cs.TrunkSwitch || block.InTrunk)
	return nil, trunkSwitch
}

/* pickIndexes
 * pickIndexes 尽可能选择多的区块返回
 */
func pickIndexes(targetSize int64, list []int) []int {
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

/* randomPickPeers
 * randomPickPeers 从现有peersStatusMap中可连接的peers中随机选取number个作为目标节点
 */
func randomPickPeers(number int64, peers *sync.Map) ([]string, error) {
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
		random := math_rand.Int63n(int64(remains)) // [0,remains)
		if random < selection {
			result[number-selection] = i
			selection--
		}
		remains--
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
func assignTaskRandomly(targetPeers []string, headersList [][]byte) (map[string][][]byte, error) {
	if len(targetPeers) == 0 {
		return nil, ErrInvalidMsg
	}
	assignMap := map[uint32][][]byte{}
	for _, blockId := range headersList {
		var index uint32
		// 随机分配给不同index
		for j := 0; j+SIZE_OF_UINT32 < len(blockId); j = j + SIZE_OF_UINT32 {
			pos := binary.BigEndian.Uint32(blockId[j : j+SIZE_OF_UINT32])
			index = index + pos
		}
		index = index % uint32(len(targetPeers))
		if _, ok := assignMap[index]; !ok {
			assignMap[index] = [][]byte{blockId}
			continue
		}
		assignMap[index] = append(assignMap[index], blockId)
	}
	peersTask := map[string][][]byte{}
	for peerIndex, taskList := range assignMap {
		peersTask[targetPeers[int(peerIndex)]] = taskList
	}
	return peersTask, nil
}

/* changeSyncBeginBlockPoint
 * changeSyncBeginBlockPoint 当当前beginBlockId无法获取同步头列表时，需要通过输入账本回溯获取新的BlockId
 * 当前提供一种方法，向前回溯一个高度，TODO: 二分查找
 */
func changeSyncBeginPointBackwards(beginBlock *pb.InternalBlock) []byte {
	return beginBlock.GetPreHash()
}

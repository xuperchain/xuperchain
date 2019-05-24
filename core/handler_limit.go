package xchaincore

import (
	xuper_p2p "github.com/xuperchain/xuperunion/p2pv2/pb"
)

// MessageDispatcher message dispatcher for message limitation
type MessageDispatcher struct {
	msgDispChan      chan *xuper_p2p.XuperMessage      // 存放消息的队列
	handlerLimitChan chan bool                         // 控制goroutine数量
	limit            uint32                            // 初始化handlerLimitChan
	handle           func(msg *xuper_p2p.XuperMessage) // 注册方法,不同类型的消息注册方法不同,用于处理消息
	factor           float64                           // 装载因子
}

// Init init an instance of MessageDispatcher
func (msgDisp *MessageDispatcher) Init(limit uint32) {
	msgDisp.limit = limit
	msgDisp.msgDispChan = make(chan *xuper_p2p.XuperMessage, 12)
	msgDisp.handlerLimitChan = make(chan bool, msgDisp.limit)
	msgDisp.factor = 0.8
}

// Stop stop MessageDispatcher service
func (msgDisp *MessageDispatcher) Stop() {
	// TODO
	close(msgDisp.msgDispChan)
}

// Start start MessageDispatcher service
func (msgDisp *MessageDispatcher) Start() {
	for {
		select {
		case msg := <-msgDisp.msgDispChan:
			msgDisp.Dispatcher(msg)
		}
	}
}

// Register add limitation for specific msg type
func (msgDisp *MessageDispatcher) Register(h func(*xuper_p2p.XuperMessage)) {
	msgDisp.handle = h
}

// Work handle msg
func (msgDisp *MessageDispatcher) Work(msg *xuper_p2p.XuperMessage) {
	defer func() {
		<-msgDisp.handlerLimitChan
	}()
	msgDisp.handle(msg)
}

// Dispatcher dispatcher msg to handle
func (msgDisp *MessageDispatcher) Dispatcher(msg *xuper_p2p.XuperMessage) {
	msgDisp.handlerLimitChan <- true
	go msgDisp.Work(msg)
}

// NearFull check if it is time to limit msg
func (msgDisp *MessageDispatcher) NearFull() bool {
	capacity, length := cap(msgDisp.msgDispChan), len(msgDisp.msgDispChan)
	if length > int(float64(capacity)*msgDisp.factor) {
		return true
	}
	return false
}

/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package common

import (
	"time"

	"github.com/xuperchain/xuperchain/core/pb"
)

//BatchChan 将单条的chan包装成batch的chan
type BatchChan struct {
	queue    chan []*pb.Transaction
	window   int //每个batch最多少item
	waitms   int //最多间隔多久打包一个batch
	itemChan chan *pb.Transaction
}

// NewBatchChan New BatchChan
func NewBatchChan(window int, waitms int, itemChan chan *pb.Transaction) *BatchChan {
	bc := &BatchChan{
		window:   window,
		waitms:   waitms,
		queue:    make(chan []*pb.Transaction),
		itemChan: itemChan,
	}
	go bc.loopMakeBatch()
	return bc
}

func (bc *BatchChan) loopMakeBatch() {
	timer := time.Tick(time.Millisecond * time.Duration(bc.waitms))
	buffer := []*pb.Transaction{}
	timeFlag := false
	for {
		select {
		case item := <-bc.itemChan:
			buffer = append(buffer, item)
		case <-timer:
			timeFlag = true
		}

		// 数量控制 + 时间控制
		// case1: 当buffer中积攒的unconfirm transactions数量超过bc.window
		// case2: 当经过bc.waitms时间窗口后
		// case1以及case2都需要转发本地buffer中的unconfirm transactions
		if len(buffer) >= bc.window || (len(buffer) > 0 && timeFlag == true) {
			bc.queue <- buffer
			buffer = []*pb.Transaction{}
			timeFlag = false
		}

	}
}

// GetQueue 得到当前打包的transaction queue
func (bc *BatchChan) GetQueue() chan []*pb.Transaction {
	return bc.queue
}

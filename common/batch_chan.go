/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package common

import (
	"time"

	"github.com/xuperchain/xuperunion/pb"
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
	for {
		select {
		case item := <-bc.itemChan:
			buffer = append(buffer, item)
		case <-timer:
		}
		if len(buffer) >= bc.window {
			bc.queue <- buffer
			buffer = []*pb.Transaction{}
		}
	}
}

// GetQueue 得到当前打包的transaction queue
func (bc *BatchChan) GetQueue() chan []*pb.Transaction {
	return bc.queue
}

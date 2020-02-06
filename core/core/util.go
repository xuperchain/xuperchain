package xchaincore

import (
	"fmt"

	"github.com/golang/protobuf/proto"

	"github.com/xuperchain/xuperchain/core/pb"
)

func produceBlockEvent(msgChan chan *pb.Event, block *pb.InternalBlock, bcname string) {
	blockStatus := &pb.BlockStatusInfo{
		Bcname:   bcname,
		Proposer: fmt.Sprintf("%s", block.GetProposer()),
		Height:   block.GetHeight(),
	}
	blockEvent := &pb.BlockEvent{
		Block: block,
	}
	payload, _ := proto.Marshal(blockEvent)
	msgChan <- &pb.Event{
		Type:        pb.EventType_BLOCK,
		Payload:     payload,
		BlockStatus: blockStatus,
	}
	for _, tx := range block.GetTransactions() {
		produceTransactionEvent(msgChan, tx, bcname, pb.TransactionStatus_CONFIRM)
	}
}

func produceTransactionEvent(msgChan chan *pb.Event, tx *pb.Transaction, bcname string, status pb.TransactionStatus) {
	txStatus := &pb.TransactionStatusInfo{
		Bcname:      bcname,
		Initiator:   tx.GetInitiator(),
		AuthRequire: tx.GetAuthRequire(),
		Status:      status,
	}
	txEvent := &pb.TransactionEvent{
		Tx: tx,
	}
	payload, _ := proto.Marshal(txEvent)
	msgChan <- &pb.Event{
		Type:     pb.EventType_TRANSACTION,
		Payload:  payload,
		TxStatus: txStatus,
	}
}

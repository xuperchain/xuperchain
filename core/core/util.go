package xchaincore

import (
	"fmt"

	"github.com/golang/protobuf/proto"

	"github.com/xuperchain/xuperchain/core/event"
	"github.com/xuperchain/xuperchain/core/pb"
)

func produceBlockEvent(eventService *event.EventService, block *pb.InternalBlock, bcname string) {
	if block == nil {
		return
	}
	blockStatus := &pb.BlockStatusInfo{
		Bcname:   bcname,
		Proposer: fmt.Sprintf("%s", block.GetProposer()),
		Height:   block.GetHeight(),
	}
	blockEvent := &pb.BlockEvent{
		Block: block,
	}
	payload, marshalErr := proto.Marshal(blockEvent)
	if marshalErr != nil {
		return
	}
	eventService.Publish(&pb.Event{
		Type:        pb.EventType_BLOCK,
		Payload:     payload,
		BlockStatus: blockStatus,
	})
	for _, tx := range block.GetTransactions() {
		produceTransactionEvent(eventService, tx, bcname, pb.TransactionStatus_CONFIRM)
	}
}

func produceTransactionEvent(eventService *event.EventService, tx *pb.Transaction, bcname string, status pb.TransactionStatus) {
	if tx == nil {
		return
	}
	txStatus := &pb.TransactionStatusInfo{
		Bcname:      bcname,
		Initiator:   tx.GetInitiator(),
		AuthRequire: tx.GetAuthRequire(),
		Status:      status,
	}
	txEvent := &pb.TransactionEvent{
		Tx: tx,
	}
	payload, marshalErr := proto.Marshal(txEvent)
	if marshalErr != nil {
		return
	}
	eventService.Publish(&pb.Event{
		Type:     pb.EventType_TRANSACTION,
		Payload:  payload,
		TxStatus: txStatus,
	})

	// Account Event
	fromAddrs, toAddrs := getFromAddrAndToAddr(tx)
	accountStatus := &pb.AccountStatusInfo{
		Bcname:   bcname,
		FromAddr: fromAddrs,
		ToAddr:   toAddrs,
		Status:   status,
	}
	eventService.Publish(&pb.Event{
		Type:          pb.EventType_ACCOUNT,
		Payload:       payload,
		AccountStatus: accountStatus,
	})
}

func getFromAddrAndToAddr(tx *pb.Transaction) ([]string, []string) {
	if tx == nil {
		return nil, nil
	}
	fromAddrs := []string{}
	toAddrs := []string{}
	for _, input := range tx.GetTxInputs() {
		if input == nil {
			continue
		}
		fromAddrs = append(fromAddrs, fmt.Sprintf("%s", input.GetFromAddr()))
	}
	for _, output := range tx.GetTxOutputs() {
		if output == nil {
			continue
		}
		toAddrs = append(toAddrs, fmt.Sprintf("%s", output.GetToAddr()))
	}

	return fromAddrs, toAddrs
}

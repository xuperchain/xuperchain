package xchaincore

import (
	"fmt"

	"github.com/golang/protobuf/proto"

	"github.com/xuperchain/xuperchain/core/pb"
)

func produceBlockEvent(msgChan chan *pb.Event, block *pb.InternalBlock, bcname string) {
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
	msgChan <- &pb.Event{
		Type:     pb.EventType_TRANSACTION,
		Payload:  payload,
		TxStatus: txStatus,
	}

	// Account Event
	fromAddrs, toAddrs := getFromAddrAndToAddr(tx)
	accountStatus := &pb.AccountStatusInfo{
		Bcname:   bcname,
		FromAddr: fromAddrs,
		ToAddr:   toAddrs,
		Status:   status,
	}
	msgChan <- &pb.Event{
		Type:          pb.EventType_ACCOUNT,
		Payload:       payload,
		AccountStatus: accountStatus,
	}
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

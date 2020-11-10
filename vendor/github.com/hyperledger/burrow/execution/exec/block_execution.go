package exec

import (
	"fmt"
	"reflect"

	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/event/query"
	"github.com/hyperledger/burrow/txs"
)

func EventStringBlockExecution(height uint64) string {
	return fmt.Sprintf("Execution/Block/%v", height)
}

// Write out TxExecutions parenthetically
func (be *BlockExecution) StreamEvents() []*StreamEvent {
	var ses []*StreamEvent
	ses = append(ses, &StreamEvent{
		BeginBlock: &BeginBlock{
			Height:            be.Height,
			PredecessorHeight: be.PredecessorHeight,
			NumTxs:            uint64(len(be.TxExecutions)),
			Header:            be.Header,
		},
	})
	for _, txe := range be.TxExecutions {
		ses = append(ses, txe.StreamEvents()...)
	}
	return append(ses, &StreamEvent{
		EndBlock: &EndBlock{
			Height: be.Height,
		},
	})
}

func (*BlockExecution) EventType() EventType {
	return TypeBlockExecution
}

func (be *BlockExecution) Tx(txEnv *txs.Envelope) *TxExecution {
	txe := NewTxExecution(txEnv)
	be.AppendTxs(txe)
	return txe
}

func (be *BlockExecution) AppendTxs(tail ...*TxExecution) {
	for i, txe := range tail {
		txe.Index = uint64(len(be.TxExecutions) + i)
		txe.Height = be.Height
	}
	be.TxExecutions = append(be.TxExecutions, tail...)
}

// Tags

func (be *BlockExecution) Get(key string) (interface{}, bool) {
	switch key {
	case event.EventIDKey:
		return EventStringBlockExecution(be.Height), true
	case event.EventTypeKey:
		return be.EventType(), true
	}
	v, ok := query.GetReflect(reflect.ValueOf(be.Header), key)
	if ok {
		return v, true
	}
	return query.GetReflect(reflect.ValueOf(be), key)
}

func QueryForBlockExecutionFromHeight(height uint64) *query.Builder {
	return QueryForBlockExecution().AndGreaterThanOrEqual(event.HeightKey, height)
}

func QueryForBlockExecution() *query.Builder {
	return query.NewBuilder().AndEquals(event.EventTypeKey, TypeBlockExecution)
}

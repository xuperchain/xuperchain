package exec

import (
	"fmt"
	"io"

	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/event/query"
)

type EventStream interface {
	Recv() (*StreamEvent, error)
}

func (ses *StreamEvents) Recv() (*StreamEvent, error) {
	if len(ses.StreamEvents) == 0 {
		return nil, io.EOF
	}
	ev := ses.StreamEvents[0]
	ses.StreamEvents = ses.StreamEvents[1:]
	return ev, nil
}

func (ev *StreamEvent) EventType() EventType {
	switch {
	case ev.BeginBlock != nil:
		return TypeBeginBlock
	case ev.BeginTx != nil:
		return TypeBeginTx
	case ev.Envelope != nil:
		return TypeEnvelope
	case ev.Event != nil:
		return ev.Event.EventType()
	case ev.EndTx != nil:
		return TypeEndTx
	case ev.EndBlock != nil:
		return TypeEndBlock
	}
	return TypeUnknown
}

func (ev *StreamEvent) Get(key string) (interface{}, bool) {
	switch key {
	case event.EventTypeKey:
		return ev.EventType(), true
	}
	// Flatten this sum type
	return query.TagsFor(
		ev.GetBeginBlock().GetHeader(),
		ev.BeginBlock,
		ev.GetBeginTx().GetTxHeader(),
		ev.BeginTx,
		ev.Envelope,
		ev.Event,
		ev.EndTx,
		ev.EndBlock).Get(key)
}

type ContinuityOpt byte

func (so ContinuityOpt) Allows(opt ContinuityOpt) bool {
	return so&opt > 0
}

// ContinuityOpt encodes the following possible relaxations in continuity
const (
	// Default - continuous blocks, txs, and events are always permitted
	Continuous ContinuityOpt = iota
	// Allows consumption of blocks where the next block has a different predecessor block to that which was last consumed
	NonConsecutiveBlocks
	// Allows consumption of transactions with non-monotonic index (within block) or a different number of transactions
	// to that which is expected
	NonConsecutiveTxs
	// Allows consumption of events with non-monotonic index (within transaction) or a different number of events
	// to that which is expected
	NonConsecutiveEvents
)

type BlockAccumulator struct {
	block *BlockExecution
	// Number of txs expected in current block
	numTxs uint64
	// Height of last block consumed that contained transactions
	previousNonEmptyBlockHeight uint64
	// Accumulator for Txs
	stack TxStack
	// Continuity requirements for the stream
	continuity ContinuityOpt
}

func NewBlockAccumulator(continuityOptions ...ContinuityOpt) *BlockAccumulator {
	continuity := Continuous
	for _, opt := range continuityOptions {
		continuity |= opt
	}
	return &BlockAccumulator{
		continuity: continuity,
		stack: TxStack{
			continuity: continuity,
		},
	}
}

func (ba *BlockAccumulator) ConsumeBlockExecution(stream EventStream) (block *BlockExecution, err error) {
	var ev *StreamEvent
	for ev, err = stream.Recv(); err == nil; ev, err = stream.Recv() {
		block, err = ba.Consume(ev)
		if err != nil {
			return nil, err
		}
		if block != nil {
			return block, nil
		}
	}
	// If we reach here then we have failed to consume a complete block
	return nil, err
}

// Consume will add the StreamEvent passed to the block accumulator and if the block complete is complete return the
// BlockExecution, otherwise will return nil
func (ba *BlockAccumulator) Consume(ev *StreamEvent) (*BlockExecution, error) {
	switch {
	case ev.BeginBlock != nil:
		if !ba.continuity.Allows(NonConsecutiveBlocks) &&
			(ba.previousNonEmptyBlockHeight > 0 && ba.previousNonEmptyBlockHeight != ev.BeginBlock.PredecessorHeight) {
			return nil, fmt.Errorf("BlockAccumulator.Consume: received non-consecutive block at height %d: "+
				"predecessor height %d, but previous (non-empty) block height was %d",
				ev.BeginBlock.Height, ev.BeginBlock.PredecessorHeight, ba.previousNonEmptyBlockHeight)
		}
		// If we are consuming blocks over the event stream (rather than from state) we may see empty blocks
		// by definition empty blocks will not be a predecessor
		if ev.BeginBlock.NumTxs > 0 {
			ba.previousNonEmptyBlockHeight = ev.BeginBlock.Height
		}
		ba.numTxs = ev.BeginBlock.NumTxs
		ba.block = &BlockExecution{
			Height:            ev.BeginBlock.Height,
			PredecessorHeight: ev.BeginBlock.PredecessorHeight,
			Header:            ev.BeginBlock.Header,
			TxExecutions:      make([]*TxExecution, 0, ba.numTxs),
		}
	case ev.BeginTx != nil, ev.Envelope != nil, ev.Event != nil, ev.EndTx != nil:
		txe, err := ba.stack.Consume(ev)
		if err != nil {
			return nil, err
		}
		if txe != nil {
			if !ba.continuity.Allows(NonConsecutiveTxs) && uint64(len(ba.block.TxExecutions)) != txe.Index {
				return nil, fmt.Errorf("BlockAccumulator.Consume recieved transaction with index %d at "+
					"position %d in the event stream", txe.Index, len(ba.block.TxExecutions))
			}
			ba.block.TxExecutions = append(ba.block.TxExecutions, txe)
		}
	case ev.EndBlock != nil:
		if !ba.continuity.Allows(NonConsecutiveTxs) && uint64(len(ba.block.TxExecutions)) != ba.numTxs {
			return nil, fmt.Errorf("BlockAccumulator.Consume did not receive the expected number of "+
				"transactions for block %d, expected: %d, received: %d",
				ba.block.Height, ba.numTxs, len(ba.block.TxExecutions))
		}
		return ba.block, nil
	}
	return nil, nil
}

// TxStack is able to consume potentially nested txs
type TxStack struct {
	// Stack of TxExecutions, top of stack is TxExecution receiving innermost events
	txes []*TxExecution
	// Track the expected number events from the BeginTx event (also a stack)
	numEvents []uint64
	// Relaxations of transaction/event continuity
	continuity ContinuityOpt
}

func (stack *TxStack) Push(beginTx *BeginTx) {
	// Put this txe in the parent position
	stack.txes = append(stack.txes, &TxExecution{
		TxHeader:  beginTx.TxHeader,
		Result:    beginTx.Result,
		Events:    make([]*Event, 0, beginTx.NumEvents),
		Exception: beginTx.Exception,
	})
	stack.numEvents = append(stack.numEvents, beginTx.NumEvents)
}

func (stack *TxStack) Peek() (*TxExecution, error) {
	if len(stack.txes) < 1 {
		return nil, fmt.Errorf("tried to peek from an empty TxStack - might be missing essential StreamEvents")
	}
	return stack.txes[len(stack.txes)-1], nil
}

func (stack *TxStack) Pop() (*TxExecution, error) {
	txe, err := stack.Peek()
	if err != nil {
		return nil, err
	}
	newLength := len(stack.txes) - 1
	stack.txes = stack.txes[:newLength]
	numEvents := stack.numEvents[newLength]
	if !stack.continuity.Allows(NonConsecutiveEvents) && uint64(len(txe.Events)) != numEvents {
		return nil, fmt.Errorf("TxStack.Pop emitted transaction %s with wrong number of events, "+
			"expected: %d, received: %d", txe.TxHash, numEvents, len(txe.Events))
	}
	stack.numEvents = stack.numEvents[:newLength]
	return txe, nil
}

func (stack *TxStack) Length() int {
	return len(stack.txes)
}

// Consume will add the StreamEvent to the transaction stack and if that completes a single outermost transaction
// returns the TxExecution otherwise will return nil
func (stack *TxStack) Consume(ev *StreamEvent) (*TxExecution, error) {
	switch {
	case ev.BeginTx != nil:
		stack.Push(ev.BeginTx)
	case ev.Envelope != nil:
		txe, err := stack.Peek()
		if err != nil {
			return nil, err
		}
		txe.Envelope = ev.Envelope
		txe.Receipt = txe.Envelope.Tx.GenerateReceipt()
	case ev.Event != nil:
		txe, err := stack.Peek()
		if err != nil {
			return nil, err
		}
		if !stack.continuity.Allows(NonConsecutiveEvents) && uint64(len(txe.Events)) != ev.Event.Header.Index {
			return nil, fmt.Errorf("TxStack.Consume recieved event with index %d at "+
				"position %d in the event stream", ev.Event.GetHeader().GetIndex(), len(txe.Events))
		}
		txe.Events = append(txe.Events, ev.Event)
	case ev.EndTx != nil:
		txe, err := stack.Pop()
		if err != nil {
			return nil, err
		}
		if txe.Envelope == nil || txe.Receipt == nil {
			return nil, fmt.Errorf("TxStack.Consume did not receive transaction envelope for transaction %s",
				txe.TxHash)
		}
		if stack.Length() == 0 {
			// This terminates the outermost transaction
			return txe, nil
		}
		// If there is a parent tx on the stack add this tx as child
		parent, err := stack.Peek()
		if err != nil {
			return nil, err
		}
		parent.TxExecutions = append(parent.TxExecutions, txe)
	}
	return nil, nil
}

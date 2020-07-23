package event

import (
	"crypto/rand"

	"github.com/xuperchain/xuperchain/core/pb"
	"github.com/xuperchain/xuperchain/core/xmodel"
)

type blockBuilder struct {
	block *pb.InternalBlock
}

func newBlockBuilder() *blockBuilder {
	return &blockBuilder{
		block: &pb.InternalBlock{
			Blockid: makeRandID(),
		},
	}
}

func (b *blockBuilder) AddTx(tx ...*pb.Transaction) *blockBuilder {
	b.block.Transactions = append(b.block.Transactions, tx...)
	return b
}

func (b *blockBuilder) Block() *pb.InternalBlock {
	return b.block
}

type txBuilder struct {
	tx     *pb.Transaction
	events []*pb.ContractEvent
}

func newTxBuilder() *txBuilder {
	return &txBuilder{
		tx: &pb.Transaction{
			Txid: makeRandID(),
		},
	}
}

func (t *txBuilder) Initiator(addr string) *txBuilder {
	t.tx.Initiator = addr
	return t
}

func (t *txBuilder) AuthRequire(addr ...string) *txBuilder {
	t.tx.AuthRequire = addr
	return t
}

func (t *txBuilder) Transfer(from, to, amount string) *txBuilder {
	input := &pb.TxInput{
		RefTxid:  makeRandID(),
		FromAddr: []byte(from),
		Amount:   []byte(amount),
	}
	output := &pb.TxOutput{
		ToAddr: []byte(to),
		Amount: []byte(amount),
	}
	t.tx.TxInputs = append(t.tx.TxInputs, input)
	t.tx.TxOutputs = append(t.tx.TxOutputs, output)
	return t
}

func (t *txBuilder) Invoke(contract, method string, events ...*pb.ContractEvent) *txBuilder {
	req := &pb.InvokeRequest{
		ModuleName:   "wasm",
		ContractName: contract,
		MethodName:   method,
	}
	t.tx.ContractRequests = append(t.tx.ContractRequests, req)
	t.events = append(t.events, events...)
	return t
}

func (t *txBuilder) eventRWSet() []*pb.TxOutputExt {
	buf, _ := xmodel.MarshalMessages(t.events)
	return []*pb.TxOutputExt{
		{
			Bucket: xmodel.TransientBucket,
			Key:    []byte("contractEvent"),
			Value:  buf,
		},
	}
}

func (t *txBuilder) Tx() *pb.Transaction {
	t.tx.TxOutputsExt = t.eventRWSet()
	return t.tx
}

func makeRandID() []byte {
	buf := make([]byte, 32)
	rand.Read(buf)
	return buf
}

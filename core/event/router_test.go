package event

import (
	"encoding/hex"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/xuperchain/xuperchain/core/pb"
)

func TestRouteBlockTopic(t *testing.T) {
	ledger := newMockBlockStore()
	block := newBlockBuilder().Block()
	ledger.AppendBlock(block)

	router := NewRounterFromChainMG(ledger)

	filter := &pb.BlockFilter{
		Range: &pb.BlockRange{
			Start: "0",
		},
	}
	buf, err := proto.Marshal(filter)
	if err != nil {
		t.Fatal(err)
	}
	iter, err := router.Subscribe(pb.SubscribeType_BLOCK, buf)
	if err != nil {
		t.Fatal(err)
	}
	defer iter.Close()
	iter.Next()
	fblock := iter.Data().(*pb.FilteredBlock)
	if fblock.GetBlockid() != hex.EncodeToString(block.GetBlockid()) {
		t.Fatalf("block not equal, expect %x got %s", block.GetBlockid(), fblock.GetBlockid())
	}
}

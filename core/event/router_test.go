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
	encfunc, iter, err := router.Subscribe(pb.SubscribeType_BLOCK, buf)
	if err != nil {
		t.Fatal(err)
	}
	defer iter.Close()
	iter.Next()
	fblock := iter.Data().(*pb.FilteredBlock)

	_, err = encfunc(fblock)
	if err != nil {
		t.Fatal(err)
	}

	if fblock.GetBlockid() != hex.EncodeToString(block.GetBlockid()) {
		t.Fatalf("block not equal, expect %x got %s", block.GetBlockid(), fblock.GetBlockid())
	}
}

func TestRouteBlockTopicRaw(t *testing.T) {
	ledger := newMockBlockStore()
	block := newBlockBuilder().Block()
	ledger.AppendBlock(block)

	router := NewRounterFromChainMG(ledger)

	filter := &pb.BlockFilter{
		Range: &pb.BlockRange{
			Start: "0",
		},
	}

	iter, err := router.RawSubscribe(pb.SubscribeType_BLOCK, filter)
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

package event

import (
	"testing"

	"github.com/xuperchain/xuperchain/core/pb"
)

func expectTxMatch(t *testing.T, tx *pb.Transaction, pbfilter *pb.BlockFilter) {
	filter, err := newBlockFilter(pbfilter)
	if err != nil {
		t.Fatal(err)
	}
	if !matchTx(filter, tx) {
		t.Fatal("tx not match")
	}
}

func expectTxNotMatch(t *testing.T, tx *pb.Transaction, pbfilter *pb.BlockFilter) {
	filter, err := newBlockFilter(pbfilter)
	if err != nil {
		t.Fatal(err)
	}
	if matchTx(filter, tx) {
		t.Fatal("unexpected tx match")
	}
}

func expectEventMatch(t *testing.T, event *pb.ContractEvent, pbfilter *pb.BlockFilter) {
	filter, err := newBlockFilter(pbfilter)
	if err != nil {
		t.Fatal(err)
	}
	if !matchEvent(filter, event) {
		t.Fatal("event not match")
	}
}

func expectEventNotMatch(t *testing.T, event *pb.ContractEvent, pbfilter *pb.BlockFilter) {
	filter, err := newBlockFilter(pbfilter)
	if err != nil {
		t.Fatal(err)
	}
	if matchEvent(filter, event) {
		t.Fatal("unexpected event match")
	}
}

func TestFilterContractName(t *testing.T) {
	tx := newTxBuilder().Invoke("counter", "increase", nil).Tx()
	t.Run("empty", func(tt *testing.T) {
		expectTxMatch(tt, tx, &pb.BlockFilter{})
	})
	t.Run("match", func(tt *testing.T) {
		expectTxMatch(tt, tx, &pb.BlockFilter{
			Contract: "counter",
		})
	})
	t.Run("notMatch", func(tt *testing.T) {
		expectTxNotMatch(tt, tx, &pb.BlockFilter{
			Contract: "erc20",
		})
	})
	t.Run("subMatch", func(tt *testing.T) {
		expectTxMatch(tt, tx, &pb.BlockFilter{
			Contract: "count",
		})
	})
	t.Run("fullMatch", func(tt *testing.T) {
		expectTxNotMatch(tt, tx, &pb.BlockFilter{
			Contract: "^count$",
		})
	})
}

func TestFilterInitiator(t *testing.T) {
	tx := newTxBuilder().Initiator("addr1").Tx()
	t.Run("empty", func(tt *testing.T) {
		expectTxMatch(tt, tx, &pb.BlockFilter{})
	})
	t.Run("notMatch", func(tt *testing.T) {
		expectTxNotMatch(tt, tx, &pb.BlockFilter{
			Initiator: "addr2",
		})
	})
	t.Run("match", func(tt *testing.T) {
		expectTxMatch(tt, tx, &pb.BlockFilter{
			Initiator: "addr1",
		})
	})
}

func TestFilterAuthRequire(t *testing.T) {
	tx := newTxBuilder().AuthRequire("addr1", "addr2").Tx()
	t.Run("empty", func(tt *testing.T) {
		expectTxMatch(tt, tx, &pb.BlockFilter{})
	})
	t.Run("notMatch", func(tt *testing.T) {
		expectTxNotMatch(tt, tx, &pb.BlockFilter{
			AuthRequire: "not_exists",
		})
	})
	t.Run("match", func(tt *testing.T) {
		expectTxMatch(tt, tx, &pb.BlockFilter{
			AuthRequire: "addr1",
		})
		expectTxMatch(tt, tx, &pb.BlockFilter{
			AuthRequire: "addr2",
		})
	})
}

func TestFilterFromAddr(t *testing.T) {
	tx := newTxBuilder().Transfer("fromAddr", "toAddr", "10").Tx()
	t.Run("empty", func(tt *testing.T) {
		expectTxMatch(tt, tx, &pb.BlockFilter{})
	})
	t.Run("notMatch", func(tt *testing.T) {
		expectTxNotMatch(tt, tx, &pb.BlockFilter{
			FromAddr: "addr2",
		})
	})
	t.Run("match", func(tt *testing.T) {
		expectTxMatch(tt, tx, &pb.BlockFilter{
			FromAddr: "fromAddr",
		})
	})
}

func TestFilterToAddr(t *testing.T) {
	tx := newTxBuilder().Transfer("fromAddr", "toAddr", "10").Tx()
	t.Run("empty", func(tt *testing.T) {
		expectTxMatch(tt, tx, &pb.BlockFilter{})
	})
	t.Run("notMatch", func(tt *testing.T) {
		expectTxNotMatch(tt, tx, &pb.BlockFilter{
			ToAddr: "addr2",
		})
	})
	t.Run("match", func(tt *testing.T) {
		expectTxMatch(tt, tx, &pb.BlockFilter{
			ToAddr: "toAddr",
		})
	})
}

func TestFilterEvent(t *testing.T) {
	event := &pb.ContractEvent{
		Contract: "counter",
		Name:     "increase",
	}

	t.Run("empty", func(tt *testing.T) {
		expectEventMatch(tt, event, &pb.BlockFilter{})
	})
	t.Run("notMatch", func(tt *testing.T) {
		expectEventNotMatch(tt, event, &pb.BlockFilter{
			EventName: "get",
		})
	})
	t.Run("match", func(tt *testing.T) {
		expectEventMatch(tt, event, &pb.BlockFilter{
			EventName: "increase",
		})
	})
}

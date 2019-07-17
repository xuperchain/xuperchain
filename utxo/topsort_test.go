package utxo

import (
	"testing"
)

func TestTopSortWithCircle(t *testing.T) {
	graph := map[string][]string{}
	graph["tx3"] = []string{"tx1", "tx2"}
	graph["tx4"] = []string{"tx5", "tx6"}
	graph["tx2"] = []string{"tx3"}
	_, circle := TopSortDFS(graph)
	if circle == nil {
		t.Error("expect circle, but no circle")
	}
}

func TestTopSortWithoutCircle(t *testing.T) {
	graph := map[string][]string{}
	graph["tx3"] = []string{"tx1", "tx2"}
	graph["tx4"] = []string{"tx5", "tx6"}
	order, circle := TopSortDFS(graph)
	if circle != nil || len(order) != 6 {
		t.Error("TestTopSortWithoutCircle error")
	}
}

func TestTopSortWithIsolatedNetwork(t *testing.T) {
	graph := map[string][]string{}
	graph["tx3"] = []string{"tx1", "tx2"}
	graph["tx4"] = []string{"tx5", "tx6"}
	graph["tx7"] = []string{}
	order, circle := TopSortDFS(graph)
	if circle != nil || len(order) != 7 {
		t.Error("TestTopSortWithIsolatedNetwork error")
	}
}

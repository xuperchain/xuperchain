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
	ret1 := "tx3tx1tx2"
	ret2 := "tx3tx2tx1"
	ret3 := "tx4tx5tx6"
	ret4 := "tx4tx6tx5"
	graph := map[string][]string{}
	graph["tx3"] = []string{"tx1", "tx2"}
	graph["tx4"] = []string{"tx5", "tx6"}
	order, circle := TopSortDFS(graph)
	if circle != nil || len(order) != 6 {
		t.Error("TestTopSortWithoutCircle error")
	}
	// tx3 tx1 tx2 | tx2 tx1
	// tx4 tx5 tx6 | tx6 tx5
	orderStr := ""
	for _, str := range order {
		orderStr += str
	}
	if orderStr != (ret1+ret3) && orderStr != (ret1+ret4) && orderStr != (ret2+ret3) && orderStr != (ret2+ret4) && orderStr != (ret3+ret1) && orderStr != (ret3+ret2) && orderStr != (ret4+ret1) && orderStr != (ret4+ret2) {
		t.Error("TestTopSortWithoutCircle error")
	}
}

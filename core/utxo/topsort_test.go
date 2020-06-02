package utxo

import (
	"fmt"
	"github.com/xuperchain/xuperchain/core/pb"
	"sort"
	"testing"
)

// 无环case
/*
 * case1: 每个DAG只有一个元素
 * case2: 所有的DAG都包含多个元素
 * case3: 有部分DAG有多个元素，有部分DAG只有一个元素
 */
// 有环case
/*
 * case1: 存在一个环，并且DAG包括不止一个元素
 * case2: 存在一个环，并且有部分DAG包括不止一个元素，有部分DAG只有一个元素
 * case3: 存在两个环及以上，并且DAG包括不止一个元素
 * case4: 存在两个环及以上，并且有部分DAG包括不止一个元素，有部分DAG只有一个元素
 */

// 有环case + case1
func TestTopSortWithCircleCase1(t *testing.T) {
	graph := map[string][]string{}
	graph["tx3"] = []string{"tx1", "tx2"}
	graph["tx4"] = []string{"tx5", "tx6"}
	graph["tx2"] = []string{"tx3"}

	_, circle, _ := TopSortDFS(graph)
	if circle == false {
		t.Error("expect circle, but no circle")
	}
}

// 有环case + case2
func TestTopSortWithCircleCase2(t *testing.T) {
	graph := map[string][]string{}
	graph["tx3"] = []string{"tx1", "tx2"}
	graph["tx4"] = []string{"tx5", "tx6"}
	graph["tx2"] = []string{"tx3"}
	graph["tx7"] = []string{}

	_, circle, _ := TopSortDFS(graph)
	if circle == false {
		t.Error("expect circle, but no circle")
	}
}

// 有环case + case3
func TestTopSortWithCircleCase3(t *testing.T) {
	graph := map[string][]string{}
	graph["tx3"] = []string{"tx1", "tx2"}
	graph["tx4"] = []string{"tx5", "tx6"}
	graph["tx2"] = []string{"tx3"}
	graph["tx5"] = []string{"tx4"}

	_, circle, _ := TopSortDFS(graph)
	if circle == false {
		t.Error("expect circle, but no circle")
	}
}

// 有环case + case4
func TestTopSortWithCircleCase4(t *testing.T) {
	graph := map[string][]string{}
	graph["tx3"] = []string{"tx1", "tx2"}
	graph["tx4"] = []string{"tx5", "tx6"}
	graph["tx2"] = []string{"tx3"}
	graph["tx5"] = []string{"tx4"}
	graph["tx7"] = []string{}

	_, circle, _ := TopSortDFS(graph)
	if circle == false {
		t.Error("expect circle, but no circle")
	}
}

// 无环case + case1
func TestTopSortWithoutCircleCase1(t *testing.T) {
	graph := map[string][]string{}
	graph["tx1"] = []string{}
	graph["tx2"] = []string{}
	graph["tx3"] = []string{}

	order, circle, childDAG := TopSortDFS(graph)
	t.Log("order->", order)
	t.Log("circle->", circle)
	t.Log("childDAG->", childDAG)
	if circle || len(order) != 3 {
		t.Error("TestTopSortWithoutCircle error")
	}
	childDAGSlice := []string{}
	// 按照childDAG拆分出多个子DAG字符串
	idx := 0
	start, end, length := 0, 0, len(childDAG)
	for idx < length {
		end += childDAG[idx]
		currChildStr := getStr(start, end, order)
		tmpStr := ""
		for _, v := range currChildStr {
			tmpStr += v
		}
		childDAGSlice = append(childDAGSlice, tmpStr)
		start = end
		idx++
	}
	sort.Strings(childDAGSlice)
	orderStr := ""
	for _, str := range childDAGSlice {
		orderStr += str
	}
	t.Log("orderStr->", orderStr)
	if orderStr != "tx1tx2tx3" {
		t.Error("TestTopSortWithoutCircle error")
	}
}

// 无环 + case2
func TestTopSortWithoutCircleCase2(t *testing.T) {
	graph := map[string][]string{}
	graph["tx3"] = []string{"tx1", "tx2"}
	graph["tx4"] = []string{"tx5", "tx6"}
	order, circle, childDAG := TopSortDFS(graph)
	if circle || len(order) != 6 || len(childDAG) != 2 {
		t.Error("TestTopSortWithoutCircle error")
	}
}

// 无环 + case3
func TestTopSortWithoutCircleCase3(t *testing.T) {
	graph := map[string][]string{}
	graph["tx3"] = []string{"tx1", "tx2"}
	graph["tx4"] = []string{"tx5", "tx6"}
	graph["tx7"] = []string{}
	order, circle, childDAG := TopSortDFS(graph)
	if circle || len(order) != 7 || len(childDAG) != 3 {
		t.Error("TestTopSortWithoutCircle error")
	}
	t.Log("order->", order)
	t.Log("circle->", circle)
	t.Log("childDAG->", childDAG)
	oo := map[string]int{}
	for idx, tx := range order {
		oo[tx] = idx
	}
	if oo["tx3"] < oo["tx1"] &&
		oo["tx3"] < oo["tx2"] &&
		oo["tx4"] < oo["tx5"] &&
		oo["tx4"] < oo["tx6"] {
		t.Log("order ok")
	} else {
		t.Error("order check failed at TestTopSortWithoutCircleCase3")
	}
}

func getStr(start int, end int, order []string) []string {
	fmt.Println("getStr->", "start->", start, " end->", end)
	ret := []string{}
	for _, v := range order[start:end] {
		ret = append(ret, v)
	}

	return ret
}

func TestTopSortWithoutCircleCase4(t *testing.T) {
	graph := map[string][]string{}
	graph["tx1"] = []string{"tx2"}
	graph["tx2"] = []string{"tx3"}
	graph["tx3"] = []string{"tx4"}
	graph["tx4"] = []string{"tx5"}
	graph["tx6"] = []string{"tx3"}

	order, circle, childDAG := TopSortDFS(graph)
	if circle || len(order) != 6 || len(childDAG) != 1 {
		t.Error("TestTopSortWithoutCircle error")
	}
	t.Log("order->", order)
	t.Log("circle->", circle)
	t.Log("childDAG->", childDAG)
}

func TestSplitBlock(t *testing.T) {
	tx1 := &pb.Transaction{Txid: []byte("tx1")}
	tx2 := &pb.Transaction{Txid: []byte("tx2"), TxInputs: []*pb.TxInput{&pb.TxInput{RefTxid: []byte("tx1")}}}
	tx3 := &pb.Transaction{Txid: []byte("tx3"), TxInputs: []*pb.TxInput{&pb.TxInput{RefTxid: []byte("tx1")}}}
	tx4 := &pb.Transaction{Txid: []byte("tx4")}
	tx5 := &pb.Transaction{Txid: []byte("tx5"), TxInputs: []*pb.TxInput{&pb.TxInput{RefTxid: []byte("tx4")}}}
	tx6 := &pb.Transaction{Txid: []byte("tx6"), TxInputs: []*pb.TxInput{&pb.TxInput{RefTxid: []byte("tx777")}}}
	block := &pb.InternalBlock{Transactions: []*pb.Transaction{tx1, tx2, tx3, tx4, tx5, tx6}}
	dags := splitToDags(block)
	if len(dags) != 3 {
		t.Error("dags count unexpected", len(dags))
	}
	if len(dags[0]) != 3 || len(dags[1]) != 2 || len(dags[2]) != 1 {
		t.Error("dag size unexpected")
	}
}

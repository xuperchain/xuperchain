package utxo

import (
	"fmt"
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
	if orderStr != "tx3tx2tx1tx4tx6tx5" {
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
	childDAGSlice := []string{}
	// 按照childDAG拆分出多个子DAG字符串
	idx := 0
	start, end, length := 0, 0, len(childDAG)
	for idx < length {
		end = start + childDAG[idx]
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
	if orderStr != "tx3tx2tx1tx4tx6tx5tx7" {
		t.Error("TestTopSortWithoutCircle error")
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

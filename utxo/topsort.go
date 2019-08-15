package utxo

// TxGraph 交易依赖关系图
type TxGraph map[string][]string

/*
*  说明：
*  'tx3' --> ['tx1', 'tx2']  tx3依赖了tx1,tx2, 也可以表示反向依赖关系:tx3被tx1,tx2依赖
*  'tx2' --> ['tx0', 'tx1']
 */

// TopSortDFS 对依赖关系图进行拓扑排序
// 输入：依赖关系图，就是个map
// 输出: order: 排序后的有序数组，依赖者排在前面，被依赖的排在后面
//       cyclic: 如果发现有环形依赖关系则输出这个数组
//
// 实现参考： https://rosettacode.org/wiki/Topological_sort#Go
func TopSortDFS(g TxGraph) (order []string, cycle bool, childDAGsSize []int) {
	// 统计每个tx的次数(包括被引用以及引用次数)
	degreeForTx := map[string]int{}
	headTx := map[string]int{}
	cyclic := []string{}
	for k, outputs := range g {
		degreeForTx[k]++
		headTx[k]++
		for _, m := range outputs {
			headTx[m]--
			if g[m] == nil {
				g[m] = []string{} //预处理一下，coinbase交易可能没有依赖
				degreeForTx[m]++
			}
			degreeForTx[m]++
		}
	}
	L := make([]string, len(g))
	i := len(L)
	temp := map[string]bool{} //临时访问标记
	perm := map[string]bool{} //永久访问标记
	var cycleFound bool
	var cycleStart string
	var visit func(string)
	visit = func(n string) {
		switch {
		case temp[n]: //临时标记里面有，说明产生环了
			cycleFound = true
			cycleStart = n
			return
		case perm[n]:
			return
		}
		temp[n] = true
		for _, m := range g[n] {
			visit(m)
			if cycleFound {
				if cycleStart > "" {
					cyclic = append(cyclic, n)
					if n == cycleStart {
						cycleStart = ""
					}
				}
				return
			}
		}
		delete(temp, n)
		perm[n] = true
		i--
		L[i] = n
	}
	// 上一个子DAG对应的起始数组索引
	lastDAGIdx := len(L)
	// 当前子DAG对应的tx个数
	currChildDAGSize := 0
	for n := range g {
		if perm[n] || len(g[n]) <= 0 {
			continue
		}
		if v, ok := headTx[n]; ok && v != 1 {
			continue
		}
		visit(n)

		currChildDAGSize = lastDAGIdx - i
		childDAGsSize = append([]int{currChildDAGSize}, childDAGsSize...)
		lastDAGIdx = i

		if cycleFound {
			return nil, true, childDAGsSize
		}
	}
	leftIdx := 0
	for k, _ := range degreeForTx {
		// 将入度和出度为0的tx分离出来
		// degreeForTx[k] == 1表示tx入度为0
		// len(g[k])表示tx的出度为0
		if degreeForTx[k] == 1 && len(g[k]) <= 0 {
			L[leftIdx] = k
			leftIdx++
			childDAGsSize = append([]int{1}, childDAGsSize...)
		}
	}
	// 存在环
	if leftIdx != i {
		return nil, true, nil
	}
	return L, false, childDAGsSize
}

package utxo

//交易依赖关系图
type TxGraph map[string][]string

//TopSortDFS 对依赖关系图进行拓扑排序
// 输入：依赖关系图，就是个map
// 输出: order: 排序后的有序数组，依赖者排在前面，被依赖的排在后面
//       cyclic: 如果发现有环形依赖关系则输出这个数组
//
// 实现参考： https://rosettacode.org/wiki/Topological_sort#Go
// 在我们映射中，RefTx是边的源点
func TopSortDFS(g TxGraph) (order []string, cyclic bool, childDAGSize []int) {
	reverseG := TxGraph{}
	for n, outputs := range g {
		for _, m := range outputs {
			if g[m] == nil {
				g[m] = []string{} //预处理一下，coinbase交易可能没有依赖
			}
			if reverseG[m] == nil {
				reverseG[m] = []string{}
			}
			reverseG[m] = append(reverseG[m], n)
		}
	}
	L := make([]string, len(g))
	i := len(L)
	temp := map[string]bool{} //临时访问标记
	perm := map[string]bool{} //永久访问标记
	var cycleFound bool
	var visit func(string)
	visit = func(n string) {
		switch {
		case temp[n]: //临时标记里面有，说明产生环了
			cycleFound = true
			return
		case perm[n]:
			return
		}
		temp[n] = true
		for _, m := range g[n] {
			visit(m)
			if cycleFound {
				cyclic = true
				return
			}
		}
		delete(temp, n)
		perm[n] = true
		i--
		L[i] = n
	}
	subGraphs := [][]string{}
	marked := map[string]bool{}
	subG := []string{}
	var dfs func(string)
	dfs = func(n string) {
		if marked[n] {
			return
		}
		marked[n] = true
		for _, m := range g[n] {
			dfs(m)
		}
		for _, m := range reverseG[n] {
			dfs(m)
		}
		subG = append(subG, n)
	}
	// dfs变量切分出多个连通的子图
	for n := range g {
		if marked[n] {
			continue
		}
		dfs(n)
		subGraphs = append(subGraphs, subG)
		subG = []string{}
	}

	childDAGSize = make([]int, len(subGraphs))
	for i, g := range subGraphs {
		//记录每个DAG子图的大小
		childDAGSize[len(subGraphs)-i-1] = len(g)
		for _, n := range g {
			if perm[n] {
				continue
			}
			visit(n)
			if cycleFound {
				return nil, cyclic, childDAGSize
			}
		}
	}
	return L, false, childDAGSize
}

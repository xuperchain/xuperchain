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
func TopSortDFS(g TxGraph) (order, cyclic []string) {
	// 先将孤立的点给分离出来(孤立的点就是它不依赖别的节点,也不被其他节点依赖)
	tmpSlice := map[string]bool{}
	// 赋值一份完整的tx, 最终剩下tx的就是不依赖别人, 也不被别人依赖
	for k, _ := range g {
		tmpSlice[k] = true
	}
	for k, outputs := range g {
		// 被依赖的tx需要删掉
		if len(outputs) > 0 {
			delete(tmpSlice, k)
		}
		for _, m := range outputs {
			// m已经需要引用父亲tx了，这种tx不是完全独立的
			if g[m] == nil {
				g[m] = []string{} //预处理一下，coinbase交易可能没有依赖
			}
			// 依赖别人的tx需要被删掉
			delete(tmpSlice, m)
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
	for n := range g {
		// 不处理不被别人依赖的tx
		// 对于依赖别人的tx，会在处理被依赖的tx时遍历到这些tx
		// 对于孤立tx，会在后面统一处理
		if perm[n] || len(g[n]) <= 0 {
			continue
		}
		visit(n)
		if cycleFound {
			return nil, cyclic
		}
	}
	// 将之前孤立的点整合到最终的返回结果中
	leftIdx := 0
	for k, _ := range tmpSlice {
		L[leftIdx] = k
		leftIdx++
	}
	return L, nil
}

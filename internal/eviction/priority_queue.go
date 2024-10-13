package eviction

type priorityqueue []*lfuEntry

type lfuEntry struct {
	index int
	entry Entry
	count int
}

func (l *lfuEntry) referenced() {
	l.count++
	l.entry.Touch()
}

func (pq priorityqueue) Less(i, j int) bool {
	if pq[i].count == pq[j].count {
		return pq[i].entry.UpdateAt.Before(*pq[j].entry.UpdateAt)
	}
	return pq[i].count < pq[j].count
}

func (pq priorityqueue) Len() int {
	return len(pq)
}

func (pq priorityqueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityqueue) Pop() interface{} {
	oldpq := *pq
	n := len(oldpq)
	entry := oldpq[n-1]

	// avoid memory leaks
	oldpq[n-1] = nil

	newpq := oldpq[0 : n-1]
	for i := 0; i < len(newpq); i++ {
		newpq[i].index = i
	}

	*pq = newpq
	return entry
}

func (pq *priorityqueue) Push(x interface{}) { // 绑定push方法，插入新元素
	entry := x.(*lfuEntry)
	entry.index = len(*pq)
	*pq = append(*pq, x.(*lfuEntry))
}

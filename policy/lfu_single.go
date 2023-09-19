package policy

// type Value interface {
// 	Len() int
// }

// type entry struct {
// 	key      string
// 	value    Value
// 	updateAt *time.Time
// }

// type lfuEntry struct {
// 	index int
// 	entry entry
// 	count int
// }

// type LfuCache struct {
// 	maxCacheSize int64
// 	nBytes       int64
// 	cache        map[string]*lfuEntry
// 	priority     *priorityqueue
// 	onEvicted    func(string, Value)
// }

// type priorityqueue []*lfuEntry

// func (pq priorityqueue) Len() int { return len(pq) }
// func (pq priorityqueue) Less(i, j int) bool {
// 	if pq[i].count == pq[j].count {
// 		return pq[i].entry.updateAt.Before(*pq[j].entry.updateAt)
// 	} else {
// 		return pq[i].count < pq[j].count
// 	}
// }
// func (pq priorityqueue) Swap(i, j int) {
// 	pq[i], pq[j] = pq[j], pq[i]
// }

// func (pq *priorityqueue) Push(x interface{}) {
// 	// 构造时 entry.index = 0，我们需要根据当前堆状态更新它的值
// 	entry := x.(*lfuEntry)
// 	entry.index = len(*pq)
// 	*pq = append(*pq, entry)
// }

// func (pq *priorityqueue) Pop() interface{} {
// 	oldpq := *pq
// 	n := len(oldpq)
// 	entry := oldpq[n-1]
// 	oldpq[n-1] = nil // avoid memory leak
// 	newpq := oldpq[:n-1]
// 	// 所有节点的 index 都还是未删除小顶堆顶点钱的索引
// 	// 比如 1 2 3 5 4 这个小根堆，node1.index=0 node2.index=1 node3.index=2 node5.index=3 node4.index=4
// 	// 删除小顶堆堆顶元素后，newpq = [2 3 5 4] 但 node2.index=1 node3.index=2 node5.index=3 node4.index=4
// 	// 所以我们需要根据最新的切片重新设置各个元素的索引
// 	for i := 0; i < len(newpq); i++ {
// 		newpq[i].index = i
// 	}
// 	// newpq [2 3 5 4] node2.index=0 node3.index=1 node5.index=2 node4.index=3
// 	*pq = newpq
// 	return entry
// }

// func NewLfuCache(maxCacheSize int64, onEvcited func(key string, value Value)) *LfuCache {
// 	queue := priorityqueue(make([]*lfuEntry, 0))
// 	return &LfuCache{
// 		maxCacheSize: maxCacheSize,
// 		nBytes:       0,
// 		cache:        make(map[string]*lfuEntry),
// 		priority:     &queue,
// 		onEvicted:    onEvcited,
// 	}
// }

// func (lf *LfuCache) Get(key string) (Value, *time.Time, bool) {
// 	if lfuEntry, ok := lf.cache[key]; ok {
// 		lfuEntry.referenced()
// 		heap.Fix(lf.priority, lfuEntry.index)
// 		return lfuEntry.entry.value, lfuEntry.entry.updateAt, true
// 	}
// 	return nil, nil, false
// }

// func (lf *LfuCache) Put(key string, value Value) {
// 	fmt.Printf("现在插入的是 key %s value %v\n", key, value)
// 	if e, ok := lf.cache[key]; ok {
// 		e.referenced()
// 		fmt.Println("未更新：当前容量：", lf.nBytes)
// 		lf.nBytes += int64(value.Len()) - int64(e.entry.value.Len())
// 		fmt.Println("更新时：当前容量：", lf.nBytes)
// 		e.entry.value = value
// 		heap.Fix(lf.priority, e.index)
// 	} else {
// 		e := &lfuEntry{
// 			index: 0,
// 			entry: entry{
// 				key:      key,
// 				value:    value,
// 				updateAt: nil,
// 			},
// 		}
// 		e.referenced()
// 		lf.nBytes += int64(len(key) + value.Len())
// 		fmt.Println("插入时：当前容量：", lf.nBytes)
// 		heap.Push(lf.priority, e)
// 		lf.cache[key] = e
// 	}

// 	for lf.maxCacheSize > 0 && lf.nBytes > lf.maxCacheSize {
// 		lf.Remove()
// 	}
// }

// // 缓存淘汰
// func (lf *LfuCache) Remove() {
// 	e := heap.Pop(lf.priority).(*lfuEntry)
// 	fmt.Println(e.entry)
// 	lf.nBytes -= int64(len(e.entry.key) + e.entry.value.Len())
// 	fmt.Println("删除后：当前容量：", lf.nBytes)
// 	delete(lf.cache, e.entry.key)
// 	if lf.onEvicted != nil {
// 		lf.onEvicted(e.entry.key, e.entry.value)
// 	}
// }

// // 更新访问计数和最新访问时间
// func (lfuEntry *lfuEntry) referenced() {
// 	lfuEntry.count++
// 	now := time.Now()
// 	lfuEntry.entry.updateAt = &now
// }

// // 返回优先队列的长度
// func (lf *LfuCache) Len() int {
// 	return lf.priority.Len()
// }

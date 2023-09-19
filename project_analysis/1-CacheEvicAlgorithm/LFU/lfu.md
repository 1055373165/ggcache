# LFU 算法设计思路与实现


## 核心结构体

- 存储条目 entry
除了 key 和 value 之外，entry 还包含了一个 updateAt 字段，用于表示该条目的最近一次访问时间。（从而实现 TTL 能力）
```go
type entry struct {
	key      string
	value    Value
	updateAt *time.Time
}
```
但是 entry 结构体中没有包含访问次数字段，因为访问次数是 LFU 算法的核心，所以需要对存储条目进行进一步封装；

- 封装后的条目 lfuEntry
```go
type lfuEntry struct {
	index int
	entry entry
	count int
}
```
lfuEntry 结构体中包含了一个 index 字段，用于记录当前条目在有限数组中的索引，以及一个 entry 字段，用于存储缓存条目，count 用于记录当前条目的访问次数。


- LFU 缓存
```go
type LfuCache struct {
	maxCacheSize int64
	nBytes       int64
	cache        map[string]*lfuEntry
	priority     *priorityqueue
	onEvicted    func(string, Value)
}
```
LFU 缓存包含了一个 maxCacheSize 字段，用于记录缓存的最大容量，一个 nBytes 字段，用于记录当前缓存中所有条目的大小，一个 cache 字段，用于存储指定 key 对应缓存条目的映射关系，一个 priority 字段，它是存储缓存条目的底层优先队列，一个 onEvicted 字段，用于记录缓存条目被删除时的回调函数。

## 底层数据结构实现---优先队列（堆）
我们需要明确一点：优先队列中存储的是封装了访问次数的缓存条目结构体的指针（因为需要修改底层结构体的访问次数）；
```go
type priorityqueue []*lfuEntry
```
我们需要手动实现优先队列的堆特性，priorityqueue 类型需要分别实现
- sort.Interface接口
  - Len() int 方法
  - Swap(i, j int) 方法
  - Less(i, j int) bool 方法
- heap.Interface接口
  - Push(x interface) 方法（必须指针接收者）
  - Pop() interface 方法（必须指针接收者）

### 实现 sort.Interface 接口
```go
func (pq priorityqueue) Len() int { return len(pq) }
func (pq priorityqueue) Less(i, j int) bool {
	if pq[i].count == pq[j].count {
		return pq[i].entry.updateAt.Before(*pq[j].entry.updateAt)
	} else {
		return pq[i].count < pq[j].count
	}
}
func (pq priorityqueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}
```

> 实现 Less 方法时需要特别注意
1. 访问次数相同时，需要根据最近访问时间来判断，即最近访问时间靠前的优先级更高
2. 这也就是说：LFU 缓存中，有两个维度的考量，分别是访问次数和最近访问时间，当访问次数相同时，最近访问的缓存条目优先级越高

### 实现 heap.Interface 接口
```go
func (pq *priorityqueue) Push(x interface{}) {
	// 构造时 entry.index = 0，我们需要根据当前堆状态更新它的值
	entry := x.(*lfuEntry)
	entry.index = len(*pq)
	*pq = append(*pq, entry)
}

func (pq *priorityqueue) Pop() interface{} {
	oldpq := *pq
	n := len(oldpq)
	entry := oldpq[n-1]
	oldpq[n-1] = nil // avoid memory leak
	newpq := oldpq[:n-1]
	// 所有节点的 index 都还是未删除小顶堆顶点钱的索引
	// 比如 1 2 3 5 4 这个小根堆，node1.index=0 node2.index=1 node3.index=2 node5.index=3 node4.index=4
	// 删除小顶堆堆顶元素后，newpq = [2 3 5 4] 但 node2.index=1 node3.index=2 node5.index=3 node4.index=4
	// 所以我们需要根据最新的切片重新设置各个元素的索引
	for i := 0; i < len(newpq); i++ {
		newpq[i].index = i
	}
	// newpq [2 3 5 4] node2.index=0 node3.index=1 node5.index=2 node4.index=3
	*pq = newpq
	return entry
}
```


## 基于工厂模式获取新的缓存对象

```go
func NewLfuCache(maxCacheSize int64, onEvcited func(key string, value Value)) *LfuCache {
	queue := priorityqueue(make([]*lfuEntry, 0))
	return &LfuCache{
		maxCacheSize: maxCacheSize,
		nBytes:       0,
		cache:        make(map[string]*lfuEntry),
		priority:     &queue,
		onEvicted:    onEvcited,
	}
}
```

## 核心方法

### Get 查询缓存

1. 去缓存查询 key 的值
2. 如果可以查询的得到，那么更新访问计数和过期时间
3. 根据新的频次调整堆的结构，重新建立堆（修复堆结构）
4. 返回缓存值、最新访问时间以及是否查询成功的标志


```go
func (lf *LfuCache) Get(key string) (Value, *time.Time, bool) {
	if lfuEntry, ok := lf.cache[key]; ok {
		lfuEntry.referenced()
		heap.Fix(lf.priority, lfuEntry.index)
		return lfuEntry.entry.value, lfuEntry.entry.updateAt, true
	}
	return nil, nil, false
}
```

### Put 添加缓存或者更新缓存

1. 先去缓存中查询
2. 如果可以查询得到，那么需要更新当前缓存值（当前已使用容量加上用新使用容量和旧使用容量的差值）
3. 更新值、更新访问计数和最新访问时间
4. 调整堆（调用 Fix 函数对条目所在的索引进行堆调整）

二. 如果查询不到，那么相当于插入一条新的 kv 对，构造一个 LfuEntry 类型的记录
三. 更新它的访问频次和上一次的更新时间；
四. 插入到堆中并更新缓存容量
五. 将 key 和对应的 LfuEntry 存入缓存

无论最开始是否从缓存中查询到 key 的值，都需要判断当前使用的缓存容量是否超过了缓存上限，如果是
则使用 LFU 缓存淘汰算法淘汰掉一些条目，直至满足缓存上限为止；因为 LFU 使用了优先队列作为数据结构，
实际上底层就是一个小根堆，按照条目的访问频次进行堆排序，相同条目按照更新时间先后进行堆排序，默认最新
更新的条目优先级更高，即更晚从堆中删除。

```go
func (lf *LfuCache) Put(key string, value Value) {
	fmt.Printf("现在插入的是 key %s value %v\n", key, value)
	if e, ok := lf.cache[key]; ok {
		e.referenced()
		fmt.Println("未更新：当前容量：", lf.nBytes)
		lf.nBytes += int64(value.Len()) - int64(e.entry.value.Len())
		fmt.Println("更新时：当前容量：", lf.nBytes)
		e.entry.value = value
		heap.Fix(lf.priority, e.index)
	} else {
		e := &lfuEntry{
			index: 0,
			entry: entry{
				key:      key,
				value:    value,
				updateAt: nil,
			},
		}
		e.referenced()
		lf.nBytes += int64(len(key) + value.Len())
		fmt.Println("插入时：当前容量：", lf.nBytes)
		heap.Push(lf.priority, e)
		lf.cache[key] = e
	}

	for lf.maxCacheSize > 0 && lf.nBytes > lf.maxCacheSize {
		lf.Remove()
	}
}
```

### Remove 淘汰缓存

```go
func (lf *LfuCache) Remove() {
	e := heap.Pop(lf.priority).(*lfuEntry)
	fmt.Println(e.entry)
	lf.nBytes -= int64(len(e.entry.key) + e.entry.value.Len())
	fmt.Println("删除时：当前容量：", lf.nBytes)
	delete(lf.cache, e.entry.key)
	if lf.onEvicted != nil {
		lf.onEvicted(e.entry.key, e.entry.value)
	}
}

### referenced 更新访问计数和访问时间
func (lfuEntry *lfuEntry) referenced() {
	lfuEntry.count++
	now := time.Now()
	lfuEntry.entry.updateAt = &now
}

func (lf *LfuCache) Len() int {
	return lf.priority.Len()
}

```
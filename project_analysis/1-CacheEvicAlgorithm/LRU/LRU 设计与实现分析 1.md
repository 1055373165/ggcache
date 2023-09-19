# LRU 缓存淘汰策略

# 简介

三种缓存淘汰算法：FIFO、LFU、LRU；

## FIFO（First In First Out）

- 淘汰策略

<aside>
🏔️ First In First Out 算法认为：最早添加的记录，其不再被使用的可能性比刚添加的记录要大，因此在缓存大小达到内存设定时，优先淘汰最早添加的记录；


</aside>

- 算法实现

这种算法的实现比较简单：只需要维护一个队列，新添加的记录添加到队尾，需要删除缓存记录时从队列移除即可。

- 缺点

在很多场景下，部分最早添加的记录也很有可能最常被访问，但是这些高频访问记录还是会因为很早添加被优先淘汰，结果导致这类数据淘汰后又添加到缓存中，导致缓存的命中率降低。

## LFU（Least Frequentry Used）

- 淘汰策略

<aside>
🏔️ LFU 算法认为：最少被使用的记录在将来被访问的概率最小，因此优先删除低频访问记录。


</aside>

- 算法实现

LFU 的算法实现需要维护一个按照访问次数排序的队列，每次访问记录时，对应记录的访问记录数加 1，队列按照新的记录数进行重排，在需要淘汰记录时选择淘汰那些访问次数少的记录，即队列尾部的那些记录。LFU 算法的命中率比较高，但缺点也很明显；

- 缺点

如果数据的访问模式发生变化，LFU 需要很长的时间才能适应，即 LFU 算法受历史数据的影响比较大。例如某个数据历史上访问次数特别高，但在某个时间点之后几乎不再使用了，但因为它的历史访问次数比较高，仍然会排在队列的前部，迟迟不能被释放。

- 分析

上面这两种算法各有其优缺点，FiFO 算法实现起来非常简单，但是对于一些可能访问旧的记录可能性更大的场景往往性能较差；比如一个公司有老员工和一些新员工，在实际业务场景中，老员工的访问次数大部分情况下都要高于新员工，即使老员工是先添加进来的。

LFU 根据访问次数对记录进行排序，命中率比较高，但是在访问模式发生变化时，一些可能不再使用的记录却不能及时从缓存中删除，受历史数据的影响比较大。所以有没有一种比较折中的方案呢？有，它就是 LRU，最近最少访问缓存淘汰算法。

## LRU（Least Recently Used）

- 淘汰策略

<aside>
🏔️ LRU认为：如果数据最近被访问，那么将来被访问的概率也会更高，因此在需要淘汰缓存条目时优先选择那些最近没有被访问记录。


</aside>

相对于进考虑时间因素的 FIFO 和仅考虑访问频率的 LFU，LRU 算法可以认为是一种相对平衡的淘汰算法。

- 算法实现

维护一个队列（双向链表模拟）来存储键值记录，使用另一种数据结构 map 维护键到双向链表的映射，当需要访问某个键的值时，根据 map 找到键在双向链表中的位置，即可在 O(1) 时间内完成值的读取、也可以在 O(1) 时间内移动和删除记录。

# LRU 算法实现

# 核心数据结构

### 1. 创建 LRU 缓存结构体

创建一个包含字典和双向链表的结构体类型 Cache，方便实现后续的增删查改操作

```go
package lru

import (
	"container/list"
)

type Cache struct {
	maxByets int64 // 当设置为 0 时表示不设置内存使用上限
	nBytes int64
	ll *list.List
	cache map[string]*list.Element
	// Value 是一个接口类型，便于统计每种类型占用的字节数
	onEvicted func(key string, value Value)
}

type Entry struct {
	key string
	value Value
}

type Value interface {
	Len() int
}

```

- 使用 go 语言标准库提供的双向链表 list.List。
- map 的定义是 map[string]*list.Element，键是字符串类型，值是双向链表中对应的节点指针。
- maxBytes 是允许使用的最大内存；
- nBytes 是当前缓存已经使用的内存空间；
- OnEvicted 是某条记录被移除时的回调函数，可以为 nil；
- 键值对以结构体形式存储（Entry），它也是双向链表节点存储的数据类型，在链表中仍然保留每个值对应的 key 是因为：在淘汰队尾节点时，需要用从字典中删除对应的映射关系。

<aside>
🏔️ （触发缓存淘汰时，我们只知道双向链表的表尾节点，如果表尾节点中没有存储 key 的话，我们就无法获取这个节点对应的 key，自然就无法根据 key 在 map 中删除 key 和双向链表节点的映射关系）


</aside>

为了通用性，我们允许值是任何实现了 Value 接口的类型，该接口只包含了一个方法 Len() int，用于返回实现了 Value 接口的具体类型所占的内存大小。即只要一个类型实现了 Len 方法提供了自己占用的内存大小，那么就相对于实现了接口 Value，就可以存储到Entry 中，进而可以作为双向链表的数据类型存储到 LRU 缓存中。

### 2. 实例化 LRU Cache

为了方便实例化 LRU Cache 结构体，提供实例化函数；

需要提供允许存储的最大内存和删除 kv 对时的回调函数（可以为 nil）。

与此同时，在我们插入新纪录是需要构造新的记录条目，在这里也给出快捷方式。

```go
/**
* 用于新建 LRU 缓存实例
*
* @param maxBytes 允许 lru 缓存占用的最大内存空间
* @param onEvicted 删除记录时触发的回调函数
* @return LRU 缓存实例
*/
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes: maxBytes,
		ll: new(list.List), // list.New()
		cache: make(map[string]*list.Element),
		onEvicted: onEvicted,
	}
}

// +++++
func NewEntry(key string, value Value) *Entry {
	return &Entry{
		key: key,
		value: value,
	}
}
```

### 3. 查询功能 Get

查找功能主要有 2 个步骤：

1. 从 map 中找到 key 对应的双向链表节点的指针；
2. 如果找到，将该节点移动到双向链表的表头位置；
3. 如果没找到，直接返回 -1，表示没有命中。

```go
/**
* 根据指定的 key 从缓存中读取它的 value
*
* @param key 要查询的键
* @return value 查询到的键的值，为 Value 接口类型
* @return ok 查询是否成功的标识，bool 类型
*/
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		// 调用 list 库提供的函数，将该节点移动到链表的表头；
		c.ll.MoveToFront(ele)
		// 存储的数据类型为 &Entry，存储到双向链表时会被转换为空接口类型
		// 因此需要对空接口类型进行类型断言
		kv := ele.Value.(*Entry)
		return kv.value, true
	}

	return nil, false
}

```

### 4. 新增 or 修改记录 Put

```go
/**
* 往 LRU 缓存中提交新的记录，如果记录不存在就新增记录，否则更新记录
*
* @param key 要插入或者更新的键
* @param value 要插入的键的值 Value 接口类型，可以提供占用的内存大小
*
*/
func (c *Cache) Put(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		kv := ele.Value.(*Entry)
		kv.value = value
		// 由于 value 类型可能发生改变，要更新缓存占用
		// 将两者差值加到 nBytes 字段上
		c.nBytes += int64(value.Len()) - int64(kv.value.Len())
		return
	}

	// 新建记录条目，插入到链表表头并返回对应的链表节点
	newEntry := NewEntry(key, value)
	ele := c.ll.PushFront(newEntry)
	// 更新 nBytes 和映射关系
	c.cache[key] = ele
	// todo: 这里实际上有 bug，默认 key 的每个字符占用一个字节，但汉字等字符一个位置占用 3 个字节；但是即使是汉字，len 计算是也已经默认包含进去了，所以不需要修改。
	// 误会解除 😁
	c.nBytes += int64(len(key)) + int64(value.Len())

	// 插入新的记录后，判断占用的内存大小是否超过限制，如果超出限制，循环删除队尾记录，直至满足内存限制
	for c.maxBytes != 0 && c.maxBytes < c.nBytes {
		c.RemoveOldest()
	}
}

```

### 5. 删除

这里删除对应了缓存淘汰，即移除最近最少访问的节点（队尾）；

```go
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*Entry)
		// 更新 nBytes & map
		nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		// 这里就是在链表存储的数据中设置 key 的作用
		delete(c.cache, kv.key)
		// 如果删除记录时的回调函数设置的不是 nil，那么调用回调函数
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

```

# 测试

最后，为了方便测试，我们实现 `Len()` 用来获取添加了多少条数据。

```go
func (c *Cache) Len() int {
	return c.ll.Len()
}

```

```go
type String string

func (d String) Len() int {
	return len(d)
}

func TestGet(t *testing.T) {
	lru := New(int64(0), nil)
	lru.Put("key1", String("1234"))

	if v, ok := lru.Get("key1"); !ok || string(v.(String)) != "1234" {
		t.Fatal("cache hit key1=1234 failed")
	}

	if _, ok := lru.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
}
```


# 整体代码

```go
package main

import (
	"container/list"
)

type Cache struct {
	maxBytes int64
	nBytes   int64
	ll       *list.List
	cache    map[string]*list.Element
	// Value 是一个接口类型，便于统计每种类型占用的字节数
	onEvicted func(key string, value Value)
}

type Entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

/**
* 用于新建 LRU 缓存实例
*
* @param maxBytes 允许 lru 缓存占用的最大内存空间
* @param onEvicted 删除记录时触发的回调函数
* @return LRU 缓存实例
 */
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        new(list.List), // list.New()
		cache:     make(map[string]*list.Element),
		onEvicted: onEvicted,
	}
}

// +++++
func NewEntry(key string, value Value) *Entry {
	return &Entry{
		key:   key,
		value: value,
	}
}

/**
* 根据指定的 key 从缓存中读取它的 value
*
* @param key 要查询的键
* @return value 查询到的键的值，为 Value 接口类型
* @return ok 查询是否成功的标识，bool 类型
 */
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		// 调用 list 库提供的函数，将该节点移动到链表的表头；
		c.ll.MoveToFront(ele)
		// 存储的数据类型为 &Entry，存储到双向链表时会被转换为空接口类型
		// 因此需要对空接口类型进行类型断言
		kv := ele.Value.(*Entry)
		return kv.value, true
	}

	return nil, false
}

/**
* 往 LRU 缓存中提交新的记录，如果记录不存在就新增记录，否则更新记录
*
* @param key 要插入或者更新的键
* @param value 要插入的键的值 Value 接口类型，可以提供占用的内存大小
*
 */
func (c *Cache) Put(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		kv := ele.Value.(*Entry)
		kv.value = value
		// 由于 value 类型可能发生改变，要更新缓存占用
		// 将两者差值加到 nBytes 字段上
		c.nBytes += int64(value.Len()) - int64(kv.value.Len())
	} else {
		// 新建记录条目，插入到链表表头并返回对应的链表节点
		newEntry := NewEntry(key, value)
		ele := c.ll.PushFront(newEntry)
		// 更新 nBytes 和映射关系
		c.cache[key] = ele
		// todo: 这里实际上有 bug，默认 key 的每个字符占用一个字节，但汉字等字符一个位置占用 3 个字节；但是即使是汉字，len 计算是也已经默认包含进去了，所以不需要修改。
		// 误会解除 😁
		c.nBytes += int64(len(key)) + int64(value.Len())
	}

	// 插入新的记录后，判断占用的内存大小是否超过限制，如果超出限制，循环删除队尾记录，直至满足内存限制
	for c.maxBytes != 0 && c.maxBytes < c.nBytes {
		c.RemoveOldest()
	}
}

func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*Entry)
		// 更新 nBytes & map
		c.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		// 这里就是在链表存储的数据中设置 key 的作用
		delete(c.cache, kv.key)
		// 如果删除记录时的回调函数设置的不是 nil，那么调用回调函数
		if c.onEvicted != nil {
			c.onEvicted(kv.key, kv.value)
		}
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}
```

# 代码更新

## 复用查询部分

```go
package main

import (
	"container/list"
)

type Cache struct {
	maxBytes int64 // 设置为 0 默认不开启内存限制
	nBytes   int64
	cache    map[string]*list.Element
	ll       *list.List

	onEvicted func(key string, value Value)
}

type Value interface {
	Len() int
}

type Entry struct {
	key   string
	value Value
}

func New(maxBytes int64, onEvicted func(key string, value Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		onEvicted: onEvicted,
		cache:     make(map[string]*list.Element),
		ll:        list.New(),
	}
}

func NewEntry(key string, value Value) *Entry {
	return &Entry{
		key:   key,
		value: value,
	}
}

func (c *Cache) search(key string) *list.Element {
	if listnode, ok := c.cache[key]; ok {
		return listnode
	}

	return nil
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	if listnode := c.search(key); listnode != nil {
		c.ll.MoveToFront(listnode)
		kv := listnode.Value.(*Entry)
		return kv.value, true
	}

	return nil, false
}

func (c *Cache) Put(key string, value Value) {
	if listnode := c.search(key); listnode != nil {
		c.ll.MoveToFront(listnode)
		kv := listnode.Value.(*Entry)
		kv.value = value
		c.nBytes = int64(value.Len()) - int64(kv.value.Len())
	} else {
		entry := NewEntry(key, value)
		ele := c.ll.PushFront(entry)
		c.cache[key] = ele
		c.nBytes += int64(len(key)) + int64(value.Len())
	}

	for c.maxBytes != 0 && c.maxBytes < c.nBytes {
		c.RemoveOldest()
	}
}

func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()

	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*Entry)
		c.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		delete(c.cache, kv.key)

		if c.onEvicted != nil {
			c.onEvicted(kv.key, kv.value)
		}
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}
```

## 更完善的测试用例（符合接口测试）

```go
package main

import (
	"reflect"
	"testing"
)

type String string

func (s String) Len() int {
	return len(s)
}

type MyInt int

func (mi MyInt) Len() int {
	return int(reflect.TypeOf(mi).Size())
}

type Person struct {
	Name string
	Age  int
}

func (p Person) Len() int {
	return int(reflect.TypeOf(p).Size())
}

func TestGet(t *testing.T) {
	lru := New(int64(0), nil)
	lru.Put("key1", String("1234"))

	if v, ok := lru.Get("key1"); !ok || v.(String) != "1234" {
		t.Fatalf("cache hit key1=1234 failed")
	}

	if _, ok := lru.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
}

func TestRemoveOldest(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "key3"
	v1, v2, v3 := String("hello"), MyInt(2), Person{Name: "smy", Age: 23}
	// 4 5 4 8 4 24
	// 设置 cap 容量为可以容纳 k2, k3 但不能同时容下 k1, k2, k3
	// cap = 4 + 8 + 4 + 24 = 40
	lru := New(int64(40), nil)
	lru.Put(k1, v1)
	lru.Put(k2, v2)
	lru.Put(k3, v3)

	if _, ok := lru.cache[k1]; ok || lru.Len() != 2 {
		t.Fatalf("RemoveOldest key1 failed")
	}

	if v, ok := lru.cache[k3]; !ok || v.Value.(*Entry).value.(Person).Name != "smy" || v.Value.(*Entry).value.(Person).Age != 23 {
		t.Fatalf("expected name=%v and age=%v but got name=%v and age=%v", "smy", 23, v.Value.(*Entry).value.(Person).Name, v.Value.(*Entry).value.(Person).Age)
	}
}

func TestOnEvicted(t *testing.T) {
	// 回调函数设置为：将所有被删除的缓存条目的键存储到一个数组当中
	keys := make([]string, 0)
	callback := func(key string, value Value) {
		keys = append(keys, key)
	}

	// cap = 10
	// key1 value1 8
	// key2 value2 8 nBytes = 16 删除 key1 [key1]
	// key3 value3 8 nBytes = 16 删除 key2 [key1 key2]
	// key4 value4 8 nBytes = 16 删除 key3 [key1 key2 key3]
	// lru cache 剩余 key4
	lru := New(int64(10), callback)
	lru.Put("key1", String("1234"))
	lru.Put("key2", String("key2"))
	lru.Put("key3", String("key3"))
	lru.Put("key4", String("key4"))

	expected := []string{"key1", "key2", "key3"}
	if !reflect.DeepEqual(expected, keys) {
		t.Fatalf("Call OnEvicted failed, expected keys equals to %v but got %v", expected, keys)
	}

	_, ok1 := lru.cache["key1"]
	_, ok2 := lru.cache["key2"]
	_, ok3 := lru.cache["key3"]

	if ok1 || ok2 || ok3 {
		t.Fatalf("expected only key4 in cache, but got other keys")
	}
}
```


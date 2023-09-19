# day1 LRU

实现 LRU 缓存策略

## 结构体 LRUCache

[View on canvas](https://app.eraser.io/workspace/dQ1bLajGiczM2lLiraLb?elements=AcCRLYTJ3MW8oclijUpM8A) 

为结构体配备的函数

NewLRUCache 初始化一个 LRUCache 对象（ 需要提供最大缓存容量 maxCache 以及回调函数）

- maxCache
- onEnvicted
- table: make(map[string]*list.Element)
- root: list.New()


## 需要注意的点

1、对哈希表的值进行抽象，实现了 Len 方法可以提供类型长度的类型均可作为哈希表的值；

```
type Value interface {
  Len() int
}

type MyInt int

func (mi MyInt) Len() int {
  return reflect.TypeOf(mi).Len()
}
```

2、LRU 缓存最大容量使用字节为单位，同时记录当前插入的 kv 对所占的总缓存容量，以便后面淘汰缓存时使用。

```
for maxCache < nBytes {
  ...
}
```

3、删除键值对时回调 onEnvicted 函数，参数分别为 string 和 Value；



## 用于存储 kv 的结构体 Entry

```
type Entry struct {
  key string
  value Value
}
```

[View on canvas](https://app.eraser.io/workspace/dQ1bLajGiczM2lLiraLb?elements=AcCRLYTJ3MW8oclijUpM8A) 

为了方便为其提供一个构造方法不为过

```
NewEntry(key string, value Value) *Entry
```


## 核心 API

[View on canvas](https://app.eraser.io/workspace/dQ1bLajGiczM2lLiraLb?elements=rwUc8oHV-gA8F9azC25wFg) 

- Get 

```
(*LRUCache) Get(key string) (Value, bool)
```

1. 去 map 中查询 key 对应的底层链表节点是否存在
	1. 如果存在，返回对应节点的指针
	2. 如果不存在，返回 nil
2. 判断查询得到的节点是否为 nil
	1. 如果为 nil，返回 nil 和 false，表示查询失败；
	2. 如果查询成功，将该节点移动到链表首部；*list.Element 类型拥有一个 Value 字段（存储了节点存放的值，类型为接口类型），需要对接口进行断言（*Entry），然后从 Entry 结构体中取出 value 的值，返回 value 值和 true，表示查询成功。

- Put

```
(*LRUCache) Put(key string, value Value) 
```

1. 同样从 map 中查询 key 对应的值是否已经在缓存中
	1. 查询成功，返回底层节点的指针
	2. 否则，返回 nil
2. 判断返回的节点指针是否为 nil
	1. 如果为 nil，则说明 key 还未缓存，因此构造一个 Entry，通过链表插入到缓存中的首部，与此同时更新缓存已经使用的容量（即需要加上新的 key、value 占用的字节大小），然后将 key 的映射信息保存到 map 中。还没完，如果插入新的 Entry 导致已经使用的缓存超出了设置的最大缓存容量，就需要淘汰掉最近最少使用的条目，即从末尾开始移除（RemoveOldest），直至缓存使用的容量小于设置的最大缓存容量为止。（lru 不满足高级局部性）
	2. 如果缓存已经存在，那么先将对应的节点从链表中移除，然后插入到链表的首部（MoveToFront），接着更新节点存储的 value 的值，并且重新计算使用的缓存容量，缓存淘汰判断逻辑和 2.a 一致。


## 最终代码

```go
package lru

import "container/list"

type LRUCache struct {
	maxCache   int64
	nBytes     int64
	root       *list.List
	m          map[string]*list.Element
	onEnvctied func(string, Value)
}

type Value interface {
	Len() int
}

func NewLRUCache(maxCache int64, onEnvicted func(string, Value)) *LRUCache {
	return &LRUCache{
		maxCache:   maxCache,
		onEnvctied: onEnvicted,
		root:       list.New(),
		m:          make(map[string]*list.Element),
	}
}

type Entry struct {
	key   string
	value Value
}

func NewEntry(key string, value Value) *Entry {
	return &Entry{key: key, value: value}
}

func (lc LRUCache) Len() int {
	return lc.root.Len()
}

func (lc LRUCache) search(key string) *list.Element {
	if e, ok := lc.m[key]; ok {
		return e
	}
	return nil
}

func (lc LRUCache) Get(key string) (Value, bool) {
	if ele := lc.search(key); ele == nil {
		return nil, false
	} else {
		// 访问需要移动到队首
		lc.root.MoveToFront(ele)
		return ele.Value.(*Entry).value, true
	}
}

func (lc *LRUCache) Put(key string, value Value) {
	if ele := lc.search(key); ele != nil {
		lc.root.MoveToFront(ele)
		entry := ele.Value.(*Entry)
		lc.nBytes += int64(value.Len() - entry.value.Len())
		entry.value = value
	} else {
		entry := NewEntry(key, value)
		kv := lc.root.PushFront(entry)
		lc.nBytes += int64(len(key) + value.Len())
		lc.m[key] = kv
	}

	for lc.maxCache > 0 && lc.maxCache < lc.nBytes {
		lc.RemoveOldest()
	}
}

func (lc *LRUCache) RemoveOldest() {
	entry_Interface := lc.root.Remove(lc.root.Back()).(*Entry)
	key, value := entry_Interface.key, entry_Interface.value
	lc.nBytes -= int64(len(key) + value.Len())
	delete(lc.m, key)

	if lc.onEnvctied != nil {
		lc.onEnvctied(key, value)
	}
}
```


# 测试代码

```go
package lru

import (
	"fmt"
	"testing"
)

type MyInt int

func (mi MyInt) Len() int {
	return 8
}

func TestLru(t *testing.T) {
	lru := NewLRUCache(18, nil)
	// value length = 8
	m1, m2, m3, m4 := MyInt(1), MyInt(2), MyInt(3), MyInt(4)
	// key string  length = 1
	// 因此 18 最多只能存两个 entry 组合
	lru.Put("1", m1)
	lru.Put("2", m2)
	m11, _ := lru.Get("1")
	if m11 != m1 {
		t.Fatalf("expected %d but got %d", m1, m11)
	}
	fmt.Println(lru.nBytes)
	// 现在 1 位于队首
	// 我们插入 3, m3 后 "2" 应该被淘汰
	lru.Put("3", m3)
	m22, ok := lru.Get("2")
	if ok != false || m22 != nil {
		t.Fatalf("lru cache failed")
	}

	// 3 1
	lru.Put("4", m4) // 1 应该被淘汰
	_, ok = lru.Get("1")
	if ok || lru.Len() != 2 {
		t.Fatal("lru cache failed")
	}
}
```


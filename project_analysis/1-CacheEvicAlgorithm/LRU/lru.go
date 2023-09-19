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

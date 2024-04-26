package lru

/*LRU缓存淘汰策略*/

import (
	"container/list"
	"time"

	"ggcache/internal/service/cachepurge/interfaces"
)

type LRUCache struct {
	maxBytes int64
	nbytes   int64
	root     *list.List
	cache    map[string]*list.Element
	// optional and executed when an entry is purged.
	// 回调函数，采用依赖注入的方式，该函数用于处理从缓存中淘汰的数据
	OnEvicted func(key string, value interfaces.Value)
}

func NewLRUCache(maxBytes int64, onEvicted func(string, interfaces.Value)) *LRUCache {
	return &LRUCache{
		maxBytes:  maxBytes,
		root:      list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

func (c *LRUCache) RemoveOldest() {
	ele := c.root.Front()
	if ele != nil {
		kv := c.root.Remove(ele).(*interfaces.Entry)
		delete(c.cache, kv.Key)
		c.nbytes -= int64(len(kv.Key)) + int64(kv.Value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.Key, kv.Value)
		}
	}
}

func (c *LRUCache) Get(key string) (value interfaces.Value, updateAt *time.Time, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.root.MoveToBack(ele)
		e := ele.Value.(*interfaces.Entry)
		e.Touch()
		return e.Value, e.UpdateAt, ok
	}
	return
}

func (c *LRUCache) Put(key string, value interfaces.Value) {
	// Update Operation
	if ele, ok := c.cache[key]; ok {
		c.root.MoveToBack(ele)
		kv := ele.Value.(*interfaces.Entry)
		kv.Touch()
		// update cache nbytes
		c.nbytes += int64(value.Len()) - int64(kv.Value.Len())
		// update cache entry's value
		kv.Value = value
	} else { // Put Operation
		kv := &interfaces.Entry{Key: key, Value: value, UpdateAt: nil}
		kv.Touch()
		ele := c.root.PushBack(kv)
		c.cache[key] = ele
		c.nbytes += int64(len(kv.Key)) + int64(kv.Value.Len())
	}
	// cache evicted trigger
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

func (c *LRUCache) Len() int {
	return c.root.Len()
}

func (c *LRUCache) CleanUp(ttl time.Duration) {
	for e := c.root.Front(); e != nil; e = e.Next() {
		if e.Value.(*interfaces.Entry).Expired(ttl) {
			kv := c.root.Remove(e).(*interfaces.Entry)
			delete(c.cache, kv.Key)
			c.nbytes -= int64(len(kv.Key)) + int64(kv.Value.Len())
			if c.OnEvicted != nil {
				c.OnEvicted(kv.Key, kv.Value)
			}
		} else {
			break
		}
	}
}

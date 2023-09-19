package policy

/*LRU缓存淘汰策略*/

import (
	"container/list"
	"time"
)

type LRUCache struct {
	maxBytes int64
	nbytes   int64
	ll       *list.List
	cache    map[string]*list.Element
	// optional and executed when an entry is purged.
	// 回调函数，采用依赖注入的方式，该函数用于处理从缓存中淘汰的数据
	OnEvicted func(key string, value Value)
}

type Value interface {
	Len() int
}

func (c *LRUCache) RemoveOldest() {
	ele := c.ll.Front()
	if ele != nil {
		kv := c.ll.Remove(ele).(*entry)
		delete(c.cache, kv.key)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *LRUCache) Get(key string) (value Value, updateAt *time.Time, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToBack(ele)
		e := ele.Value.(*entry)
		e.touch()
		//更新数据过期时间
		return e.value, e.updateAt, ok
	}
	return
}

func (c *LRUCache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToBack(ele)
		//更新value
		kv := ele.Value.(*entry)
		kv.touch()
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		kv := &entry{key, value, nil}
		kv.touch()
		ele := c.ll.PushBack(kv)
		c.cache[key] = ele
		c.nbytes += int64(len(kv.key)) + int64(kv.value.Len())
	}
	//内存溢出进行删除
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

func (c *LRUCache) Len() int {
	return c.ll.Len()
}

func (c *LRUCache) CleanUp(ttl time.Duration) {
	for e := c.ll.Front(); e != nil; e = e.Next() {
		if e.Value.(*entry).expired(ttl) {
			kv := c.ll.Remove(e).(*entry)
			delete(c.cache, kv.key)
			c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
			if c.OnEvicted != nil {
				c.OnEvicted(kv.key, kv.value)
			}
		} else {
			break
		}
	}
}

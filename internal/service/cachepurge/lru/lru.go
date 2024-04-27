package lru

/*LRU缓存淘汰策略*/

import (
	"container/list"
	"sync"
	"time"

	"github.com/1055373165/ggcache/config"
	"github.com/1055373165/ggcache/internal/service/cachepurge/interfaces"
	"github.com/1055373165/ggcache/utils/logger"
)

type LRUCache struct {
	maxBytes int64
	nbytes   int64
	root     *list.List
	mu       sync.RWMutex
	cache    map[string]*list.Element
	// optional and executed when an entry is purged.
	// 回调函数，采用依赖注入的方式，该函数用于处理从缓存中淘汰的数据
	OnEvicted func(key string, value interfaces.Value)
}

func NewLRUCache(maxBytes int64, onEvicted func(string, interfaces.Value)) *LRUCache {
	l := &LRUCache{
		maxBytes:  maxBytes,
		root:      list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}

	ttl := time.Duration(config.Conf.Services["groupcache"].TTL) * time.Second

	go func() {
		ticker := time.NewTicker(2 * time.Minute)
		defer ticker.Stop()

		for {
			<-ticker.C
			l.CleanUp(ttl)
			logger.LogrusObj.Warnf("触发过期缓存清理后台任务...")
		}
	}()

	return l
}

func (c *LRUCache) RemoveOldest() {
	ele := c.root.Front()
	if ele != nil {
		c.mu.Lock()
		if ele == nil { // 还有 CleanUp 并发 goroutine 的过期淘汰策略，因此需要进行并发安全双检，否则对 nil interface 进行断言直接触发 panic
			c.mu.Unlock()
			return
		}
		kv := c.root.Remove(ele).(*interfaces.Entry)
		delete(c.cache, kv.Key)
		c.nbytes -= int64(len(kv.Key)) + int64(kv.Value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.Key, kv.Value)
		}
		c.mu.Unlock()
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
		c.mu.Lock()
		if e.Value == nil { // 还有缓存自动淘汰策略 RemoveOldest 的存在，需要进行并发安全双检
			c.mu.Unlock()
			continue
		}
		if e.Value.(*interfaces.Entry).Expired(ttl) {
			kv := c.root.Remove(e).(*interfaces.Entry)
			delete(c.cache, kv.Key)
			c.nbytes -= int64(len(kv.Key)) + int64(kv.Value.Len())
			if c.OnEvicted != nil {
				c.OnEvicted(kv.Key, kv.Value) // 定义了缓存被淘汰时的回调函数，比如更新数据库、释放资源、发送更新通知等
			}
		}
	}
	c.mu.Unlock()
}

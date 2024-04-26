package service

import (
	"sync"

	"ggcache/utils/logger"

	"ggcache/internal/service/policy"
)

// cache 模块负责提供对lru模块的并发控制

// 给 lru 上层并发上一层锁
type cache struct {
	mu           sync.Mutex
	lru          *policy.LRUCache
	maxCacheSize int64 // 保证 lru 一定初始化
}

func newCache(cacheSize int64) *cache {
	return &cache{
		maxCacheSize: cacheSize,
	}
}

// 并发控制
func (c *cache) set(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil { // 100 MB 内存缓存
		c.lru = policy.New("lru", 100*2<<20, nil).(*policy.LRUCache)
	}
	c.lru.Add(key, value)
}
func (c *cache) get(key string) (ByteView, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = policy.New("lru", 100*2<<20, nil).(*policy.LRUCache)
	}

	if v, _, ok := c.lru.Get(key); ok { // Get 返回值是 Value 接口，直接类型断言
		return v.(ByteView), true
	} else {
		return ByteView{}, false
	}
}

func (c *cache) put(key string, val ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil { // 策略类模式
		c.lru = policy.New("lru", 100*2<<20, nil).(*policy.LRUCache)
	}
	logger.LogrusObj.Infof("存入数据库之后压入缓存, (key, value)=(%s, %s)", key, val)
	c.lru.Add(key, val)
}

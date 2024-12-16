package cache

import (
	"sync"

	"github.com/1055373165/ggcache/pkg/common/logger"

	"github.com/1055373165/ggcache/internal/cache/eviction"
)

// Cache is a concurrent safe cache that evicts least recently used items.
type Cache struct {
	mu           sync.RWMutex
	strategy     eviction.CacheStrategy
	maxCacheSize int64
}

// NewCache creates a new Cache with the given maximum bytes capacity.
func NewCache(strategy string, cacheSize int64) *Cache {
	onEvicted := func(key string, val eviction.Value) {
		logger.LogrusObj.Infof("缓存条目 [%s:%s] 被淘汰", key, val)
	}

	s, err := eviction.New(strategy, cacheSize, onEvicted)
	if err != nil {
		logger.LogrusObj.Errorf("eviction.New failed: %s", err)
		return nil
	}

	return &Cache{
		maxCacheSize: cacheSize,
		strategy:     s,
	}
}

// add adds a value to the cache.
func (c *Cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.strategy.Put(key, value)
}

// get looks up a key's value from the cache.
func (c *Cache) get(key string) (ByteView, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if v, _, ok := c.strategy.Get(key); ok {
		return v.(ByteView), true
	}
	return ByteView{}, false
}

// put adds a value to the cache.
func (c *Cache) put(key string, val ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()

	logger.LogrusObj.Infof("存入数据库之后压入缓存, (key, value)=(%s, %s)", key, val)
	c.strategy.Put(key, val)
}

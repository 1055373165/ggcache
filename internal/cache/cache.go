package cache

import (
	"fmt"
	"sync"

	"github.com/1055373165/ggcache/pkg/common/logger"

	"github.com/1055373165/ggcache/internal/cache/eviction"
)

// Cache is a concurrent safe cache that evicts least recently used items.
type cache struct {
	mu           sync.RWMutex
	strategy     eviction.CacheStrategy
	maxCacheSize int64
}

// NewCache creates a new Cache with the given maximum bytes capacity.
// It returns an error if the strategy is invalid or if cacheSize is <= 0.
func NewCache(strategy string, cacheSize int64) (*cache, error) {
	if cacheSize <= 0 {
		return nil, fmt.Errorf("cache size must be positive, got %d", cacheSize)
	}

	onEvicted := func(key string, val eviction.Value) {
		logger.LogrusObj.Infof("Cache entry evicted [key=%s]", key)
	}

	s, err := eviction.New(strategy, cacheSize, onEvicted)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache strategy: %w", err)
	}

	return &cache{
		maxCacheSize: cacheSize,
		strategy:     s,
	}, nil
}

// Get looks up a key's value from the cache.
func (c *cache) get(key string) (value ByteView, ok bool) {
	if c == nil {
		return ByteView{}, false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()

	if v, _, exists := c.strategy.Get(key); exists {
		if bv, ok := v.(ByteView); ok {
			return bv, true
		}
	}
	return ByteView{}, false
}

// put adds a value to the cache.
func (c *cache) put(key string, val ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()

	logger.LogrusObj.Infof("存入数据库之后压入缓存, (key, value)=(%s, %s)", key, val)
	c.strategy.Put(key, val)
}

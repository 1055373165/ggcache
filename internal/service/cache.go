package service

import (
	"log"
	"sync"

	"ggcache/internal/service/cachepurge"
	"ggcache/internal/service/cachepurge/interfaces"
)

/*
	the cache module is responsible for providing concurrency control for cachepurge modules, such as lru、lfu、fifo.
*/

// add a mutex lock to lru to achieve concurrency safety
type cache struct {
	mu           sync.Mutex
	strategy     interfaces.CacheStrategy
	maxCacheSize int64
}

func newCache(strategy string, cacheSize int64) *cache {
	// defaut cache purge strategy is lru
	if strategy == "" {
		strategy = "lru"
	}

	// default cache purge threshold is 2^10
	if cacheSize == 0 {
		cacheSize = 2 << 10
	}

	return &cache{
		maxCacheSize: cacheSize,
		strategy:     cachepurge.New(strategy, cacheSize, nil),
	}
}

func (c *cache) get(key string) (ByteView, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// lru evited policy is default
	if c.strategy == nil {
		c.strategy = cachepurge.New("lru", 2<<10, nil)
	}
	// the return value of Get is the Value interface, direct type assertion
	if v, _, ok := c.strategy.Get(key); ok {
		log.Printf("cache hit, key: %v, value: %v\n", key, v)
		return v.(ByteView), true
	} else {
		return ByteView{}, false
	}
}

// set and put semantics is equivalent to CacheStrategy.Put method.
func (c *cache) set(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.strategy == nil {
		c.strategy = cachepurge.New("lru", 2<<10, nil)
	}
	c.strategy.Put(key, value)
}

func (c *cache) put(key string, val ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.strategy == nil {
		c.strategy = cachepurge.New("lru", 2<<10, nil)
	}
	c.strategy.Put(key, val)
}

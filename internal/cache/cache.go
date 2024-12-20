// Package cache implements a distributed caching system with support for various eviction strategies.
package cache

import (
	"fmt"
	"sync"
	"time"

	"github.com/1055373165/ggcache/internal/cache/eviction"
	"github.com/1055373165/ggcache/internal/metrics"
	"github.com/1055373165/ggcache/pkg/common/logger"
)

// cache represents a concurrent-safe cache that supports different eviction strategies.
// The zero value for cache is not usable; use NewCache to create a cache.
type cache struct {
	mu       sync.RWMutex // protects strategy
	strategy eviction.CacheStrategy
	maxBytes int64
}

// NewCache creates a new cache with the specified eviction strategy and maximum size in bytes.
// It returns an error if the strategy is invalid or if maxBytes is not positive.
func NewCache(strategy string, maxBytes int64) (*cache, error) {
	if maxBytes <= 0 {
		return nil, fmt.Errorf("cache size must be positive, got %d", maxBytes)
	}

	onEvicted := func(key string, val eviction.Value) {
		logger.LogrusObj.Infof("Cache entry evicted: key=%s", key)
	}

	s, err := eviction.New(strategy, maxBytes, onEvicted)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache strategy: %w", err)
	}

	return &cache{
		maxBytes: maxBytes,
		strategy: s,
	}, nil
}

// It returns the value and whether the key was found.
func (c *cache) get(key string) (ByteView, bool) {
	if c == nil {
		return ByteView{}, false
	}

	start := time.Now()
	defer func() {
		metrics.ObserveRequestDuration("get", time.Since(start).Seconds())
	}()

	c.mu.RLock()
	defer c.mu.RUnlock()

	if v, _, exists := c.strategy.Get(key); exists {
		if bv, ok := v.(ByteView); ok {
			metrics.RecordCacheHit()
			return bv, true
		}
		logger.LogrusObj.Warnf("Invalid cache value type for key=%s", key)
	}
	metrics.RecordCacheMiss()
	return ByteView{}, false
}

// put adds a key-value pair to the cache.
// If the key already exists, its value will be updated.
func (c *cache) put(key string, value ByteView) {
	if c == nil {
		return
	}

	start := time.Now()
	defer func() {
		metrics.ObserveRequestDuration("put", time.Since(start).Seconds())
	}()

	c.mu.Lock()
	defer c.mu.Unlock()

	logger.LogrusObj.Infof("Update to cache: key=%s, value=%v", key, value)
	c.strategy.Put(key, value)
}

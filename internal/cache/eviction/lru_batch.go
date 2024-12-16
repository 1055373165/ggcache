package eviction

import (
	"container/list"
	"sync"
	"time"
)

const (
	defaultBatchSize = 100 // Default size for batch operations
)

// CacheUseLRUBatch implements a Least Recently Used (LRU) cache with batch processing.
type CacheUseLRUBatch struct {
	maxBytes        int64
	nbytes          int64
	root            *list.List
	mu              sync.RWMutex
	cache           map[string]*list.Element
	OnEvicted       func(key string, value Value)
	cleanupInterval time.Duration
	ttl             time.Duration
	stopCleanup     chan struct{}
	batchSize       int
}

// NewCacheUseLRUBatch creates a new LRU cache with batch processing capabilities.
func NewCacheUseLRUBatch(maxBytes int64, onEvicted func(string, Value)) *CacheUseLRUBatch {
	c := &CacheUseLRUBatch{
		maxBytes:        maxBytes,
		root:            list.New(),
		cache:           make(map[string]*list.Element),
		OnEvicted:       onEvicted,
		cleanupInterval: defaultCleanupInterval,
		ttl:             defaultTTL,
		stopCleanup:     make(chan struct{}),
		batchSize:       defaultBatchSize,
	}

	go c.cleanupRoutine()
	return c
}

// SetBatchSize sets the size for batch operations.
func (c *CacheUseLRUBatch) SetBatchSize(size int) {
	if size > 0 {
		c.batchSize = size
	}
}

// SetTTL sets the time-to-live for cache entries.
func (c *CacheUseLRUBatch) SetTTL(ttl time.Duration) {
	c.ttl = ttl
}

// SetCleanupInterval sets the interval between cleanup runs.
func (c *CacheUseLRUBatch) SetCleanupInterval(interval time.Duration) {
	c.cleanupInterval = interval
}

// Stop stops the cleanup routine.
func (c *CacheUseLRUBatch) Stop() {
	close(c.stopCleanup)
}

// cleanupRoutine periodically cleans up expired entries.
func (c *CacheUseLRUBatch) cleanupRoutine() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.CleanUp(c.ttl)
		case <-c.stopCleanup:
			return
		}
	}
}

// CleanUp removes all expired entries from the cache in batches.
func (c *CacheUseLRUBatch) CleanUp(ttl time.Duration) {
	for {
		expired := make([]*list.Element, 0, c.batchSize)

		// First pass: identify expired entries with read lock
		c.mu.RLock()
		var next *list.Element
		count := 0
		for e := c.root.Front(); e != nil && count < c.batchSize; e = next {
			next = e.Next()
			if e.Value != nil && e.Value.(*Entry).Expired(ttl) {
				expired = append(expired, e)
			}
			count++
		}
		hasMore := next != nil
		c.mu.RUnlock()

		// Second pass: remove expired entries with write lock
		if len(expired) > 0 {
			c.mu.Lock()
			for _, e := range expired {
				if e.Value != nil && e.Value.(*Entry).Expired(ttl) {
					c.removeElement(e)
				}
			}
			c.mu.Unlock()
		}

		if !hasMore {
			break
		}
	}
}

// Get retrieves a value from the cache.
func (c *CacheUseLRUBatch) Get(key string) (value Value, updateAt time.Time, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if ele, ok := c.cache[key]; ok {
		c.root.MoveToBack(ele)
		e := ele.Value.(*Entry)
		e.Touch()
		return e.Value, e.UpdateAt, true
	}
	return nil, time.Time{}, false
}

// Put adds or updates a value in the cache.
func (c *CacheUseLRUBatch) Put(key string, value Value) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ele, ok := c.cache[key]; ok {
		c.root.MoveToBack(ele)
		entry := ele.Value.(*Entry)
		c.nbytes += int64(value.Len()) - int64(entry.Value.Len())
		entry.Value = value
		entry.Touch()
	} else {
		entry := &Entry{
			Key:      key,
			Value:    value,
			UpdateAt: time.Now(),
		}
		ele := c.root.PushBack(entry)
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}

	// Batch eviction
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		if !c.removeOldestBatch() {
			break
		}
	}
}

// removeOldestBatch removes a batch of least recently used items.
func (c *CacheUseLRUBatch) removeOldestBatch() bool {
	if c.root.Len() == 0 {
		return false
	}

	removed := 0
	for e := c.root.Front(); e != nil && removed < c.batchSize; e = c.root.Front() {
		c.removeElement(e)
		removed++
	}

	return removed > 0
}

// removeElement removes an element from the cache.
func (c *CacheUseLRUBatch) removeElement(e *list.Element) {
	entry := c.root.Remove(e).(*Entry)
	delete(c.cache, entry.Key)
	c.nbytes -= int64(len(entry.Key)) + int64(entry.Value.Len())
	if c.OnEvicted != nil {
		c.OnEvicted(entry.Key, entry.Value)
	}
}

// Len returns the number of items in the cache.
func (c *CacheUseLRUBatch) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.root.Len()
}

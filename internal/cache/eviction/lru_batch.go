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
	cache           map[string]*list.Element
	OnEvicted       func(key string, value Value)
	cleanupInterval time.Duration
	ttl             time.Duration
	stopCleanup     chan struct{}
	batchSize       int
	mu              sync.RWMutex
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
		batchSize:       defaultBatchSize,
	}
	return c
}

// Start starts the cleanup routine.
func (c *CacheUseLRUBatch) Start() {
	if c.stopCleanup == nil {
		c.stopCleanup = make(chan struct{})
		go c.cleanupRoutine()
	}
}

// SetBatchSize sets the size for batch operations.
func (c *CacheUseLRUBatch) SetBatchSize(size int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if size > 0 {
		c.batchSize = size
	}
}

// SetTTL sets the time-to-live for cache entries.
func (c *CacheUseLRUBatch) SetTTL(ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ttl = ttl
}

// SetCleanupInterval sets the interval between cleanup runs.
func (c *CacheUseLRUBatch) SetCleanupInterval(interval time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.stopCleanup != nil {
		close(c.stopCleanup)
	}
	c.cleanupInterval = interval
	c.Start()
}

// Stop stops the cleanup routine.
func (c *CacheUseLRUBatch) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.stopCleanup != nil {
		close(c.stopCleanup)
		c.stopCleanup = nil
	}
}

// cleanupRoutine periodically cleans up expired entries.
func (c *CacheUseLRUBatch) cleanupRoutine() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if c.ttl <= 0 {
				continue
			}

			// Try to acquire lock, skip this round if can't get it
			if !c.mu.TryLock() {
				continue
			}

			now := time.Now()
			for elem := c.root.Back(); elem != nil; {
				entry := elem.Value.(*Entry)
				if now.Sub(entry.UpdateAt) < c.ttl {
					break
				}
				nextElem := elem.Prev()
				delete(c.cache, entry.Key)
				c.root.Remove(elem)
				c.nbytes -= int64(len(entry.Key)) + int64(entry.Value.Len())
				if c.OnEvicted != nil {
					c.OnEvicted(entry.Key, entry.Value)
				}
				elem = nextElem
			}
			c.mu.Unlock()

		case <-c.stopCleanup:
			return
		}
	}
}

// Get retrieves a value from the cache.
func (c *CacheUseLRUBatch) Get(key string) (value Value, updateAt time.Time, ok bool) {
	c.mu.Lock() // 使用写锁，因为 MoveToFront 会修改链表结构
	defer c.mu.Unlock()

	if elem, hit := c.cache[key]; hit {
		entry := elem.Value.(*Entry)
		c.root.MoveToFront(elem)
		return entry.Value, entry.UpdateAt, true
	}
	return nil, time.Time{}, false
}

// Put adds or updates a value in the cache.
func (c *CacheUseLRUBatch) Put(key string, value Value) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.cache[key]; ok {
		c.root.MoveToFront(elem)
		entry := elem.Value.(*Entry)
		c.nbytes += int64(value.Len()) - int64(entry.Value.Len())
		entry.Value = value
		entry.UpdateAt = time.Now()
	} else {
		entry := &Entry{
			Key:      key,
			Value:    value,
			UpdateAt: time.Now(),
		}
		elem := c.root.PushFront(entry)
		c.cache[key] = elem
		c.nbytes += int64(len(key)) + int64(value.Len())
	}

	for c.maxBytes != 0 && c.nbytes > c.maxBytes {
		c.removeOldestBatch()
	}
}

// removeOldestBatch removes a batch of least recently used items.
func (c *CacheUseLRUBatch) removeOldestBatch() bool {
	if c.root.Len() == 0 {
		return false
	}

	removed := 0
	for removed < c.batchSize {
		elem := c.root.Back()
		if elem == nil {
			break
		}
		entry := elem.Value.(*Entry)
		delete(c.cache, entry.Key)
		c.root.Remove(elem)
		c.nbytes -= int64(len(entry.Key)) + int64(entry.Value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(entry.Key, entry.Value)
		}
		removed++
	}

	return removed > 0
}

// CleanUp removes all expired entries from the cache.
func (c *CacheUseLRUBatch) CleanUp(ttl time.Duration) {
	if ttl <= 0 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for elem := c.root.Back(); elem != nil; {
		entry := elem.Value.(*Entry)
		if now.Sub(entry.UpdateAt) < ttl {
			break
		}
		nextElem := elem.Prev()
		delete(c.cache, entry.Key)
		c.root.Remove(elem)
		c.nbytes -= int64(len(entry.Key)) + int64(entry.Value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(entry.Key, entry.Value)
		}
		elem = nextElem
	}
}

// Len returns the number of items in the cache.
func (c *CacheUseLRUBatch) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.root.Len()
}

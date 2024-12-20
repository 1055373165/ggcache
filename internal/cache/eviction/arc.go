// Package eviction provides cache eviction strategies.
package eviction

import (
	"container/list"
	"sync"
	"time"

	"github.com/1055373165/ggcache/internal/metrics"
	"github.com/1055373165/ggcache/pkg/common/logger"
)

// CacheUseARC implements the Adaptive Replacement Cache (ARC) algorithm.
// ARC maintains four lists:
// - T1: Contains pages that have been accessed exactly once recently (recency)
// - T2: Contains pages accessed at least twice recently (frequency)
// - B1: Ghost entries for pages that were in T1 (history for recency)
// - B2: Ghost entries for pages that were in T2 (history for frequency)
type CacheUseARC struct {
	mu sync.RWMutex

	// Cache capacity
	maxBytes int64
	nbytes   int64

	// Target size for T1 (p)
	p int64

	// Main lists for items present in cache
	t1 *list.List // Recent items
	t2 *list.List // Frequent items

	// Ghost lists for tracking history
	b1 *list.List // Ghost entries for T1
	b2 *list.List // Ghost entries for T2

	// Maps for O(1) lookup
	cache     map[string]*list.Element
	ghost     map[string]*list.Element
	OnEvicted func(key string, value Value)

	// TTL related fields
	cleanupInterval time.Duration
	ttl             time.Duration
	stopCleanup     chan struct{}
}

// NewCacheUseARC creates a new ARC cache.
func NewCacheUseARC(maxBytes int64, onEvicted func(string, Value)) *CacheUseARC {
	c := &CacheUseARC{
		maxBytes:        maxBytes,
		nbytes:          0,
		p:               0,
		t1:              list.New(),
		t2:              list.New(),
		b1:              list.New(),
		b2:              list.New(),
		cache:           make(map[string]*list.Element),
		ghost:           make(map[string]*list.Element),
		OnEvicted:       onEvicted,
		cleanupInterval: time.Minute,
		stopCleanup:     make(chan struct{}),
	}
	metrics.UpdateCacheSize(0)      // Initialize cache size to 0
	metrics.UpdateCacheItemCount(0) // Initialize item count to 0
	c.updateARCMetrics()            // Initialize ARC metrics

	logger.LogrusObj.Warnf("NewCacheUseARC: maxBytes=%d", maxBytes)

	go c.cleanupRoutine()
	return c
}

// arcEntry represents an entry in the ARC cache
type arcEntry struct {
	Entry
	inT2 bool // true if in T2, false if in T1
}

// ghostEntry represents an entry in the ghost lists
type ghostEntry struct {
	key  string
	inB2 bool // true if in B2, false if in B1
}

// Get retrieves a value from the cache
func (c *CacheUseARC) Get(key string) (value Value, updateAt time.Time, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ele, hit := c.cache[key]; hit {
		entry := ele.Value.(*arcEntry)
		if !entry.inT2 {
			// Move from T1 to T2
			c.t1.Remove(ele)
			entry.inT2 = true
			c.t2.PushFront(entry)
			c.cache[key] = c.t2.Front()
		} else {
			// Already in T2, move to front
			c.t2.MoveToFront(ele)
		}
		entry.Touch()
		return entry.Value, entry.UpdateAt, true
	}

	return nil, time.Time{}, false
}

// Put adds a value to the cache
func (c *CacheUseARC) Put(key string, value Value) {
	c.mu.Lock()
	defer c.mu.Unlock()

	newSize := int64(len(key)) + int64(value.Len())
	if newSize > c.maxBytes {
		return // Value too large
	}

	if ele, exists := c.cache[key]; exists {
		entry := ele.Value.(*arcEntry)
		oldSize := int64(len(key)) + int64(entry.Value.Len())
		c.nbytes = c.nbytes - oldSize + newSize

		entry.Value = value
		entry.Touch()

		// Move from T1 to T2 if needed
		if !entry.inT2 {
			c.t1.Remove(ele)
			entry.inT2 = true
			ele = c.t2.PushFront(entry)
			c.cache[key] = ele
		} else {
			c.t2.MoveToFront(ele)
		}
		metrics.UpdateCacheSize(c.nbytes)
		metrics.UpdateCacheItemCount(int64(len(c.cache)))
		c.updateARCMetrics()
		return
	}

	// Check ghost cache
	if ghost, exists := c.ghost[key]; exists {
		ghostEntry := ghost.Value.(*ghostEntry)
		if ghostEntry.inB2 {
			c.p = min(c.p+max(int64(c.b1.Len())/int64(c.b2.Len()), 1), c.maxBytes)
		} else {
			c.p = max(c.p-max(int64(c.b2.Len())/int64(c.b1.Len()), 1), 0)
		}
		delete(c.ghost, key)
		if ghostEntry.inB2 {
			c.b2.Remove(ghost)
		} else {
			c.b1.Remove(ghost)
		}
	}

	// Make space if needed
	for c.nbytes+newSize > c.maxBytes {
		c.evict()
		metrics.RecordEviction()
	}

	// Add new entry to T1
	entry := &arcEntry{
		Entry: Entry{
			Key:      key,
			Value:    value,
			UpdateAt: time.Now(),
		},
		inT2: false,
	}
	ele := c.t1.PushFront(entry)
	c.cache[key] = ele
	c.nbytes += newSize
	metrics.UpdateCacheSize(c.nbytes)
	metrics.UpdateCacheItemCount(int64(len(c.cache)))
	c.updateARCMetrics()
}

// evict removes one entry based on the ARC algorithm
func (c *CacheUseARC) evict() {
	// T2 has exceeded the space it should occupy.
	if int64(c.t1.Len()) > 0 && (int64(c.t2.Len()) > c.p || (c.t2.Len() == 0 && c.b2.Len() == 0)) {
		// Evict from T1
		ele := c.t1.Back()
		entry := ele.Value.(*arcEntry)
		c.removeEntry(ele, entry, true)
	} else if c.t2.Len() > 0 {
		// Evict from T2
		ele := c.t2.Back()
		entry := ele.Value.(*arcEntry)
		c.removeEntry(ele, entry, false)
	}
}

// removeEntry handles the removal of an entry from the cache
func (c *CacheUseARC) removeEntry(ele *list.Element, entry *arcEntry, fromT1 bool) {
	c.nbytes -= int64(len(entry.Key)) + int64(entry.Value.Len())
	delete(c.cache, entry.Key)
	metrics.UpdateCacheSize(c.nbytes)
	metrics.UpdateCacheItemCount(int64(len(c.cache)))

	if fromT1 {
		c.t1.Remove(ele)
		// Add to B1
		ghost := &ghostEntry{key: entry.Key, inB2: false}
		ghostEle := c.b1.PushFront(ghost)
		c.ghost[entry.Key] = ghostEle
	} else {
		c.t2.Remove(ele)
		// Add to B2
		ghost := &ghostEntry{key: entry.Key, inB2: true}
		ghostEle := c.b2.PushFront(ghost)
		c.ghost[entry.Key] = ghostEle
	}

	// Maintain ghost list sizes
	for int64(c.b1.Len()) > c.maxBytes {
		ghost := c.b1.Back()
		delete(c.ghost, ghost.Value.(*ghostEntry).key)
		c.b1.Remove(ghost)
	}
	for int64(c.b2.Len()) > c.maxBytes {
		ghost := c.b2.Back()
		delete(c.ghost, ghost.Value.(*ghostEntry).key)
		c.b2.Remove(ghost)
	}

	if c.OnEvicted != nil {
		c.OnEvicted(entry.Key, entry.Value)
	}

	c.updateARCMetrics()
}

// updateARCMetrics updates ARC-specific metrics
func (c *CacheUseARC) updateARCMetrics() {
	metrics.UpdateARCMetrics(c.t1.Len(), c.t2.Len(), c.b1.Len(), c.b2.Len(), int(c.p))
}

// cleanupRoutine periodically removes expired entries
func (c *CacheUseARC) cleanupRoutine() {
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

// CleanUp removes expired entries from the cache
func (c *CacheUseARC) CleanUp(ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Clean T1
	for e := c.t1.Front(); e != nil; {
		next := e.Next()
		if e.Value.(*arcEntry).Expired(ttl) {
			c.removeEntry(e, e.Value.(*arcEntry), true)
		}
		e = next
	}

	// Clean T2
	for e := c.t2.Front(); e != nil; {
		next := e.Next()
		if e.Value.(*arcEntry).Expired(ttl) {
			c.removeEntry(e, e.Value.(*arcEntry), false)
		}
		e = next
	}
}

// SetTTL sets the time-to-live for cache entries
func (c *CacheUseARC) SetTTL(ttl time.Duration) {
	c.ttl = ttl
}

// SetCleanupInterval sets the interval between cleanup runs
func (c *CacheUseARC) SetCleanupInterval(interval time.Duration) {
	c.cleanupInterval = interval
}

// Stop stops the cleanup routine
func (c *CacheUseARC) Stop() {
	close(c.stopCleanup)
}

// Len returns the number of items in the cache
func (c *CacheUseARC) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.t1.Len() + c.t2.Len()
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

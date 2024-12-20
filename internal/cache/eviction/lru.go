package eviction

import (
	"container/list"
	"hash/fnv"
	"sync"
	"time"
)

const (
	defaultCleanupInterval = 2 * time.Minute  // Default interval for cleanup routine
	defaultTTL             = 10 * time.Minute // Default TTL for cache entries
	defaultNumSegments     = 16               // Default number of segments for the cache
)

// segment represents a portion of the cache with its own lock
type segment struct {
	mu        sync.RWMutex
	maxBytes  int64
	nbytes    int64
	ll        *list.List
	cache     map[string]*list.Element
	OnEvicted func(key string, value Value)
}

// CacheUseLRU implements a segmented Least Recently Used (LRU) cache.
// It maintains multiple segments, each with its own lock, to reduce lock contention.
type CacheUseLRU struct {
	segments        []*segment
	numSegments     int
	cleanupInterval time.Duration
	ttl             time.Duration
	stopCleanup     chan struct{}
	mu              sync.RWMutex
}

// NewCacheUseLRU creates a new segmented LRU cache with the specified maximum size and eviction callback.
func NewCacheUseLRU(maxBytes int64, onEvicted func(string, Value)) *CacheUseLRU {
	c := &CacheUseLRU{
		segments:        make([]*segment, defaultNumSegments),
		numSegments:     defaultNumSegments,
		cleanupInterval: defaultCleanupInterval,
		ttl:             defaultTTL,
		stopCleanup:     make(chan struct{}),
	}

	// Initialize segments
	segmentMaxBytes := maxBytes / int64(defaultNumSegments)
	for i := 0; i < defaultNumSegments; i++ {
		c.segments[i] = &segment{
			maxBytes:  segmentMaxBytes,
			ll:        list.New(),
			cache:     make(map[string]*list.Element),
			OnEvicted: onEvicted,
		}
	}

	// Start cleanup routine
	go c.cleanupRoutine()

	return c
}

// getSegment returns the appropriate segment for a given key.
// Use the FNV hash algorithm to determine which segment the key belongs to.
func (c *CacheUseLRU) getSegment(key string) *segment {
	h := fnv.New32a()
	h.Write([]byte(key))
	return c.segments[h.Sum32()%uint32(c.numSegments)]
}

// SetTTL sets the time-to-live for cache entries.
func (c *CacheUseLRU) SetTTL(ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ttl = ttl
}

// SetCleanupInterval sets the interval between cleanup runs.
func (c *CacheUseLRU) SetCleanupInterval(interval time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// Stop existing cleanup routine
	close(c.stopCleanup)

	// Create new stop channel and set new interval
	c.stopCleanup = make(chan struct{})
	c.cleanupInterval = interval

	// Start new cleanup routine
	go c.cleanupRoutine()
}

// Stop stops the cleanup routine.
func (c *CacheUseLRU) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()
	close(c.stopCleanup)
}

// cleanupRoutine periodically cleans up expired entries across all segments.
func (c *CacheUseLRU) cleanupRoutine() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			currentTTL := c.ttl // capture current TTL value
			c.CleanUp(currentTTL)
		case <-c.stopCleanup:
			return
		}
	}
}

// cleanupSegment removes expired entries from a single segment
func (c *CacheUseLRU) cleanupSegment(seg *segment, ttl time.Duration) {
	seg.mu.Lock()
	defer seg.mu.Unlock()

	var next *list.Element
	for e := seg.ll.Front(); e != nil; e = next {
		next = e.Next()
		if e.Value == nil {
			continue
		}
		if e.Value.(*Entry).Expired(ttl) {
			seg.removeElement(e)
		}
	}
}

// Get retrieves a value from the cache.
func (c *CacheUseLRU) Get(key string) (value Value, updateAt time.Time, ok bool) {
	seg := c.getSegment(key)
	seg.mu.RLock()
	defer seg.mu.RUnlock()

	if ele, ok := seg.cache[key]; ok {
		seg.ll.MoveToBack(ele)
		e := ele.Value.(*Entry)
		e.Touch()
		return e.Value, e.UpdateAt, true
	}
	return nil, time.Time{}, false
}

// Put adds or updates a value in the cache.
func (c *CacheUseLRU) Put(key string, value Value) {
	seg := c.getSegment(key)
	seg.mu.Lock()
	defer seg.mu.Unlock()

	if ele, ok := seg.cache[key]; ok {
		seg.ll.MoveToBack(ele)
		entry := ele.Value.(*Entry)
		seg.nbytes += int64(value.Len()) - int64(entry.Value.Len())
		entry.Value = value
		entry.Touch()
	} else {
		entry := &Entry{
			Key:      key,
			Value:    value,
			UpdateAt: time.Now(),
		}
		ele := seg.ll.PushBack(entry)
		seg.cache[key] = ele
		seg.nbytes += int64(len(key)) + int64(value.Len())
	}

	for seg.maxBytes != 0 && seg.maxBytes < seg.nbytes {
		seg.removeOldest()
	}
}

func (c *CacheUseLRU) CleanUp(ttl time.Duration) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, seg := range c.segments {
		c.cleanupSegment(seg, ttl)
	}
}

// removeOldest removes the least recently used item from a segment.
func (seg *segment) removeOldest() {
	if ele := seg.ll.Front(); ele != nil {
		seg.removeElement(ele)
	}
}

// removeElement removes an element from a segment.
func (seg *segment) removeElement(e *list.Element) {
	seg.ll.Remove(e)
	entry := e.Value.(*Entry)
	delete(seg.cache, entry.Key)
	seg.nbytes -= int64(len(entry.Key)) + int64(entry.Value.Len())
	if seg.OnEvicted != nil {
		seg.OnEvicted(entry.Key, entry.Value)
	}
}

// Len returns the total number of items in the cache.
func (c *CacheUseLRU) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	total := 0
	for _, seg := range c.segments {
		seg.mu.RLock()
		total += seg.ll.Len()
		seg.mu.RUnlock()
	}
	return total
}

package eviction

import (
	"container/heap"
	"time"
)

// CacheUseLFU implements a Least Frequently Used (LFU) cache.
// It maintains both a hash table for O(1) lookups and a priority queue
// for efficient removal of least frequently used items.
type CacheUseLFU struct {
	nbytes    int64                         // Current size in bytes
	maxBytes  int64                         // Maximum allowed size in bytes (0 means unlimited)
	cache     map[string]*lfuEntry          // Hash table for O(1) lookups
	pq        *priorityQueue                // Priority queue for LFU ordering
	OnEvicted func(key string, value Value) // Optional callback when an entry is evicted
}

// NewCacheUseLFU creates a new LFU cache with the specified maximum size and eviction callback.
func NewCacheUseLFU(maxBytes int64, onEvicted func(string, Value)) *CacheUseLFU {
	queue := priorityQueue(make([]*lfuEntry, 0))
	return &CacheUseLFU{
		maxBytes:  maxBytes,
		pq:        &queue,
		cache:     make(map[string]*lfuEntry),
		OnEvicted: onEvicted,
	}
}

// Get retrieves a value from the cache.
// It returns the value, its last update time, and whether the key was found.
// If the key exists, its access count is incremented.
func (p *CacheUseLFU) Get(key string) (Value, time.Time, bool) {
	if e, ok := p.cache[key]; ok {
		e.referenced()
		heap.Fix(p.pq, e.index)
		return e.entry.Value, e.entry.UpdateAt, ok
	}
	return nil, time.Time{}, false
}

// Put adds or updates a value in the cache.
// If the key already exists, its value is updated and access count is incremented.
// If the key is new, it is added to both the hash table and priority queue.
// If adding the new entry would exceed maxBytes, least frequently used entries
// are removed until the cache size is within bounds.
func (p *CacheUseLFU) Put(key string, value Value) {
	if e, ok := p.cache[key]; ok {
		p.nbytes += int64(value.Len()) - int64(e.entry.Value.Len())
		e.entry.Value = value
		e.referenced()
		heap.Fix(p.pq, e.index)
		return
	}

	// Add new entry
	e := &lfuEntry{
		entry: Entry{
			Key:      key,
			Value:    value,
			UpdateAt: time.Now(),
		},
	}
	e.referenced()
	heap.Push(p.pq, e)
	p.cache[key] = e
	p.nbytes += int64(len(e.entry.Key)) + int64(e.entry.Value.Len())

	// Remove least frequently used entries if cache exceeds size limit
	for p.maxBytes != 0 && p.maxBytes < p.nbytes {
		p.Remove()
	}
}

// CleanUp removes all expired entries from the cache.
// An entry is considered expired if its last update time plus the TTL
// is before the current time.
func (p *CacheUseLFU) CleanUp(ttl time.Duration) {
	if p.pq == nil {
		return
	}

	// Create a list of indices to remove
	var toRemove []int
	for i, e := range *p.pq {
		if e.entry.Expired(ttl) {
			toRemove = append(toRemove, i)
		}
	}

	// Remove entries in reverse order to maintain heap indices
	for i := len(toRemove) - 1; i >= 0; i-- {
		idx := toRemove[i]
		if idx < len(*p.pq) { // Safety check
			expiredEntry := heap.Remove(p.pq, idx).(*lfuEntry)
			delete(p.cache, expiredEntry.entry.Key)
			p.nbytes -= int64(len(expiredEntry.entry.Key)) + int64(expiredEntry.entry.Value.Len())
			if p.OnEvicted != nil {
				p.OnEvicted(expiredEntry.entry.Key, expiredEntry.entry.Value)
			}
		}
	}
}

// Len returns the number of items in the cache.
func (p *CacheUseLFU) Len() int {
	return p.pq.Len()
}

// Remove removes and returns the least frequently used item from the cache.
// If there are multiple items with the same frequency, the least recently used one is removed.
func (p *CacheUseLFU) Remove() {
	e := heap.Pop(p.pq).(*lfuEntry)
	delete(p.cache, e.entry.Key)
	p.nbytes -= int64(len(e.entry.Key)) + int64(e.entry.Value.Len())
	if p.OnEvicted != nil {
		p.OnEvicted(e.entry.Key, e.entry.Value)
	}
}

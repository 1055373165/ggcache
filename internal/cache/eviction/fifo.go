package eviction

import (
	"container/list"
	"sync"
	"time"
)

// CacheUseFIFO implements a First-In-First-Out (FIFO) cache.
// It maintains both a hash table for O(1) lookups and a doubly linked list
// for efficient removal of the oldest items.
type CacheUseFIFO struct {
	maxBytes  int64                         // Maximum allowed size in bytes (0 means unlimited)
	nbytes    int64                         // Current size in bytes
	ll        *list.List                    // Doubly linked list for FIFO ordering
	cache     map[string]*list.Element      // Hash table for O(1) lookups
	mu        sync.RWMutex                  // Protects shared resources
	OnEvicted func(key string, value Value) // Optional callback when an entry is evicted
}

// NewCacheUseFIFO creates a new FIFO cache with the specified maximum size and eviction callback.
func NewCacheUseFIFO(maxBytes int64, onEvicted func(string, Value)) *CacheUseFIFO {
	return &CacheUseFIFO{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Get retrieves a value from the cache.
// It returns the value, its last update time, and whether the key was found.
// Unlike LRU, accessing an item does not affect its position in the eviction order.
func (cuf *CacheUseFIFO) Get(key string) (Value, time.Time, bool) {
	cuf.mu.RLock()
	defer cuf.mu.RUnlock()

	if ele, ok := cuf.cache[key]; ok {
		e := ele.Value.(*Entry)
		return e.Value, e.UpdateAt, true
	}
	return nil, time.Time{}, false
}

// Put adds or updates a value in the cache.
// If the key already exists, its value is updated but its position remains unchanged.
// If the key is new, it is added to the back of the list.
// If adding the new entry would exceed maxBytes, oldest entries are removed
// until the cache size is within bounds.
func (cuf *CacheUseFIFO) Put(key string, value Value) {
	cuf.mu.Lock()
	defer cuf.mu.Unlock()

	if ele, ok := cuf.cache[key]; ok {
		entry := ele.Value.(*Entry)
		cuf.nbytes += int64(value.Len()) - int64(entry.Value.Len())
		entry.Value = value
		entry.Touch()
		return
	}

	// Add new entry
	newEntry := &Entry{
		Key:      key,
		Value:    value,
		UpdateAt: time.Now(),
	}
	ele := cuf.ll.PushBack(newEntry)
	cuf.cache[key] = ele
	cuf.nbytes += int64(len(newEntry.Key)) + int64(newEntry.Value.Len())

	// Remove oldest entries if cache exceeds size limit
	for cuf.maxBytes != 0 && cuf.maxBytes < cuf.nbytes {
		cuf.removeFront()
	}
}

// removeFront removes the oldest item from the cache.
// Caller must hold the lock.
func (cuf *CacheUseFIFO) removeFront() {
	if ele := cuf.ll.Front(); ele != nil {
		cuf.removeElement(ele)
	}
}

// RemoveFront removes the oldest item from the cache.
// This method is thread-safe and can be called externally.
func (cuf *CacheUseFIFO) RemoveFront() {
	cuf.mu.Lock()
	defer cuf.mu.Unlock()
	cuf.removeFront()
}

// CleanUp removes all expired entries from the cache.
// An entry is considered expired if its last update time plus the TTL
// is before the current time.
func (cuf *CacheUseFIFO) CleanUp(ttl time.Duration) {
	cuf.mu.Lock()
	defer cuf.mu.Unlock()

	var next *list.Element
	for e := cuf.ll.Front(); e != nil; e = next {
		next = e.Next() // Save next pointer before potential removal
		if e.Value.(*Entry).Expired(ttl) {
			cuf.removeElement(e)
		}
	}
}

// Len returns the number of items in the cache.
func (cuf *CacheUseFIFO) Len() int {
	cuf.mu.RLock()
	defer cuf.mu.RUnlock()
	return cuf.ll.Len()
}

// removeElement removes an element from the cache, updating the size
// and calling the eviction callback if set.
// Caller must hold the lock.
func (cuf *CacheUseFIFO) removeElement(e *list.Element) {
	entry := cuf.ll.Remove(e).(*Entry)
	delete(cuf.cache, entry.Key)
	cuf.nbytes -= int64(len(entry.Key)) + int64(entry.Value.Len())
	if cuf.OnEvicted != nil {
		cuf.OnEvicted(entry.Key, entry.Value)
	}
}

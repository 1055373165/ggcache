package eviction

import (
	"container/heap"
	"time"
)

type lfuCache struct {
	nbytes    int64
	maxBytes  int64
	cache     map[string]*lfuEntry
	pq        *priorityqueue
	OnEvicted func(key string, value Value)
}

// implement interface.Value interface
func (p *lfuCache) Len() int {
	return p.pq.Len()
}

func NewLFUCache(maxBytes int64, onEvicted func(string, Value)) *lfuCache {
	queue := priorityqueue(make([]*lfuEntry, 0))
	return &lfuCache{
		maxBytes:  maxBytes,
		pq:        &queue,
		cache:     make(map[string]*lfuEntry),
		OnEvicted: onEvicted,
	}
}

func (p *lfuCache) Get(key string) (value Value, updateAt *time.Time, ok bool) {
	if e, ok := p.cache[key]; ok {
		e.referenced()
		heap.Fix(p.pq, e.index)
		return e.entry.Value, e.entry.UpdateAt, ok
	}
	return
}

func (p *lfuCache) Put(key string, value Value) {
	if e, ok := p.cache[key]; ok {
		// update memory usage (if old value len > new value len, nbytes increse)
		p.nbytes += int64(value.Len()) - int64(e.entry.Value.Len())
		e.entry.Value = value
		e.referenced()
		heap.Fix(p.pq, e.index)
	} else {
		e := &lfuEntry{0, Entry{Key: key, Value: value, UpdateAt: nil}, 0}
		e.referenced()
		heap.Push(p.pq, e)
		p.cache[key] = e
		// update memory usage
		p.nbytes += int64(len(e.entry.Key)) + int64(e.entry.Value.Len())
	}
	// cache purge trigger
	for p.maxBytes != 0 && p.maxBytes < p.nbytes {
		p.Remove()
	}
}

// clear expired cache
func (p *lfuCache) CleanUp(ttl time.Duration) {
	for _, e := range *p.pq {
		if e.entry.Expired(ttl) {
			kv := heap.Remove(p.pq, e.index).(*lfuEntry).entry
			delete(p.cache, kv.Key)
			p.nbytes -= int64(len(kv.Key)) + int64(kv.Value.Len())
			if p.OnEvicted != nil {
				p.OnEvicted(kv.Key, kv.Value)
			}
		}
	}
}

// cache purge operation
func (p *lfuCache) Remove() {
	e := heap.Pop(p.pq).(*lfuEntry)
	delete(p.cache, e.entry.Key)
	p.nbytes -= int64(len(e.entry.Key)) + int64(e.entry.Value.Len())
	if p.OnEvicted != nil {
		p.OnEvicted(e.entry.Key, e.entry.Value)
	}
}

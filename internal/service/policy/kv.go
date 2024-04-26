package policy

import (
	"container/list"
	"time"
)

type entry struct {
	key      string
	value    Value
	updateAt *time.Time
}

// ttl
func (ele *entry) expired(duration time.Duration) (ok bool) {
	if ele.updateAt == nil {
		ok = false
	} else {
		ok = ele.updateAt.Add(duration).Before(time.Now())
	}
	return
}

// ttl
func (ele *entry) touch() {
	//ele.updateAt=time.Now()
	nowTime := time.Now()
	ele.updateAt = &nowTime
}

func New(name string, maxBytes int64, onEvicted func(string, Value)) Interface {

	if name == "fifo" {
		return newFifoCache(maxBytes, onEvicted)
	}
	if name == "lru" {
		return newLruCache(maxBytes, onEvicted)
	}
	if name == "lfu" {
		return newLfuCache(maxBytes, onEvicted)
	}

	return nil
}

func newLruCache(maxBytes int64, onEvicted func(string, Value)) *LRUCache {

	return &LRUCache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

func newFifoCache(maxBytes int64, onEvicted func(string, Value)) *fifoCahce {

	return &fifoCahce{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

func newLfuCache(maxBytes int64, onEvicted func(string, Value)) *lfuCache {
	queue := priorityqueue(make([]*lfuEntry, 0))
	return &lfuCache{
		maxBytes:  maxBytes,
		pq:        &queue,
		cache:     make(map[string]*lfuEntry),
		OnEvicted: onEvicted,
	}
}

type Interface interface {
	Get(string) (Value, *time.Time, bool)
	Add(string, Value)
	CleanUp(ttl time.Duration)
	Len() int
}

package policy

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

func (p *lfuCache) Get(key string) (value Value, updateAt *time.Time, ok bool) {
	if e, ok := p.cache[key]; ok {
		e.referenced()
		heap.Fix(p.pq, e.index)
		return e.entry.value, e.entry.updateAt, ok
	}
	return
}

func (p *lfuCache) Add(key string, value Value) {
	if e, ok := p.cache[key]; ok {
		//更新value
		p.nbytes += int64(value.Len()) - int64(e.entry.value.Len())
		e.entry.value = value
		e.referenced()
		heap.Fix(p.pq, e.index)
	} else {
		e := &lfuEntry{0, entry{key, value, nil}, 0}
		e.referenced()
		heap.Push(p.pq, e)
		p.cache[key] = e
		p.nbytes += int64(len(e.entry.key)) + int64(e.entry.value.Len())
	}

	for p.maxBytes != 0 && p.maxBytes < p.nbytes {
		p.Remove()
	}
}

func (p *lfuCache) CleanUp(ttl time.Duration) {
	for _, e := range *p.pq {
		if e.entry.expired(ttl) {
			kv := heap.Remove(p.pq, e.index).(*lfuEntry).entry
			delete(p.cache, kv.key)
			p.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
			if p.OnEvicted != nil {
				p.OnEvicted(kv.key, kv.value)
			}
		}
	}
}

func (p *lfuCache) Remove() {
	e := heap.Pop(p.pq).(*lfuEntry)
	delete(p.cache, e.entry.key)
	p.nbytes -= int64(len(e.entry.key)) + int64(e.entry.value.Len())
	if p.OnEvicted != nil {
		p.OnEvicted(e.entry.key, e.entry.value)
	}
}

func (p *lfuCache) Len() int {
	return p.pq.Len()
}

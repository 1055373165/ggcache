package fifo

import (
	"container/list"
	"time"
)

type entry struct {
	key      string
	value    Value
	updateAt *time.Time
}

type Value interface {
	Len() int
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
	return newFifoCache(maxBytes, onEvicted)
}

func newFifoCache(maxBytes int64, onEvicted func(string, Value)) *fifoCahce {

	return &fifoCahce{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

type Interface interface {
	Get(string) (Value, *time.Time, bool)
	Add(string, Value)
	CleanUp(ttl time.Duration)
	Len() int
}

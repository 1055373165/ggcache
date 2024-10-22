package eviction

import (
	"container/list"
	"time"
)

type fifoCahce struct {
	maxBytes int64
	nbytes   int64
	ll       *list.List
	cache    map[string]*list.Element
	// optional and executed when an entry is purged.
	// 回调函数，采用依赖注入的方式，该函数用于处理从缓存中淘汰的数据
	OnEvicted func(key string, value Value)
}

func NewFIFOCache(maxBytes int64, onEvicted func(string, Value)) *fifoCahce {
	return &fifoCahce{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

func (f *fifoCahce) Get(key string) (value Value, updateAt *time.Time, ok bool) {
	if ele, ok := f.cache[key]; ok {
		e := ele.Value.(*Entry)
		return e.Value, e.UpdateAt, true
	}
	return

}

func (f *fifoCahce) Put(key string, value Value) {
	if ele, ok := f.cache[key]; ok {
		//更新缓存
		f.nbytes += int64(value.Len()) - int64(ele.Value.(*Entry).Value.Len())
		ele.Value.(*Entry).Value = value
	} else {
		kv := &Entry{Key: key, Value: value, UpdateAt: nil}
		kv.Touch()
		ele := f.ll.PushBack(kv)
		f.cache[key] = ele
		f.nbytes += int64(len(kv.Key)) + int64(kv.Value.Len())
	}

	for f.maxBytes != 0 && f.maxBytes < f.nbytes {
		f.RemoveFront()
	}
}

func (f *fifoCahce) RemoveFront() {
	ele := f.ll.Front()
	if ele != nil {
		kv := f.ll.Remove(ele).(*Entry)
		delete(f.cache, kv.Key)
		f.nbytes -= int64(len(kv.Key)) + int64(kv.Value.Len())
		if f.OnEvicted != nil {
			f.OnEvicted(kv.Key, kv.Value)
		}
	}
}

func (f *fifoCahce) CleanUp(ttl time.Duration) {
	for e := f.ll.Front(); e != nil; e = e.Next() {
		if e.Value.(*Entry).Expired(ttl) {
			kv := f.ll.Remove(e).(*Entry)
			delete(f.cache, kv.Key)
			f.nbytes -= int64(len(kv.Key)) + int64(kv.Value.Len())
			if f.OnEvicted != nil {
				f.OnEvicted(kv.Key, kv.Value)
			}
		} else {
			break
		}
	}
}

func (f *fifoCahce) Len() int {
	return f.ll.Len()
}

package policy

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

func (f *fifoCahce) Get(key string) (value Value, updateAt *time.Time, ok bool) {
	if ele, ok := f.cache[key]; ok {
		e := ele.Value.(*entry)
		return e.value, e.updateAt, true
	}
	return

}

func (f *fifoCahce) Add(key string, value Value) {
	if ele, ok := f.cache[key]; ok {
		//更新缓存
		f.nbytes += int64(value.Len()) - int64(ele.Value.(*entry).value.Len())
		ele.Value.(*entry).value = value
	} else {
		kv := &entry{key, value, nil}
		kv.touch()
		ele := f.ll.PushBack(kv)
		f.cache[key] = ele
		f.nbytes += int64(len(kv.key)) + int64(kv.value.Len())
	}
	for f.maxBytes != 0 && f.maxBytes < f.nbytes {
		f.RemoveFront()
	}
}

func (f *fifoCahce) RemoveFront() {
	ele := f.ll.Front()
	if ele != nil {
		kv := f.ll.Remove(ele).(*entry)
		delete(f.cache, kv.key)
		f.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if f.OnEvicted != nil {
			f.OnEvicted(kv.key, kv.value)
		}
	}
}

func (f *fifoCahce) CleanUp(ttl time.Duration) {
	for e := f.ll.Front(); e != nil; e = e.Next() {
		if e.Value.(*entry).expired(ttl) {
			kv := f.ll.Remove(e).(*entry)
			delete(f.cache, kv.key)
			f.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
			if f.OnEvicted != nil {
				f.OnEvicted(kv.key, kv.value)
			}
		} else {
			break
		}
	}
}

func (f *fifoCahce) Len() int {
	return f.ll.Len()
}

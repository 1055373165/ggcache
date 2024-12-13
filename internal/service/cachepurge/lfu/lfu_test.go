package lfu

import (
	"testing"
)

type String string

func (d String) Len() int {
	return len(d)
}
func TestPriorityQueue_Get(t *testing.T) {
	lfu := NewLFUCache(10, nil)

	lfu.Put("key1", String("1234"))
	if v, _, ok := lfu.Get("key1"); !ok || string(v.(String)) != "1234" {
		t.Fatalf("cache hit key1=1234 failed")
	}

	if _, _, ok := lfu.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
}

func TestPriorityQueue_Remove(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "k3"
	v1, v2, v3 := "value1", "value2", "v3"
	curCap := len(k1 + k2 + v1 + v2)

	lfu := NewLFUCache(int64(curCap), nil)
	lfu.Put(k1, String(v1))
	lfu.Put(k1, String(v1))
	lfu.Put(k2, String(v2))
	lfu.Put(k3, String(v3))

	if _, _, ok := lfu.Get("key2"); ok || lfu.Len() != 2 {
		t.Fatalf("Removeoldest key1 failed")
	}
}

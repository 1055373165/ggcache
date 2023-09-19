package lfu

import (
	"fmt"
	"testing"
)

type String string

func (s String) Len() int {
	return len(s)
}
func TestLfu(t *testing.T) {
	lfucache := NewLfuCache(10, nil)
	// 凑成 9，无法再容纳其他的条目（计数 1）
	lfucache.Put("123", String("12345"))
	lfucache.Get("123")
	fmt.Println("nbytes", lfucache.nBytes)
	expect1 := 1
	if lfucache.Len() != expect1 {
		t.Fatalf("expected len %d, got %d", expect1, lfucache.Len())
	}

	// 超出容量，被删除
	lfucache.Put("234", String("1234567"))
	fmt.Println("nbytes", lfucache.nBytes)
	if _, _, ok := lfucache.Get("234"); ok {
		t.Fatal("lfu unused policy failed, should be delete key=234")
	}

	lfucache.Put("1", String("1"))
	fmt.Println("nbytes", lfucache.nBytes)
	lfucache.Get("1")
	lfucache.Get("1")
	// key = 1 count = 3

	if lfucache.Len() != 2 {
		t.Fatal("lfu policy is not valid")
	}

	// 超出容量，2 被删除
	lfucache.Put("2", String("1234"))
	expect := []bool{true, false}
	_, _, ok1 := lfucache.Get("1")
	_, _, ok2 := lfucache.Get("2")
	if ok1 != expect[0] || ok2 != expect[1] {
		t.Fatal("should delete key=2 and reserve key=1")
	}
	fmt.Println("1 频次", lfucache.cache["1"].count)
	lfucache.Get("123")
	lfucache.Get("123")
	lfucache.Get("123")
	lfucache.Get("123")
	fmt.Println("123 频次", lfucache.cache["123"].count)
	// key123.count=4

	lfucache.Put("123", String("123456789"))

	// 123 访问频次为 7,1 访问频次为 4，所以应该会先删除 key=1，但是此时仍然超出容量限制，因此 key=123 也会被删除
	if lfucache.Len() != 0 {
		t.Fatalf("lru cache policy is not valid")
	}
}

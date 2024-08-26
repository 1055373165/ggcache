package lfu

import (
	"container/heap"
	"fmt"
	"testing"

	"github.com/1055373165/ggcache/internal/service/eviction/strategy"
)

func Test_priorityqueue_Pop(t *testing.T) {
	pq := priorityqueue([]*lfuEntry{})

	//heap.Init(&pq)
	for i := 0; i < 10; i++ {
		heap.Push(&pq, &lfuEntry{0, strategy.Entry{}, i})
	}

	for pq.Len() != 0 {
		e := heap.Pop(&pq).(*lfuEntry)
		fmt.Println(e.count)
	}
}

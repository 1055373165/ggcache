package policy

import (
	"container/heap"
	"fmt"
	"testing"
)

func Test_priorityqueue_Pop(t *testing.T) {
	pq := priorityqueue([]*lfuEntry{})
	//heap.Init(&pq)
	for i := 0; i < 10; i++ {
		heap.Push(&pq, &lfuEntry{0, entry{}, i})
	}
	for pq.Len() != 0 {
		e := heap.Pop(&pq).(*lfuEntry)
		fmt.Println(e.count)
	}

}

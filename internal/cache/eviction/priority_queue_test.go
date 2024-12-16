package eviction

import (
	"container/heap"
	"testing"
	"time"
)

func TestPriorityQueue_Basic(t *testing.T) {
	t.Run("empty queue operations", func(t *testing.T) {
		pq := &priorityQueue{}
		heap.Init(pq)

		if pq.Len() != 0 {
			t.Errorf("Expected empty queue, got len = %d", pq.Len())
		}
	})

	t.Run("push and pop", func(t *testing.T) {
		pq := &priorityQueue{}
		heap.Init(pq)

		// Push entries with different counts
		entries := []*lfuEntry{
			{count: 3, entry: Entry{Key: "k3"}},
			{count: 1, entry: Entry{Key: "k1"}},
			{count: 2, entry: Entry{Key: "k2"}},
		}

		for _, e := range entries {
			heap.Push(pq, e)
		}

		// Verify min-heap property
		if pq.Len() != 3 {
			t.Errorf("Expected queue length 3, got %d", pq.Len())
		}

		// Pop should return entries in ascending order of count
		expected := []int{1, 2, 3}
		for i, want := range expected {
			got := heap.Pop(pq).(*lfuEntry)
			if got.count != want {
				t.Errorf("Pop %d: got count %d, want %d", i, got.count, want)
			}
		}
	})
}

func TestPriorityQueue_Ordering(t *testing.T) {
	tests := []struct {
		name     string
		entries  []*lfuEntry
		wantKeys []string
	}{
		{
			name: "order by count",
			entries: []*lfuEntry{
				{count: 3, entry: Entry{Key: "k3"}},
				{count: 1, entry: Entry{Key: "k1"}},
				{count: 2, entry: Entry{Key: "k2"}},
			},
			wantKeys: []string{"k1", "k2", "k3"},
		},
		{
			name: "equal counts ordered by time",
			entries: []*lfuEntry{
				{
					count: 1,
					entry: Entry{Key: "k1", UpdateAt: time.Now().Add(2 * time.Second)},
				},
				{
					count: 1,
					entry: Entry{Key: "k2", UpdateAt: time.Now().Add(1 * time.Second)},
				},
				{
					count: 1,
					entry: Entry{Key: "k3", UpdateAt: time.Now()},
				},
			},
			wantKeys: []string{"k3", "k2", "k1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pq := &priorityQueue{}
			heap.Init(pq)

			// Push all entries
			for _, e := range tt.entries {
				heap.Push(pq, e)
			}

			// Pop and verify order
			for i, wantKey := range tt.wantKeys {
				got := heap.Pop(pq).(*lfuEntry)
				if got.entry.Key != wantKey {
					t.Errorf("Pop %d: got key %s, want %s", i, got.entry.Key, wantKey)
				}
			}
		})
	}
}

func TestPriorityQueue_Referenced(t *testing.T) {
	entry := &lfuEntry{
		count: 1,
		entry: Entry{Key: "test"},
	}

	// Record initial state
	initialCount := entry.count
	initialTime := entry.entry.UpdateAt

	// Wait a bit to ensure time difference
	time.Sleep(time.Millisecond)

	// Call referenced
	entry.referenced()

	// Verify count increased
	if entry.count != initialCount+1 {
		t.Errorf("Count: got %d, want %d", entry.count, initialCount+1)
	}

	// Verify timestamp updated
	if !entry.entry.UpdateAt.After(initialTime) {
		t.Error("UpdateAt time was not updated")
	}
}

func TestPriorityQueue_IndexManagement(t *testing.T) {
	pq := &priorityQueue{}
	heap.Init(pq)

	// Push some entries
	entries := []*lfuEntry{
		{count: 3, entry: Entry{Key: "k3"}},
		{count: 1, entry: Entry{Key: "k1"}},
		{count: 2, entry: Entry{Key: "k2"}},
	}

	for _, e := range entries {
		heap.Push(pq, e)
	}

	// Verify indices are maintained
	for i := 0; i < pq.Len(); i++ {
		if (*pq)[i].index != i {
			t.Errorf("Index mismatch at position %d: got %d", i, (*pq)[i].index)
		}
	}

	// Pop an entry and verify its index is set to -1
	popped := heap.Pop(pq).(*lfuEntry)
	if popped.index != -1 {
		t.Errorf("Popped entry index: got %d, want -1", popped.index)
	}
}

func TestPriorityQueue_HeapInterface(t *testing.T) {
	// This test verifies that our priorityQueue correctly implements heap.Interface
	var _ heap.Interface = &priorityQueue{} // Compile-time verification

	pq := &priorityQueue{}
	heap.Init(pq)

	// Test all heap operations in sequence
	heap.Push(pq, &lfuEntry{count: 2})
	heap.Push(pq, &lfuEntry{count: 1})
	heap.Push(pq, &lfuEntry{count: 3})

	// Verify heap property is maintained
	prev := -1
	for pq.Len() > 0 {
		entry := heap.Pop(pq).(*lfuEntry)
		if prev != -1 && entry.count < prev {
			t.Error("Heap property violated: elements not in ascending order")
		}
		prev = entry.count
	}
}

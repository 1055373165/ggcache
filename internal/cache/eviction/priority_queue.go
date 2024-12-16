package eviction

// priorityQueue implements a min-heap of lfuEntry items.
// It is used by the LFU cache to maintain the order of cache entries
// based on their access frequency and last update time.
type priorityQueue []*lfuEntry

// lfuEntry represents an entry in the LFU cache.
type lfuEntry struct {
	index int   // The index of the entry in the heap
	entry Entry // The actual cache entry containing key, value and update time
	count int   // Number of times this entry has been accessed
}

// referenced increments the access count of the entry and updates its timestamp.
func (l *lfuEntry) referenced() {
	l.count++
	l.entry.Touch()
}

// Less implements heap.Interface.
// Entries are ordered first by access count, then by update time for equal counts.
func (pq priorityQueue) Less(i, j int) bool {
	if pq[i].count == pq[j].count {
		return pq[i].entry.UpdateAt.Before(pq[j].entry.UpdateAt)
	}
	return pq[i].count < pq[j].count
}

// Len implements heap.Interface.
func (pq priorityQueue) Len() int {
	return len(pq)
}

// Swap implements heap.Interface.
// When swapping elements, their indices must be updated to maintain heap consistency.
func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

// Push implements heap.Interface.
// The pushed element's index is set to its position in the heap.
func (pq *priorityQueue) Push(x interface{}) {
	entry := x.(*lfuEntry)
	entry.index = len(*pq)
	*pq = append(*pq, entry)
}

// Pop implements heap.Interface.
// The popped element is removed from the heap and its memory is properly cleaned up.
func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	entry := old[n-1]
	old[n-1] = nil   // Avoid memory leak
	entry.index = -1 // For safety
	*pq = old[0 : n-1]
	return entry
}

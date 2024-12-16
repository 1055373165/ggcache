// Package cache implements a distributed cache system with various features.
package cache

import (
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
)

// Hash maps bytes to uint32
type Hash func(data []byte) uint32

// Map constains all hashed keys
type ConsistentMap struct {
	mu       sync.RWMutex
	hash     Hash           // Hash function to use
	replicas int            // Number of virtual nodes per real node
	keys     []int          // Sorted hash keys
	hashMap  map[int]string // Map from hash key to real node name
}

// New creates a Map instance with given replicas count and hash function.
// If hash is nil, crc32.ChecksumIEEE is used.
func NewConsistentHash(replicas int, fn Hash) *ConsistentMap {
	m := &ConsistentMap{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add adds nodes to the hash ring.
func (m *ConsistentMap) AddNodes(nodes ...string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, node := range nodes {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + node)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = node
		}
	}
	sort.Ints(m.keys)
}

// Get returns the closest node in the hash ring to the provided key.
func (m *ConsistentMap) GetNode(key string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))

	// Binary search for appropriate replica
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	// Wrap around to first replica if necessary
	if idx == len(m.keys) {
		idx = 0
	}

	return m.hashMap[m.keys[idx]]
}

// Remove removes a node from the hash ring.
func (m *ConsistentMap) RemoveNode(node string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := 0; i < m.replicas; i++ {
		hash := int(m.hash([]byte(strconv.Itoa(i) + node)))
		idx := sort.SearchInts(m.keys, hash)
		m.keys = append(m.keys[:idx], m.keys[idx+1:]...)
		delete(m.hashMap, hash)
	}
}

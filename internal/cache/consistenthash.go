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

// GetNode returns the closest node in the hash ring to the provided key.
// Returns empty string if the hash ring is empty or key is empty.
func (m *ConsistentMap) GetNode(key string) string {
	if key == "" {
		return ""
	}

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

// RemoveNode removes a node and all its replicas from the hash ring.
// It's safe to call this method even if the node doesn't exist.
func (m *ConsistentMap) RemoveNode(node string) {
	if node == "" {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Calculate all possible hashes for the node's replicas
	var hashesToRemove []int
	for i := 0; i < m.replicas; i++ {
		hash := int(m.hash([]byte(strconv.Itoa(i) + node)))
		if _, exists := m.hashMap[hash]; exists {
			hashesToRemove = append(hashesToRemove, hash)
			delete(m.hashMap, hash)
		}
	}

	// Remove the hashes from the sorted keys slice
	if len(hashesToRemove) > 0 {
		newKeys := make([]int, 0, len(m.keys)-len(hashesToRemove))
		for _, k := range m.keys {
			shouldKeep := true
			for _, h := range hashesToRemove {
				if k == h {
					shouldKeep = false
					break
				}
			}
			if shouldKeep {
				newKeys = append(newKeys, k)
			}
		}
		m.keys = newKeys
	}
}

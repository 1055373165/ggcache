// Package cache implements a distributed cache system with various features.
package cache

import (
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
)

// Hash defines a function that generates a hash value for the given data.
type Hash func(data []byte) uint32

// ConsistentMap implements consistent hashing to distribute keys across nodes.
// It maintains a ring of virtual nodes to ensure even distribution.
type ConsistentMap struct {
	mu       sync.RWMutex
	hash     Hash           // hash function to use
	replicas int            // number of virtual nodes per real node
	keys     []int          // sorted list of hash keys
	hashMap  map[int]string // maps virtual nodes to real nodes
}

// NewConsistentHash creates a ConsistentMap with the specified number of replicas
// and an optional hash function. If fn is nil, uses crc32.ChecksumIEEE.
func NewConsistentHash(replicas int, fn Hash) *ConsistentMap {
	if replicas <= 0 {
		replicas = 1
	}

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

// AddNodes adds the specified nodes to the hash ring.
// Each node is replicated multiple times for better distribution.
func (m *ConsistentMap) AddNodes(nodes ...string) {
	if len(nodes) == 0 {
		return
	}

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

// GetNode returns the node responsible for the given key.
// Returns empty string if the hash ring is empty or key is invalid.
func (m *ConsistentMap) GetNode(key string) string {
	if key == "" || m == nil {
		return ""
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	// If we reached the end, wrap around to the first node
	if idx == len(m.keys) {
		idx = 0
	}

	return m.hashMap[m.keys[idx]]
}

// RemoveNode removes a node and its replicas from the hash ring.
// This operation is safe even if the node doesn't exist.
func (m *ConsistentMap) RemoveNode(node string) {
	if node == "" || m == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Find all hashes for this node's replicas
	hashesToRemove := make([]int, 0, m.replicas)
	for i := 0; i < m.replicas; i++ {
		hash := int(m.hash([]byte(strconv.Itoa(i) + node)))
		if _, exists := m.hashMap[hash]; exists {
			hashesToRemove = append(hashesToRemove, hash)
			delete(m.hashMap, hash)
		}
	}

	// If no hashes were found, nothing to remove
	if len(hashesToRemove) == 0 {
		return
	}

	// Create a new slice without the removed hashes
	newKeys := make([]int, 0, len(m.keys)-len(hashesToRemove))
	for _, k := range m.keys {
		if !containsInt(hashesToRemove, k) {
			newKeys = append(newKeys, k)
		}
	}
	m.keys = newKeys
}

// containsInt returns true if x is present in the sorted slice nums.
func containsInt(nums []int, x int) bool {
	for _, n := range nums {
		if n == x {
			return true
		}
	}
	return false
}

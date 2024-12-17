package cache

import (
	"fmt"
	"hash/crc32"
	"strconv"
	"testing"
)

func TestNewConsistentHash(t *testing.T) {
	tests := []struct {
		name     string
		replicas int
		wantRep  int
	}{
		{"zero replicas", 0, 1},
		{"negative replicas", -1, 1},
		{"valid replicas", 3, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewConsistentHash(tt.replicas, nil)
			if got.replicas != tt.wantRep {
				t.Errorf("NewConsistentHash(%d) got replicas = %d, want %d",
					tt.replicas, got.replicas, tt.wantRep)
			}
			if got.hash == nil {
				t.Error("NewConsistentHash() hash function is nil")
			}
		})
	}
}

func TestConsistentHash_AddNodes(t *testing.T) {
	ch := NewConsistentHash(3, func(data []byte) uint32 {
		i, _ := strconv.Atoi(string(data))
		return uint32(i)
	})

	// Test adding single node
	t.Run("add single node", func(t *testing.T) {
		ch.AddNodes("6")
		if len(ch.keys) != 3 { // 3 replicas
			t.Errorf("got %d keys, want 3", len(ch.keys))
		}
	})

	// Test adding multiple nodes
	t.Run("add multiple nodes", func(t *testing.T) {
		ch = NewConsistentHash(3, nil)
		ch.AddNodes("2", "4", "6")
		if len(ch.keys) != 9 { // 3 nodes * 3 replicas
			t.Errorf("got %d keys, want 9", len(ch.keys))
		}
	})

	// Test keys are sorted
	t.Run("keys are sorted", func(t *testing.T) {
		for i := 1; i < len(ch.keys); i++ {
			if ch.keys[i] < ch.keys[i-1] {
				t.Errorf("keys not sorted: keys[%d] = %d < keys[%d] = %d",
					i, ch.keys[i], i-1, ch.keys[i-1])
			}
		}
	})
}

func TestConsistentHash_GetNode(t *testing.T) {
	ch := NewConsistentHash(3, func(data []byte) uint32 {
		return crc32.ChecksumIEEE(data)
	})

	// Test empty hash ring
	t.Run("empty ring", func(t *testing.T) {
		if got := ch.GetNode("key"); got != "" {
			t.Errorf("empty ring got %q, want empty string", got)
		}
	})

	nodes := []string{"node1", "node2", "node3"}
	ch.AddNodes(nodes...)

	tests := []struct {
		name    string
		key     string
		wantNil bool
	}{
		{"empty key", "", true},
		{"normal key", "test_key", false},
		{"special chars", "!@#$%^", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ch.GetNode(tt.key)
			if tt.wantNil && got != "" {
				t.Errorf("GetNode(%q) = %q, want empty string", tt.key, got)
			}
			if !tt.wantNil && got == "" {
				t.Errorf("GetNode(%q) = empty string, want non-empty", tt.key)
			}
		})
	}

	// Test consistent hashing property
	t.Run("consistency", func(t *testing.T) {
		key := "test_key"
		node1 := ch.GetNode(key)
		node2 := ch.GetNode(key)
		if node1 != node2 {
			t.Errorf("inconsistent hashing: got %q and %q for same key", node1, node2)
		}
	})
}

func TestConsistentHash_RemoveNode(t *testing.T) {
	ch := NewConsistentHash(3, nil)
	nodes := []string{"node1", "node2", "node3"}
	ch.AddNodes(nodes...)

	initialKeys := len(ch.keys)

	tests := []struct {
		name       string
		removeNode string
		wantKeys   int
	}{
		{"remove existing", "node1", initialKeys - 3},     // 3 replicas less
		{"remove non-existing", "node4", initialKeys - 3}, // no change from previous
		{"remove empty", "", initialKeys - 3},             // no change
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch.RemoveNode(tt.removeNode)
			if len(ch.keys) != tt.wantKeys {
				t.Errorf("after removing %q got %d keys, want %d",
					tt.removeNode, len(ch.keys), tt.wantKeys)
			}
		})
	}

	// Verify remaining nodes are still accessible
	t.Run("remaining nodes accessible", func(t *testing.T) {
		node := ch.GetNode("test_key")
		if node == "node1" {
			t.Error("got removed node")
		}
	})
}

func TestConsistentHash(t *testing.T) {
	// Use a custom hash function for deterministic testing
	hashFunc := func(data []byte) uint32 {
		// Simple hash function for testing that maps "key1" and "key2" to first node
		// and "key3" and "key4" to second node
		if len(data) == 0 {
			return 0
		}
		if int(data[len(data)-1])%2 == 0 {
			fmt.Println(data[len(data)-1])
			return 2 // will map to first node
		}
		return 4 // will map to second node
	}

	ch := NewConsistentHash(1, hashFunc)

	// Add test nodes in sorted order
	nodes := []string{"A", "B"}
	ch.AddNodes(nodes...)

	// Test node distribution
	testCases := []struct {
		key  string
		want string
	}{
		{"key1", "A"}, // odd hash, maps to first node
		{"key3", "A"}, // odd hash, maps to first node
		{"key2", "B"}, // even hash, maps to second node
		{"key4", "B"}, // even hash, maps to second node
	}

	for _, tc := range testCases {
		t.Run(tc.key, func(t *testing.T) {
			got := ch.GetNode(tc.key)
			if got != tc.want {
				t.Errorf("GetNode(%q) = %q; want %q", tc.key, got, tc.want)
			}
		})
	}

	// Test node removal
	ch.RemoveNode("A")

	// After removing A, all keys should map to B
	for _, tc := range testCases {
		t.Run(tc.key+" after removal", func(t *testing.T) {
			got := ch.GetNode(tc.key)
			if got != "B" {
				t.Errorf("After removing A, GetNode(%q) = %q; want B", tc.key, got)
			}
		})
	}
}

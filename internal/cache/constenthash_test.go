package cache

import (
	"testing"

	"github.com/1055373165/ggcache/config"
)

func TestConsistentHash(t *testing.T) {
	config.InitClientV3Config()

	// Use crc32.ChecksumIEEE hash algorithm
	ch := NewConsistentHash(2, nil)

	// Add test nodes
	ch.AddNodes([]string{"2", "4"}...)

	// Test node distribution
	testCases := []struct {
		key  string
		want string
	}{
		{"key1", "2"},
		{"key2", "2"},
		{"key3", "4"},
		{"key4", "4"},
	}

	for _, tc := range testCases {
		t.Run(tc.key, func(t *testing.T) {
			if got := ch.GetNode(tc.key); got != tc.want {
				t.Errorf("GetNode(%q) = %q; want %q", tc.key, got, tc.want)
			}
		})
	}

	// Test node removal
	ch.RemoveNode("2")
	if got := ch.GetNode("key1"); got != "4" {
		t.Errorf("After removing node 2, GetNode(key1) = %q; want '4'", got)
	}
}

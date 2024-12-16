package eviction

import (
	"testing"
	"time"
)

func TestCacheUseLFU_Basic(t *testing.T) {
	t.Run("creation", func(t *testing.T) {
		lfu := NewCacheUseLFU(100, nil)
		if lfu == nil {
			t.Error("Failed to create LFU cache")
		}
		if lfu.Len() != 0 {
			t.Errorf("New cache should be empty, got len = %d", lfu.Len())
		}
	})

	t.Run("basic operations", func(t *testing.T) {
		lfu := NewCacheUseLFU(1024, nil)

		// Test Put and Get
		lfu.Put("key1", String("value1"))
		if v, _, ok := lfu.Get("key1"); !ok || string(v.(String)) != "value1" {
			t.Errorf("Get after Put failed, got %v, want %v", v, "value1")
		}

		// Test missing key
		if _, _, ok := lfu.Get("missing"); ok {
			t.Error("Get with missing key should return false")
		}

		// Test update existing key
		lfu.Put("key1", String("value2"))
		if v, _, ok := lfu.Get("key1"); !ok || string(v.(String)) != "value2" {
			t.Errorf("Get after update failed, got %v, want %v", v, "value2")
		}
	})
}

func TestCacheUseLFU_FrequencyBehavior(t *testing.T) {
	tests := []struct {
		name     string
		maxBytes int64
		ops      []struct {
			op    string // "put" or "get"
			key   string
			value string
		}
		wantPresent []string // keys that should be present after operations
	}{
		{
			name:     "least frequently used eviction",
			maxBytes: 20, // Only enough for 2 entries
			ops: []struct {
				op    string
				key   string
				value string
			}{
				{"put", "k1", "v1"}, // count will be 1
				{"put", "k2", "v2"}, // count will be 1
				{"get", "k2", ""},   // k2 count becomes 2
				{"put", "k3", "v3"}, // should evict k1 (count=1)
			},
			wantPresent: []string{"k2", "k3"},
		},
		{
			name:     "frequency tie-breaking by time",
			maxBytes: 20,
			ops: []struct {
				op    string
				key   string
				value string
			}{
				{"put", "k1", "v1"},
				{"put", "k2", "v2"},
				{"get", "k1", ""},
				{"get", "k2", ""},
				{"put", "k3", "v3"}, // should evict older entry with count 1
			},
			wantPresent: []string{"k2", "k3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lfu := NewCacheUseLFU(tt.maxBytes, nil)

			for _, op := range tt.ops {
				switch op.op {
				case "put":
					lfu.Put(op.key, String(op.value))
				case "get":
					lfu.Get(op.key)
				}
			}

			// Verify present keys
			for _, key := range tt.wantPresent {
				if _, _, ok := lfu.Get(key); !ok {
					t.Errorf("Key %s should be present", key)
				}
			}
		})
	}
}

func TestCacheUseLFU_MemoryManagement(t *testing.T) {
	evicted := make(map[string]Value)
	onEvicted := func(key string, value Value) {
		evicted[key] = value
	}

	lfu := NewCacheUseLFU(10, onEvicted) // Only enough for 2 entries

	// Add entries that should fit
	lfu.Put("k1", String("v1"))
	lfu.Put("k2", String("v2"))

	if got := lfu.Len(); got != 2 {
		t.Errorf("Cache length = %d, want 2", got)
	}

	// Add another entry that should trigger eviction
	lfu.Put("k3", String("v3"))

	if got := lfu.Len(); got != 2 {
		t.Errorf("Cache length after eviction = %d, want 2", got)
	}

	if _, ok := evicted["k1"]; !ok {
		t.Error("Eviction callback should have been called for k1")
	}
}

func TestCacheUseLFU_CleanUp(t *testing.T) {
	lfu := NewCacheUseLFU(1024, nil)

	// Add some entries
	lfu.Put("k1", String("v1"))
	lfu.Put("k2", String("v2"))
	lfu.Put("k3", String("v3"))

	// Access some entries to vary their frequencies and ensure update times are set
	_, _, ok1 := lfu.Get("k1")
	_, _, ok2 := lfu.Get("k1")
	_, _, ok3 := lfu.Get("k2")

	if !ok1 || !ok2 || !ok3 {
		t.Fatal("Failed to get entries that should exist")
	}

	// Wait for entries to expire
	time.Sleep(10 * time.Millisecond)

	// Clean up with small TTL to force expiration
	lfu.CleanUp(5 * time.Millisecond)

	// All entries should be cleaned up
	if lfu.Len() != 0 {
		t.Errorf("CleanUp failed, cache should be empty, got len = %d", lfu.Len())
	}
}

func TestCacheUseLFU_ZeroSize(t *testing.T) {
	lfu := NewCacheUseLFU(0, nil) // Zero size means unlimited

	// Should be able to add any number of entries
	for i := 0; i < 100; i++ {
		key := String("k" + string(rune(i+'0')))
		lfu.Put(string(key), key)
	}

	if lfu.Len() != 100 {
		t.Errorf("Expected 100 entries in unlimited cache, got %d", lfu.Len())
	}
}

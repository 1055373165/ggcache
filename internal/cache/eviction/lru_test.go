package eviction

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestCacheUseLRU_Basic(t *testing.T) {
	t.Run("creation", func(t *testing.T) {
		lru := NewCacheUseLRU(100, nil)
		if lru == nil {
			t.Error("Failed to create LRU cache")
		}
		if lru.Len() != 0 {
			t.Errorf("New cache should be empty, got len = %d", lru.Len())
		}
	})

	t.Run("basic operations", func(t *testing.T) {
		lru := NewCacheUseLRU(1024, nil)

		// Test Put and Get
		lru.Put("key1", String("value1"))
		if v, _, ok := lru.Get("key1"); !ok || string(v.(String)) != "value1" {
			t.Errorf("Get after Put failed, got %v, want %v", v, "value1")
		}

		// Test missing key
		if _, _, ok := lru.Get("missing"); ok {
			t.Error("Get with missing key should return false")
		}

		// Test update existing key
		lru.Put("key1", String("value2"))
		if v, _, ok := lru.Get("key1"); !ok || string(v.(String)) != "value2" {
			t.Errorf("Get after update failed, got %v, want %v", v, "value2")
		}
	})
}

func TestCacheUseLRU_EvictionOrder(t *testing.T) {
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
			name:     "least recently used eviction",
			maxBytes: 512, // Increased to account for segmentation (16 segments * 32 bytes)
			ops: []struct {
				op    string
				key   string
				value string
			}{
				{"put", "k1", "v1"},
				{"put", "k2", "v2"},
				{"get", "k1", ""},   // k1 becomes most recently used
				{"put", "k3", "v3"}, // should evict k2
			},
			wantPresent: []string{"k1", "k3"},
		},
		{
			name:     "update makes entry most recent",
			maxBytes: 512, // Increased to account for segmentation
			ops: []struct {
				op    string
				key   string
				value string
			}{
				{"put", "k1", "v1"},
				{"put", "k2", "v2"},
				{"put", "k1", "v1-new"}, // k1 becomes most recent
				{"put", "k3", "v3"},     // should evict k2
			},
			wantPresent: []string{"k1", "k3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lru := NewCacheUseLRU(tt.maxBytes, nil)
			defer lru.Stop() // Ensure cleanup goroutine is stopped

			for _, op := range tt.ops {
				switch op.op {
				case "put":
					lru.Put(op.key, String(op.value))
				case "get":
					lru.Get(op.key)
				}
			}

			// Verify present keys
			for _, key := range tt.wantPresent {
				if _, _, ok := lru.Get(key); !ok {
					t.Errorf("Key %s should be present", key)
				}
			}
		})
	}
}

func TestCacheUseLRU_MemoryManagement(t *testing.T) {
	var mu sync.Mutex
	evicted := make(map[string]Value)
	onEvicted := func(key string, value Value) {
		mu.Lock()
		evicted[key] = value
		mu.Unlock()
	}

	// Calculate exact memory requirements:
	// Each entry needs: len(key) + len(value) bytes
	// For keys like "key-0", "key-1": 5 bytes each
	// For values "1234", "5678", "9012": 4 bytes each
	// So each entry takes ~9 bytes
	// With 16 bytes per segment (256/16), we can fit 1 entry comfortably
	lru := NewCacheUseLRU(256, onEvicted)
	defer lru.Stop()

	// Find 3 keys that hash to the same segment
	keys := findKeysInSameSegment(lru, 3)
	k1, k2, k3 := keys[0], keys[1], keys[2]

	// Add first entry (~9 bytes)
	v1 := String("1234")
	lru.Put(k1, v1)

	// Add second entry (~9 bytes), should trigger eviction of k1
	v2 := String("5678")
	lru.Put(k2, v2)

	// Verify only k2 is in cache
	if _, _, ok := lru.Get(k1); ok {
		t.Error("k1 should have been evicted")
	}
	if _, _, ok := lru.Get(k2); !ok {
		t.Error("k2 should be in cache")
	}

	mu.Lock()
	_, evictedK1 := evicted[k1]
	mu.Unlock()

	if !evictedK1 {
		t.Error("Eviction callback should have been called for k1")
	}

	// Add third entry, should trigger eviction of k2
	v3 := String("9012")
	lru.Put(k3, v3)

	// Verify only k3 is in cache
	if _, _, ok := lru.Get(k1); ok {
		t.Error("k1 should still be evicted")
	}
	if _, _, ok := lru.Get(k2); ok {
		t.Error("k2 should have been evicted")
	}
	if _, _, ok := lru.Get(k3); !ok {
		t.Error("k3 should be in cache")
	}
}

func TestCacheUseLRU_CleanUp(t *testing.T) {
	lru := NewCacheUseLRU(1024, nil)
	lru.SetCleanupInterval(50 * time.Millisecond)
	lru.SetTTL(100 * time.Millisecond)

	// Add some entries
	lru.Put("k1", String("v1"))
	lru.Put("k2", String("v2"))
	lru.Put("k3", String("v3"))

	// Access some entries to vary their last access time
	time.Sleep(20 * time.Millisecond)
	_, _, ok1 := lru.Get("k1")
	time.Sleep(20 * time.Millisecond)
	_, _, ok2 := lru.Get("k2")
	time.Sleep(20 * time.Millisecond)
	_, _, ok3 := lru.Get("k3")

	if !ok1 || !ok2 || !ok3 {
		t.Fatal("Failed to get entries that should exist")
	}

	// Wait for cleanup to occur
	time.Sleep(150 * time.Millisecond)

	// All entries should be cleaned up
	if lru.Len() != 0 {
		t.Errorf("CleanUp failed, cache should be empty, got len = %d", lru.Len())
	}
}

func TestCacheUseLRU_Concurrent(t *testing.T) {
	lru := NewCacheUseLRU(1024, nil)
	var wg sync.WaitGroup
	numOps := 1000
	numGoroutines := 10

	// Concurrent writes and reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(2)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				key := String("k" + string(rune(j%100+'0')))
				lru.Put(string(key), key)
			}
		}(i)

		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				key := "k" + string(rune(j%100+'0'))
				lru.Get(key)
			}
		}(i)
	}

	wg.Wait()
}

func TestCacheUseLRU_ZeroSize(t *testing.T) {
	lru := NewCacheUseLRU(0, nil) // Zero size means unlimited

	// Should be able to add any number of entries
	for i := 0; i < 100; i++ {
		key := String("k" + string(rune(i+'0')))
		lru.Put(string(key), key)
	}

	if lru.Len() != 100 {
		t.Errorf("Expected 100 entries in unlimited cache, got %d", lru.Len())
	}
}

func TestCacheUseLRU_SetTTL(t *testing.T) {
	lru := NewCacheUseLRU(1024, nil)

	// Add an entry
	lru.Put("key1", String("value1"))

	// Set a very short TTL
	lru.SetTTL(10 * time.Millisecond)

	// Verify entry exists
	if _, _, ok := lru.Get("key1"); !ok {
		t.Error("Entry should exist before TTL expiration")
	}

	// Wait for TTL to expire
	time.Sleep(20 * time.Millisecond)

	// Clean up expired entries
	lru.CleanUp(10 * time.Millisecond)

	// Verify entry has been removed
	if _, _, ok := lru.Get("key1"); ok {
		t.Error("Entry should have been removed after TTL expiration")
	}
}

func TestCacheUseLRU_SetCleanupInterval(t *testing.T) {
	lru := NewCacheUseLRU(1024, nil)

	// Set a very short cleanup interval and TTL first
	shortInterval := 10 * time.Millisecond
	lru.SetCleanupInterval(shortInterval)
	lru.SetTTL(5 * time.Millisecond)

	// Add an entry after setting TTL
	lru.Put("key1", String("value1"))

	// Wait for cleanup to occur
	time.Sleep(20 * time.Millisecond)

	// Entry should be automatically cleaned up due to the short cleanup interval
	if _, _, ok := lru.Get("key1"); ok {
		t.Error("Entry should have been automatically cleaned up")
	}
}

func TestCacheUseLRU_Stop(t *testing.T) {
	lru := NewCacheUseLRU(1024, nil)

	// Add an entry
	lru.Put("key1", String("value1"))

	// Set short TTL and cleanup interval
	lru.SetTTL(5 * time.Millisecond)
	lru.SetCleanupInterval(10 * time.Millisecond)

	// Stop the cleanup routine
	lru.Stop()

	// Wait for what would have been multiple cleanup cycles
	time.Sleep(30 * time.Millisecond)

	// Entry should still exist because cleanup routine was stopped
	if _, _, ok := lru.Get("key1"); !ok {
		t.Error("Entry should still exist after stopping cleanup routine")
	}
}

// findKeysInSameSegment returns n keys that hash to the same segment
func findKeysInSameSegment(lru *CacheUseLRU, n int) []string {
	seen := make(map[*segment][]string)
	i := 0
	for {
		key := fmt.Sprintf("key-%d", i)
		seg := lru.getSegment(key)
		seen[seg] = append(seen[seg], key)
		if len(seen[seg]) == n {
			return seen[seg]
		}
		i++
	}
}

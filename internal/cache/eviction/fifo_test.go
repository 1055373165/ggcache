package eviction

import (
	"sync"
	"testing"
	"time"
)

// String is a test helper type that implements Value interface.
type String string

func (d String) Len() int {
	return len(d)
}

func (d String) String() string {
	return string(d)
}

func TestNewCacheUseFIFO(t *testing.T) {
	tests := []struct {
		name      string
		maxBytes  int64
		wantCache bool
	}{
		{
			name:      "valid cache",
			maxBytes:  100,
			wantCache: true,
		},
		{
			name:      "zero size cache",
			maxBytes:  0,
			wantCache: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := NewCacheUseFIFO(tt.maxBytes, nil)
			if (cache != nil) != tt.wantCache {
				t.Errorf("NewFIFOCache() = %v, want %v", cache != nil, tt.wantCache)
			}
		})
	}
}

func TestCacheUseFIFO_Operations(t *testing.T) {
	t.Run("basic operations", func(t *testing.T) {
		cache := NewCacheUseFIFO(1024, nil)

		// Test Put and Get
		cache.Put("key1", String("value1"))
		if v, _, ok := cache.Get("key1"); !ok || string(v.(String)) != "value1" {
			t.Errorf("Get after Put failed, got %v, want %v", v, "value1")
		}

		// Test missing key
		if _, _, ok := cache.Get("missing"); ok {
			t.Error("Get with missing key should return false")
		}

		// Test update existing key
		cache.Put("key1", String("value2"))
		if v, _, ok := cache.Get("key1"); !ok || string(v.(String)) != "value2" {
			t.Errorf("Get after update failed, got %v, want %v", v, "value2")
		}
	})

	t.Run("eviction", func(t *testing.T) {
		evicted := make(map[string]Value)
		onEvicted := func(key string, value Value) {
			evicted[key] = value
		}

		cache := NewCacheUseFIFO(10, onEvicted) // Only enough for 2 entries

		cache.Put("k1", String("v1")) // 4 bytes
		cache.Put("k2", String("v2")) // 4 bytes
		cache.Put("k3", String("v3")) // Should trigger eviction of k1

		if _, _, ok := cache.Get("k1"); ok {
			t.Error("k1 should have been evicted")
		}

		if _, ok := evicted["k1"]; !ok {
			t.Error("eviction callback should have been called for k1")
		}
	})
}

func TestCacheUseFIFO_CleanUp(t *testing.T) {
	cache := NewCacheUseFIFO(1024, nil)

	// Add some entries
	cache.Put("k1", String("v1"))
	cache.Put("k2", String("v2"))
	cache.Put("k3", String("v3"))

	// Wait for entries to expire
	time.Sleep(10 * time.Millisecond)

	// Clean up with small TTL to force expiration
	cache.CleanUp(5 * time.Millisecond)

	// All entries should be cleaned up
	if cache.Len() != 0 {
		t.Errorf("CleanUp failed, cache should be empty, got len = %d", cache.Len())
	}
}

func TestCacheUseFIFO_Concurrent(t *testing.T) {
	cache := NewCacheUseFIFO(1024, nil)
	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := String("k" + string(rune(i+'0')))
			cache.Put(string(key), key)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := String("k" + string(rune(i+'0')))
			cache.Get(string(key))
		}(i)
	}

	wg.Wait()
}

func TestCacheUseFIFO_MaxBytes(t *testing.T) {
	tests := []struct {
		name     string
		maxBytes int64
		puts     []struct {
			key   string
			value String
		}
		wantLen int
	}{
		{
			name:     "respect max bytes",
			maxBytes: 10,
			puts: []struct {
				key   string
				value String
			}{
				{"k1", "v1"}, // 4 bytes
				{"k2", "v2"}, // 4 bytes
				{"k3", "v3"}, // 4 bytes (should trigger eviction)
			},
			wantLen: 2,
		},
		{
			name:     "unlimited cache",
			maxBytes: 0,
			puts: []struct {
				key   string
				value String
			}{
				{"k1", "v1"},
				{"k2", "v2"},
				{"k3", "v3"},
			},
			wantLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := NewCacheUseFIFO(tt.maxBytes, nil)

			for _, put := range tt.puts {
				cache.Put(put.key, put.value)
			}

			if got := cache.Len(); got != tt.wantLen {
				t.Errorf("cache.Len() = %v, want %v", got, tt.wantLen)
			}
		})
	}
}

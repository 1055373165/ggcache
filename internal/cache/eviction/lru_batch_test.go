package eviction

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestCacheUseLRUBatch_Basic(t *testing.T) {
	t.Run("creation and configuration", func(t *testing.T) {
		lru := NewCacheUseLRUBatch(100, nil)
		if lru == nil {
			t.Error("Failed to create LRU batch cache")
		}

		// Test batch size configuration
		lru.SetBatchSize(50)
		lru.SetTTL(time.Minute)
		lru.SetCleanupInterval(time.Minute)
	})

	t.Run("basic operations", func(t *testing.T) {
		lru := NewCacheUseLRUBatch(1024, nil)

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

func TestCacheUseLRUBatch_BatchEviction(t *testing.T) {
	var mu sync.Mutex
	evicted := make(map[string]Value)
	onEvicted := func(key string, value Value) {
		mu.Lock()
		evicted[key] = value
		mu.Unlock()
	}

	lru := NewCacheUseLRUBatch(30, onEvicted)
	lru.SetBatchSize(2) // Small batch size for testing
	// Add entries that exceed cache size
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("k%d", i)
		lru.Put(key, String(fmt.Sprintf("v%d", i)))
	}

	// Wait for eviction to complete
	time.Sleep(10 * time.Millisecond)

	// Check that older entries were evicted
	mu.Lock()
	evictedCount := len(evicted)
	mu.Unlock()

	if evictedCount == 0 {
		t.Error("No entries were evicted")
	}
}

func TestCacheUseLRUBatch_BatchCleanup(t *testing.T) {
	lru := NewCacheUseLRUBatch(1024, nil)
	lru.SetBatchSize(2)
	lru.SetCleanupInterval(50 * time.Millisecond)
	lru.SetTTL(100 * time.Millisecond)

	// Add entries
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("k%d", i)
		lru.Put(key, String(fmt.Sprintf("v%d", i)))
		if i%2 == 0 {
			time.Sleep(20 * time.Millisecond)
		}
	}

	// Wait for cleanup to occur
	time.Sleep(200 * time.Millisecond)

	// All entries should be cleaned up
	if lru.Len() != 0 {
		t.Errorf("Batch cleanup failed, cache should be empty, got len = %d", lru.Len())
	}
}

func TestCacheUseLRUBatch_ConcurrentBatchOperations(t *testing.T) {
	lru := NewCacheUseLRUBatch(1024, nil)
	lru.SetBatchSize(10)

	var wg sync.WaitGroup
	numOps := 1000
	numGoroutines := 10

	// Concurrent batch operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(2)

		// Writer goroutine
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				key := fmt.Sprintf("k%d-%d", id, j)
				lru.Put(key, String(fmt.Sprintf("v%d-%d", id, j)))
			}
		}(i)

		// Reader goroutine
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				key := fmt.Sprintf("k%d-%d", id, j)
				lru.Get(key)
			}
		}(i)
	}

	wg.Wait()
}

func TestCacheUseLRUBatch_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	lru := NewCacheUseLRUBatch(1<<20, nil) // 1MB cache
	lru.SetBatchSize(100)

	var wg sync.WaitGroup
	numOps := 10000
	numGoroutines := 20

	start := time.Now()

	// Multiple concurrent readers and writers
	for i := 0; i < numGoroutines; i++ {
		wg.Add(3)

		// Writer
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				key := fmt.Sprintf("w%d-%d", id, j)
				lru.Put(key, String(fmt.Sprintf("value-%d-%d", id, j)))
			}
		}(i)

		// Reader 1
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				key := fmt.Sprintf("w%d-%d", id, j)
				lru.Get(key)
			}
		}(i)

		// Reader 2 (reading with different pattern)
		go func(id int) {
			defer wg.Done()
			for j := numOps - 1; j >= 0; j-- {
				key := fmt.Sprintf("w%d-%d", (id+1)%numGoroutines, j)
				lru.Get(key)
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)
	opsPerSec := float64(numOps*numGoroutines*3) / duration.Seconds()

	t.Logf("Stress test completed: %d operations in %v (%.2f ops/sec)",
		numOps*numGoroutines*3, duration, opsPerSec)
}

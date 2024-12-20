package eviction

import (
	"fmt"
	"math/rand"
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
		defer lru.Stop()
	})

	t.Run("basic operations", func(t *testing.T) {
		lru := NewCacheUseLRUBatch(1024, nil)
		defer lru.Stop()

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
	defer lru.Stop()
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
	defer lru.Stop()
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

func TestCacheUseLRUBatch_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	lru := NewCacheUseLRUBatch(1<<20, nil) // 1MB cache
	defer lru.Stop()
	lru.SetBatchSize(50) // 减小批量大小

	var wg sync.WaitGroup
	numOps := 1000     // 减少操作次数
	numGoroutines := 5 // 减少并发数

	// 创建一个通道来同步所有 goroutine 的开始时间
	start := make(chan struct{})

	// 记录开始时间
	startTime := time.Now()

	// 启动写入 goroutine
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			<-start // 等待开始信号
			for j := 0; j < numOps; j++ {
				key := fmt.Sprintf("w%d-%d", id, j)
				value := String(fmt.Sprintf("value-%d-%d", id, j))
				lru.Put(key, value)
				// 添加短暂随机延迟，避免过度竞争
				if j%10 == 0 {
					time.Sleep(time.Duration(rand.Intn(100)) * time.Microsecond)
				}
			}
		}(i)
	}

	// 启动读取 goroutine
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			<-start // 等待开始信号
			for j := 0; j < numOps; j++ {
				key := fmt.Sprintf("w%d-%d", id, j)
				_, _, _ = lru.Get(key)
				if j%10 == 0 {
					time.Sleep(time.Duration(rand.Intn(100)) * time.Microsecond)
				}
			}
		}(i)
	}

	// 启动交错读取 goroutine
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			<-start // 等待开始信号
			for j := 0; j < numOps; j++ {
				key := fmt.Sprintf("w%d-%d", (id+1)%numGoroutines, j)
				_, _, _ = lru.Get(key)
				if j%10 == 0 {
					time.Sleep(time.Duration(rand.Intn(100)) * time.Microsecond)
				}
			}
		}(i)
	}

	// 发送开始信号
	close(start)

	// 等待所有 goroutine 完成
	wg.Wait()

	duration := time.Since(startTime)
	totalOps := numOps * numGoroutines * 3
	opsPerSec := float64(totalOps) / duration.Seconds()

	t.Logf("Stress test completed: %d operations in %v (%.2f ops/sec)",
		totalOps, duration, opsPerSec)

	// 验证缓存状态
	if lru.Len() > int(1<<20/100) {
		t.Errorf("Cache size too large: %d", lru.Len())
	}
}

func TestCacheUseLRUBatch_ConcurrentCleanup(t *testing.T) {
	lru := NewCacheUseLRUBatch(1<<20, nil)
	defer lru.Stop()

	// 设置较短的清理间隔和 TTL
	lru.SetTTL(100 * time.Millisecond)
	lru.SetCleanupInterval(50 * time.Millisecond)
	lru.SetBatchSize(10)

	var wg sync.WaitGroup
	numOps := 100
	numGoroutines := 3

	// 启动写入 goroutine
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				key := fmt.Sprintf("w%d-%d", id, j)
				value := String(fmt.Sprintf("value-%d-%d", id, j))
				lru.Put(key, value)
				time.Sleep(time.Millisecond)
			}
		}(i)
	}

	// 启动读取 goroutine
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				key := fmt.Sprintf("w%d-%d", id, j)
				_, _, _ = lru.Get(key)
				time.Sleep(time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
	time.Sleep(200 * time.Millisecond) // 等待最后的清理完成

	// 验证缓存状态
	size := lru.Len()
	if size > numOps*numGoroutines {
		t.Errorf("Cache size too large: %d", size)
	}
}

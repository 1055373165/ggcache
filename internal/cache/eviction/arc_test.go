package eviction

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

type ArcString string

func (s ArcString) Len() int {
	return len(s)
}

func (s ArcString) String() string {
	return string(s)
}

func TestCacheUseARC_Basic(t *testing.T) {
	arc := NewCacheUseARC(1024, nil)

	// Test basic put and get
	arc.Put("key1", ArcString("value1"))
	if val, _, ok := arc.Get("key1"); !ok || val.(ArcString).String() != "value1" {
		t.Error("Failed to get value after put")
	}

	// Test update
	arc.Put("key1", String("value2"))
	if val, _, ok := arc.Get("key1"); !ok || val.(String).String() != "value2" {
		t.Error("Failed to update value")
	}

	// Test missing key
	if _, _, ok := arc.Get("missing"); ok {
		t.Error("Got value for missing key")
	}
}

func TestCacheUseARC_Eviction(t *testing.T) {
	var evicted []string
	onEvicted := func(key string, _ Value) {
		evicted = append(evicted, key)
	}

	// Create a small cache
	arc := NewCacheUseARC(20, onEvicted)

	// Add entries that exceed cache size
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("key%d", i)
		arc.Put(key, String(fmt.Sprintf("value%d", i)))
	}

	// Check that some entries were evicted
	if len(evicted) == 0 {
		t.Error("No entries were evicted")
	}
}

func TestCacheUseARC_TTL(t *testing.T) {
	arc := NewCacheUseARC(1024, nil)
	arc.SetTTL(50 * time.Millisecond)
	arc.SetCleanupInterval(10 * time.Millisecond)

	// Add an entry
	arc.Put("key1", String("value1"))

	// Verify it exists
	if _, _, ok := arc.Get("key1"); !ok {
		t.Error("Entry should exist")
	}

	// Wait for TTL to expire
	time.Sleep(100 * time.Millisecond)

	// Entry should be gone
	if _, _, ok := arc.Get("key1"); ok {
		t.Error("Entry should have been cleaned up")
	}
}

func TestCacheUseARC_Concurrency(t *testing.T) {
	arc := NewCacheUseARC(1024, nil)
	var wg sync.WaitGroup

	// Concurrent puts
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Add(-1)
			key := fmt.Sprintf("key%d", i)
			arc.Put(key, String(fmt.Sprintf("value%d", i)))
		}(i)
	}

	// Concurrent gets
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Add(-1)
			key := fmt.Sprintf("key%d", i)
			arc.Get(key)
		}(i)
	}

	wg.Wait()
}

func TestCacheUseARC_AdaptiveBehavior(t *testing.T) {
	arc := NewCacheUseARC(100, nil)

	// Phase 1: Add entries that will be accessed once
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("once%d", i)
		arc.Put(key, String(fmt.Sprintf("value%d", i)))
		arc.Get(key) // Access once
	}

	// Phase 2: Add entries that will be accessed multiple times
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("multi%d", i)
		arc.Put(key, String(fmt.Sprintf("value%d", i)))
		for j := 0; j < 3; j++ {
			arc.Get(key) // Access multiple times
		}
	}

	// Phase 3: Add more entries to force eviction
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("new%d", i)
		arc.Put(key, String(fmt.Sprintf("value%d", i)))
	}

	// Check if frequently accessed entries are still in cache
	frequentHits := 0
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("multi%d", i)
		if _, _, ok := arc.Get(key); ok {
			frequentHits++
		}
	}

	// ARC should favor keeping frequently accessed entries
	if frequentHits < 3 {
		t.Errorf("ARC didn't properly adapt: only %d frequent entries remained", frequentHits)
	}
}

func TestCacheUseARC_GhostCache(t *testing.T) {
	arc := NewCacheUseARC(50, nil)

	// Phase 1: Add and access entries
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("key%d", i)
		arc.Put(key, String(fmt.Sprintf("value%d", i)))
		arc.Get(key)
	}

	// Phase 2: Add more entries to force eviction
	for i := 5; i < 10; i++ {
		key := fmt.Sprintf("key%d", i)
		arc.Put(key, String(fmt.Sprintf("value%d", i)))
	}

	// Phase 3: Re-add evicted entries
	hits := 0
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("key%d", i)
		arc.Put(key, String(fmt.Sprintf("value%d", i)))
		if _, _, ok := arc.Get(key); ok {
			hits++
		}
	}

	// Ghost cache should help retain these entries
	if hits < 3 {
		t.Errorf("Ghost cache not working effectively: only %d hits", hits)
	}
}

// Package eviction provides cache eviction strategies including FIFO, LRU, and LFU.
// Each strategy implements different algorithms for determining which entries to remove
// when the cache reaches its capacity.
package eviction

import (
	"fmt"
	"strings"
	"time"
)

// EvictionType represents the type of eviction strategy using integer enum
type EvictionType int

const (
	// EvictionLRU represents Least Recently Used strategy
	EvictionLRU EvictionType = iota
	// EvictionLFU represents Least Frequently Used strategy
	EvictionLFU
	// EvictionFIFO represents First In First Out strategy
	EvictionFIFO
	// EvictionARC represents Adaptive Replacement Cache strategy
	EvictionARC
)

// String returns the string representation of EvictionType
func (e EvictionType) String() string {
	switch e {
	case EvictionLRU:
		return "lru"
	case EvictionLFU:
		return "lfu"
	case EvictionFIFO:
		return "fifo"
	case EvictionARC:
		return "arc"
	default:
		return "unknown"
	}
}

// StringToEvictionType converts a string to EvictionType
func StringToEvictionType(s string) (EvictionType, error) {
	switch strings.ToLower(s) {
	case "lru":
		return EvictionLRU, nil
	case "lfu":
		return EvictionLFU, nil
	case "fifo":
		return EvictionFIFO, nil
	case "arc":
		return EvictionARC, nil
	default:
		return EvictionLRU, fmt.Errorf("invalid eviction type: %s", s)
	}
}

// IsValid checks if the EvictionType is valid
func (e EvictionType) IsValid() bool {
	return e >= EvictionLRU && e <= EvictionARC
}

// Value represents a value that can be stored in the cache.
// It must provide its size through the Len method.
type Value interface {
	// Len returns the size of the value in bytes.
	Len() int
}

// CacheStrategy defines the interface that all cache eviction strategies must implement.
type CacheStrategy interface {
	// Get retrieves a value from the cache.
	// Returns the value, its last update time, and whether it was found.
	Get(key string) (value Value, updateTime time.Time, found bool)

	// Put adds or updates a value in the cache.
	// If adding the value would exceed the cache's size limit,
	// one or more entries will be evicted according to the strategy.
	Put(key string, value Value)

	// CleanUp removes expired entries from the cache.
	// An entry is considered expired if its last update time plus ttl
	// is before the current time.
	CleanUp(ttl time.Duration)

	// Len returns the number of items in the cache.
	Len() int
}

// Entry represents a cache entry with its metadata.
type Entry struct {
	Key      string    // The key used to identify the entry
	Value    Value     // The stored value
	UpdateAt time.Time // Last time the entry was accessed or modified
}

// Expired checks if the entry has expired based on the given duration.
func (e *Entry) Expired(duration time.Duration) bool {
	if e.UpdateAt.IsZero() {
		return false // Never expires if update time is not set
	}
	return e.UpdateAt.Add(duration).Before(time.Now())
}

// Touch updates the entry's last access time to now.
func (e *Entry) Touch() {
	e.UpdateAt = time.Now()
}

// CacheConfig represents the configuration for a cache
type CacheConfig struct {
	MaxBytes        int64         `json:"max_bytes"`
	EvictionType    EvictionType  `json:"eviction_type"`
	CleanupInterval time.Duration `json:"cleanup_interval"`
}

// New creates a new cache with the specified eviction strategy.
// Returns nil and an error if the strategy name is invalid.
func New(name string, maxBytes int64, onEvicted func(string, Value)) (CacheStrategy, error) {
	evictionType, err := StringToEvictionType(name)
	if err != nil {
		return nil, err
	}
	switch evictionType {
	case EvictionLRU:
		return NewCacheUseLRU(maxBytes, onEvicted), nil
	case EvictionLFU:
		return NewCacheUseLFU(maxBytes, onEvicted), nil
	case EvictionFIFO:
		return NewCacheUseFIFO(maxBytes, onEvicted), nil
	case EvictionARC:
		return NewCacheUseARC(maxBytes, onEvicted), nil
	default:
		return nil, fmt.Errorf("unsupported cache strategy: %q", name)
	}
}

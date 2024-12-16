// Package cache implements a distributed cache system with various features.
package cache

import (
	"sync"
	"time"

	"github.com/1055373165/ggcache/utils/logger"
)

// Call represents an in-flight or completed request to call a function.
type Call struct {
	wg    sync.WaitGroup // Used to wait for the function call to complete
	value interface{}    // The value returned by the function
	err   error         // Any error returned by the function
}

// cachedValue represents a cached result with an expiration time.
type cachedValue struct {
	value   interface{} // The cached value
	expires time.Time   // When this value expires
}

// SingleFlight manages function calls to prevent duplicate simultaneous calls.
// It includes caching of results with TTL support.
type SingleFlight struct {
	mu      sync.RWMutex                // Guards m and cache
	m       map[string]*Call            // Keyed by function key
	cache   map[string]*cachedValue     // Cache of function results
	ttl     time.Duration              // How long to cache results
	ticker  *time.Ticker               // For cache cleanup
}

// NewSingleFlight creates a new SingleFlight instance with the specified TTL.
// It starts a background goroutine to clean expired cache entries.
func NewSingleFlight(ttl time.Duration) *SingleFlight {
	sf := &SingleFlight{
		m:     make(map[string]*Call),
		cache: make(map[string]*cachedValue),
		ttl:   ttl,
	}
	sf.ticker = time.NewTicker(ttl)
	go sf.cacheCleaner()
	return sf
}

// Do executes the given function if there's no in-flight execution.
// If there is an in-flight execution, it waits for and returns that result.
// Results are cached for the configured TTL.
//
// Parameters:
//   - key: Unique identifier for this function call
//   - fn: The function to execute
//
// Returns the function's result and any error encountered.
func (sf *SingleFlight) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	sf.mu.RLock()

	// Check cache first
	if cv, ok := sf.cache[key]; ok && time.Now().Before(cv.expires) {
		sf.mu.RUnlock()
		return cv.value, nil
	}

	// Check for in-flight calls
	if c, ok := sf.m[key]; ok {
		sf.mu.RUnlock()
		logger.LogrusObj.Warnf("%s is already being fetched, waiting for result", key)
		c.wg.Wait()
		return c.value, c.err
	}

	sf.mu.RUnlock()

	// Create new call
	c := new(Call)
	c.wg.Add(1)

	sf.mu.Lock()
	sf.m[key] = c
	sf.mu.Unlock()

	// Execute function
	c.value, c.err = fn()
	c.wg.Done()

	// Update cache and cleanup
	sf.mu.Lock()
	delete(sf.m, key)
	if c.err == nil {
		sf.cache[key] = &cachedValue{
			value:   c.value,
			expires: time.Now().Add(sf.ttl),
		}
	}
	sf.mu.Unlock()

	return c.value, c.err
}

// cacheCleaner runs periodically to remove expired entries from the cache.
func (sf *SingleFlight) cacheCleaner() {
	for range sf.ticker.C {
		sf.mu.Lock()
		for key, cv := range sf.cache {
			if time.Now().After(cv.expires) {
				delete(sf.cache, key)
			}
		}
		sf.mu.Unlock()
	}
}

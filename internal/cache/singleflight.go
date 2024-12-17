// Package cache implements a distributed cache system with various features.
package cache

import (
	"sync"
	"time"

	"github.com/1055373165/ggcache/pkg/common/logger"
)

// Call represents an in-flight or completed request to call a function.
// It contains the value and error returned by the function, as well as a WaitGroup to wait for the function call to complete.
type Call struct {
	wg    sync.WaitGroup // Guards value and err
	value interface{}    // The value returned by the function
	err   error          // Any error that occurred
}

// cachedValue represents a cached result with an expiration time.
type cachedValue struct {
	value   interface{} // The cached value
	expires time.Time   // When this value expires
}

// SingleFlight manages function calls to prevent duplicate simultaneous calls.
// It ensures that only one execution of the same function with the same key happens at a time, sharing the result with all callers.
// It also includes caching of results with TTL support.
type SingleFlight struct {
	mu     sync.RWMutex            // Protects m and cache
	m      map[string]*Call        // Keyed by function key
	cache  map[string]*cachedValue // Cache of function results
	ttl    time.Duration           // How long to cache results
	ticker *time.Ticker            // For cache cleanup
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

// Do executes the given function if it's not already being executed for the given key.
// If a duplicate call is made, the duplicate caller waits for the original to complete and receives the same results.
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

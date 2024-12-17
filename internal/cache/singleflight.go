// Package cache implements a distributed cache system with various features.
package cache

import (
	"context"
	"sync"
	"time"
)

// Result represents the result of a function call.
type Result struct {
	Value interface{}
	Err   error
}

// call represents an in-flight or completed function call.
type call struct {
	done chan struct{} // Signals when the call is complete
	res  Result        // The result of the call
}

// cacheEntry represents a cached result with expiration.
type cacheEntry struct {
	result  Result
	expires time.Time
}

// FlightGroup manages function calls to prevent duplicate simultaneous calls.
// It ensures that only one execution of the same function with the same key
// happens at a time, sharing the result with all callers.
type FlightGroup struct {
	mu      sync.RWMutex
	calls   map[string]*call      // Active calls
	cache   map[string]cacheEntry // Results cache
	ttl     time.Duration         // Cache TTL
	cleanup *time.Ticker          // Cleanup ticker
	done    chan struct{}         // Signals shutdown
}

// NewGroup creates a new Group with the specified cache TTL.
// The cleanup interval is set to 1/4 of the TTL duration.
func NewFlightGroup(ttl time.Duration) *FlightGroup {
	if ttl <= 0 {
		ttl = time.Minute // Default TTL
	}

	g := &FlightGroup{
		calls:   make(map[string]*call),
		cache:   make(map[string]cacheEntry),
		ttl:     ttl,
		cleanup: time.NewTicker(ttl / 4),
		done:    make(chan struct{}),
	}

	go g.cleanupLoop()
	return g
}

// Do executes the given function if it's not already being executed.
// If there's a duplicate call, the caller waits for the original to complete.
// Results are cached according to the TTL.
func (g *FlightGroup) Do(ctx context.Context, key string, fn func() (interface{}, error)) (interface{}, error) {
	// Check cache first (read lock)
	if value, ok := g.checkCache(key); ok {
		return value.Value, value.Err
	}

	// Get or create call (write lock)
	c, created := g.createCall(key)
	if !created {
		return g.waitForCall(ctx, c)
	}

	// Execute function and store result
	defer g.finishCall(key, c)

	return g.executeAndCache(ctx, key, fn, c)
}

func (g *FlightGroup) checkCache(key string) (Result, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if entry, ok := g.cache[key]; ok && time.Now().Before(entry.expires) {
		return entry.result, true
	}
	return Result{}, false
}

func (g *FlightGroup) createCall(key string) (*call, bool) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if c, ok := g.calls[key]; ok {
		return c, false
	}

	c := &call{done: make(chan struct{})}
	g.calls[key] = c
	return c, true
}

func (g *FlightGroup) waitForCall(ctx context.Context, c *call) (interface{}, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.done:
		return c.res.Value, c.res.Err
	}
}

func (g *FlightGroup) executeAndCache(ctx context.Context, key string, fn func() (interface{}, error), c *call) (interface{}, error) {
	// Run the function
	value, err := g.executeWithContext(ctx, fn)
	result := Result{Value: value, Err: err}
	c.res = result

	// Cache successful results
	if err == nil {
		g.mu.Lock()
		g.cache[key] = cacheEntry{
			result:  result,
			expires: time.Now().Add(g.ttl),
		}
		g.mu.Unlock()
	}

	return value, err
}

func (g *FlightGroup) executeWithContext(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	var (
		value interface{}
		err   error
		done  = make(chan struct{})
	)

	go func() {
		value, err = fn()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-done:
		return value, err
	}
}

func (g *FlightGroup) finishCall(key string, c *call) {
	g.mu.Lock()
	delete(g.calls, key)
	g.mu.Unlock()
	close(c.done)
}

// cleanupLoop periodically removes expired cache entries.
func (g *FlightGroup) cleanupLoop() {
	for {
		select {
		case <-g.done:
			return
		case <-g.cleanup.C:
			g.removeExpiredEntries()
		}
	}
}

func (g *FlightGroup) removeExpiredEntries() {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now()
	for key, entry := range g.cache {
		if now.After(entry.expires) {
			delete(g.cache, key)
		}
	}
}

// Stop gracefully shuts down the Group and stops the cleanup goroutine.
func (g *FlightGroup) Stop() {
	g.cleanup.Stop()
	close(g.done)

	g.mu.Lock()
	defer g.mu.Unlock()

	// Clear all data
	g.calls = make(map[string]*call)
	g.cache = make(map[string]cacheEntry)
}

// ForceEvict removes an entry from the cache regardless of its expiration.
func (g *FlightGroup) ForceEvict(key string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.cache, key)
}

// Stats returns current statistics about the FlightGroup.
func (g *FlightGroup) Stats() map[string]interface{} {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return map[string]interface{}{
		"active_calls":    len(g.calls),
		"cached_entries":  len(g.cache),
		"ttl_seconds":     g.ttl.Seconds(),
		"cleanup_seconds": (g.ttl / 4).Seconds(),
	}
}

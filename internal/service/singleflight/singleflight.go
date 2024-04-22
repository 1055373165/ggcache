package singleflight

import (
	"sync"
)

type Call struct {
	wg    sync.WaitGroup
	value interface{}
	err   error
}

type SingleFlight struct {
	mu sync.Mutex
	m  map[string]*Call
}

/*
Use Single Flight to re-encapsulate the query when the Group cache misses.
During concurrent requests, only one request will call the query in the form of goroutine,
and all other requests during concurrent queries will be blocked and waiting.
*/
func (sf *SingleFlight) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	sf.mu.Lock()

	if sf.m == nil {
		sf.m = make(map[string]*Call)
	}

	if c, ok := sf.m[key]; ok {
		// the key for concurrent requests is set
		sf.mu.Unlock()
		// blocking waiting for query flight goroutine to return
		c.wg.Wait()
		// the result value has been stored in the Call structure.
		return c.value, c.err
	}

	c := new(Call)
	sf.m[key] = c
	c.wg.Add(1)
	sf.mu.Unlock()

	// open a query, c.value and c.err receive the return value.
	c.value, c.err = fn()
	c.wg.Done()

	// to ensure that we can always get a newer value
	sf.mu.Lock()
	delete(sf.m, key)
	sf.mu.Unlock()

	return c.value, c.err
}

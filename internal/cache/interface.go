package cache

// Picker is the interface that must be implemented to locate peers.
// It uses consistent hashing to determine which node should handle a specific key.
type Picker interface {
	// Pick returns the fetcher for the peer that should handle the given key.
	// If the key should be handled by the current node, returns (nil, false).
	Pick(key string) (Fetcher, bool)
}

// Fetcher is the interface that wraps the basic Fetch method.
// Each distributed node must implement this interface to support peer-to-peer cache retrieval.
type Fetcher interface {
	// Fetch retrieves the value for key from the specified group's cache.
	// Returns the value as bytes and any error encountered.
	Fetch(group string, key string) ([]byte, error)
}

// Retriever is the interface that wraps the basic retrieve method.
// It provides the ability to fetch data from a backing store when cache misses occur.
type Retriever interface {
	// retrieve fetches data for the given key from the backing store.
	retrieve(key string) ([]byte, error)
}

// RetrieveFunc is an adapter to allow the use of ordinary functions as Retrievers.
// This is a common pattern in Go that allows simple functions to satisfy an interface.
type RetrieveFunc func(key string) ([]byte, error)

// retrieve calls f(key), implementing the Retriever interface.
func (f RetrieveFunc) retrieve(key string) ([]byte, error) {
	return f(key)
}

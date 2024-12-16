// Package cache implements a distributed cache system with various features.
package cache

// Picker is the interface that must be implemented to locate
// the peer that owns a specific key.
type Picker interface {
	// PickPeer returns the address of the node that should handle the given key.
	PickPeer(key string) (peer Fetcher, ok bool)
}

// Fetcher is the interface that must be implemented to fetch data from a peer.
type Fetcher interface {
	// Fetch retrieves the value for the given key from the specified group.
	// Returns the value as bytes and any error encountered.
	Fetch(group string, key string) ([]byte, error)
}

// Retriever is the interface that wraps the basic Retrieve method.
type Retriever interface {
	// Retrieve fetches the value for the given key from the backend store.
	Retrieve(key string) ([]byte, error)
}

// RetrieveFunc is an adapter to allow the use of ordinary functions as Retrievers.
// If f is a function with the appropriate signature, RetrieveFunc(f) is a
// Retriever that calls f.
type RetrieveFunc func(key string) ([]byte, error)

// Retrieve calls f(key).
func (f RetrieveFunc) Retrieve(key string) ([]byte, error) {
	return f(key)
}

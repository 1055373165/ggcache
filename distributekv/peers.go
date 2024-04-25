package distributekv

// Picker defines the ability to obtain distributed nodes
type Picker interface {
	Pick(key string) (Fetcher, bool)
}

// Fetcher defines the ability to obtain cache from the remote end, so each Peer should implement this interface
type Fetcher interface {
	Fetch(group string, key string) ([]byte, error)
}

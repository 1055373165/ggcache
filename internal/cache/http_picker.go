package cache

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/1055373165/ggcache/pkg/common/logger"
)

// Verify HTTPPool implements Picker interface at compile time
var _ Picker = (*HTTPPool)(nil)

const (
	defaultBasePath = "/_ggcache/"
	apiServerAddr   = "127.0.0.1:9999"
)

// HTTPPool implements an HTTP-based peer picker for distributed caching.
// It uses a consistent hash ring to distribute cache keys across peers.
type HTTPPool struct {
	address      string                  // Server's base URL (e.g., "http://example.net:8000")
	basePath     string                  // URL prefix for cache endpoints
	mu           sync.Mutex              // Guards peers and httpFetchers
	peers        *ConsistentMap          // Consistent hash map for peer selection
	httpFetchers map[string]*httpFetcher // Maps peer URLs to their HTTP clients
}

// NewHTTPPool creates a new HTTP-based peer pool with the given server address.
func NewHTTPPool(address string) *HTTPPool {
	return &HTTPPool{
		address:  address,
		basePath: defaultBasePath,
	}
}

// Log formats and outputs a log message with the server's address.
func (p *HTTPPool) Log(format string, v ...interface{}) {
	logger.LogrusObj.Infof("[Server %s] %s", p.address, fmt.Sprintf(format, v...))
}

// ServeHTTP handles HTTP requests for cache operations.
// The URL format is: /<basepath>/<groupname>/<key>
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		http.Error(w, "invalid cache endpoint", http.StatusBadRequest)
		return
	}

	p.Log("%s %s", r.Method, r.URL.Path)

	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request format, expected /<basepath>/<groupname>/<key>", http.StatusBadRequest)
		return
	}

	groupName, key := parts[0], parts[1]
	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.Bytes())
}

// Pick implements the Picker interface.
// It selects a peer based on the given key and returns the corresponding HTTP client.
func (p *HTTPPool) Pick(key string) (Fetcher, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	peerAddress := p.peers.GetNode(key)
	if peerAddress == p.address {
		// upper layer get the value of the key locally after receiving false
		return nil, false
	}

	logger.LogrusObj.Infof("[dispatcher peer %s] pick remote peer: %s", apiServerAddr, peerAddress)
	return p.httpFetchers[peerAddress], true
}

// UpdatePeers updates the peer list and rebuilds the consistent hash ring.
func (p *HTTPPool) UpdatePeers(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.peers = NewConsistentHash(defaultReplicas, nil)
	p.peers.AddNodes(peers...)
	p.httpFetchers = make(map[string]*httpFetcher, len(peers))

	for _, peer := range peers {
		p.httpFetchers[peer] = &httpFetcher{
			baseURL: peer + p.basePath, // such "http://10.0.0.1:9999/_ggcache/"
		}
	}
}

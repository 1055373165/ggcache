// Package cache implements a distributed cache system with various features.
package cache

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/1055373165/ggcache/internal/service/consistenthash"
	"github.com/1055373165/ggcache/utils/logger"
)

// Ensure HTTPPool implements Picker interface
var _ Picker = (*HTTPPool)(nil)

const (
	defaultBasePath = "/_ggcache/" // Default URL prefix for cache endpoints
	apiServerAddr   = "127.0.0.1:9999"
	defaultReplicas = 50
)

// Picker is the interface that must be implemented to locate
// the peer that owns a specific key.
type Picker interface {
	PickPeer(key string) (Fetcher, bool)
}

// Fetcher is the interface that must be implemented to fetch data from a peer.
type Fetcher interface {
	Fetch(group string, key string) ([]byte, error)
}

// HTTPPool implements PeerPicker for a pool of HTTP peers.
type HTTPPool struct {
	self     string                 // This peer's base URL (e.g., "https://example.net:8000")
	basePath string                 // URL path prefix for cache endpoints
	mu       sync.Mutex            // Guards peers and httpGetters
	peers    *consistenthash.Map   // Consistent hash map for peer selection
	fetchers map[string]*httpClient // Keyed by peer URL (e.g., "http://10.0.0.2:8008")
}

// NewHTTPPool initializes an HTTP pool of peers with the local base URL.
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
		fetchers: make(map[string]*httpClient),
	}
}

// Log formats and prints a message for debugging.
func (p *HTTPPool) Log(format string, v ...interface{}) {
	logger.LogrusObj.Infof(format, v...)
}

// ServeHTTP handles HTTP requests for cache operations.
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}

	p.Log("%s %s", r.Method, r.URL.Path)

	// Parse request path: /<basepath>/<groupname>/<key>
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request format, expected /<basepath>/<groupname>/<key>", http.StatusBadRequest)
		return
	}

	groupName, key := parts[0], parts[1]
	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusBadRequest)
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

// UpdatePeers updates the pool's list of peers.
// It creates a consistent hash ring of peers and initializes fetchers for each.
func (p *HTTPPool) UpdatePeers(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.fetchers = make(map[string]*httpClient, len(peers))
	for _, peer := range peers {
		p.fetchers[peer] = &httpClient{baseURL: peer + p.basePath}
	}
}

// PickPeer picks a peer according to key.
// It returns (nil, false) if no peer was picked.
func (p *HTTPPool) PickPeer(key string) (Fetcher, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.fetchers[peer], true
	}
	return nil, false
}

// httpClient is a client to fetch data from other nodes.
type httpClient struct {
	baseURL string // base URL of remote node
}

// Fetch gets data from remote peer via HTTP.
func (h *httpClient) Fetch(group string, key string) ([]byte, error) {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(group),
		url.QueryEscape(key),
	)
	
	res, err := http.Get(u)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from %s: %v", u, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}

	return bytes, nil
}

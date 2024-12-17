package cache

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/1055373165/ggcache/pkg/common/logger"
)

var _ Picker = (*HTTPPool)(nil)

const (
	defaultBasePath = "/_ggcache/"
	apiServerAddr   = "127.0.0.1:9999"
)

// HTTPPool implements an HTTP-based peer picker for distributed caching.
// It uses a consistent hash ring to distribute cache keys across peers.
type HTTPPool struct {
	currentServer string
	basePath      string
	peerSelector  *ConsistentMap
	fetcherMap    map[string]*httpFetcher
	mu            sync.Mutex
}

// NewHTTPPool creates a new HTTP-based peer pool with the given server address.
func NewHTTPPool(srvAddr string) *HTTPPool {
	return &HTTPPool{
		currentServer: srvAddr,
		basePath:      defaultBasePath,
	}
}

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

func (p *HTTPPool) Log(format string, v ...interface{}) {
	logger.LogrusObj.Infof("[Server %s] %s", p.currentServer, fmt.Sprintf(format, v...))
}

// Pick implements the Picker interface.
// It selects a peer based on the given key and returns the corresponding HTTP client.
func (p *HTTPPool) Pick(key string) (Fetcher, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	peerAddress := p.peerSelector.GetNode(key)
	if peerAddress == p.currentServer {
		// upper layer get the value of the key locally after receiving false
		return nil, false
	}

	logger.LogrusObj.Infof("[request forward by peer %s], pick remote peer: %s", p.currentServer, peerAddress)

	return p.fetcherMap[peerAddress], true
}

// UpdatePeers updates the peer list and rebuilds the consistent hash ring.
func (p *HTTPPool) UpdatePeers(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.peerSelector = NewConsistentHash(defaultReplicas, nil)
	p.peerSelector.AddNodes(peers...)
	p.fetcherMap = make(map[string]*httpFetcher, len(peers))

	for _, peer := range peers {
		p.fetcherMap[peer] = &httpFetcher{
			baseURL: peer + p.basePath, // such "http://10.0.0.1:9999/_ggcache/"
		}
	}
}

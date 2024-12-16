// Package cache implements a distributed cache system with various features.
package cache

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/1055373165/ggcache/utils/logger"
)

const (
	defaultBasePath = "/_geecache/"
	defaultReplicas = 50
)

// HTTPPool implements PeerPicker for a pool of HTTP peers.
type HTTPPool struct {
	self     string     // this peer's base URL (e.g. "https://example.net:8000")
	basePath string     // e.g. "/_geecache/"
	peers    *Map       // keyed by e.g. "http://10.0.0.2:8008"
	clients  map[string]*httpClient
}

// NewHTTPPool initializes an HTTP pool of peers with the given self address.
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
		clients:  make(map[string]*httpClient),
	}
}

// Log formats log message with server name.
func (p *HTTPPool) Log(format string, v ...interface{}) {
	logger.LogrusObj.Infof("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// ServeHTTP handles all http requests.
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

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
	w.Write(view.ByteSlice())
}

// Set updates the pool's list of peers.
func (p *HTTPPool) Set(peers ...string) {
	p.peers = New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.clients = make(map[string]*httpClient, len(peers))
	for _, peer := range peers {
		p.clients[peer] = &httpClient{baseURL: peer + p.basePath}
	}
}

// PickPeer picks a peer according to key.
func (p *HTTPPool) PickPeer(key string) (Fetcher, bool) {
	peerAddr := p.peers.Get(key)
	if peerAddr == "" || peerAddr == p.self {
		return nil, false
	}
	p.Log("Pick peer %s", peerAddr)
	return p.clients[peerAddr], true
}

// StartHTTPServer starts an HTTP server at addr.
func StartHTTPServer(addr string, addrs []string, gm map[string]*Group) {
	peers := NewHTTPPool(addr)
	peers.Set(addrs...)
	
	for _, g := range gm {
		g.RegisterPeers(peers)
	}
	
	logger.LogrusObj.Infof("geecache is running at %s", addr)
	logger.LogrusObj.Fatal(http.ListenAndServe(addr[7:], peers))
}

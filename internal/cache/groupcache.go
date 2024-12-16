package cache

import (
	"errors"
	"sync"

	"github.com/1055373165/ggcache/pkg/common/logger"
)

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// Group represents a cache namespace and associated data loading spread over.
type Group struct {
	name      string
	getter    Getter
	mainCache *Cache
	peers     Picker
}

// NewGroup creates a new instance of Group.
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		logger.LogrusObj.Panic("nil Getter")
	}

	mu.Lock()
	defer mu.Unlock()

	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: NewCache("lru", cacheBytes),
	}
	groups[name] = g
	return g
}

// GetGroup returns the named group previously created with NewGroup, or
// nil if there's no such group.
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// Get value for a key from cache.
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, errors.New("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		logger.LogrusObj.Infof("[GeeCache] hit")
		return v, nil
	}

	return g.load(key)
}

// RegisterPeers registers a Picker for choosing remote peer.
func (g *Group) RegisterPeers(peers Picker) {
	if g.peers != nil {
		logger.LogrusObj.Panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

// load loads key either by invoking the getter locally
// or by sending it to other peers.
func (g *Group) load(key string) (value ByteView, err error) {
	if g.peers != nil {
		if peer, ok := g.peers.PickPeer(key); ok {
			if value, err = g.getFromPeer(peer, key); err == nil {
				return value, nil
			}
			logger.LogrusObj.Warnf("[GeeCache] Failed to get from peer: %v", err)
		}
	}

	return g.getLocally(key)
}

// getFromPeer gets value from peer cache.
func (g *Group) getFromPeer(peer Fetcher, key string) (ByteView, error) {
	bytes, err := peer.Fetch(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}

// getLocally calls the getter to get value and adds it to cache.
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	value := ByteView{b: bytes}
	g.populateCache(key, value)
	return value, nil
}

// populateCache adds value to cache.
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

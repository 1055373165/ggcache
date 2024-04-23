package service

import (
	"errors"
	"ggcache/internal/middleware/logger"
	"ggcache/internal/service/singleflight"
	"sync"
)

const (
	apiServer  = "http://127.0.0.1:9999"
	bindServer = "http://127.0.0.1:8001"
)

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

type Group struct {
	name      string
	cache     *cache
	retriever Retriever
	locator   Picker
	flight    *singleflight.SingleFlight
}

func NewGroup(name string, strategy string, maxBytes int64, retriever Retriever) *Group {
	if retriever == nil {
		panic("backend database retrieve ability must be provided, otherwise, the cache system has no meaning.")
	}

	if _, ok := groups[name]; ok {
		return groups[name]
	}

	g := &Group{
		name:      name,
		cache:     newCache(strategy, maxBytes),
		retriever: retriever,
		flight:    &singleflight.SingleFlight{},
	}

	mu.Lock()
	groups[name] = g
	mu.Unlock()

	return g
}

// the new Group has not been populated with Picker, so you can specify to register node locator for Group.
func (g *Group) RegisterPickerForGroup(p Picker) {
	if g.locator != nil {
		panic("group has been registered node locator")
	}
	g.locator = p
}

// group cache core operation 1 : Get
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, errors.New("to query the key value, you must provide the key")
	}

	// cache hit
	if value, ok := g.cache.get(key); ok {
		return value, nil
	}

	// cache missed, try retrieve value from backend database
	return g.load(key)
}

func (g *Group) getFromPeer(peer Fetcher, key string) (ByteView, error) {
	bytes, err := peer.Fetch(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}

func (g *Group) load(key string) (ByteView, error) {
	// singleflight: concurrent access control
	// similar to decorator pattern

	/*
		1. flight use sync.Once underlying to ensure the same key request will only be sent once.
		2. if this group have Picker(distributed kv node locator), ask which node it should be sent to(consistenthash algo.).
		3. call fetcher's Fetch method to query key's value and return bytes deep copy.
	*/
	view, err := g.flight.Do(key, func() (interface{}, error) {
		if g.locator != nil {
			if fetcher, ok := g.locator.Pick(key); ok {
				// send request
				logger.Logger.Infof("send request to fetcher %v", fetcher)
				bytes, err := fetcher.Fetch(g.name, key)
				if err == nil { // success path(cache hit)
					return ByteView{b: cloneBytes(bytes)}, nil
				}
			}
		}

		// query database
		return g.backSource(key)
	})

	// query success
	if err == nil {
		return view.(ByteView), nil
	}

	return ByteView{}, err
}

/*
1. because the cache misses, return to the database to query the value of key
2. populate the cache with the latest queried values to avoid going back to the database next time
*/
func (g *Group) backSource(key string) (ByteView, error) {
	bytes, err := g.retriever.retrieve(key)
	if err != nil {
		return ByteView{}, err
	}

	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.cache.put(key, value)
}

func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	return groups[name]
}

// todo
func DestroyGroup(name string) {
	g := GetGroup(name)
	if g == nil {
		return
	}
}

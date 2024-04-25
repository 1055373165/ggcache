package distributekv

import (
	"errors"

	"github.com/1055373165/Distributed_KV_Store/logger"

	"sync"

	"github.com/1055373165/Distributed_KV_Store/distributekv/singleflight"
)

// The Groupcache module provides a higher level of abstraction than cache and implements the ability to fill caches and name partition caches.
var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// Retriever requires the object to implement the ability to obtain data from the data source
type Retriever interface {
	retrieve(string) ([]byte, error)
}

type RetrieveFunc func(key string) ([]byte, error)

/*
RetrieveFunc implements the retrieve method, that is, implements the Retriver interface so that any anonymous function func through RetrieverFunc ( func ) forced type conversion, the ability to achieve the Retriver interface.
This is also reflected in the gin framework inside the HandlerFunc type encapsulation anonymous function, http type handler forced conversion can be directly used as gin Handler.
*/
func (f RetrieveFunc) retrieve(key string) ([]byte, error) {
	return f(key)
}

// Group provides the ability to name and manage caches and fill caches
type Group struct {
	name      string
	cache     *cache
	retriever Retriever
	server    Picker
	flight    *singleflight.SingleFlight
}

// NewGroup Creates a new cache space.
func NewGroup(name string, maxBytes int64, retriever Retriever) *Group {
	if retriever == nil {
		panic("Group Retriver must be existed!")
	}

	g := &Group{
		name:      name,
		cache:     newCache(maxBytes),
		retriever: retriever,
		flight:    &singleflight.SingleFlight{},
	}

	mu.Lock()
	groups[name] = g
	mu.Unlock()

	return g
}

// RegisterServer registers server Picker for Group
func (g *Group) RegisterServer(p Picker) {
	if g.server != nil {
		panic("group had been registed server")
	}
	g.server = p
}

// GetGroup to get the Group object of the corresponding namespace (to manage the actual cache)
func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	return groups[name]
}

func DestroryGroup(name string) {
	g := GetGroup(name)
	if g != nil {
		svr := g.server.(*Server)
		svr.Stop()

		delete(groups, name)
		logger.Logger.Info("Destrory cache [%s %s]", name, svr.Addr)
	}
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, errors.New("key must be existed")
	}

	if value, ok := g.cache.get(key); ok {
		logger.Logger.Infof("Group %s cache hit, key %s...", g.name, key)
		return value, nil
	}

	// cache missing
	return g.load(key)
}

func (g *Group) load(key string) (ByteView, error) {
	// singleFlight
	view, err := g.flight.Do(key, func() (interface{}, error) {
		if g.server != nil {
			if fetcher, ok := g.server.Pick(key); ok {
				bytes, err := fetcher.Fetch(g.name, key)
				if err == nil {
					return ByteView{b: cloneBytes(bytes)}, nil
				}
				logger.Logger.Info("fetch key %s failed, error: %s\n", fetcher, err.Error())
			}
		}
		return g.getLocally(key)
	})

	if err == nil {
		return view.(ByteView), nil
	}

	return ByteView{}, err
}

// GetLocally retrieves data from Retriever and populates it into the cache
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.retriever.retrieve(key)
	if err != nil {
		return ByteView{}, err
	}

	value := ByteView{b: cloneBytes(bytes)}

	g.populateCache(key, value)

	return value, nil
}

// Populate Cache populates the cache with data queried from the underlying database
func (g *Group) populateCache(key string, value ByteView) {
	g.cache.put(key, value)
}

package internal

import (
	"errors"
	"time"

	"github.com/1055373165/ggcache/utils/logger"
	"gorm.io/gorm"

	"sync"
)

// The Groupcache module provides a higher level of abstraction than cache and implements the ability to fill caches and name partition caches.
var (
	mu           sync.RWMutex
	GroupManager = make(map[string]*Group)
)

// Group provides the ability to name and manage caches and fill caches
type Group struct {
	name      string
	cache     *cache
	retriever Retriever
	server    Picker
	flight    *SingleFlight
}

// NewGroup Creates a new cache space.
func NewGroup(name string, strategy string, maxBytes int64, retriever Retriever) *Group {
	if retriever == nil {
		panic("Group Retriver must be existed!")
	}

	if _, ok := GroupManager[name]; ok {
		return GroupManager[name]
	}

	g := &Group{
		name:      name,
		cache:     newCache(strategy, maxBytes),
		retriever: retriever,
		flight:    NewSingleFlight(10 * time.Second),
	}

	mu.Lock()
	GroupManager[name] = g
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
	return GroupManager[name]
}

func DestroryGroup(name string) {
	g := GetGroup(name)
	if g != nil {
		svr := g.server.(*Server)
		svr.Stop()

		delete(GroupManager, name)
	}
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, errors.New("key must be existed")
	}

	if value, ok := g.cache.get(key); ok {
		logger.LogrusObj.Infof("Group %s 缓存命中..., key %s...", g.name, key)
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
				logger.LogrusObj.Warnf("fetch key %s failed, error: %s\n", fetcher, err.Error())
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.LogrusObj.Warnf("对于不存在的 key, 为了防止缓存穿透, 先存入缓存中并设置合理过期时间")
			g.cache.put(key, ByteView{})
		}
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

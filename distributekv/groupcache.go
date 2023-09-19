package distributekv

import (
	"errors"

	"github.com/1055373165/groupcache/middleware/logger"

	"sync"

	"github.com/1055373165/groupcache/distributekv/singleflight"
)

// groupcache 模块提供比 cache 更高一层的抽象能力
// 实现了填充缓存、命名划分缓存的能力
var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// Retriever 要求对象实现从数据源获取数据的能力
type Retriever interface {
	retrieve(string) ([]byte, error)
}

type RetrieveFunc func(key string) ([]byte, error)

// RetrieveFunc 实现了 retrieve 方法，即实现了 Retriver 接口
// 使得任意匿名函数 func 通过 RetrieverFunc(func) 强制类型转换后，实现了 Retriver 接口的能力
// 这个在 gin 框架里面的 HandlerFunc 类型封装匿名函数时也有所体现，http 类型的 handler 强制转换后直接可以作为 gin 的 Handler 使用
func (f RetrieveFunc) retrieve(key string) ([]byte, error) {
	return f(key)
}

// Group 提供了命名管理缓存、填充缓存的能力
type Group struct {
	name      string
	cache     *cache
	retriever Retriever
	server    Picker
	flight    *singleflight.SingleFlight
}

// NewGroup 新创建一个缓存空间
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

// RegisterServer 为 Group 注册 server
func (g *Group) RegisterServer(p Picker) {
	if g.server != nil {
		panic("group had been registed server")
	}
	g.server = p
}

// GetGroup 获取对应命名空间的 Group 对象（对实际缓存进行管理）
func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	return groups[name]
}

func DestroryGroup(name string) {
	g := GetGroup(name)
	if g != nil {
		svr := g.server.(*Server)
		// 停止服务
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
		logger.Logger.Info("cache hit...")
		return value, nil
	}

	// cache missing, get it another way
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
		// 如果目前只有单节点，那么从本地数据库查询
		return g.getLocally(key)
	})

	if err == nil {
		return view.(ByteView), nil
	}
	return ByteView{}, err
}

// getLocally 向 Retriever 取回数据并填充至缓存中
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.retriever.retrieve(key)
	if err != nil {
		return ByteView{}, err
	}

	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

// populateCache 将从底层数据库中查询到的数据填充到缓存中
func (g *Group) populateCache(key string, value ByteView) {
	g.cache.put(key, value)
}

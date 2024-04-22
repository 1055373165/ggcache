package cachepurge

import (
	"ggcache/internal/service/cachepurge/fifo"
	"ggcache/internal/service/cachepurge/interfaces"
	"ggcache/internal/service/cachepurge/lfu"
	"ggcache/internal/service/cachepurge/lru"
	"strings"
)

func New(name string, maxBytes int64, onEvicted func(string, interfaces.Value)) interfaces.CacheStrategy {
	name = strings.ToLower(name)
	switch name {
	case "fifo":
		return fifo.NewFIFOCache(maxBytes, onEvicted)
	case "lru":
		return lru.NewLRUCache(maxBytes, onEvicted)
	case "lfu":
		return lfu.NewLFUCache(maxBytes, onEvicted)
	}
	return nil
}

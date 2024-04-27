package cachepurge

import (
	"strings"

	"github.com/1055373165/ggcache/internal/service/cachepurge/fifo"

	"github.com/1055373165/ggcache/internal/service/cachepurge/interfaces"

	"github.com/1055373165/ggcache/internal/service/cachepurge/lfu"

	"github.com/1055373165/ggcache/internal/service/cachepurge/lru"
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

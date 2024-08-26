package eviction

import (
	"strings"

	"github.com/1055373165/ggcache/internal/service/eviction/fifo"

	"github.com/1055373165/ggcache/internal/service/eviction/lfu"

	"github.com/1055373165/ggcache/internal/service/eviction/lru"
	"github.com/1055373165/ggcache/internal/service/eviction/strategy"
)

func New(name string, maxBytes int64, onEvicted func(string, strategy.Value)) strategy.EvictionStrategy {
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

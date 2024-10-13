package eviction

import (
	"strings"
)

func New(name string, maxBytes int64, onEvicted func(string, Value)) EvictionStrategy {
	name = strings.ToLower(name)
	switch name {
	case "fifo":
		return NewFIFOCache(maxBytes, onEvicted)
	case "lru":
		return NewLRUCache(maxBytes, onEvicted)
	case "lfu":
		return NewLFUCache(maxBytes, onEvicted)
	}
	return nil
}

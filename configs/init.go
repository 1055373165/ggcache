package conf

import (
	"ggcache/internal/middleware/logger"
	db "ggcache/internal/middleware/mysql"
	"sync"
)

var (
	mu sync.Mutex
)

func Init() {
	logger.Init()
	mu.Lock()
	db.Init()
	mu.Unlock()
}

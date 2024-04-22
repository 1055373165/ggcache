package conf

import (
	"ggcache/internal/middleware/logger"
	db "ggcache/internal/middleware/mysql"
)

func Init() {
	logger.Init()
	db.Init()
}

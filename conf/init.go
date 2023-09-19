package conf

import (
	"github.com/1055373165/groupcache/middleware/db"
	"github.com/1055373165/groupcache/middleware/logger"
)

func Init() {
	logger.Init()
	db.Init()
}

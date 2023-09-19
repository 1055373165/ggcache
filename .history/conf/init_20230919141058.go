package conf

import (
	"github.com/1055373165/groupcache/db"
	"github.com/1055373165/groupcache/log"
)

func Init() {
	logger.Init()
	db.Init()
}

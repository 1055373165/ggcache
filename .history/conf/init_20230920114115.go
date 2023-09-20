package conf

import (
	"github.com/1055373165/Distributed_KV_Store/db"
	"github.com/1055373165/Distributed_KV_Store/logger"
)

func Init() {
	logger.Init()
	db.Init()
}

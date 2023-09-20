package conf

import (
	"github.com/1055373165/distributekv/db"
	"github.com/1055373165/distributekv/logger"
)

func Init() {
	logger.Init()
	db.Init()
}

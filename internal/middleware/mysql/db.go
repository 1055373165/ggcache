package mysql

import (
	"os"

	"ggcache/internal/middleware/logger"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Student struct {
	gorm.Model
	Name  string `json:"name"`
	Score string `json:"score"`
}

var DB *gorm.DB

func Init() {
	var err error
	DB, err = gorm.Open(mysql.Open(os.Getenv("DSN")), &gorm.Config{})
	if err != nil {
		logger.Logger.Info(err.Error())
	}

	err = DB.AutoMigrate(&Student{})
	if err != nil {
		logger.Logger.Info(err.Error())
	}
}

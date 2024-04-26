package dao

import (
	"os"

	"ggcache/internal/pkg/student/model"
	"ggcache/utils/logger"
)

func migration() {
	if IsHasTable("student") {
		return
	}

	err := _db.Set("gorm:table_options", "charset=utf8mb4").
		AutoMigrate(
			&model.Student{},
		)

	if err != nil {
		logger.LogrusObj.Infoln("register table failed")
		os.Exit(0)
	}

	logger.LogrusObj.Infoln("register table success")
}

func IsHasTable(tableName string) bool {
	return _db.Migrator().HasTable(tableName)
}

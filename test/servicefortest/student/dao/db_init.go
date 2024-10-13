package dao

import (
	"context"

	"strings"

	"github.com/1055373165/ggcache/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var _db *gorm.DB

func InitDB() {
	conf := config.Conf.Mysql

	// username:password@tcp(host:port)/database?charset=xx&parseTime=xx&loc=xx
	dsn := strings.Join([]string{conf.UserName, ":", conf.Password, "@tcp(", conf.Host, ":", conf.Port, ")/", conf.Database, "?charset=", conf.Charset, "&parseTime=", "true", "&loc=", "Local"}, "")
	err := Database(dsn)
	if err != nil {
		panic(err)
	}
}

func Database(connStr string) error {
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       connStr,
		DefaultStringSize:         256,   // Default length of String type fields
		DisableDatetimePrecision:  true,  // Disable datetime precision
		DontSupportRenameIndex:    true,  // When renaming an index, delete and create a new one
		DontSupportRenameColumn:   true,  // Rename the column with `change`
		SkipInitializeWithVersion: false, // Automatically configure based on version
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})

	if err != nil {
		panic(err)
	}

	// sqlDB, _ := db.DB()
	// sqlDB.SetMaxIdleConns(20)
	// sqlDB.SetMaxOpenConns(100)
	// sqlDB.SetConnMaxLifetime(30 * time.Second)
	_db = db
	migration()
	return err
}

func NewDBClient(ctx context.Context) *gorm.DB {
	db := _db
	return db.WithContext(ctx)
}

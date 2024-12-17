package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/1055373165/ggcache/config"
	"github.com/1055373165/ggcache/pkg/common/logger"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var _db *gorm.DB

type DBConfig struct {
	Host         string
	Port         string
	Database     string
	Username     string
	Password     string
	Charset      string
	MaxIdleConns int
	MaxOpenConns int
	MaxLifetime  time.Duration
}

// InitDB initializes the database connection
func InitDB() error {
	cfg := DBConfig{
		Host:         config.Conf.Mysql.Host,
		Port:         config.Conf.Mysql.Port,
		Database:     config.Conf.Mysql.Database,
		Username:     config.Conf.Mysql.UserName,
		Password:     config.Conf.Mysql.Password,
		Charset:      config.Conf.Mysql.Charset,
		MaxIdleConns: 10,
		MaxOpenConns: 100,
		MaxLifetime:  time.Hour,
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=true&loc=Local",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database, cfg.Charset)

	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       dsn,
		DefaultStringSize:         256,
		DisableDatetimePrecision:  true,
		DontSupportRenameIndex:    true,
		DontSupportRenameColumn:   true,
		SkipInitializeWithVersion: false,
	}), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		DisableForeignKeyConstraintWhenMigrating: true,
	})

	if err != nil {
		logger.LogrusObj.Errorf("failed to connect to database: %v", err)
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.LogrusObj.Errorf("failed to get sql.DB: %v", err)
		return fmt.Errorf("failed to get sql.DB: %v", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.MaxLifetime)

	_db = db
	migration()

	// Initialize database for testing
	InitilizeTestData()

	return nil
}

func NewDBClient(ctx context.Context) *gorm.DB {
	if _db == nil {
		panic("database not initialized")
	}
	return _db.WithContext(ctx)
}

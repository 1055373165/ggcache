package mysql

import (
	"os"
	"strconv"

	"ggcache/internal/middleware/logger"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Student struct {
	gorm.Model
	Name  string `gorm:"type:varchar(255);uniqueIndex" json:"name"`
	Score string `gorm:"type:varchar(255)" json:"score"`
}

const (
	TestEntriesNumber = 1000
)

var (
	DB *gorm.DB
)

func Init() {
	if DB != nil {
		return
	}

	var err error
	// 1. connect to the database
	DB, err = gorm.Open(mysql.Open(os.Getenv("DSN")), &gorm.Config{})
	if err != nil {
		logger.Logger.Fatalf("failed connect to mysql %v", err)
	}

	if DB.Migrator().HasTable(&Student{}) {
		return
	}

	// 2. create a table structure based on a custom structure
	err = DB.AutoMigrate(&Student{})
	if err != nil {
		logger.Logger.Fatalf("failed auto migrate table students %v", err)
	}

	// 3. insert test data
	InitDataWithGroup()
}

func InitDataWithGroup() {
	// 先往数据库中存一些数据（慢速数据库）
	DB.Create(&Student{Name: "李四", Score: "100"})
	DB.Create(&Student{Name: "张三", Score: "10"})
	DB.Create(&Student{Name: "王五", Score: "1000"})
	DB.Create(&Student{Name: "李六", Score: "10000"})
	DB.Create(&Student{Name: "赵七", Score: "100000"})
	DB.Create(&Student{Name: "孙八", Score: "1000000"})
	DB.Create(&Student{Name: "钱九", Score: "10000000"})
	DB.Create(&Student{Name: "周十", Score: "100000000"})
	DB.Create(&Student{Name: "hi", Score: "100"})
	DB.Create(&Student{Name: "hihi", Score: "1000"})
	DB.Create(&Student{Name: "hihihi", Score: "10000"})
	DB.Create(&Student{Name: "hihihihi", Score: "100000"})
	DB.Create(&Student{Name: "hihihihihi", Score: "1000000"})
	DB.Create(&Student{Name: "oh", Score: "100"})
	DB.Create(&Student{Name: "my", Score: "1000"})
	DB.Create(&Student{Name: "ohmy", Score: "10000"})
	DB.Create(&Student{Name: "one", Score: "10"})
	DB.Create(&Student{Name: "two", Score: "100"})
	DB.Create(&Student{Name: "three", Score: "1000"})
	DB.Create(&Student{Name: "four", Score: "100000"})
	DB.Create(&Student{Name: "five", Score: "1000000"})
	DB.Create(&Student{Name: "six", Score: "10000000"})
	DB.Create(&Student{Name: "unknown", Score: "0000000"})
	for i := 0; i < TestEntriesNumber; i++ {
		name, score := strconv.Itoa(i), strconv.Itoa((i+1)*10)
		DB.Create(&Student{Name: name, Score: score})
	}
	logger.Logger.Info("数据库数据插入成功...")
}

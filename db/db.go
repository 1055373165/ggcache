package db

import (
	"os"

	"github.com/1055373165/Distributed_KV_Store/logger"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Student struct {
	ID          uint   `gorm:"primarykey"`
	Name        string `gorm:"type:varchar(100);index;comment:学生姓名"`
	Score       string `gorm:"type:varchar(100);comment:学生分数"`
	Grade       string `gorm:"type:varchar(50);comment:学生年级;default:''"`
	Email       string `gorm:"type:varchar(100);comment:学生邮箱;default:''"`
	PhoneNumber string `gorm:"type:varchar(20);comment:学生电话号码;default:''"`
}

func (Student) Table() string {
	return "student"
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

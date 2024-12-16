package dao

import (
	"context"

	stuPb "github.com/1055373165/ggcache/api/studentpb"
	"github.com/1055373165/ggcache/internal/bussiness/student/model"
	"github.com/1055373165/ggcache/pkg/common/logger"

	"gorm.io/gorm"
)

// The test only requires simple insertion and query of student information.
type StudentDao struct {
	*gorm.DB
}

func NewStudentDao(ctx context.Context) *StudentDao {
	return &StudentDao{NewDBClient(ctx)}
}

func (dao *StudentDao) ShowStudentInfo(req *stuPb.StudentRequest) (*model.Student, error) {
	var s model.Student
	err := dao.Model(&model.Student{}).Where("name=?", req.Name).First(&s).Error
	return &s, err
}

func (dao *StudentDao) CreateStudent(req *stuPb.StudentRequest) error {
	var student model.Student
	student.Name = req.Name
	student.Score = req.Score
	student.Email = req.Email
	student.Grade = req.Grade
	student.PhoneNumber = req.PhoneNumber

	var err error
	if err = dao.Model(&model.Student{}).Create(&student).Error; err != nil {
		logger.LogrusObj.Error("Insert User Error: ", err.Error())
		return err
	}
	return nil
}

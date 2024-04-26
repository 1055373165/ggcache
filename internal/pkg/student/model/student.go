package model

type Student struct {
	ID          uint    `gorm:"primarykey"`
	Name        string  `gorm:"type:varchar(100);index;comment:学生姓名"`
	Score       float64 `gorm:"type:varchar(100);comment:学生分数"`
	Grade       string  `gorm:"type:varchar(50);comment:学生年级;default:''"`
	Email       string  `gorm:"type:varchar(100);comment:学生邮箱;default:''"`
	PhoneNumber string  `gorm:"type:varchar(20);comment:学生电话号码;default:''"`
}

func (Student) Table() string {
	return "student"
}

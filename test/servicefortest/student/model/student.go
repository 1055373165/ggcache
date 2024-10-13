package model

// Suppose the most common operation in the business logic is
// to query a student's score based on the student's name.
// In order to improve query efficiency, create a joint index for (name, score)
// to reduce the performance consumption caused by a large number of table returns).
type Student struct {
	ID          uint    `gorm:"primarykey"`
	Name        string  `gorm:"type:varchar(100);index:,score:idx_name_score"`
	Score       float64 `gorm:"type:decimal(10,2);index:idx_name_score,priority:2;comment:学生分数"`
	Grade       string  `gorm:"type:varchar(50);comment:学生年级;default:''"`
	Email       string  `gorm:"type:varchar(100);comment:学生邮箱;default:''"`
	PhoneNumber string  `gorm:"type:varchar(20);comment:学生电话号码;default:''"`
}

func (Student) Table() string {
	return "student"
}

package model

// 假设业务逻辑中最常用操作是根据学生姓名查询学生的分数
// 为了提高查询效率，为 (name, score) 建立一个联合索引（减少大量回表带来的性能消耗）
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

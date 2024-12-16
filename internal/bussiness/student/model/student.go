package model

type Student struct {
	ID          uint    `gorm:"primarykey"`
	Name        string  `gorm:"type:varchar(100);index:idx_name_score"`
	Score       float32 `gorm:"type:decimal(10,2);index:idx_name_score"`
	Email       string  `gorm:"type:varchar(100)"`
	Grade       string  `gorm:"type:varchar(20)"`
	PhoneNumber string  `gorm:"type:varchar(20)"`
}

func (Student) Table() string {
	return "student"
}

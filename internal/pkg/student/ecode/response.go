package ecode

type Response struct {
	Status    int     `json:"status"`
	StudentId int64   `json:"student_id"`
	Name      string  `json:"name"`
	Score     float64 `json:"score"`
}

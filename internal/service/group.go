package service

import (
	"errors"
	"ggcache/internal/middleware/logger"
	db "ggcache/internal/middleware/mysql"
)

func NewGroupInstance(groupName string) *Group {
	g := NewGroup(groupName, "lru", 2<<10, RetrieveFunc(func(key string) ([]byte, error) {
		// 从后端数据库中查找
		logger.Logger.Info("进入 RetrieveFunc, 数据库中查询....")

		var scores []*db.Student
		db.DB.Where("name = ?", key).Find(&scores)
		if len(scores) == 0 {
			logger.Logger.Info("后端数据库中也查询不到...")
			return []byte{}, errors.New("record not found")
		}

		logger.Logger.Infof("成功从后端数据库中查询到学生 %s 的分数：%s", key, scores[0].Score)
		return []byte(scores[0].Score), nil
	}))
	return g
}

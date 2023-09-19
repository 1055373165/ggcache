package distributekv

import (
	"errors"

	"github.com/1055373165/distributekv/db"
	"github.com/1055373165/distributekv/logger"
)

func NewGroupInstance(groupname string) *Group {
	g := NewGroup(groupname, 1<<10, RetrieveFunc(func(key string) ([]byte, error) {
		// 从后端数据库中查找
		logger.Logger.Info("进入 GetterFunc，数据库中查询....")
		var scores []*db.Student
		db.DB.Where("name = ?", key).Find(&scores)
		if len(scores) == 0 {
			logger.Logger.Info("后端数据库中也查询不到...")
			return []byte{}, errors.New("record not found")
		}

		logger.Logger.Infof("成功从后端数据库中查询到学生 %s 的分数：%s", key, scores[0].Score)
		return []byte(scores[0].Score), nil
	}))

	groups[groupname] = g
	InitDataWithGroup(g)
	return g
}

func InitDataWithGroup(g *Group) {
	dataPrefix := g.name
	// 先往数据库中存一些数据（慢速数据库）
	db.DB.Create(&db.Student{Name: dataPrefix + "张三", Score: "100"})
	db.DB.Create(&db.Student{Name: dataPrefix + "李四", Score: "1000"})
	db.DB.Create(&db.Student{Name: dataPrefix + "王五", Score: "10000"})
	db.DB.Create(&db.Student{Name: dataPrefix + "hihihi", Score: "100"})
	db.DB.Create(&db.Student{Name: dataPrefix + "hi", Score: "1000"})
	db.DB.Create(&db.Student{Name: dataPrefix + "hihi", Score: "10000"})
	logger.Logger.Info("数据库数据插入成功...")
	// 往缓存中存储一些元素
	g.cache.put(dataPrefix+"abc", ByteView{b: []byte("123")})
	g.cache.put(dataPrefix+"bcd", ByteView{b: []byte("234")})
	g.cache.put(dataPrefix+"cde", ByteView{b: []byte("345")})
	logger.Logger.Info("缓存数据插入成功...")
}

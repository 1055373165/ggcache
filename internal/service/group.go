package service

import (
	"context"
	"errors"
	"math/rand"
	"strconv"
	"time"

	pb "github.com/1055373165/ggcache/api/studentpb"
	"github.com/1055373165/ggcache/internal/pkg/student/dao"
	"github.com/1055373165/ggcache/utils/logger"

	"gorm.io/gorm"
)

func NewGroupManager(groupnames []string, currentPeerAddr string) map[string]*Group {
	// 为每个 group 构造一个 Group 实例
	for i := 0; i < len(groupnames); i++ {
		g := NewGroup(groupnames[i], "lru", 100*2*20, RetrieveFunc(func(key string) ([]byte, error) {
			start := time.Now()
			dao := dao.NewStudentDao(context.Background())
			stus, err := dao.ShowStudentInfo(&pb.StudentRequest{
				Name:  key,
				Score: rand.Float32(),
			})

			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return []byte{}, gorm.ErrRecordNotFound
				} else {
					return []byte{}, err
				}
			} else {
				logger.LogrusObj.Infof("成功从后端数据库中查询到学生 %s 的分数：%v", key, stus.Score)
				logger.LogrusObj.Warnf("查询数据库总耗时: %v ms", time.Since(start).Milliseconds())
			}
			// "123.79"
			return []byte(strconv.FormatFloat(stus.Score, 'f', 2, 64)), nil
		}))

		GroupManager[groupnames[i]] = g
	}

	// 统一使用 group manager 进行管理
	return GroupManager
}

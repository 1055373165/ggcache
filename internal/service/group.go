package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	pb "ggcache/api/studentpb"
	"ggcache/internal/pkg/student/dao"
	"ggcache/utils/logger"

	"gorm.io/gorm"
)

func NewGroupManager(groupnames []string, currentPeerAddr string) map[string]*Group {

	// 为每个 group 构造一个 Group 实例
	for i := 0; i < len(groupnames); i++ {
		g := NewGroup(groupnames[i], 100*2*20, RetrieveFunc(func(key string) ([]byte, error) {
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
		// InitDataWithGroup(g)
	}

	// 统一使用 group manager 进行管理

	// 模拟数据插入，等到数据库中记录行数足够时注释掉

	return GroupManager
}

func InitDataWithGroup(g *Group) {
	dao := dao.NewStudentDao(context.Background())
	// 中文测试数据
	names := []string{
		"王五", "赵四", "李雷", "张三", "刘六", "陈七", "杨八", "吴九", "周十", "徐二",
		"孙明", "朱琪", "马华", "胡京", "郭士", "何东", "高北", "罗成", "林松", "赖林",
		"郑帅", "黄蓉", "韩梅", "顾桂", "汪松", "施云", "文希", "向荣", "梁宝", "宋江",
		"唐伯", "许利", "魏明", "蒋华", "沈丹", "韦石", "昌平", "苏波", "金山", "侯月",
		"邓光", "曹志", "彭波", "曾峰", "田野", "樊瑞", "程心", "袁思", "陆雨", "邹渊",
	}
	for _, name := range names {
		dao.CreateStudent(&pb.StudentRequest{
			Name:  name,
			Score: rand.Float32(),
		})
	}
	for i := 0; i < 1000; i++ {
		dao.CreateStudent(&pb.StudentRequest{
			Name:  fmt.Sprintf("%d", i),
			Score: rand.Float32(),
		})
	}
	logger.LogrusObj.Info("数据库数据插入成功...")
}

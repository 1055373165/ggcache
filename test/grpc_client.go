package main

import (
	"context"
	"errors"
	"math"
	"math/rand/v2"

	pb "github.com/1055373165/ggcache/api/groupcachepb"
	stuPb "github.com/1055373165/ggcache/api/studentpb"
	"github.com/1055373165/ggcache/config"
	"github.com/1055373165/ggcache/internal/bussiness/student/dao"
	"github.com/1055373165/ggcache/pkg/common/logger"
	"github.com/1055373165/ggcache/pkg/etcd/discovery"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gorm.io/gorm"
)

const (
	MaxRetries = 3
)

// 第一次重试等待 1s
// 第二次重试等待 2s
// 第三次重试等待 4s
func backoff(retry int) int {
	return int(math.Pow(float64(2), float64(retry)))
}

func main() {
	config.InitConfig()
	dao.InitDB()

	cli, err := clientv3.New(config.DefaultEtcdConfig)
	if err != nil {
		panic(err)
	}

	// 服务发现（直接根据服务名字获取与服务的虚拟端连接）
	conn, err := discovery.Discovery(cli, config.Conf.Services["groupcache"].Name)
	if err != nil {
		panic(err)
	}
	logger.LogrusObj.Debug("Discovery continue")
	client_stub := pb.NewGroupCacheClient(conn)

	// 为了模拟实际查询情况，将 names 数组中存在的名字和不存在的名字的分部比例设置为 10 : 1
	// 然后为了尽可能随机，使用洗牌算法将数组内所有学生姓名打散
	// 存在的名字中文 1000 个，英文 1000 个（已经存到数据库中了，测试集在主程序下方提供）
	// 查询请求使用 100 个已经存在中文名和 100 个已经存在的英文名
	// 使用不存在名字中文 10 个，英文 10 个

	unExistChineseNames := dao.GenerateChineseNames(10)
	unExistEnglishNames := dao.GenerateEnglishNames(10)

	dbExistChineseNames := dao.GetGenerateChineseNames()
	dbExistEnglishNames := dao.GetGenerateEnglishNames()

	names := make([]string, 0, 20+len(*dbExistChineseNames)+len(*dbExistEnglishNames))
	names = append(names, unExistChineseNames...)
	names = append(names, unExistEnglishNames...)
	names = append(names, *dbExistChineseNames...)
	names = append(names, *dbExistEnglishNames...)

	// 打散
	rand.Shuffle(len(names), func(i, j int) {
		names[i], names[j] = names[j], names[i]
	})

	dao := dao.NewStudentDao(context.Background())
	for _, name := range names {
		dao.CreateStudent(&stuPb.StudentRequest{
			Name:  name,
			Score: rand.Float32() * 100,
		})
	}

	for {
		for _, name := range names {
			searchFunc := func() (*pb.GetResponse, error) {
				return client_stub.Get(context.Background(), &pb.GetRequest{
					Group: "scores",
					Key:   name,
				})
			}

			for i := 0; i < MaxRetries; i++ {
				resp, err := searchFunc()
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						logger.LogrusObj.Errorf("学生 %s 信息不在数据库中", name)
						break
					}
					logger.LogrusObj.Errorf("本次查询学生 %s 分数的 rpc 调用出现故障，重试次数 %d", name, i+1)
					// time.Sleep(time.Second * time.Duration(backoff(i))) // 退避算法
				} else {
					logger.LogrusObj.Infof("rpc 调用成功, 学生 %s 的成绩为 %s", name, string(resp.Value))
					break
				}
			}
		}
	}
}

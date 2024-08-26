package main

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	pb "github.com/1055373165/ggcache/api/groupcachepb"
	"github.com/1055373165/ggcache/config"
	discovery "github.com/1055373165/ggcache/internal/discovery"
	"github.com/1055373165/ggcache/pkg/student/dao"
	"github.com/1055373165/ggcache/utils/logger"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	ErrRPCCallNotFound  = "rpc error: code = Unknown desc = record not found"
	MaxRetries          = 3
	InitialRetryWaitSec = 1
)

const (
	NotFoundStatus Status = iota // 说明服务器没有查询到指定名字学生的分数
	ErrorStatus                  // 说明服务器出现问题
)

type Status int

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
	names := []string{"王五", "张三", "李四", "王二", "不存在", "赵六", "李奇"}
	for i := 0; i < 10; i++ {
		names = append(names, fmt.Sprintf("不存在%d", i))
	}
	for i := 0; i < 100; i++ {
		names = append(names, fmt.Sprintf("%d", i))
	}
	for i := 0; i < 100; i++ {
		names = append(names, fmt.Sprintf("%d", i))
	}
	for i := 0; i < 100; i++ {
		names = append(names, fmt.Sprintf("%d", i))
	}
	for i := 0; i < 100; i++ {
		names = append(names, fmt.Sprintf("%d", i))
	}
	// 打散
	rand.Shuffle(len(names), func(i, j int) {
		names[i], names[j] = names[j], names[i]
	})

	for {
		for _, name := range names {
			searchFunc := func() (*pb.GetResponse, error) {
				return client_stub.Get(context.Background(), &pb.GetRequest{
					Group: "scores",
					Key:   name,
				})
			}

			for i := 0; i < MaxRetries; i++ { // 重试机制
				resp, err := searchFunc()
				if err != nil {
					if ErrorHandle(err) == NotFoundStatus {
						logger.LogrusObj.Warnf("查询不到学生 %s 的成绩", name)
						break
					}
					logger.LogrusObj.Errorf("本次查询学生 %s 分数的 rpc 调用出现故障，重试次数 %d", name, i+1)
					waitTime := time.Duration(InitialRetryWaitSec*(1<<uint(i))) * time.Second // 退避算法
					time.Sleep(waitTime)
				} else {
					logger.LogrusObj.Infof("rpc 调用成功, 学生 %s 的成绩为 %s", name, string(resp.Value))
					break
				}
			}
		}
	}
}

func ErrorHandle(err error) Status {
	if err.Error() == ErrRPCCallNotFound {
		return NotFoundStatus
	}
	return ErrorStatus
}

// 第一次重试等待 1s
// 第二次重试等待 2s
// 第三次重试等待 4s
func backoff(retry int) int {
	return int(math.Pow(float64(2), float64(retry)))
}

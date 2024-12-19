package main

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	pb "github.com/1055373165/ggcache/api/groupcachepb"
	"github.com/1055373165/ggcache/config"
	"github.com/1055373165/ggcache/internal/bussiness/student/dao"
	"github.com/1055373165/ggcache/pkg/common/logger"
	discovery "github.com/1055373165/ggcache/pkg/etcd/discovery"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

type GGCacheClient struct {
	etcdCli     *clientv3.Client
	conn        *grpc.ClientConn
	client      pb.GroupCacheClient
	serviceName string
	connected   bool
	mu          sync.RWMutex
}

func NewGGCacheClient(etcdCli *clientv3.Client, serviceName string) (*GGCacheClient, error) {
	client := &GGCacheClient{
		etcdCli:     etcdCli,
		serviceName: serviceName,
	}
	if err := client.connect(); err != nil {
		return nil, err
	}
	return client, nil
}

func (c *GGCacheClient) connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	conn, err := discovery.Discovery(c.etcdCli, c.serviceName)
	if err != nil {
		return fmt.Errorf("failed to discover service: %v", err)
	}

	c.conn = conn
	c.client = pb.NewGroupCacheClient(conn)
	c.connected = true
	return nil
}

func (c *GGCacheClient) Get(ctx context.Context, group, key string) (*pb.GetResponse, error) {
	c.mu.RLock()
	if !c.connected {
		c.mu.RUnlock()
		if err := c.connect(); err != nil {
			return nil, err
		}
	} else {
		c.mu.RUnlock()
	}

	var resp *pb.GetResponse
	var lastErr error

	retries := 0
	maxRetries := 3
	for retries < maxRetries {
		select {
		case <-ctx.Done():
			if lastErr != nil {
				return nil, lastErr
			}
			return nil, ctx.Err()
		default:
			var err error
			resp, err = c.client.Get(ctx, &pb.GetRequest{
				Group: group,
				Key:   key,
			})

			if err != nil {
				if status.Code(err) == codes.Unavailable {
					// 连接断开，尝试重连
					c.mu.Lock()
					c.connected = false
					c.mu.Unlock()
					if reconnErr := c.connect(); reconnErr != nil {
						lastErr = reconnErr
						// 使用指数退避等待
						waitTime := time.Duration(backoff(retries)) * time.Second
						time.Sleep(waitTime)
						retries++
						continue
					}
				}
				lastErr = err
				retries++
				continue
			}
			return resp, nil
		}
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("max retries exceeded")
}

func main() {
	config.InitConfig()
	dao.InitDB()

	cli, err := clientv3.New(config.DefaultEtcdConfig)
	if err != nil {
		panic(err)
	}

	ggcacheClient, err := NewGGCacheClient(cli, config.Conf.Services["groupcache"].Name)
	if err != nil {
		panic(err)
	}

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
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			resp, err := ggcacheClient.Get(ctx, "scores", name)
			if err != nil {
				if ErrorHandle(err) == NotFoundStatus {
					logger.LogrusObj.Warnf("查询不到学生 %s 的成绩", name)
				} else {
					logger.LogrusObj.Errorf("查询学生 %s 分数失败: %v", name, err)
				}
			} else {
				logger.LogrusObj.Infof("查询成功, 学生 %s 的成绩为 %s", name, string(resp.Value))
			}
			cancel()
			time.Sleep(100 * time.Millisecond) // 控制请求速率
		}
	}
}

func ErrorHandle(err error) Status {
	if err.Error() == ErrRPCCallNotFound {
		return NotFoundStatus
	}
	return ErrorStatus
}

// First retry wait 1s
// Second retry wait 2s
// The third retry waits for 4 seconds
func backoff(retry int) int {
	return int(math.Pow(float64(2), float64(retry)))
}

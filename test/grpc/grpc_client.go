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

	var lastErr error
	for retry := 0; retry < MaxRetries; retry++ {
		resp, err := c.client.Get(ctx, &pb.GetRequest{
			Group: group,
			Key:   key,
		})

		if err == nil {
			return resp, nil
		}

		lastErr = err
		if status.Code(err) == codes.Unavailable {
			// 连接断开，尝试重连
			c.mu.Lock()
			c.connected = false
			c.mu.Unlock()

			if reconnErr := c.connect(); reconnErr != nil {
				lastErr = reconnErr
			}
		}

		// 使用指数退避等待
		waitTime := time.Duration(backoff(retry)) * time.Second
		logger.LogrusObj.Warnf("第 %d 次重试失败，等待 %v 后重试: %v", retry+1, waitTime, err)
		time.Sleep(waitTime)
	}

	return nil, fmt.Errorf("max retries exceeded: %v", lastErr)
}

func main() {
	config.InitConfig()
	if err := dao.InitDB(); err != nil {
		logger.LogrusObj.Fatalf("Failed to initialize database: %v", err)
	}

	cli, err := clientv3.New(config.DefaultEtcdConfig)
	if err != nil {
		panic(err)
	}

	ggcacheClient, err := NewGGCacheClient(cli, config.Conf.Services["groupcache"].Name)
	if err != nil {
		panic(err)
	}

	// 构造热点数据（20%的key承载80%的访问）
	// 选择成绩最高的学生作为热点数据
	hotKeys := []string{
		"Emma Moore",        // 9910分
		"郭上",                // 9818分
		"Emma Williams",     // 9710分
		"冯怎样",               // 9537分
		"Emily Franklin",    // 9533分
		"杨里",                // 9643分
		"Hannah Harris",     // 9297分
		"宋它",                // 9239分
		"Joseph Walker",     // 9251分
		"Noah Hall",         // 9360分
		"周就",                // 9363分
		"陈是的",               // 9373分
		"赵到",                // 9502分
		"杨是",                // 9536分
		"Matthew Jefferson", // 9629分
	}

	// 构造长尾数据（80%的key承载20%的访问）
	// 使用分数较低的学生数据
	coldKeys := make([]string, 0)
	// 使用ID 1014-1100范围的学生（避开已经在热点数据中的学生）
	usedIds := map[string]bool{
		"Emma Moore": true, "郭上": true, "Emma Williams": true,
		"冯怎样": true, "Emily Franklin": true, "杨里": true,
		"Hannah Harris": true, "宋它": true, "Joseph Walker": true,
		"Noah Hall": true, "周就": true, "陈是的": true,
		"赵到": true, "杨是": true, "Matthew Jefferson": true,
	}

	// 添加英文名学生
	englishNames := []string{
		"David Lopez", "Daniel Green", "Chloe Scott", "Joseph Clark",
		"John Young", "Samuel Davis", "Isabella Hill", "Emily Baker",
		"Ava Jenkins", "Grace Adams", "Samuel Morris", "Joseph Bell",
		"Zoe Howard", "John Anderson", "Chloe Miller", "Samuel Collins",
		"Sophia Sanchez", "Joseph Scott", "Christopher Jenkins", "William White",
	}

	// 添加中文名学生
	chineseNames := []string{
		"李说", "林它", "宋什么", "刘好像", "杨什么",
		"胡着", "郭看", "刘好", "萧她", "刘没",
		"赵你", "董没", "朱了", "王不", "陈一",
	}

	// 合并长尾数据，避免添加已在热点数据中的学生
	for _, name := range englishNames {
		if !usedIds[name] {
			coldKeys = append(coldKeys, name)
		}
	}

	for _, name := range chineseNames {
		if !usedIds[name] {
			coldKeys = append(coldKeys, name)
		}
	}

	// 添加更多的学生ID来达到足够的数量
	for i := 1100; i < 1400; i++ {
		studentId := fmt.Sprintf("student_%d", i)
		if !usedIds[studentId] {
			coldKeys = append(coldKeys, studentId)
		}
	}

	// 构造最终的请求序列
	totalRequests := 100 // 总请求数
	names := make([]string, 0, totalRequests)

	// 添加热点请求（80%的访问量）
	hotRequestCount := int(float64(totalRequests) * 0.8) // 80%的请求量
	for i := 0; i < hotRequestCount; i++ {
		names = append(names, hotKeys[i%len(hotKeys)])
	}

	// 添加长尾请求（20%的访问量）
	coldRequestCount := totalRequests - hotRequestCount // 剩余20%的请求量
	for i := 0; i < coldRequestCount; i++ {
		names = append(names, coldKeys[i%len(coldKeys)])
	}

	// 随机打散请求顺序
	rand.Shuffle(len(names), func(i, j int) {
		names[i], names[j] = names[j], names[i]
	})

	for {
		for _, name := range names {
			ctx := context.Background()
			resp, err := ggcacheClient.Get(ctx, "scores", name)
			if err != nil {
				if ErrorHandle(err) == NotFoundStatus {
					logger.LogrusObj.Warnf("查询不到学生 %s 的成绩", name)
					continue
				}
				logger.LogrusObj.Errorf("查询学生 %s 分数失败: %v", name, err)
				return // 如果不是 NotFound 错误，直接退出程序
			}
			logger.LogrusObj.Infof("查询成功, 学生 %s 的成绩为 %s", name, string(resp.Value))
		}
		time.Sleep(time.Millisecond * 100)
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

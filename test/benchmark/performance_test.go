package main

import (
	"context"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	pb "github.com/1055373165/ggcache/api/groupcachepb"
	"github.com/1055373165/ggcache/config"
	"github.com/1055373165/ggcache/internal/bussiness/student/dao"
	"github.com/1055373165/ggcache/pkg/common/logger"
	"github.com/1055373165/ggcache/pkg/etcd/discovery"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/exp/rand"
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

type Metrics struct {
	totalRequests   int64
	successRequests int64
	failedRequests  int64
	totalLatency    int64 // 单位：纳秒
	maxLatency      int64
	minLatency      int64
	cacheHits       int64
	cacheMisses     int64
}

type workloadGenerator struct {
	hotKeys  []string
	coldKeys []string
}

func newWorkloadGenerator() *workloadGenerator {
	// 热点数据（20%的key承载80%的访问）
	hotKeys := []string{"张三", "李四", "王五", "赵六", "王二"}

	// 长尾数据（80%的key承载20%的访问）
	coldKeys := make([]string, 0)
	for i := 0; i < 20; i++ {
		coldKeys = append(coldKeys, fmt.Sprintf("student_%d", i))
	}

	return &workloadGenerator{
		hotKeys:  hotKeys,
		coldKeys: coldKeys,
	}
}

func (w *workloadGenerator) getKey() string {
	// 80%的概率访问热点数据
	if rand.Float64() < 0.8 {
		return w.hotKeys[rand.Intn(len(w.hotKeys))]
	}
	// 20%的概率访问长尾数据
	return w.coldKeys[rand.Intn(len(w.coldKeys))]
}

func BenchmarkCache(b *testing.B) {
	config.InitConfig()
	dao.InitDB()
	// 初始化 etcd 客户端
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379"},
	})
	if err != nil {
		b.Fatal(err)
	}
	defer cli.Close()

	// 创建缓存客户端
	client, err := NewGGCacheClient(cli, config.Conf.Services["groupcache"].Name)
	if err != nil {
		b.Fatal(err)
	}

	tests := []struct {
		name        string
		concurrency int
		duration    time.Duration
		keySpace    int // 用于生成不同的key范围
	}{
		{"Low_Concurrency", 10, 30 * time.Second, 1000},
		{"Medium_Concurrency", 50, 30 * time.Second, 1000},
		{"High_Concurrency", 100, 30 * time.Second, 1000},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			runWorkload(b, client, tt.concurrency, tt.duration)
		})
	}
}

func runWorkload(b *testing.B, client *GGCacheClient, concurrency int, duration time.Duration) {
	metrics := &Metrics{
		minLatency: int64(^uint64(0) >> 1), // 设置为最大值
	}
	wg := sync.WaitGroup{}

	// 创建工作负载生成器
	generator := newWorkloadGenerator()

	// 启动指标报告协程
	stopReport := make(chan struct{})
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				reportMetrics(metrics)
			case <-stopReport:
				return
			}
		}
	}()

	// 启动工作负载
	start := time.Now()
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for time.Since(start) < duration {
				ctx := context.Background()

				// 使用工作负载生成器获取符合二八原则的key
				key := generator.getKey()

				reqStart := time.Now()
				resp, err := client.Get(ctx, "scores", key)
				latency := time.Since(reqStart)

				atomic.AddInt64(&metrics.totalRequests, 1)
				atomic.AddInt64(&metrics.totalLatency, latency.Nanoseconds())

				if err != nil {
					atomic.AddInt64(&metrics.failedRequests, 1)
					continue
				}

				atomic.AddInt64(&metrics.successRequests, 1)
				if isCacheHit(resp) {
					atomic.AddInt64(&metrics.cacheHits, 1)
				} else {
					atomic.AddInt64(&metrics.cacheMisses, 1)
				}

				// 更新最大最小延迟
				updateLatencyStats(metrics, latency.Nanoseconds())
			}
		}()
	}

	wg.Wait()
	close(stopReport)
	printFinalMetrics(b, metrics, duration)
}

func reportMetrics(metrics *Metrics) {
	total := atomic.LoadInt64(&metrics.totalRequests)
	if total == 0 {
		return // 如果还没有请求，直接返回
	}

	success := atomic.LoadInt64(&metrics.successRequests)
	failed := atomic.LoadInt64(&metrics.failedRequests)
	hits := atomic.LoadInt64(&metrics.cacheHits)
	misses := atomic.LoadInt64(&metrics.cacheMisses)

	var avgLatency int64
	if total > 0 {
		avgLatency = atomic.LoadInt64(&metrics.totalLatency) / total
	}

	// 避免除零错误
	var hitRate float64
	if hits+misses > 0 {
		hitRate = float64(hits) / float64(hits+misses) * 100
	}

	logger.LogrusObj.Infof(
		"Current QPS: %d, Avg Latency: %dms, Success Rate: %.2f%%, Cache Hit Rate: %.2f%%, failed requests: %d",
		total,
		avgLatency/1e6,
		float64(success)/float64(total)*100,
		hitRate,
		failed,
	)
}

func printFinalMetrics(b *testing.B, metrics *Metrics, duration time.Duration) {
	total := atomic.LoadInt64(&metrics.totalRequests)
	if total == 0 {
		b.Log("No requests processed during the test")
		return
	}

	success := atomic.LoadInt64(&metrics.successRequests)
	failed := atomic.LoadInt64(&metrics.failedRequests)
	hits := atomic.LoadInt64(&metrics.cacheHits)
	misses := atomic.LoadInt64(&metrics.cacheMisses)

	var avgLatency int64
	if total > 0 {
		avgLatency = atomic.LoadInt64(&metrics.totalLatency) / total
	}

	// 避免除零错误
	var hitRate float64
	if hits+misses > 0 {
		hitRate = float64(hits) / float64(hits+misses) * 100
	}

	b.ReportMetric(float64(total)/duration.Seconds(), "req/s")
	b.ReportMetric(float64(avgLatency)/1e6, "avg_latency_ms")
	b.ReportMetric(float64(metrics.maxLatency)/1e6, "max_latency_ms")
	b.ReportMetric(float64(metrics.minLatency)/1e6, "min_latency_ms")
	b.ReportMetric(float64(success)/float64(total)*100, "success_rate_%")
	b.ReportMetric(hitRate, "cache_hit_%")
	b.ReportMetric(float64(failed), "failed_requests")
}

func updateLatencyStats(metrics *Metrics, latency int64) {
	// 更新最大最小延迟
	for {
		oldMax := atomic.LoadInt64(&metrics.maxLatency)
		if latency <= oldMax || atomic.CompareAndSwapInt64(&metrics.maxLatency, oldMax, latency) {
			break
		}
	}

	for {
		oldMin := atomic.LoadInt64(&metrics.minLatency)
		if latency >= oldMin || atomic.CompareAndSwapInt64(&metrics.minLatency, oldMin, latency) {
			break
		}
	}
}

func isCacheHit(resp *pb.GetResponse) bool {
	return resp.Value != nil
}

// First retry wait 1s
// Second retry wait 2s
// The third retry waits for 4 seconds
func backoff(retry int) int {
	return int(math.Pow(float64(2), float64(retry)))
}

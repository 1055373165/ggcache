package cache

import (
	"context"
	"sync"

	"fmt"
	"time"

	pb "github.com/1055373165/ggcache/api/groupcachepb"
	"github.com/1055373165/ggcache/pkg/common/logger"
	"github.com/1055373165/ggcache/pkg/etcd/discovery"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var _ Fetcher = (*Client)(nil)

type Client struct {
	serviceName string
	conn        *clientv3.Client
	mu          sync.RWMutex // 保护连接状态
}

func NewClient(serviceName string) *Client {
	cli, err := clientv3.NewFromURL("http://localhost:2379")
	if err != nil {
		return nil
	}
	return &Client{
		serviceName: serviceName,
		conn:        cli,
	}
}

// Fetch gets the corresponding cache value from remote peer
func (c *Client) Fetch(group string, key string) ([]byte, error) {
	// Discover services and obtain connection to services
	conn, err := discovery.Discovery(c.conn, c.serviceName)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	grpcClient := pb.NewGroupCacheClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Now()
	resp, err := grpcClient.Get(ctx, &pb.GetRequest{
		Group: group,
		Key:   key,
	})
	if err != nil {
		return nil, fmt.Errorf("could not get %s/%s from peer %s", group, key, c.serviceName)
	}

	logger.LogrusObj.Debugf("the duration of this grpc Call is: %v ms", time.Since(start).Milliseconds())

	return resp.Value, nil
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

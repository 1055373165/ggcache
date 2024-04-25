package distributekv

import (
	"context"

	"fmt"
	"time"

	services "github.com/1055373165/Distributed_KV_Store/etcd"

	pb "github.com/1055373165/Distributed_KV_Store/grpc/groupcachepb"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// The client module implements groupcache's ability to access other remote nodes to fetch caches.
type client struct {
	serviceName string // 服务名称 groupcache/ip:addr
}

// Fetch gets the corresponding cache value from remote peer
func (c *client) Fetch(group string, key string) ([]byte, error) {
	cli, err := clientv3.New(services.DefaultEtcdConfig)
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	// Discover services and obtain connection to services
	conn, err := services.EtcdDial(cli, c.serviceName)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	grpcClient := pb.NewGroupCacheClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// 使用带有超时自动取消的上下文和指定请求调用客户端的 Get 方法发起 rpc 请求调用
	resp, err := grpcClient.Get(ctx, &pb.GetRequest{
		Group: group,
		Key:   key,
	})
	if err != nil {
		return nil, fmt.Errorf("could not get %s/%s from perr %s", group, key, c.serviceName)
	}

	return resp.Value, nil
}

func NewClient(service string) *client {
	return &client{serviceName: service}
}

// 测试 client 是否实现了 Fetcher 接口
var _ Fetcher = (*client)(nil)

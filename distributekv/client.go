package distributekv

import (
	"context"

	"fmt"
	"time"

	services "github.com/1055373165/groupcache/etcd"

	pb "github.com/1055373165/groupcache/grpc/groupcachepb"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// client 模块实现了 groupcache 访问其他远程节点从而获取缓存的能力
type client struct {
	name string // 服务名称 gcache/ip:addr
}

// Fetch 从 remote peer 获取对应的缓存值
func (c *client) Fetch(group string, key string) ([]byte, error) {
	// 创建一个 etcd client
	cli, err := clientv3.New(services.DefaultEtcdConfig)
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	// 发现服务，取得与服务的链接
	conn, err := services.EtcdDial(cli, c.name)
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
		return nil, fmt.Errorf("could not get %s/%s from perr %s", group, key, c.name)
	}

	return resp.Value, nil
}

func NewClient(service string) *client {
	return &client{name: service}
}

// 测试 client 是否实现了 Fetcher 接口
var _ Fetcher = (*client)(nil)

package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	pb "github.com/1055373165/ggcache/api/groupcachepb"
	"github.com/1055373165/ggcache/configs"
	"github.com/1055373165/ggcache/internal/bussiness/student/dao"
	"github.com/1055373165/ggcache/pkg/common/logger"
	"github.com/1055373165/ggcache/pkg/etcd/discovery"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

type Client struct {
	etcdCli     *clientv3.Client
	conn        *grpc.ClientConn
	client      pb.GroupCacheClient
	serviceName string
	connected   bool
	mu          sync.RWMutex
}

func NewClient(etcdCli *clientv3.Client, serviceName string) (*Client, error) {
	client := &Client{
		etcdCli:     etcdCli,
		serviceName: serviceName,
	}
	if err := client.connect(); err != nil {
		return nil, err
	}
	return client, nil
}

func (c *Client) connect() error {
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

// This small helper repeatedly queries the running ggcache cluster for the
// key "周就" in group "scores" once per second.  It is meant for manual
// debugging / observation rather than automated tests.
func TestSingleKeyClient() {
	configs.InitConfig()
	if err := dao.InitDB(); err != nil {
		logger.LogrusObj.Fatalf("Failed to initialize database: %v", err)
	}

	cli, err := clientv3.New(configs.DefaultEtcdConfig)
	if err != nil {
		panic(err)
	}

	// Change this if you want to hit another node; the picker will forward
	// the request if the key belongs elsewhere.
	ggcacheClient, err := NewClient(cli, configs.Conf.Services["groupcache"].Name)
	if err != nil {
		logger.LogrusObj.Fatalf("failed to dial server: %v", err)
	}
	conn, err := discovery.Discovery(ggcacheClient.etcdCli, "GroupCache")
	if err != nil {
		logger.LogrusObj.Fatalf("failed to dial server: %v", err)
	}
	defer conn.Close()

	client := pb.NewGroupCacheClient(conn)

	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		resp, err := client.Get(ctx, &pb.GetRequest{Group: "scores", Key: "Hannah Harris"})
		cancel()
		if err != nil {
			logger.LogrusObj.Errorf("RPC error: %v", err)
		} else {
			logger.LogrusObj.Infof("Hannah Harris => %s", string(resp.Value))
		}
		time.Sleep(2 * time.Second)
	}
}

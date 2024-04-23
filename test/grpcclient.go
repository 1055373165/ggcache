package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	pb "ggcache/api"
	conf "ggcache/configs"

	etcdservice "ggcache/internal/middleware/etcd"

	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	group   = "scores"
	service = "ggcache"
)

func main() {
	conf.Init()
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdservice.DefaultEtcdConfig.Endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		panic("etcd client v3 create failed" + err.Error())
	}

	// get the client connection to the specified service, also called grpc channel
	conn, err := etcdservice.Discovery(cli, service)
	if err != nil {
		return
	}
	defer conn.Close()

	// the client stub implements the same methods as the grpc server
	clientStub := pb.NewGroupCacheClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// populate requests in protobuf format
	var wg sync.WaitGroup
	names := []string{"张三", "李四", "one", "two", "1", "3"}
	wg.Add(len(names))
	for _, name := range names {
		go func(name string) {
			defer wg.Done()
			resp, err := clientStub.Get(ctx, &pb.GetRequest{Key: name, Group: group})
			if err != nil {
				fmt.Printf("failed to get %s's score via rpc request\n", name)
			} else {
				fmt.Printf("the rpc call succeeds, and the score of %s is %s\n", name, string(resp.GetValue()))
			}
		}(name)
	}
	wg.Wait()
}

package main

import (
	"context"
	"fmt"

	"github.com/1055373165/Distributed_KV_Store/conf"
	pb "github.com/1055373165/Distributed_KV_Store/grpc/groupcachepb"
	"github.com/1055373165/Distributed_KV_Store/logger"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	etcdUrl     = "http://localhost:2379"
	serviceName = "groupcache"
)

const ErrRPCCallNotFound = "rpc error: code = Unknown desc = record not found"

func main() {
	//bd := &ChihuoBuilder{addrs: map[string][]string{"/api": []string{"localhost:8001", "localhost:8002", "localhost:8003"}}}
	//resolver.Register(bd)
	conf.Init()
	etcdClient, err := clientv3.NewFromURL(etcdUrl)
	if err != nil {
		panic(err)
	}
	etcdResolver, err := resolver.NewBuilder(etcdClient)
	if err != nil {
		panic(err)
	}
	conn, err := grpc.Dial(fmt.Sprintf("etcd:///%s", serviceName), grpc.WithResolvers(etcdResolver), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, roundrobin.Name)))

	if err != nil {
		fmt.Printf("err: %v", err)
		return
	}

	ServerClient := pb.NewGroupCacheClient(conn)

	names := []string{}
	for i := 0; i < 500; i++ {
		names = append(names, fmt.Sprintf("%d", i))
	}
	for {
		for _, name := range names {
			resp, err := ServerClient.Get(context.Background(), &pb.GetRequest{
				Group: "scores",
				Key:   "李四",
			})
			if err != nil {
				if ErrRPCCallNotFound != err.Error() {
					logger.Logger.Fatalf("rpc call failed, err: %v", err)
					return
				} else {
					logger.Logger.Warnf("没有查询到学生 %s 的成绩", name)
				}
			} else {
				logger.Logger.Infof("rpc 调用成功, 学生 %s 的成绩为 %s", name, string(resp.Value))
			}
		}
	}
}

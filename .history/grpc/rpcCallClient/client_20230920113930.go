package main

import (
	"context"
	"time"

	"github.com/1055373165/Distributed_KV_Store/conf"
	pb "github.com/1055373165/Distributed_KV_Store/grpc/groupcachepb"
	"github.com/1055373165/Distributed_KV_Store/logger"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conf.Init()
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := cli.Get(ctx, "clusters", clientv3.WithPrefix())
	if err != nil {
		logger.Logger.Error("从 etcd 获取 grpc 通道失败")
		return
	}
	logger.Logger.Info("从 etcd 获取 grpc 通道成功")

	addr := string(resp.Kvs[0].Value)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		logger.Logger.Error("获取 grpc 通道失败")
		return
	}
	client_stub := pb.NewGroupCacheClient(conn)

	response, err := client_stub.Get(ctx, &pb.GetRequest{Key: "scores张三", Group: "scores"})
	if err != nil {
		logger.Logger.Info("没有查询到这个人的记录", err.Error())
		return
	}
	logger.Logger.Infof("成功从 RPC 返回调用结果：%s\n", string(response.GetValue()))
}

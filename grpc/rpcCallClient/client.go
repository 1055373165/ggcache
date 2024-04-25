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

const ErrRPCCallNotFound = "rpc error: code = Unknown desc = record not found"

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

	logger.Logger.Info("从 etcd 获取服务实例地址")
	addr := string(resp.Kvs[0].Value)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		logger.Logger.Error("获取 grpc 通道失败")
		return
	}
	client_stub := pb.NewGroupCacheClient(conn)

	names := []string{"张三", "1", "2", "3", "王五", "李四", "不存在1", "不存在2"}
	for _, name := range names {
		response, err := client_stub.Get(ctx, &pb.GetRequest{Key: name, Group: "scores"})
		if err != nil && err.Error() == ErrRPCCallNotFound {
			logger.Logger.Infof("没有查询到学生 '%s' 的成绩, rpc 调用返回的错误信息: %v", name, err.Error())
		} else if err != nil {
			panic(err)
		}
		logger.Logger.Infof("成功从 RPC 返回学生 %s 分数的调用结果：%s\n", name, string(response.GetValue()))
	}

}

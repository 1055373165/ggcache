package main

import (
	"context"
	"math/rand"
	"strings"
	"time"

	pb "github.com/1055373165/ggcache/api/groupcachepb"
	"github.com/1055373165/ggcache/config"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/1055373165/ggcache/utils/logger"
)

const ErrRPCCallNotFound = "rpc error: code = Unknown desc = record not found"

func main() {
	config.InitConfig()
	cli, err := clientv3.New(config.DefaultEtcdConfig)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := cli.Get(ctx, "clusters", clientv3.WithPrefix())
	if err != nil {
		logger.LogrusObj.Error("从 etcd 获取 grpc 通道失败")
		return
	}

	var addrs []string
	for i := 0; i < int(resp.Count); i++ {
		addrs = append(addrs, string(resp.Kvs[i].Value))
	}
	logger.LogrusObj.Debugf("从 etcd 获取的地址列表: [%v]", addrs)
	addr := string(addrs[rand.Intn(3)])
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		logger.LogrusObj.Error("获取 grpc 通道失败")
		return
	}
	client_stub := pb.NewGroupCacheClient(conn)

	names := []string{"张三", "1", "2", "3", "王五", "李四", "不存在1", "不存在2"}
	for _, name := range names {
		response, err := client_stub.Get(ctx, &pb.GetRequest{Key: name, Group: "scores"})
		if err != nil {
			if strings.Compare(err.Error(), ErrRPCCallNotFound) == 0 {
				logger.LogrusObj.Warnf("没有查询到学生 '%s' 的成绩, rpc 调用返回的信息: %v", name, err.Error())
			} else {
				logger.LogrusObj.Errorf("rpc 调用出现问题: %v", err.Error())
			}
		} else {
			logger.LogrusObj.Infof("成功从 RPC 返回学生 %s 分数 %s", name, string(response.GetValue()))
		}
	}
}
